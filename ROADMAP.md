# SDS eBPF Roadmap

Build this repository as a serious open-source systems artifact: a runtime-derived understanding layer for modern infrastructure.

The project goal is not to be "yet another telemetry pipeline." The goal is to produce a reusable artifact that reconstructs live system structure from runtime evidence and exposes that structure for architecture, security, and operations workflows.

This roadmap is downstream of the project charter in [docs/VISION.md](docs/VISION.md).

## Project Shape

The repository should mature into four durable layers:

1. Instrumentation
   eBPF sensors for process, network, container, and selected syscall activity.
2. Runtime Model
   Canonical events, identity resolution, enrichment, and temporal correlation.
3. Knowledge Graph
   Incremental graph construction over hosts, processes, containers, pods, services, and endpoints.
4. Analysis Surface
   Queries, diagrams, drift reports, dependency inspection, and grounded reasoning tools.

## Current State

This repo is an early scaffold.

- The architectural intent is strong.
- The implementation is partial.
- The docs currently go farther than the code.
- The next priority is to make the sensor and data model layers real before broadening the stack.

## Guiding Principles

- Engineering first. Every high-level claim must correspond to a concrete subsystem.
- Truth from runtime. Static manifests are context, not ground truth.
- Reproducibility matters. The artifact must be benchmarkable, replayable, and testable.
- Incremental graph construction. Avoid batch-only thinking.
- Honest scope. Build the minimum infrastructure needed to support the paper-grade claims.

## Release Plan

### Milestone 0: Repo Hardening

Goal: make the repository credible as a public engineering project.

- Fix build reproducibility for the Go agent and eBPF toolchain.
- Add CI for lint, unit tests, and Linux build validation.
- Add a real license file and contributor-facing docs.
- Rewrite status language so README claims match reality.

Exit criteria:
- Clean bootstrap on a documented Linux environment.
- Repeatable local developer workflow.
- CI green on every push.

### Milestone 1: Instrumentation MVP

Goal: produce real runtime events from a Linux host or containerized workload.

- Replace placeholder collector flow with real eBPF program loading.
- Generate and embed bindings with `bpf2go`.
- Emit versioned NDJSON events with host metadata.
- Support process exec/exit and TCP connect first.
- Add selected syscall visibility only after validation.

Exit criteria:
- Real events, not mocks.
- Deterministic sample output from a controlled workload.
- Basic parser tests and smoke tests.

### Milestone 2: Canonical Schema And Identity

Goal: define the substrate the rest of the system depends on.

- Formalize event schema and versioning.
- Add stable IDs for hosts, workloads, processes, and endpoints.
- Add container and pod identity enrichment.
- Define provenance and confidence fields for every inferred entity.

Exit criteria:
- Schema doc checked into the repo.
- Backward-compatible event evolution policy.
- Replay support from saved event logs.

### Milestone 3: Graph Pipeline

Goal: build the first useful system model from runtime evidence.

- Create graph builder service.
- Materialize nodes and edges incrementally.
- Support process lineage, containment, and network dependency edges.
- Persist graph into one primary graph backend first.

Exit criteria:
- Controlled microservice deployment becomes a queryable graph.
- Graph updates reflect runtime changes.
- First architecture diagram can be generated from graph data.

### Milestone 4: Evaluation Harness

Goal: make the project paper-grade.

- Build benchmark workloads and repeatable experiments.
- Measure CPU, memory, storage, event volume, and update latency.
- Compare sparse sampling modes and recovery fidelity.
- Create ground-truth datasets for dependency and topology evaluation.

Exit criteria:
- Reproducible benchmark scripts.
- Evaluation tables and plots checked into the repo.
- Clear methodology for artifact reviewers.

### Milestone 5: Analysis Surface

Goal: expose the graph in ways that matter to users and reviewers.

- Implement dependency and blast-radius queries.
- Generate architecture and change diagrams.
- Add drift and undocumented-service detection.
- Add optional reasoning layer grounded strictly in graph context.

Exit criteria:
- End-to-end demo from workload to query to diagram.
- Use-case walkthroughs for architecture, security, and operations.

### Milestone 6: Artifact Release

Goal: ship a polished public artifact.

- Package a one-command demo environment.
- Add artifact setup docs and review guide.
- Pin dependencies and publish versioned releases.
- Record known limitations and threat model.

Exit criteria:
- A reviewer can reproduce the system end to end.
- The repo can back a paper submission without hand-waving.

## Target Repository Layout

This is a better long-term structure than the current mixed prototype layout:

```text
/
├── agent/               # eBPF programs and host-side collector
├── schema/              # Canonical event schemas and graph model docs
├── graph/               # Graph builder and graph query services
├── eval/                # Benchmarks, workloads, replay harness, result scripts
├── deploy/              # Docker Compose, Kubernetes, local demo environments
├── docs/                # Research, design, operations, and artifact docs
├── examples/            # Small demo workloads and sample output
├── scripts/             # Developer and evaluation automation
└── ROADMAP.md
```

## Workstreams

These should become GitHub project epics:

- Sensor Runtime
- Schema And Identity
- Graph Construction
- Evaluation And Benchmarking
- Analysis And Visualization
- DX, CI, And Release Engineering
- Paper And Artifact Documentation

## Suggested GitHub Project Board

Use a simple project board with these columns:

- Backlog
- Ready
- In Progress
- Validation
- Blocked
- Done

Use issue labels such as:

- `area/agent`
- `area/schema`
- `area/graph`
- `area/eval`
- `area/docs`
- `area/release`
- `kind/bug`
- `kind/feature`
- `kind/research`
- `priority/p0`
- `priority/p1`

## Initial Epic Issues

Create these first:

1. Make the Go agent build reliably on Linux and CI.
2. Introduce `bpf2go` generation and real object loading.
3. Define canonical NDJSON event schema v1.
4. Implement host and workload identity enrichment.
5. Create graph schema for runtime entities and edges.
6. Build minimal graph ingester from replayed event files.
7. Stand up one reproducible demo workload.
8. Add benchmark harness for overhead and fidelity measurements.
9. Generate first runtime-derived architecture diagram.
10. Publish artifact setup and reviewer guide.

## What Not To Build Yet

Avoid these until Milestones 1 through 4 are solid:

- Multiple databases for the sake of architecture diagrams
- Broad LLM features without grounded graph context
- A large streaming stack before replay and local ingestion work
- Security claims that are not backed by measured coverage

## Definition Of A Solid Artifact

This project becomes solid when:

- the agent produces trustworthy runtime events,
- the graph pipeline reconstructs useful structure,
- the evaluation is reproducible,
- the docs are honest and complete,
- and the repository can be cloned and exercised by an external reviewer without private context.
