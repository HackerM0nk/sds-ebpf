# Novelty Map

## Purpose

This document positions SDS eBPF against adjacent prior work and identifies which paper claims are actually defensible.

The goal is not to downplay the project. The goal is to prevent SDS from making weak novelty claims when there is a stronger systems contribution available.

## Executive View

SDS is not novel because it uses:

- eBPF for observability,
- graphs to represent dependencies,
- or an assistant layer to explain system structure.

All of those already exist in meaningful forms.

SDS becomes publishable if it demonstrates a stronger claim:

- sparse or adaptive runtime sensing can recover a useful, continuously updated architecture graph,
- under realistic partial visibility,
- with explicit confidence and provenance,
- at materially lower operational cost than full tracing or broad application instrumentation.

That is the boundary between a project with a compelling artifact and a project with a defensible paper.

## What Is Already Known

### 1. Runtime tracing and causal reconstruction are established

The broad idea of reconstructing distributed behavior from runtime evidence is not new.

- `X-Trace` framed cross-layer tracing as recovering a comprehensive view of service behavior and task trees across complex systems.
- `Pivot Tracing` showed that distributed-system monitoring can correlate behavior across components at runtime without redeploying code.

Implication for SDS:

SDS should not claim to have invented runtime reconstruction, cross-layer monitoring, or runtime correlation.

### 2. Whole-system provenance and graph-shaped runtime evidence are established

Systems such as `CamFlow` already treat low-level events as graph-structured provenance with explicit system-wide capture and downstream analysis value.

Implication for SDS:

SDS should not claim that representing runtime evidence as a graph is itself novel.

### 3. Service maps already exist in observability tooling

Current open-source tooling already builds service maps from telemetry:

- `OpenTelemetry servicegraphconnector` builds service-interaction maps from traces.
- `Cilium Hubble UI` provides service dependency and connectivity maps at L3/L4/L7.
- `Pixie` uses eBPF to auto-collect telemetry and expose service-level views in Kubernetes.

Implication for SDS:

SDS should not claim novelty as "we built a service map" or "we built graph-based observability."

### 4. eBPF-based zero-code tracing is an active area

Recent research already explores non-invasive tracing with eBPF:

- `Nahida` focuses on in-band distributed tracing with eBPF.
- `CrossTrace` focuses on efficient cross-thread and cross-service span correlation without source changes.
- `Retrofitting Service Dependency Discovery in Distributed Systems` targets process-level dependency reconstruction under NAT and partial deployment constraints.

Implication for SDS:

SDS should not lead with "zero-code eBPF tracing for microservices" as its novelty claim.

## Claims That Are Weak

These claims are real product statements, but weak paper statements:

### Weak claim 1

`We built an eBPF observability platform for distributed systems.`

Why weak:

- too broad,
- heavily covered by industry systems,
- not a clear scientific contribution.

### Weak claim 2

`We build a knowledge graph from runtime telemetry.`

Why weak:

- graph construction from runtime evidence already appears in provenance, tracing, and service-map systems,
- the interesting question is accuracy, confidence, and cost, not just graph existence.

### Weak claim 3

`We use GraphRAG and LLMs to explain infrastructure.`

Why weak:

- easy to dismiss as application glue,
- not a strong systems contribution,
- difficult to evaluate rigorously unless the substrate is already strong.

### Weak claim 4

`We automatically discover dependencies between services.`

Why weak:

- service discovery and dependency inference are already known problem areas,
- reviewers will immediately ask: compared to what, with what accuracy, at what overhead, and under what assumptions?

## Claims That Could Be Strong

These are more defensible because they identify a specific gap.

### Strong claim 1

`Sparse kernel-level telemetry is sufficient to recover a useful runtime architecture graph for modern service deployments at low overhead.`

Why stronger:

- measurable,
- comparative,
- naturally evaluated against full tracing or denser sensing.

### Strong claim 2

`A confidence-aware runtime graph can remain useful under partial deployment and incomplete visibility.`

Why stronger:

- tackles a real deployment constraint,
- aligns with the original SDS vision of runtime truth under imperfect observation,
- distinguishes SDS from systems that assume complete instrumentation.

### Strong claim 3

`A layered runtime substrate can separate evidence, identity resolution, graph inference, and explanation well enough to support multiple downstream workflows from one measured system model.`

Why stronger:

- broad enough for a platform paper,
- still rooted in architecture and engineering choices,
- lets SDS stay ambitious without becoming vague.

## Candidate Paper Theses

### Thesis A: Sparse Runtime Architecture Reconstruction

SDS shows that sparse eBPF sensing can reconstruct useful system structure without continuous full-fidelity tracing.

