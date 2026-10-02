#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
WORKLOAD_DIR="$ROOT_DIR/eval/workloads/polar-bookshop"
COMPOSE_FILE="$WORKLOAD_DIR/docker-compose.yml"
PROJECT_NAME="${PROJECT_NAME:-sds-polar-demo}"
OBSERVER_IMAGE="${OBSERVER_IMAGE:-sds-observer:latest}"
CAPTURE_SECONDS="${CAPTURE_SECONDS:-120}"
OUTPUT_DIR="${OUTPUT_DIR:-$WORKLOAD_DIR/output}"
CAPTURE_DIR="$OUTPUT_DIR/capture"
EVAL_DIR="$OUTPUT_DIR/eval"
OBSERVER_NAME="${OBSERVER_NAME:-sds-polar-observer-demo}"
KEEP_RUNNING="${KEEP_RUNNING:-0}"
KEYCLOAK_PORT="${POLAR_KEYCLOAK_HOST_PORT:-18080}"
CATALOG_PORT="${POLAR_CATALOG_HOST_PORT:-19001}"
EDGE_PORT="${POLAR_EDGE_HOST_PORT:-19000}"

compose_cmd() {
  if command -v docker-compose >/dev/null 2>&1; then
    docker-compose "$@"
    return
  fi

  docker compose "$@"
}

link_or_copy() {
  local src="$1"
  local dst="$2"

  ln -f "$src" "$dst" 2>/dev/null || cp "$src" "$dst"
}

wait_for_url() {
  local url="$1"
  local attempts="${2:-90}"
  local sleep_seconds="${3:-2}"

  for ((i = 1; i <= attempts; i++)); do
    if curl -fsS "$url" >/dev/null 2>&1; then
      return 0
    fi
    sleep "$sleep_seconds"
  done

  echo "timed out waiting for $url" >&2
  return 1
}

wait_for_running_container() {
  local name="$1"
  local attempts="${2:-60}"
  local sleep_seconds="${3:-2}"
  local status

  for ((i = 1; i <= attempts; i++)); do
    status="$(docker inspect -f '{{.State.Status}}' "$name" 2>/dev/null || true)"
    if [[ "$status" == "running" ]]; then
      return 0
    fi
    if [[ "$status" == "exited" || "$status" == "dead" ]]; then
      docker logs --tail 80 "$name" >&2 || true
      echo "container $name entered state $status" >&2
      return 1
    fi
    sleep "$sleep_seconds"
  done

  echo "timed out waiting for container $name to be running" >&2
  return 1
}

cleanup() {
  if [[ "$KEEP_RUNNING" != "1" ]]; then
    docker rm -f "$OBSERVER_NAME" >/dev/null 2>&1 || true
    compose_cmd -f "$COMPOSE_FILE" -p "$PROJECT_NAME" down -v --remove-orphans >/dev/null 2>&1 || true
  fi
}

trap cleanup EXIT

mkdir -p "$CAPTURE_DIR" "$EVAL_DIR"

if ! docker image inspect "$OBSERVER_IMAGE" >/dev/null 2>&1; then
  docker build -t "$OBSERVER_IMAGE" -f "$ROOT_DIR/docker/Dockerfile.agent" "$ROOT_DIR"
fi

docker rm -f "$OBSERVER_NAME" >/dev/null 2>&1 || true
compose_cmd -f "$COMPOSE_FILE" -p "$PROJECT_NAME" down -v --remove-orphans >/dev/null 2>&1 || true
compose_cmd -f "$COMPOSE_FILE" -p "$PROJECT_NAME" pull

docker run -d \
  --name "$OBSERVER_NAME" \
  --privileged \
  --pid=host \
  --network=host \
  -v /sys/kernel/debug:/sys/kernel/debug:ro \
  -v /sys/fs/bpf:/sys/fs/bpf \
  -v /var/run/docker.sock:/var/run/docker.sock:ro \
  -v "$CAPTURE_DIR:/var/log/sds-observer" \
  "$OBSERVER_IMAGE" \
  --output /var/log/sds-observer \
  --sampling-on "$CAPTURE_SECONDS" \
  --sampling-off 0 \
  --run-once \
  --verbose >/dev/null

sleep 3

compose_cmd -f "$COMPOSE_FILE" -p "$PROJECT_NAME" up -d \
  polar-postgres \
  polar-redis \
  polar-rabbitmq \
  polar-keycloak \
  polar-ui

