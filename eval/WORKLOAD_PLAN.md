# Workload Plan

## Purpose

This document defines the staged workload plan for SDS evaluation.

The evaluation strategy should not rely on a single benchmark family. SDS needs:

- at least one workload with documented intended architecture,
- at least one workload with research-benchmark credibility,
- and at least one moderately manageable integration target for rapid iteration.

That combination gives SDS both engineering usefulness and paper-quality evaluation value.

## Selection Principles

Workloads should be chosen based on the following criteria:

- `Architecture clarity`
  The intended service structure should be documented well enough to derive expected dependencies.

- `Runtime richness`
  The workload should exercise process, network, and possibly asynchronous behavior.

- `Reproducibility`
  It should be runnable with pinned artifacts and documented setup.

- `Deployment variety`
  SDS should not depend on only one environment model.

- `Ground-truth quality`
  It should be practical to derive expected topology and dependency edges from manifests, configs, and scenario design.

## Recommended Staged Workloads

### Stage 1: Polar Bookshop

Source:

- Manning book page: https://www.manning.com/books/cloud-native-spring-in-action
- Official code repo: https://github.com/ThomasVitale/cloud-native-spring-in-action

Why first:

- book-backed and human-readable,
- modern cloud-native application structure,
- moderate complexity,
- suitable for getting SDS from local runtime capture to graph output quickly.

What it is good for:

- first end-to-end SDS demo,
- identity enrichment and service naming,
- runtime versus documented topology comparison,
- architecture-diagram generation from live evidence.

Expected use in SDS:

- first workload that reviewers can understand quickly,
- first workload used for screenshots and walkthroughs,
- first workload used to validate graph semantics.

### Stage 2: FTGO

Source:

- Official repo: https://github.com/microservices-patterns/ftgo-application
- Supporting book context: https://microservices.io/post/microservices/patterns/data/2019/07/15/ftgo-database-per-service.html

Why second:

- richer architecture than a simple demo app,
- includes API gateway, Kafka, sagas, CQRS, and event-driven patterns,
- strong fit for SDS because intended architecture is documented in the book ecosystem.

Why not first:

- higher operational complexity,
- more memory-intensive,
- documentation in the repo itself is intentionally sparse.

What it is good for:

- service dependency reconstruction beyond simple HTTP,
- testing SDS against event-driven and gateway-oriented flows,
- evaluating how well SDS handles richer architecture patterns.

Expected use in SDS:

- deeper architecture case study,
- internal dependency recovery evaluation,
- graph drift and partial-visibility scenarios.

### Stage 3: Online Boutique

Source:

- Official repo: https://github.com/GoogleCloudPlatform/microservices-demo

Why third:

- widely recognized microservice demo,
- polyglot services,
- Kubernetes-native deployment,
- gRPC-heavy communication patterns.

What it is good for:

- broader external credibility,
- cross-language deployment behavior,
- Kubernetes-focused evaluation,
- testing SDS against a known cloud-native application used widely in demos and studies.

Expected use in SDS:

- benchmark-style public demo workload,
- Kubernetes graph recovery experiments,
- comparison of declared manifests to runtime-observed communication.

### Stage 4: DeathStarBench

Source:

- Official suite: https://github.com/delimitrou/DeathStarBench
- Benchmark paper reference from repo README: ASPLOS 2019

Why fourth:

- strongest research-benchmark credibility among the candidates,
- multiple end-to-end services with different communication patterns,
- standard point of comparison in microservices systems work.

Why not first:

- heavier setup and operational burden,
- less friendly as the first SDS integration target,
- better suited after the graph and identity layers are stable.

What it is good for:

- paper-grade experiments,
- overhead and fidelity measurements,
- robustness under higher concurrency and more diverse service graphs.

Expected use in SDS:

- benchmark evaluation,
- sparse versus denser sensing experiments,
- reproducibility and artifact-review validation.

## Optional Supplemental Workload

### Microservices in Action examples

Source:

- Repo: https://github.com/morganjbruce/microservices-in-action

Use only as a supplemental source, not the primary evaluation target.

Why:

- useful as book-backed example material,
- less attractive than Polar Bookshop or FTGO as the main SDS case study,
- more appropriate for sidecar examples or sanity checks than for the core paper story.

## Final Recommendation

SDS should use this progression:

1. `Polar Bookshop` for first integration and first end-to-end demo.
2. `FTGO` for richer documented architecture and event-driven behavior.
3. `Online Boutique` for broad community recognition and Kubernetes realism.
4. `DeathStarBench` for benchmark-grade evaluation.

This sequence balances:

- ease of integration,
- architecture richness,
- and research credibility.

## Evaluation Roles Per Workload

### Polar Bookshop

Role:

- engineering integration target
- architecture walkthrough target
- first graph-quality case study

### FTGO

Role:

- documented-architecture case study
- richer dependency reconstruction target
- event-driven architecture stress case

### Online Boutique

Role:

- public demo workload
- polyglot Kubernetes target
- external reproducibility baseline

### DeathStarBench

Role:

- research benchmark
- performance and robustness target
- paper tables and comparisons

## Ground Truth Strategy

Ground truth must not be guessed from README prose alone.

For each workload, SDS should derive:

1. `Declared topology`
   From manifests, Compose files, Helm charts, service definitions, and the workload's own architecture docs.

2. `Scenario-specific expected edges`
   From controlled request flows and scenario scripts.

3. `Observed runtime edges`
   From SDS canonical events and graph output.

4. `Disagreement set`
   Edges present only in docs, only at runtime, or only under specific scenarios.

This lets SDS evaluate not only recovery accuracy, but also architecture drift and undocumented behavior.

## Suggested Repo Shape

The evaluation area should evolve toward:

```text
eval/
├── workloads/
│   ├── polar-bookshop/
│   ├── ftgo/
│   ├── online-boutique/
│   └── deathstarbench/
├── traces/
├── ground-truth/
├── scenarios/
└── output/
```

For each workload, SDS should eventually keep:

- a source pointer,
- setup notes,
- scenario definitions,
- expected dependency edges,
- and captured canonical traces.

## Near-Term Build Order

### 1. Polar Bookshop

Build first:

- workload setup notes,
- service inventory,
- expected edges file,
- one scripted traffic scenario,
- one captured SDS trace,
- one runtime graph snapshot.

### 2. FTGO

Build second:

- service inventory,
- dependency hypotheses,
- gateway and async path scenarios,
- richer graph and drift checks.

### 3. Online Boutique

Build third:

- Kubernetes deployment path,
- gRPC and service-mesh-aware scenario design,
- cluster-level identity checks.

### 4. DeathStarBench

Build fourth:

- selected application choice,
- pinned version,
- benchmark traffic generation,
- overhead and fidelity experiment scripts.

## Success Condition

The workload plan succeeds when SDS can demonstrate:

- one understandable book-backed architecture,
- one richer book-backed architecture,
- one mainstream public microservice app,
- and one recognized benchmark suite,

all feeding the same canonical event and graph pipeline.
