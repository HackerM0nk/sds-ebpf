// SPDX-License-Identifier: GPL-2.0
// eBPF Observer - Sparse sampling agent for system telemetry
// Monitors: processes, syscalls, network connections, container metadata

#include "vmlinux.h"
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_tracing.h>
#include <bpf/bpf_core_read.h>

#define TASK_COMM_LEN 16
#define MAX_FILENAME_LEN 256
#define MAX_SYSCALL_NAME 32

// Event types
enum event_type {
    EVENT_PROCESS_EXEC = 1,
    EVENT_PROCESS_EXIT = 2,
    EVENT_SYSCALL = 3,
    EVENT_TCP_CONNECT = 4,
};

// Process exec event
struct process_event {
    u32 type;
    u64 timestamp;
    u32 pid;
    u32 ppid;
    u32 uid;
    u32 gid;
    char comm[TASK_COMM_LEN];
    char filename[MAX_FILENAME_LEN];
    u64 cgroup_id;
};

// Syscall event (sampling key syscalls)
struct syscall_event {
    u32 type;
    u64 timestamp;
    u32 pid;
    u32 uid;
    char comm[TASK_COMM_LEN];
    u64 syscall_id;
    char syscall_name[MAX_SYSCALL_NAME];
    u64 duration_ns;
};

// TCP connect event
struct tcp_event {
    u32 type;
    u64 timestamp;
    u32 pid;
    u32 uid;
    char comm[TASK_COMM_LEN];
    u32 saddr;
    u32 daddr;
    u16 sport;
    u16 dport;
    u64 cgroup_id;
};

// Process exit event
struct exit_event {
    u32 type;
    u64 timestamp;
    u32 pid;
    u32 exit_code;
};

// Ring buffer for sending events to user-space
struct {
    __uint(type, BPF_MAP_TYPE_RINGBUF);
    __uint(max_entries, 256 * 1024); // 256KB ring buffer
} events SEC(".maps");

// Map to track syscall entry times (for duration calculation)
struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __uint(max_entries, 10240);
    __type(key, u64); // pid_tgid
    __type(value, u64); // entry timestamp
} syscall_entry_times SEC(".maps");

// Helper to get container cgroup ID
static __always_inline u64 get_cgroup_id(struct task_struct *task) {
    struct css_set *cgroups;
    struct cgroup *cgroup;

    cgroups = BPF_CORE_READ(task, cgroups);
    if (!cgroups)
        return 0;

    // Get the cgroup from css_set (simplified)
    // In production, you'd want to read specific cgroup subsystems
    cgroup = BPF_CORE_READ(cgroups, dfl_cgrp);
    if (!cgroup)
        return 0;

    return BPF_CORE_READ(cgroup, kn, id);
}

// Tracepoint: sched_process_exec
SEC("tp/sched/sched_process_exec")
int handle_exec(struct trace_event_raw_sched_process_exec *ctx) {
    struct process_event *event;
    struct task_struct *task;
    u64 pid_tgid = bpf_get_current_pid_tgid();
    u32 pid = pid_tgid >> 32;
    u64 uid_gid = bpf_get_current_uid_gid();

    // Reserve space in ring buffer
    event = bpf_ringbuf_reserve(&events, sizeof(*event), 0);
    if (!event)
        return 0;

    task = (struct task_struct *)bpf_get_current_task();

    event->type = EVENT_PROCESS_EXEC;
    event->timestamp = bpf_ktime_get_ns();
    event->pid = pid;
    event->ppid = BPF_CORE_READ(task, real_parent, tgid);
    event->uid = uid_gid & 0xFFFFFFFF;
    event->gid = uid_gid >> 32;
    event->cgroup_id = get_cgroup_id(task);

    bpf_get_current_comm(&event->comm, sizeof(event->comm));
    bpf_probe_read_str(&event->filename, sizeof(event->filename),
                       (void *)ctx->filename);

    bpf_ringbuf_submit(event, 0);
    return 0;
}

// Tracepoint: sched_process_exit
SEC("tp/sched/sched_process_exit")
int handle_exit(struct trace_event_raw_sched_process_template *ctx) {
    struct exit_event *event;
    struct task_struct *task;
    u64 pid_tgid = bpf_get_current_pid_tgid();
    u32 pid = pid_tgid >> 32;

    event = bpf_ringbuf_reserve(&events, sizeof(*event), 0);
    if (!event)
        return 0;

    task = (struct task_struct *)bpf_get_current_task();

    event->type = EVENT_PROCESS_EXIT;
    event->timestamp = bpf_ktime_get_ns();
    event->pid = pid;
    event->exit_code = BPF_CORE_READ(task, exit_code);

    bpf_ringbuf_submit(event, 0);
    return 0;
}