wait_for_running_container "${PROJECT_NAME}-polar-keycloak-1"
wait_for_url "http://localhost:${KEYCLOAK_PORT}/realms/PolarBookshop/.well-known/openid-configuration" 150 2

compose_cmd -f "$COMPOSE_FILE" -p "$PROJECT_NAME" up -d --no-deps config-service
wait_for_running_container "${PROJECT_NAME}-config-service-1"
sleep 10

compose_cmd -f "$COMPOSE_FILE" -p "$PROJECT_NAME" up -d --no-deps catalog-service
wait_for_running_container "${PROJECT_NAME}-catalog-service-1"
wait_for_url "http://localhost:${CATALOG_PORT}/books" 180 2

compose_cmd -f "$COMPOSE_FILE" -p "$PROJECT_NAME" up -d --no-deps order-service
wait_for_running_container "${PROJECT_NAME}-order-service-1"
sleep 20

compose_cmd -f "$COMPOSE_FILE" -p "$PROJECT_NAME" up -d --no-deps edge-service
wait_for_running_container "${PROJECT_NAME}-edge-service-1"
wait_for_url "http://localhost:${EDGE_PORT}/books" 180 2

bash "$WORKLOAD_DIR/scenarios/smoke.sh"

docker wait "$OBSERVER_NAME" >/dev/null

EVENTS_PATH="$(find "$CAPTURE_DIR" -maxdepth 1 -name 'events_*.ndjson' | sort | tail -n1)"
if [[ -z "$EVENTS_PATH" ]]; then
  echo "no canonical event file produced in $CAPTURE_DIR" >&2
  exit 1
fi

RUN_ID="${RUN_ID:-$(basename "$EVENTS_PATH" .ndjson)}"
RUN_DIR="$OUTPUT_DIR/runs/$RUN_ID"

GRAPH_NAMESPACE="$PROJECT_NAME" \
  bash "$ROOT_DIR/scripts/run_eval_loop.sh" \
  "$EVENTS_PATH" \
  "$WORKLOAD_DIR/ground_truth_edges.json" \
  "$EVAL_DIR"

VALIDATION_PATH="$EVAL_DIR/artifact_validation.json"
set +e
python3 "$ROOT_DIR/scripts/validate_demo_artifact.py" \
  --events "$EVENTS_PATH" \
  --graph "$EVAL_DIR/runtime_graph.json" \
  --eval-summary "$EVAL_DIR/eval_summary.json" \
  --acceptance "$WORKLOAD_DIR/acceptance.json" \
  --output "$VALIDATION_PATH"
VALIDATION_STATUS=$?
set -e

mkdir -p "$RUN_DIR"
link_or_copy "$EVENTS_PATH" "$RUN_DIR/$(basename "$EVENTS_PATH")"
cp "$EVAL_DIR/runtime_graph.json" "$RUN_DIR/runtime_graph.json"
cp "$EVAL_DIR/eval_summary.json" "$RUN_DIR/eval_summary.json"
cp "$VALIDATION_PATH" "$RUN_DIR/artifact_validation.json"
cp "$WORKLOAD_DIR/acceptance.json" "$RUN_DIR/acceptance.json"

python3 "$ROOT_DIR/scripts/write_demo_artifact_manifest.py" \
  --workload polar-bookshop \
  --namespace "$PROJECT_NAME" \
  --events "$RUN_DIR/$(basename "$EVENTS_PATH")" \
  --graph "$RUN_DIR/runtime_graph.json" \
  --eval-summary "$RUN_DIR/eval_summary.json" \
  --validation "$RUN_DIR/artifact_validation.json" \
  --acceptance "$RUN_DIR/acceptance.json" \
  --output "$RUN_DIR/artifact_manifest.json"

cp "$RUN_DIR/artifact_manifest.json" "$OUTPUT_DIR/latest_artifact.json"

echo "Polar demo artifacts:"
echo "  events:    $RUN_DIR/$(basename "$EVENTS_PATH")"
echo "  graph:     $RUN_DIR/runtime_graph.json"
echo "  eval:      $RUN_DIR/eval_summary.json"
echo "  check:     $RUN_DIR/artifact_validation.json"
echo "  manifest:  $RUN_DIR/artifact_manifest.json"

exit "$VALIDATION_STATUS"
