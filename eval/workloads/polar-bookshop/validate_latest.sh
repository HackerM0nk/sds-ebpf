#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
WORKLOAD_DIR="$ROOT_DIR/eval/workloads/polar-bookshop"
OUTPUT_DIR="$WORKLOAD_DIR/output"
LATEST_MANIFEST="$OUTPUT_DIR/latest_artifact.json"

link_or_copy() {
  local src="$1"
  local dst="$2"

  ln -f "$src" "$dst" 2>/dev/null || cp "$src" "$dst"
}

if [[ ! -f "$LATEST_MANIFEST" ]]; then
  CAPTURE_DIR="$OUTPUT_DIR/capture"
  EVAL_DIR="$OUTPUT_DIR/eval"
  EVENTS_PATH="$(find "$CAPTURE_DIR" -maxdepth 1 -name 'events_*.ndjson' | sort | tail -n1)"

  if [[ -z "$EVENTS_PATH" ]]; then
    echo "no bundled artifact manifest found and no legacy capture artifact found under $CAPTURE_DIR" >&2
    exit 1
  fi

  RUN_ID="$(basename "$EVENTS_PATH" .ndjson)"
  RUN_DIR="$OUTPUT_DIR/runs/$RUN_ID"
  mkdir -p "$RUN_DIR"

  link_or_copy "$EVENTS_PATH" "$RUN_DIR/$(basename "$EVENTS_PATH")"
  cp "$EVAL_DIR/runtime_graph.json" "$RUN_DIR/runtime_graph.json"
  cp "$EVAL_DIR/eval_summary.json" "$RUN_DIR/eval_summary.json"
  cp "$WORKLOAD_DIR/acceptance.json" "$RUN_DIR/acceptance.json"

  set +e
  python3 "$ROOT_DIR/scripts/validate_demo_artifact.py" \
    --events "$RUN_DIR/$(basename "$EVENTS_PATH")" \
    --graph "$RUN_DIR/runtime_graph.json" \
    --eval-summary "$RUN_DIR/eval_summary.json" \
    --acceptance "$RUN_DIR/acceptance.json" \
    --output "$RUN_DIR/artifact_validation.json"
  VALIDATION_STATUS=$?
  set -e

  python3 "$ROOT_DIR/scripts/write_demo_artifact_manifest.py" \
    --workload polar-bookshop \
    --namespace sds-polar-demo \
    --events "$RUN_DIR/$(basename "$EVENTS_PATH")" \
    --graph "$RUN_DIR/runtime_graph.json" \
    --eval-summary "$RUN_DIR/eval_summary.json" \
    --validation "$RUN_DIR/artifact_validation.json" \
    --acceptance "$RUN_DIR/acceptance.json" \
    --output "$RUN_DIR/artifact_manifest.json"

  cp "$RUN_DIR/artifact_manifest.json" "$LATEST_MANIFEST"
  exit "$VALIDATION_STATUS"
fi

mapfile -t ARTIFACT_PATHS < <(
  python3 - "$LATEST_MANIFEST" <<'PY'
import json
import sys

manifest = json.load(open(sys.argv[1], "r", encoding="utf-8"))
artifacts = manifest["artifacts"]
print(artifacts["events"]["path"])
print(artifacts["graph"]["path"])
print(artifacts["eval_summary"]["path"])
print(artifacts["acceptance"]["path"])
print(artifacts["validation"]["path"])
PY
)

EVENTS_PATH="${ARTIFACT_PATHS[0]}"
GRAPH_PATH="${ARTIFACT_PATHS[1]}"
EVAL_SUMMARY_PATH="${ARTIFACT_PATHS[2]}"
ACCEPTANCE_PATH="${ARTIFACT_PATHS[3]}"
VALIDATION_PATH="${ARTIFACT_PATHS[4]}"

python3 "$ROOT_DIR/scripts/validate_demo_artifact.py" \
  --events "$EVENTS_PATH" \
  --graph "$GRAPH_PATH" \
  --eval-summary "$EVAL_SUMMARY_PATH" \
  --acceptance "$ACCEPTANCE_PATH" \
  --output "$VALIDATION_PATH"
