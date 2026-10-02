# Artifact Spec v0.1

## Purpose

This document defines the first buildable slice of SDS eBPF as a research artifact.

The goal of v0.1 is not to deliver the full platform vision. The goal is to establish the first credible end-to-end path that later versions can extend without changing the project thesis.

## v0.1 Artifact Boundary

Version 0.1 covers exactly this pipeline:

1. Runtime signals are represented as canonical NDJSON events.
2. Canonical events are validated against a stable contract.
3. Events are transformed into a graph snapshot.
4. The resulting graph is evaluated against a known ground-truth dependency set.

This is the smallest artifact slice that exercises:

- schema design,
- identity representation,
- graph construction,
- and evaluation logic.

## What v0.1 Must Prove

v0.1 should prove that the repository can support a principled engineering loop around runtime system reconstruction.

It must make the following concrete:

- a canonical event interface,
- a graph model,
- a replayable harness,
- a measurable output,
- and an automated loop that can run in CI.

## In Scope

- Canonical event schema v1
- Service and endpoint oriented graph model
- Sample runtime event corpus
- Graph construction script
- Evaluation script for edge recovery quality
- Root make targets and CI hook for the loop
- Documentation for how this loop fits the broader artifact

## Out Of Scope

The following are explicitly not required for v0.1:

- live eBPF-to-graph streaming,
- production-ready graph storage,
- Kubernetes-level enrichment,
- adaptive sampling logic,
- assistant or LLM features,
- large-scale benchmark workloads,
- and finalized paper figures.

## Data Path

The v0.1 data path is:

```text
canonical_events.ndjson
  -> validate_canonical_events.py
  -> build_runtime_graph.py
  -> runtime_graph.json
  -> evaluate_runtime_graph.py
  -> eval_summary.json
```

This path must be replayable, deterministic, and easy to inspect.

## Canonical Event Contract

The canonical event contract for v0.1 is defined in:

- [canonical_event_v1.schema.json](../schema/events/canonical_event_v1.schema.json)

Every event must carry enough context to support downstream reconstruction:

- schema version,
- host identity,
- process identity,
- workload identity,
- sensor identity,
- event kind,
- and event payload.

## Graph Contract

The graph contract for v0.1 is defined in:

- [GRAPH_MODEL.md](../schema/GRAPH_MODEL.md)

The initial graph focuses on:

- `Service` nodes,
- `Endpoint` nodes,
- `ProcessInstance` nodes,
- `DEPENDS_ON` edges,
- `CONNECTS_TO_EXTERNAL` edges,
- and `SPAWNED_BY` edges where process lineage is known.

## Evaluation Contract

The v0.1 evaluation contract is intentionally narrow:

- compare predicted graph edges to known expected edges,
- compute precision, recall, and F1,
- emit the result as machine-readable JSON,
- and fail loudly on malformed input.

This is enough to establish the research loop without pretending the system is complete.

## Exit Criteria

v0.1 is complete when:

- canonical events can be validated automatically,
- graph snapshots can be produced from replayed events,
- graph outputs can be evaluated against ground truth,
- the loop runs locally and in CI,
- and the repo documents this flow clearly.

## Relation To The Broader System

v0.1 is the contract layer and harness layer for the larger vision.

Later versions should extend it like this:

- v0.2: real agent output mapped into canonical events
- v0.3: incremental graph builder with one graph backend
- v0.4: benchmark workloads and measured overhead
- v0.5: richer analysis queries and diagram generation

The key principle is that later milestones should reuse the v0.1 contracts rather than replacing them casually.
