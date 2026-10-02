# Research Blueprint

## Thesis

Modern infrastructure lacks a living, machine-readable model of runtime reality.

Static manifests, dashboards, logs, and architecture documents describe intent or fragments of behavior, but they do not maintain a durable representation of what is actually executing, communicating, and changing in a system. This project aims to build a runtime-native substrate that continuously reconstructs that reality from low-level evidence and makes it queryable.

## Artifact Goal

Build an open, reproducible systems artifact that:

- observes runtime behavior with low overhead,
- converts raw events into stable identities and relationships,
- materializes those relationships as a temporal knowledge graph,
- and supports architecture, security, and operations workflows on top of that graph.

The artifact should stand on its own as an engineering system even if the paper framing changes.

## Design Principles

- Runtime truth over declared intent.
- Clear provenance for every inferred entity and edge.
- Low-overhead sensing with measurable tradeoffs.
- Incremental graph updates instead of periodic full recomputation.
- Strong separation between measurement, inference, and explanation.
- Reproducibility as a first-class requirement.

## System Layers

### 1. Instrumentation Layer

The sensing substrate should capture the minimum useful runtime evidence needed to reconstruct system structure.

Initial signal classes:

- Process exec and exit
- TCP connect and listen
- Selected syscall activity
- DNS activity where feasible
- Container and cgroup identity
- Host and namespace context

Later signal classes:

- File access for selected sensitive paths
- Privilege transitions
- IPC and local socket relationships

### 2. Canonical Event Layer

All downstream components depend on a stable event model.

Each event should include:

- schema version
- event ID
- observation timestamp
- host identity
- process identity
- workload identity when known
- source sensor type
- event payload
- provenance metadata

The event layer should support both live ingestion and file replay.

### 3. Identity And Enrichment Layer

Raw runtime evidence is not useful enough on its own. The artifact must resolve identities across execution contexts.

Required mappings:

- process instance to binary
- process instance to container
- container to pod or workload
- network endpoint to internal or external entity
- repeated process instances to service-level identity

This layer should explicitly track confidence and ambiguity rather than pretending every mapping is exact.

### 4. Graph Construction Layer

The graph is the central systems abstraction.

Candidate node types:

- `Host`
- `Binary`
- `ProcessInstance`
- `Container`
- `Pod`
- `Service`
- `Endpoint`
- `Database`
- `User`
- `PrivilegeDomain`

Candidate edge types:

- `SPAWNED_BY`
- `RUNS_ON`
- `RUNS_IN`
- `LISTENS_ON`
- `CONNECTS_TO`
- `RESOLVES`
- `DEPENDS_ON`
- `ACCESSES`
- `RUNS_AS`

Every node and edge should carry:

- `first_seen`
- `last_seen`
- `evidence_count`
- `confidence`
- `source_event_types`

### 5. Analysis Layer

Once the graph exists, the artifact should support queries such as:

- What services currently depend on this endpoint?
- What changed in the architecture over the last hour?
- Which processes talk to external infrastructure?
- Which workloads run with elevated privileges?
- What is the likely blast radius if this service fails?

### 6. Presentation Layer

The presentation layer should remain grounded in measured system state.

Outputs:

- runtime architecture diagrams
- dependency summaries
- change and drift reports
- documented query workflows
- optional natural-language interface grounded only in graph context

## Research Questions

The project should be broad, but the questions must remain concrete.

1. Can low-level runtime sensing recover useful system structure at acceptable overhead?
2. How accurately can runtime events reconstruct process, container, and service dependencies?
3. How quickly can a temporal graph reflect real changes in workload behavior?
4. Which operational and security questions become easier to answer once the graph exists?
5. What are the limits of sparse or adaptive sensing in highly dynamic systems?

## Evaluation Strategy

The paper should evaluate the artifact along multiple dimensions.

### Overhead

Measure:

- CPU cost
- memory footprint
- event volume
- storage growth
- effect of different sensing and sampling modes

### Fidelity

Measure:

- precision and recall of recovered nodes
- precision and recall of recovered dependency edges
- identity resolution quality
- false positives from ambiguous or transient behavior

### Freshness

Measure:

- delay from runtime event to graph update
- delay from topology change to visible model change

### Robustness

Measure:

- short-lived process capture
- restart and churn behavior
- burst traffic handling
- degraded behavior under event loss or partial visibility

### Utility

Measure:

- ability to answer dependency and blast-radius questions
- ability to find undocumented services or edges
- comparison against static deployment metadata alone

## Experimental Environments

Use multiple environments so the artifact is not tied to one story.

- Small local Compose-based microservice stack
- Container benchmark applications with known topology
- A more dynamic orchestration setting once the baseline is stable

Ground truth should come from controlled deployments, known workload structure, and explicit scenario design rather than assumed documentation quality.

The current staged workload plan is documented in [../eval/WORKLOAD_PLAN.md](../eval/WORKLOAD_PLAN.md).

## Paper Positioning

Keep the paper broad, but center it on the systems contribution.

Good framing options:

- runtime architecture reconstruction
- continuous systems knowledge graph generation
- low-overhead runtime substrate for infrastructure understanding

The graph and runtime model are the core contribution. Any reasoning interface is downstream value, not the primary scientific claim.

For a sharper breakdown of strong and weak paper claims, see [NOVELTY_MAP.md](./NOVELTY_MAP.md).

## Threats And Limitations

The paper should be candid about:

- incomplete visibility under sparse sensing
- ambiguity in mapping processes to service identities
- short-lived or encrypted interactions
- environment-specific metadata gaps
- reliance on Linux and eBPF support

These limitations do not weaken the project; hiding them would.

## Artifact Quality Bar

To qualify as a serious public artifact, the repository should provide:

- reproducible build instructions
- one-command demo or scripted setup
- replayable sample traces
- benchmark scripts
- documented schemas
- honest status and limitation docs
- release tagging and versioned interfaces

## Near-Term Priorities

1. Make the sensing path real and testable.
2. Define canonical schema and provenance.
3. Build the first graph pipeline from replayed events.
4. Create benchmark workloads and evaluation scripts.
5. Expose a small but meaningful set of analysis queries.

## Success Condition

This project succeeds when an external reviewer can clone the repository, run a documented workload, produce a runtime-derived graph, inspect the resulting architecture, and reproduce the evaluation without needing private guidance.
