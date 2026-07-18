#!/usr/bin/env bash
set -euo pipefail

AXIS_COMPANION_BASE_URL="${AXIS_COMPANION_BASE_URL:-}"
AXIS_COMPANION_API_KEY="${AXIS_COMPANION_API_KEY:-}"
PONTI_BASE_URL="${PONTI_BASE_URL:-}"
PONTI_ORG_ID="${PONTI_ORG_ID:-}"
PONTI_AXIS_API_KEY="${PONTI_AXIS_API_KEY:-}"
AXIS_ACTOR_ID="${AXIS_ACTOR_ID:-ponti-smoke}"
AXIS_SCOPES="${AXIS_SCOPES:-companion:products:admin companion:runtime:admin companion:capabilities:admin companion:connectors:execute ponti:insights:read}"

require_env() {
  local name="$1"
  local value="${!name:-}"
  if [[ -z "${value}" ]]; then
    echo "missing required env: ${name}" >&2
    exit 2
  fi
}

axis_headers=()

setup_axis_headers() {
  axis_headers=(
    -H "Accept: application/json"
    -H "X-API-Key: ${AXIS_COMPANION_API_KEY}"
    -H "X-Org-ID: ${PONTI_ORG_ID}"
    -H "X-User-ID: ${AXIS_ACTOR_ID}"
    -H "X-On-Behalf-Of: ${AXIS_ACTOR_ID}"
    -H "X-Product-Surface: ponti"
    -H "X-Auth-Scopes: ${AXIS_SCOPES}"
  )
}

axis_request() {
  local method="$1"
  local path="$2"
  local body="${3:-}"
  local url="${AXIS_COMPANION_BASE_URL%/}${path}"
  if [[ -n "${body}" ]]; then
    curl -fsS -X "${method}" "${url}" \
      "${axis_headers[@]}" \
      -H "Content-Type: application/json" \
      --data "${body}"
  else
    curl -fsS -X "${method}" "${url}" "${axis_headers[@]}"
  fi
}

validate_ponti_direct() {
  local raw
  raw="$(curl -fsS "${PONTI_BASE_URL%/}/api/v1/capabilities" \
    -H "Accept: application/json" \
    -H "Authorization: Bearer ${PONTI_AXIS_API_KEY}")"
  RAW_JSON="${raw}" python3 - <<'PY'
import json
import os
import sys

data = json.loads(os.environ["RAW_JSON"])
items = data.get("items")
if not isinstance(items, list) or not items:
    print("Ponti capabilities response must contain items[]", file=sys.stderr)
    sys.exit(1)
manifest = next((m for m in items if isinstance(m, dict) and m.get("id") == "ponti.insights"), None)
if not manifest:
    print("Ponti capabilities response must include ponti.insights", file=sys.stderr)
    sys.exit(1)
tools = {t.get("name") for t in manifest.get("tools", []) if isinstance(t, dict)}
missing = {"ponti.insights.list", "ponti.insights.summary", "ponti.insights.explain"} - tools
if missing:
    print(f"Ponti capabilities missing tools: {sorted(missing)}", file=sys.stderr)
    sys.exit(1)
PY
}

validate_ponti_capability_discovery() {
  local raw
  raw="$(axis_request GET "/v1/connectors/capabilities")"
  RAW_JSON="${raw}" python3 - <<'PY'
import json
import os
import sys

data = json.loads(os.environ["RAW_JSON"])
for connector in data.get("connectors", []):
    caps = connector.get("capabilities", [])
    for cap in caps:
        op = cap.get("operation") or cap.get("id")
        if op == "ponti.insights.summary":
            sys.exit(0)
print("Ponti connector with ponti.insights.summary not found. Start Axis Companion with PONTI_BASE_URL and PONTI_API_KEY, then run POST /v1/connectors/refresh.", file=sys.stderr)
sys.exit(1)
PY
}

