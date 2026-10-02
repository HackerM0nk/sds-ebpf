# Polar Bookshop Demo Artifact

This workload is the first runnable SDS demo artifact.

It packages a documented microservice application, a bounded runtime capture, a controlled traffic scenario, and workload-specific ground truth so SDS can move beyond the sample replay harness.

## Source

- Application repo: https://github.com/ThomasVitale/cloud-native-spring-in-action
- Book context: https://www.manning.com/books/cloud-native-spring-in-action

The SDS workload package uses the published Polar Bookshop container images and vendors the small bootstrapping assets needed to keep the demo self-contained:

- local Config Server property files
- PostgreSQL init SQL
- RabbitMQ credentials
- a minimal Keycloak realm for runtime authentication

## What The Demo Does

1. Starts the SDS observer in a bounded single-cycle capture mode.
2. Brings up the Polar Bookshop stack with Docker Compose.
3. Exercises a deterministic scenario:
   - public catalog reads through `edge-service`
   - direct catalog reads against `catalog-service`
   - authenticated order submission against `order-service`
4. Builds a runtime graph filtered to the workload namespace.
5. Evaluates the inferred service edges against workload-specific ground truth.
6. Validates the artifact against workload acceptance gates.

## Run

From the repository root:

```bash
make polar-demo
```

## Artifacts

The demo writes artifacts under `eval/workloads/polar-bookshop/output/`:

- `capture/events_*.ndjson`
  Canonical SDS runtime events from the bounded capture window.
- `runs/<run-id>/runtime_graph.json`
  Namespace-filtered graph snapshot for a specific bounded workload run.
- `runs/<run-id>/eval_summary.json`
  Precision/recall summary against `ground_truth_edges.json`.
- `runs/<run-id>/artifact_validation.json`
  Acceptance-gate summary for whether the run is actually a valid Phase 1 demo artifact.
- `runs/<run-id>/artifact_manifest.json`
  Immutable bundle manifest with exact artifact paths, hashes, and top-line metrics.
- `latest_artifact.json`
  Pointer manifest for the most recent bundled artifact used by `make polar-validate`.

## Notes

- The observer relies on Docker runtime metadata for service naming, so the container runtime socket is mounted read-only into the observer container.
- The graph build is filtered to the Compose project namespace to avoid unrelated host activity polluting the demo result.
- The acceptance contract for the workload is defined in `acceptance.json`. This is the harness-level definition of what must be true before SDS treats a run as a valid demo artifact.
- `make polar-validate` re-validates the latest bundled artifact rather than guessing from loose files in `output/`.
