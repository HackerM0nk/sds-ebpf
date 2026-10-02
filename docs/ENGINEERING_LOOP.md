# Engineering Loop

## Purpose

This document defines the development loop for SDS eBPF as a research artifact.

The project should not evolve through ad hoc feature work. Each iteration should tighten the path from runtime evidence to measurable system understanding.

## The Loop

Every meaningful engineering cycle should move through these stages:

1. Specify
   Update the contract, schema, or system boundary before changing behavior.

2. Implement
   Add or modify instrumentation, normalization, graph, or evaluation logic.

3. Replay
   Run the system on sample traces, synthetic traces, or benchmark workloads.

4. Evaluate
   Measure correctness, overhead, or reconstruction quality.

5. Inspect
   Review graph output, unexpected edges, dropped identities, and failure cases.

6. Refine
   Change the sensor, schema, inference logic, or benchmark scenario based on evidence.

7. Document
   Record what changed, what improved, what regressed, and what remains unknown.

## Required Outputs Per Iteration

Any substantial subsystem change should come with at least one of the following:

- a schema update,
- a sample trace or replay corpus update,
- a benchmark or evaluation update,
- a graph output example,
- or a documented limitation discovered through testing.

## Local Commands

The current root loop is:

```bash
make research-loop
```

This runs:

- Go userspace tests for the agent
- Go userspace build for the agent
- sample canonical-event validation
- sample graph reconstruction
- sample graph evaluation

Use narrower commands while iterating:

```bash
make agent-go-test
make agent-go-build
make validate-sample
make build-sample-graph
make eval-sample
```

For the first workload-backed artifact path, use:

```bash
make polar-demo
make polar-validate
```

`make polar-demo` should eventually become the primary Phase 1 harness:

- capture a bounded live workload trace,
- build a graph,
- evaluate edge recovery,
- validate artifact acceptance gates,
- and bundle the run into an explicit artifact manifest.

## Current Meaning Of Success

For the current repo state, success means:

- the contracts are explicit,
- the replay harness is deterministic,
- and the output graph can be scored automatically.

This is intentionally narrower than the full project vision.

## Definition Of Done For A Milestone

A milestone is not done because code exists.

A milestone is done when:

- it has a defined contract,
- it has a reproducible command path,
- it has a measurable output,
- and the docs state both its value and its limitations.

## Anti-Patterns

Avoid these failure modes:

- adding sensing features without updating the canonical schema,
- adding databases before the graph contract is stable,
- adding broad claims without benchmark evidence,
- and treating demo output as evaluation.

## How This Evolves

As the real agent becomes trustworthy, this loop should expand from sample replay to live capture:

1. capture real workload events,
2. normalize into canonical events,
3. build incremental graph state,
4. evaluate against benchmark ground truth,
5. and publish the resulting measurements.

The loop should remain visible in the repository even as the internals get more sophisticated.
