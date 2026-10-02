# SDS Execution Plan

## Purpose

This document turns the project vision into an execution sequence.

The key question is not "what could SDS become?" It is:

- what must be built first,
- what depends on what,
- what is already good enough,
- and what is still too weak to support a serious artifact or paper.

This plan is intentionally critical. It focuses on the gaps that still separate the repo from a solid public artifact.

## Current Reality

SDS now has four meaningful pieces:

1. A real-ish host collector path
   The observer can load eBPF programs, attach probes, and emit canonical NDJSON events.

2. A contract layer
   Canonical events, graph model docs, validation scripts, and a replayable sample evaluation loop exist.

3. A first workload package
   Polar Bookshop is packaged as the first demo workload with scenario traffic and workload-specific ground truth.

4. A partial runtime identity layer
   Docker Compose service naming is now available through runtime metadata enrichment.

That is real progress, but it is not yet an artifact-grade system.

## Hard Truths

These are the most important weaknesses right now.

### 1. The graph layer is still a script, not a system

The current graph build path is an offline Python snapshot builder.

That is fine for contract validation, but it is not yet:

- incremental,
- queryable as a service,
- stateful across captures,
- or capable of temporal change analysis.

### 2. Identity resolution is still narrow

Docker Compose enrichment is a good first step, but the identity story is still thin.

It is not yet robust for:

- Kubernetes,
- containerd without Docker labels,
- short-lived processes,
- process lineage gaps,
- or ambiguous internal endpoints.

### 3. The live demo is packaged but not fully proven

The Polar demo runner now exists, but the first live run showed a real issue:

- cold image pulls consumed the bounded capture window.

That specific problem has been addressed by pre-pulling before capture, but the full warm-cache run still needs to be completed and checked end to end.

### 4. Sensor correctness is not yet defensible enough

The collector is no longer mock-driven, but the sensing layer still needs more trust work:

- exec coverage is still unreliable on some kernels,
- syscall semantics are still broad and noisy,
- event meaning is not yet confidence-scored from source conditions,
- and we do not yet have a clear "known good" sensor matrix by environment.

### 5. The paper claim is broader than the implementation

The repo vision is correctly broad. The implementation is not yet broad in a defensible way.

So the build must stay focused on the substrate:

- evidence,
- identity,
- graph,
- evaluation,
- and reproducibility.

## What Must Be True Before SDS Is A Solid Artifact

SDS becomes solid when all of the following are true:

1. A reviewer can run one workload-backed demo from scratch and get:
   - canonical events,
   - a graph,
   - an evaluation summary,
   - and a clear explanation of what was inferred.

2. The observer emits trustworthy enough runtime evidence that service recovery is not mostly accidental.

3. The system can preserve and use identity across multiple captures, not just one-shot file conversion.

4. The graph layer can answer a small but real set of architecture questions.

5. The evaluation loop includes both fidelity and overhead, not just graph edge matching.

## Execution Strategy

We should build SDS in six strict phases.

Do not skip phases. Do not broaden early. Do not build the reasoning layer before the graph substrate is real.

## Phase 1: Stabilize The Demo Artifact

### Goal

Make the Polar workload demo reproducible and trustworthy enough to serve as the first public artifact path.

### Why This Comes First

Right now, the repo has a demo package but not yet a proven one-command artifact loop. That is the shortest path to something externally legible.

### Work

1. Complete a full warm-cache `make polar-demo` run and inspect the outputs.
2. Verify that the capture window contains workload behavior, not mostly boot noise.
3. Check whether the inferred edges match the intended workload graph.
4. Tighten the scenario if needed so it deterministically exercises:
   - `edge-service -> catalog-service`
   - `order-service -> catalog-service`
   - `order-service -> polar-rabbitmq`
   - any startup dependencies we decide belong in ground truth
5. Split "startup edges" from "scenario edges" in the ground truth if they are being conflated.
6. Record a known-good demo run in docs with exact outputs and expected F1 range.

### Exit Criteria

- `make polar-demo` completes from a warm local image cache.
- It produces canonical events, `runtime_graph.json`, and `eval_summary.json`.
- The graph is dominated by workload services, not unrelated host noise.
- The evaluation result is stable enough to use in docs and a paper artifact appendix.

## Phase 2: Make The Event Substrate Defensible

### Goal

Improve the evidence layer until the graph output is something we can trust, not just inspect.

### Why This Is Second

The graph can only be as good as the event semantics beneath it.

### Work

1. Define event source guarantees by sensor:
   - what each event means,
   - when it can be missing,
   - and which inference confidence it should support.
2. Add explicit provenance and confidence to canonical events where missing.
3. Add richer tests for:
   - timestamp normalization,
   - TCP destination interpretation,
   - cgroup/container identity extraction,
   - and fallback behavior when metadata is unavailable.
4. Reduce noise in syscall collection or clearly scope which syscall signals are actually used downstream.
5. Add a compatibility matrix for environments:
   - native Linux,
   - Docker-on-Linux,
   - Docker Desktop / Colima style kernels.
6. Address exec lineage reliability:
   - either improve the probe strategy,
   - or formalize an alternative lineage fallback.

### Exit Criteria

- The meaning of each used event type is documented and tested.
- Canonical events carry enough provenance to explain why an edge exists.
- Unsupported or unreliable environments degrade explicitly rather than silently.

## Phase 3: Build Identity Resolution v1

### Goal

Make service and workload identity robust enough that the graph is not Docker-Compose-specific.

