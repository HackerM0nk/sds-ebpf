# Evaluation Harness

## Purpose

This directory contains the first reproducible evaluation loop for the repository.

It does not attempt to benchmark the final system yet. It exists to make the contract between canonical events, graph reconstruction, and measurable output explicit and runnable.

## Contents

- `sample/canonical_events.ndjson`
  Sample canonical runtime events for replay.
- `sample/ground_truth_edges.json`
  Expected service and external dependency edges.
- `output/`
  Generated graph snapshots and evaluation summaries.

## Run

From the repository root:

```bash
make eval-sample
```

Or run the loop directly:

```bash
bash scripts/run_eval_loop.sh
```

The first workload-backed demo artifact is available as:

```bash
make polar-demo
```

See [workloads/polar-bookshop/README.md](./workloads/polar-bookshop/README.md) for the packaged runtime capture, scenario, and workload-specific ground truth.

## What The Loop Produces

- `eval/output/runtime_graph.json`
  Graph snapshot generated from replayed canonical events.
- `eval/output/eval_summary.json`
  Precision, recall, and F1 against the sample ground truth.

## Why This Matters

This harness is the seed of the broader research methodology.

As the real agent matures, this directory should expand to include:

- benchmark workloads,
- real captured event traces,
- scenario-specific ground truth,
- and performance measurements.

The staged workload selection for that expansion is documented in [WORKLOAD_PLAN.md](./WORKLOAD_PLAN.md).