ensure_ponti_connector_id() {
  local raw
  raw="$(axis_request GET "/v1/connectors")"
  local existing
  existing="$(RAW_JSON="${raw}" python3 - <<'PY'
import json
import os

data = json.loads(os.environ["RAW_JSON"])
for connector in data.get("connectors", []):
    if connector.get("kind") == "ponti":
        print(connector.get("id", ""))
        break
PY
)"
  if [[ -n "${existing}" ]]; then
    echo "${existing}"
    return
  fi
  raw="$(axis_request POST "/v1/connectors" '{
    "name": "Ponti",
    "kind": "ponti",
    "enabled": true,
    "config": {}
  }')"
  RAW_JSON="${raw}" python3 - <<'PY'
import json
import os
import sys

data = json.loads(os.environ["RAW_JSON"])
connector_id = data.get("id")
if not connector_id:
    print(f"created connector response missing id: {data}", file=sys.stderr)
    sys.exit(1)
print(connector_id)
PY
}

validate_execution() {
  local raw="$1"
  local expected_operation="$2"
  RAW_JSON="${raw}" EXPECTED_OPERATION="${expected_operation}" python3 - <<'PY'
import json
import os
import sys

data = json.loads(os.environ["RAW_JSON"])
expected = os.environ["EXPECTED_OPERATION"]
if data.get("operation") != expected:
    print(f"operation mismatch: {data.get('operation')} != {expected}", file=sys.stderr)
    sys.exit(1)
if data.get("status") != "success":
    print(f"execution status must be success: {data}", file=sys.stderr)
    sys.exit(1)
if data.get("org_id") != os.environ.get("PONTI_ORG_ID"):
    print(f"org_id mismatch: {data.get('org_id')} != {os.environ.get('PONTI_ORG_ID')}", file=sys.stderr)
    sys.exit(1)
result = data.get("result") or {}
if expected.endswith(".summary"):
    if "summary" not in result or "evidence" not in result:
        print(f"summary result missing summary/evidence: {result}", file=sys.stderr)
        sys.exit(1)
else:
    if "items" not in result or "evidence" not in result:
        print(f"list result missing items/evidence: {result}", file=sys.stderr)
        sys.exit(1)
evidence = data.get("evidence") or {}
action_binding = evidence.get("action_binding") or {}
if evidence.get("product_surface") != "ponti" and action_binding.get("product_surface") != "ponti":
    print(f"execution evidence missing product_surface=ponti: {evidence}", file=sys.stderr)
    sys.exit(1)
PY
}

require_env AXIS_COMPANION_BASE_URL
require_env AXIS_COMPANION_API_KEY
require_env PONTI_BASE_URL
require_env PONTI_ORG_ID
require_env PONTI_AXIS_API_KEY
setup_axis_headers

echo "Validating Ponti product capabilities directly..."
validate_ponti_direct

echo "Running Ponti onboarding against Axis..."
"$(dirname "$0")/onboard-ponti.sh"

echo "Refreshing Axis connectors..."
axis_request POST "/v1/connectors/refresh" '{}' >/dev/null

echo "Validating Ponti capability discovery in Axis..."
validate_ponti_capability_discovery

echo "Ensuring persisted Ponti connector config in Axis..."
CONNECTOR_ID="$(ensure_ponti_connector_id)"
if [[ -z "${CONNECTOR_ID}" ]]; then
  echo "Ponti connector id is empty" >&2
  exit 1
fi

echo "Executing ponti.insights.summary through Axis connector ${CONNECTOR_ID}..."
summary_raw="$(axis_request POST "/v1/connectors/execute" '{
  "connector_id": "'"${CONNECTOR_ID}"'",
  "operation": "ponti.insights.summary",
  "payload": {}
}')"
validate_execution "${summary_raw}" "ponti.insights.summary"

echo "Executing ponti.insights.list through Axis connector ${CONNECTOR_ID}..."
list_raw="$(axis_request POST "/v1/connectors/execute" '{
  "connector_id": "'"${CONNECTOR_ID}"'",
  "operation": "ponti.insights.list",
  "payload": {
    "limit": 5,
    "include_resolved": true
  }
}')"
validate_execution "${list_raw}" "ponti.insights.list"

echo "Ponti Axis read-only smoke complete."
