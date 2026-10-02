// vmlinux.h - Kernel type definitions for eBPF CO-RE
// This is a minimal stub. Generate the full version using:
//   bpftool btf dump file /sys/kernel/btf/vmlinux format c > vmlinux.h
//
// For local development, you can use a pre-generated vmlinux.h from:
//   https://github.com/libbpf/libbpf-bootstrap/tree/master/examples/c

#ifndef __VMLINUX_H__
#define __VMLINUX_H__

typedef unsigned char __u8;
typedef unsigned short __u16;
typedef unsigned int __u32;
typedef unsigned long long __u64;
typedef signed short __s16;
typedef signed int __s32;
typedef signed long long __s64;

typedef signed char s8;
typedef signed short s16;
typedef signed int s32;
typedef signed long long s64;

typedef __u8 u8;
typedef __u16 u16;
typedef __u32 u32;
typedef __u64 u64;
typedef __u16 __le16;
typedef __u16 __be16;
typedef __u32 __be32;
typedef __u64 __be64;
typedef __u32 __wsum;

enum bpf_map_type {
    BPF_MAP_TYPE_UNSPEC = 0,
    BPF_MAP_TYPE_HASH = 1,
    BPF_MAP_TYPE_ARRAY = 2,
    BPF_MAP_TYPE_RINGBUF = 27,
};

enum {
    BPF_ANY = 0,
    BPF_NOEXIST = 1,
    BPF_EXIST = 2,
    BPF_F_LOCK = 4,
};

// Minimal kernel structures needed for our eBPF programs
// In production, use full vmlinux.h generated from target kernel

struct task_struct {
    int pid;
    int tgid;
    int exit_code;
    struct task_struct *real_parent;
    struct css_set *cgroups;
    char comm[16];
};

struct sock_common {
    __u32 skc_rcv_saddr;
    __u32 skc_daddr;
    __u16 skc_num;
    __u16 skc_dport;
    __u16 skc_family;
};

struct sock {
    struct sock_common __sk_common;
};

struct css_set {
    struct cgroup *dfl_cgrp;
};

struct cgroup {
    struct kernfs_node *kn;
};

struct kernfs_node {
    u64 id;
};

// Tracepoint structures
struct trace_event_raw_sched_process_exec {
    u64 unused;
    int pid;
    int old_pid;
    const char *filename;
};

struct trace_event_raw_sched_process_template {
    u64 unused;
    char comm[16];
    int pid;
};

struct bpf_raw_tracepoint_args {
    u64 args[6];
};

struct user_pt_regs {
    u64 regs[31];
    u64 sp;
    u64 pc;
};

#endif /* __VMLINUX_H__ */