### Why This Is Third

Identity is the bridge between evidence and graph structure. Without it, everything above the collector stays brittle.

### Work

1. Refactor identity enrichment into a first-class subsystem, not just helper code.
2. Support multiple identity sources:
   - process metadata,
   - cgroup metadata,
   - Docker labels,
   - address-to-container mapping,
   - and eventually Kubernetes metadata.
3. Distinguish:
   - process instance,
   - container instance,
   - workload instance,
   - and logical service.
4. Add explicit ambiguity handling:
   - unknown destination,
   - multiple candidate services,
   - unresolved short-lived process.
5. Define when an internal IP should become:
   - a `DEPENDS_ON` edge,
   - an internal endpoint edge,
   - or remain unresolved.

### Exit Criteria

- Service naming is stable across repeated runs of the demo workload.
- Internal destination recovery is good enough that dependency edges are mostly resolved to services.
- The identity layer can report unresolved cases explicitly.

## Phase 4: Replace The Graph Script With A Real Graph Layer

### Goal

Move from an offline snapshot script to a minimal but real graph subsystem.

### Why This Is The Structural Pivot

This is where SDS stops being "collector + scripts" and becomes a systems artifact.

### Work

1. Extract the graph-building logic into a library or service with a stable API.
2. Support incremental updates from event streams or replay files.
3. Store temporal properties and evidence counts per node and edge.
4. Keep provenance and confidence attached to every inferred edge.
5. Add a minimal query surface for:
   - `what talks to what`
   - `what depends on X`
   - `what changed between two captures`
6. Keep the backend simple:
   - file-backed state or one embedded store first,
   - external graph database later.

### Exit Criteria

- The graph can be built incrementally from live or replayed events.
- It can answer a small query set without custom one-off scripts.
- Repeated captures update the same logical model instead of producing only disconnected snapshots.

## Phase 5: Build The Evaluation Loop Reviewers Will Respect

### Goal

Make SDS evaluable as a systems artifact, not just demoable.

### Why This Comes Before "AI" Or Fancy Surfaces

Without measured evidence, the paper collapses into product language.

### Work

1. Separate evaluation dimensions:
   - graph fidelity,
   - event coverage,
   - update freshness,
   - overhead,
   - and robustness.
2. Add workload-specific traces and result snapshots under `eval/`.
3. Distinguish startup topology from scenario-induced behavior in ground truth.
4. Add at least one second workload after Polar:
   - FTGO or Online Boutique.
5. Add overhead measurement scripts:
   - observer CPU,
   - memory,
   - event counts,
   - capture latency.
6. Add failure-mode experiments:
   - partial visibility,
   - short capture windows,
   - service restart churn.

### Exit Criteria

- SDS has reproducible workload-backed evaluation outputs checked into the repo.
- There is at least one accuracy table and one overhead table that can support a paper section.
- The evaluation methodology is documented clearly enough for artifact reviewers.

## Phase 6: Build The Public Artifact Surface

### Goal

Package the system so an outsider can understand and exercise it without prior context.

### Work

1. Add an artifact review guide with:
   - prerequisites,
   - exact commands,
   - expected outputs,
   - troubleshooting.
2. Make the workload-backed demo the primary repo entry point.
3. Add diagram generation from graph output.
4. Add a small query CLI or script layer.
5. Clean up docs so the repo no longer mixes future vision with current capability.

### Exit Criteria

- The repo has one obvious artifact path.
- The docs reflect current reality cleanly.
- A reviewer can reproduce the system end to end from the repository alone.

## What We Should Not Do Yet

These are tempting, but premature.

### Do Not Build The LLM Layer Yet

It will distract from the graph substrate and weaken the paper if introduced too early.

### Do Not Introduce Kafka / ClickHouse / Neo4j Prematurely

Those are scaling choices, not proof of correctness.

### Do Not Add More Sensors Before The Existing Ones Are Defensible

Breadth without trust will produce a noisier system, not a better one.

### Do Not Expand To Kubernetes Before Docker Identity Is Clean

Kubernetes enrichment will amplify any weakness already present in the identity model.

## Recommended Immediate Sprint

This is the next tight loop to execute.

### Sprint 1

1. Finish a full warm-cache `make polar-demo` run.
2. Inspect the resulting event corpus and graph.
3. Tighten the Polar ground truth into:
   - startup edges,
   - scenario edges,
   - and optional edges.
4. Add one regression test around graph recovery from a captured Polar trace.
5. Document the exact known-good artifact output.

### Sprint 2

1. Strengthen event provenance and confidence.
2. Improve identity resolution for unresolved internal endpoints.
3. Formalize the graph-building logic as a reusable module.

### Sprint 3

1. Add a second workload.
2. Add overhead measurements.
3. Start building the first paper-quality evaluation tables.

## Definition Of Success For The Next Month

Within the next month, the repo should be able to claim all of this honestly:

- SDS can capture runtime events from a real workload.
- SDS can recover a useful service dependency graph from those events.
- SDS can evaluate that graph against workload-specific ground truth.
- SDS can reproduce that loop for at least one public microservice system.
- SDS has the beginnings of a real runtime graph substrate, not just a demo collector.

## Relationship To Other Docs

- [Vision](./VISION.md) defines what SDS is trying to become.
- [Artifact Spec v0.1](./ARTIFACT_SPEC_V0_1.md) defines the current minimum artifact slice.
- [Roadmap](../ROADMAP.md) defines the broader milestone structure.
- This document defines the critical execution order from here.