// Raw tracepoint: sys_enter (syscall monitoring)
// We'll sample common syscalls: open, read, write, connect, etc.
SEC("raw_tp/sys_enter")
int handle_syscall_enter(struct bpf_raw_tracepoint_args *ctx) {
    u64 pid_tgid = bpf_get_current_pid_tgid();
    u64 timestamp = bpf_ktime_get_ns();

    // Store entry time for duration calculation
    bpf_map_update_elem(&syscall_entry_times, &pid_tgid, &timestamp, BPF_ANY);

    return 0;
}

// Raw tracepoint: sys_exit
SEC("raw_tp/sys_exit")
int handle_syscall_exit(struct bpf_raw_tracepoint_args *ctx) {
    struct syscall_event *event;
    u64 pid_tgid = bpf_get_current_pid_tgid();
    u32 pid = pid_tgid >> 32;
    u64 *entry_ts;
    u64 exit_ts = bpf_ktime_get_ns();
    u64 uid_gid = bpf_get_current_uid_gid();

    // Sample only ~10% of syscalls (reduce overhead)
    if ((exit_ts % 10) != 0)
        return 0;

    entry_ts = bpf_map_lookup_elem(&syscall_entry_times, &pid_tgid);
    if (!entry_ts)
        return 0;

    event = bpf_ringbuf_reserve(&events, sizeof(*event), 0);
    if (!event) {
        bpf_map_delete_elem(&syscall_entry_times, &pid_tgid);
        return 0;
    }

    event->type = EVENT_SYSCALL;
    event->timestamp = exit_ts;
    event->pid = pid;
    event->uid = uid_gid & 0xFFFFFFFF;
    event->syscall_id = ctx->args[1]; // syscall number
    event->duration_ns = exit_ts - *entry_ts;

    bpf_get_current_comm(&event->comm, sizeof(event->comm));

    // Map common syscall IDs to names (simplified)
    switch (event->syscall_id) {
        case 0: __builtin_memcpy(event->syscall_name, "read", 5); break;
        case 1: __builtin_memcpy(event->syscall_name, "write", 6); break;
        case 2: __builtin_memcpy(event->syscall_name, "open", 5); break;
        case 3: __builtin_memcpy(event->syscall_name, "close", 6); break;
        case 42: __builtin_memcpy(event->syscall_name, "connect", 8); break;
        case 43: __builtin_memcpy(event->syscall_name, "accept", 7); break;
        default: __builtin_memcpy(event->syscall_name, "unknown", 8); break;
    }

    bpf_ringbuf_submit(event, 0);
    bpf_map_delete_elem(&syscall_entry_times, &pid_tgid);

    return 0;
}

// Kprobe: tcp_connect (track TCP connections)
SEC("kprobe/tcp_connect")
int BPF_KPROBE(trace_tcp_connect, struct sock *sk) {
    struct tcp_event *event;
    struct task_struct *task;
    u64 pid_tgid = bpf_get_current_pid_tgid();
    u32 pid = pid_tgid >> 32;
    u64 uid_gid = bpf_get_current_uid_gid();

    event = bpf_ringbuf_reserve(&events, sizeof(*event), 0);
    if (!event)
        return 0;

    task = (struct task_struct *)bpf_get_current_task();

    event->type = EVENT_TCP_CONNECT;
    event->timestamp = bpf_ktime_get_ns();
    event->pid = pid;
    event->uid = uid_gid & 0xFFFFFFFF;
    event->cgroup_id = get_cgroup_id(task);

    bpf_get_current_comm(&event->comm, sizeof(event->comm));

    // Read socket addresses (IPv4 only for simplicity)
    struct sock_common *sk_common = &sk->__sk_common;
    event->saddr = BPF_CORE_READ(sk_common, skc_rcv_saddr);
    event->daddr = BPF_CORE_READ(sk_common, skc_daddr);
    event->sport = BPF_CORE_READ(sk_common, skc_num);
    event->dport = bpf_ntohs(BPF_CORE_READ(sk_common, skc_dport));

    bpf_ringbuf_submit(event, 0);
    return 0;
}

char LICENSE[] SEC("license") = "GPL";