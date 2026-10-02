# Graph Model

## Purpose

This document defines the initial graph model for SDS eBPF.

The graph is the central representation of system structure derived from runtime evidence. The model should remain stable enough to support queries, evaluation, and replay even as the sensing layer evolves.

## Design Rules

- Nodes and edges represent runtime-observed or runtime-inferred entities.
- Every edge should be traceable to one or more source events.
- Temporal properties are first-class.
- Confidence should be explicit where inference is involved.

## Node Types

### Service

Represents a stable workload or application identity above the process-instance level.

Suggested key:

```text
service:<service-name>
```

Core properties:

- `service`
- `namespace`
- `first_seen`
- `last_seen`
- `evidence_count`

### Endpoint

Represents an external or unresolved network destination.

Suggested key:

```text
endpoint:<name-or-address>:<port>
```

Core properties:

- `address`
- `port`
- `scope`
- `first_seen`
- `last_seen`
- `evidence_count`

### ProcessInstance

Represents a concrete process observed at runtime.

Suggested key:

```text
process:<host-id>:<pid>
```

Core properties:

- `host_id`
- `pid`
- `ppid`
- `comm`
- `binary`
- `uid`
- `first_seen`
- `last_seen`

## Edge Types

### DEPENDS_ON

Represents an internal service-to-service dependency inferred from runtime communication.

Suggested key:

```text
DEPENDS_ON(service:<src>, service:<dst>)
```

Core properties:

- `first_seen`
- `last_seen`
- `evidence_count`
- `source_event_types`
- `confidence`

### CONNECTS_TO_EXTERNAL

Represents communication from a service to an external or unresolved endpoint.

Core properties:

- `first_seen`
- `last_seen`
- `evidence_count`
- `source_event_types`
- `confidence`

### SPAWNED_BY

Represents observed process lineage.

Core properties:

- `first_seen`
- `last_seen`
- `evidence_count`
- `source_event_types`

## Graph Snapshot Format

The current replay harness emits graph snapshots as JSON with this shape:

```json
{
  "schema_version": "sds.graph.v1",
  "generated_at": "2026-03-21T18:30:00Z",
  "nodes": [],
  "edges": []
}
```

Each node entry contains:

- `id`
- `type`
- `properties`

Each edge entry contains:

- `kind`
- `src`
- `dst`
- `properties`

## Notes

This model is intentionally small.

It is sufficient for:

- dependency reconstruction,
- service and endpoint analysis,
- sample evaluation loops,
- and early architecture visualization.

It should grow only when new node or edge types materially improve the system's ability to represent runtime truth.
