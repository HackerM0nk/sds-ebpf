#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
INPUT_PATH="${1:-$ROOT_DIR/eval/sample/canonical_events.ndjson}"
GROUND_TRUTH_PATH="${2:-$ROOT_DIR/eval/sample/ground_truth_edges.json}"
OUTPUT_DIR="${3:-$ROOT_DIR/eval/output}"
GRAPH_NAMESPACE="${GRAPH_NAMESPACE:-${4:-}}"

mkdir -p "$OUTPUT_DIR"

python3 "$ROOT_DIR/scripts/validate_canonical_events.py" --input "$INPUT_PATH"
if [[ -n "$GRAPH_NAMESPACE" ]]; then
  python3 "$ROOT_DIR/scripts/build_runtime_graph.py" --input "$INPUT_PATH" --output "$OUTPUT_DIR/runtime_graph.json" --namespace "$GRAPH_NAMESPACE"
else
  python3 "$ROOT_DIR/scripts/build_runtime_graph.py" --input "$INPUT_PATH" --output "$OUTPUT_DIR/runtime_graph.json"
fi
python3 "$ROOT_DIR/scripts/evaluate_runtime_graph.py" --graph "$OUTPUT_DIR/runtime_graph.json" --ground-truth "$GROUND_TRUTH_PATH" --output "$OUTPUT_DIR/eval_summary.json"

echo "wrote evaluation artifacts to $OUTPUT_DIR"