Core contribution:

- overhead versus fidelity tradeoff for dependency and architecture recovery.

What must be proven:

- graph precision and recall,
- update freshness,
- overhead under different sampling modes,
- degradation behavior under churn and partial capture.

Strength:

- clear systems story,
- direct evaluation path,
- strongest fit for current implementation trajectory.

Risk:

- if graph fidelity is weak, the paper collapses quickly.

### Thesis B: Confidence-Aware System Knowledge Graph Under Partial Visibility

SDS builds a temporal graph from incomplete runtime evidence and explicitly models uncertainty, provenance, and ambiguity.

Core contribution:

- graph semantics and inference discipline for partial deployment and partial observation.

What must be proven:

- confidence tracking improves usability or correctness,
- partial deployment still yields valuable graph structure,
- ambiguity is exposed rather than hidden.

Strength:

- more differentiated than plain service mapping,
- matches the long-term SDS architecture.

Risk:

- harder to evaluate cleanly than Thesis A unless the confidence model is rigorous.

### Thesis C: Runtime Systems Intelligence Substrate

SDS is a general artifact that turns runtime evidence into a reusable system model for architecture, security, and operations.

Core contribution:

- unified platform architecture across sensing, identity, graphing, and analysis.

What must be proven:

- the same graph substrate supports multiple serious tasks,
- architecture, security, and operations queries all work from one model,
- the artifact is reproducible and broadly useful.

Strength:

- most aligned with the repo vision.

Risk:

- easiest to make sound like a product brochure,
- novelty can blur unless anchored by one sharper technical result.

## Strongest Thesis For SDS

The strongest thesis today is:

`Thesis A, strengthened by part of Thesis B.`

In other words:

SDS should primarily argue that low-overhead sparse runtime sensing can recover a useful architecture graph, while explicitly modeling provenance and uncertainty when visibility is incomplete.

That gives SDS:

- a crisp measurement story,
- a real systems contribution,
- and a platform direction that can later support broader analysis and explanation layers.

It also aligns with the project's original README vision without forcing the first paper to prove everything at once.

## Recommended Novelty Boundary

SDS should describe itself as:

`a runtime architecture reconstruction substrate built from sparse kernel evidence, with explicit confidence and incremental graph construction`

SDS should avoid describing itself as:

- the first graph observability system,
- the first eBPF observability system,
- the first system to discover service dependencies,
- or an LLM-native infrastructure reasoning platform.

## What Must Be Built To Earn The Claim

To support the strongest thesis, SDS needs the following technical pieces:

1. Real sensing across process, network, and containment events.
2. Stable identity resolution from process and cgroup evidence to workload and service identity.
3. Incremental graph materialization with provenance and confidence fields.
4. Workload-specific ground truth for topology and dependency evaluation.
5. Comparative experiments across sparse, denser, and partial-deployment settings.

Without those pieces, the paper remains vision-heavy.

## Relation To Workload Choice

The novelty claim should drive workload selection.

The first workloads should make it possible to evaluate:

- declared architecture versus runtime-discovered architecture,
- partial visibility and missing identity context,
- internal versus external dependency recovery,
- and overhead under realistic traffic.

That is why SDS should use both:

- book-backed microservice applications with documented intent,
- and benchmark suites with established research credibility.

See [../eval/WORKLOAD_PLAN.md](../eval/WORKLOAD_PLAN.md).

## Source Pointers

The following sources are useful anchors for positioning:

- X-Trace: https://www.usenix.org/conference/nsdi-07/x-trace-pervasive-network-tracing-framework
- Pivot Tracing: https://www.usenix.org/publications/login/spring2016/mace
- CamFlow: https://camflow.org/publications/socc-2017.pdf
- OpenTelemetry connectors: https://opentelemetry.io/docs/collector/components/connector/
- OpenTelemetry service graph connector: https://pkg.go.dev/github.com/open-telemetry/opentelemetry-collector-contrib/connector/servicegraphconnector
- Cilium Hubble observability: https://docs.cilium.io/en/stable/observability/hubble/index.html
- Cilium Hubble UI: https://docs.cilium.io/en/stable/observability/hubble/hubble-ui.html
- Pixie overview: https://docs.px.dev/about-pixie/what-is-pixie/
- Pixie and eBPF: https://docs.px.dev/about-pixie/pixie-ebpf/
- Nahida: https://arxiv.org/abs/2311.09032
- CrossTrace: https://arxiv.org/abs/2508.11342
- Retrofitting Service Dependency Discovery in Distributed Systems: https://arxiv.org/abs/2510.15490
