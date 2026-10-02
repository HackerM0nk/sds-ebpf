#!/usr/bin/env bash
set -euo pipefail

WORKLOAD_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
EDGE_PORT="${POLAR_EDGE_HOST_PORT:-19000}"
CATALOG_PORT="${POLAR_CATALOG_HOST_PORT:-19001}"
ORDER_PORT="${POLAR_ORDER_HOST_PORT:-19002}"
KEYCLOAK_PORT="${POLAR_KEYCLOAK_HOST_PORT:-18080}"
OIDC_ISSUER="http://polar-keycloak:8080/realms/PolarBookshop"
OIDC_KID="kpxalusuCh8JSkU3"
OIDC_PRIVATE_KEY="$WORKLOAD_DIR/oidc/private.pem"

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

mint_access_token() {
  local now exp unsigned signature

  now="$(date +%s)"
  exp="$((now + 3600))"
  unsigned="$(
    python3 - "$OIDC_KID" "$OIDC_ISSUER" "$now" "$exp" <<'PY'
import base64
import json
import sys


def encode_json(payload):
    raw = json.dumps(payload, separators=(",", ":"), sort_keys=True).encode("utf-8")
    return base64.urlsafe_b64encode(raw).rstrip(b"=").decode("ascii")


kid, issuer, issued_at, expires_at = sys.argv[1], sys.argv[2], int(sys.argv[3]), int(sys.argv[4])
header = {"alg": "RS256", "kid": kid, "typ": "JWT"}
payload = {
    "aud": ["account", "catalog-service", "edge-service", "order-service"],
    "azp": "sds-cli",
    "exp": expires_at,
    "iat": issued_at,
    "iss": issuer,
    "preferred_username": "bjorn",
    "realm_access": {"roles": ["customer"]},
    "scope": "openid profile orders",
    "scp": ["openid", "profile", "orders"],
    "sub": "bjorn",
}
print(f"{encode_json(header)}.{encode_json(payload)}")
PY
  )"

  signature="$(
    printf '%s' "$unsigned" \
      | openssl dgst -binary -sha256 -sign "$OIDC_PRIVATE_KEY" \
      | python3 - <<'PY'
import base64
import sys

print(base64.urlsafe_b64encode(sys.stdin.buffer.read()).rstrip(b"=").decode("ascii"))
PY
  )"

  printf '%s.%s' "$unsigned" "$signature"
}

wait_for_url "http://localhost:${KEYCLOAK_PORT}/realms/PolarBookshop/.well-known/openid-configuration"
wait_for_url "http://localhost:${CATALOG_PORT}/books"
wait_for_url "http://localhost:${EDGE_PORT}/books"

CATALOG_RESPONSE="$(curl -fsS "http://localhost:${CATALOG_PORT}/books")"
BOOK_ISBN="$(
  CATALOG_RESPONSE="$CATALOG_RESPONSE" python3 - <<'PY'
import json
import os

payload = json.loads(os.environ["CATALOG_RESPONSE"])
first = payload[0]
print(first["isbn"])
PY
)"

ACCESS_TOKEN="$(mint_access_token)"

curl -fsS "http://localhost:${EDGE_PORT}/books" >/dev/null
curl -fsS "http://localhost:${CATALOG_PORT}/books" >/dev/null

curl -fsS \
  -X POST \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H 'Content-Type: application/json' \
  -d "{\"isbn\":\"$BOOK_ISBN\",\"quantity\":1}" \
  "http://localhost:${ORDER_PORT}/orders" >/dev/null

curl -fsS \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  "http://localhost:${ORDER_PORT}/orders" >/dev/null

echo "Polar Bookshop scenario completed using ISBN $BOOK_ISBN"
