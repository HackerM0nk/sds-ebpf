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

typedef signed char s8;
typedef signed short s16;
typedef signed int s32;
typedef signed long long s64;

typedef __u8 u8;
typedef __u16 u16;
typedef __u32 u32;
typedef __u64 u64;

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

#endif /* __VMLINUX_H__ */