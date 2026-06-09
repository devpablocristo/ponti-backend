#!/usr/bin/env bash
set -euo pipefail

AXIS_COMPANION_BASE_URL="${AXIS_COMPANION_BASE_URL:-}"
AXIS_COMPANION_API_KEY="${AXIS_COMPANION_API_KEY:-}"
PONTI_BASE_URL="${PONTI_BASE_URL:-}"
PONTI_ORG_ID="${PONTI_ORG_ID:-}"
PONTI_AXIS_API_KEY="${PONTI_AXIS_API_KEY:-}"
PONTI_PROJECT_ID="${PONTI_PROJECT_ID:-1}"
AXIS_ACTOR_ID="${AXIS_ACTOR_ID:-ponti-draft-smoke}"
AXIS_SCOPES="${AXIS_SCOPES:-companion:products:admin companion:runtime:admin companion:capabilities:admin companion:connectors:execute ponti:insights:read ponti:actions:prepare}"

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

axis_request_allow_failure() {
  local method="$1"
  local path="$2"
  local body="${3:-}"
  local url="${AXIS_COMPANION_BASE_URL%/}${path}"
  local response_file
  response_file="$(mktemp)"
  local status
  status="$(curl -sS -o "${response_file}" -w "%{http_code}" -X "${method}" "${url}" \
    "${axis_headers[@]}" \
    -H "Content-Type: application/json" \
    --data "${body}")"
  printf '%s\n%s\n' "${status}" "$(cat "${response_file}")"
  rm -f "${response_file}"
}

validate_ponti_direct() {
  local raw
  raw="$(curl -fsS "${PONTI_BASE_URL%/}/api/v1/capabilities" \
    -H "Accept: application/json" \
    -H "X-Tenant-Id: ${PONTI_ORG_ID}" \
    -H "Authorization: Bearer ${PONTI_AXIS_API_KEY}")"
  RAW_JSON="${raw}" python3 - <<'PY'
import json
import os
import sys

data = json.loads(os.environ["RAW_JSON"])
manifest = next((m for m in data.get("items", []) if isinstance(m, dict) and m.get("id") == "ponti.insights"), None)
if not manifest:
    print("Ponti capabilities response must include ponti.insights", file=sys.stderr)
    sys.exit(1)
tools = {t.get("name"): t for t in manifest.get("tools", []) if isinstance(t, dict)}
missing = {
    "ponti.insight.resolve.prepare",
    "ponti.workorder.draft.prepare",
    "ponti.stock_adjustment.prepare",
} - set(tools)
if missing:
    print(f"Ponti capabilities missing draft tools: {sorted(missing)}", file=sys.stderr)
    sys.exit(1)
for name in sorted(missing):
    print(name)
for name in ("ponti.insight.resolve.prepare", "ponti.workorder.draft.prepare", "ponti.stock_adjustment.prepare"):
    tool = tools[name]
    governance = tool.get("governance") or {}
    if tool.get("mode") != "write" or tool.get("side_effect") is not True:
        print(f"{name} must be write side_effect=true: {tool}", file=sys.stderr)
        sys.exit(1)
    if governance.get("requires_approval") is not True or governance.get("action_type") != "agent.capability.invoke":
        print(f"{name} must require Nexus approval: {tool}", file=sys.stderr)
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

validate_axis_discovery() {
  local raw
  raw="$(axis_request GET "/v1/connectors/capabilities?include_writes=true")"
  RAW_JSON="${raw}" python3 - <<'PY'
import json
import os
import sys

data = json.loads(os.environ["RAW_JSON"])
found = {}
for connector in data.get("connectors", []):
    if connector.get("kind") != "ponti":
        continue
    for cap in connector.get("capabilities", []):
        op = cap.get("operation") or cap.get("id")
        if op in {"ponti.insight.resolve.prepare", "ponti.workorder.draft.prepare", "ponti.stock_adjustment.prepare"}:
            found[op] = cap
missing = {"ponti.insight.resolve.prepare", "ponti.workorder.draft.prepare", "ponti.stock_adjustment.prepare"} - set(found)
if missing:
    print(f"Axis Ponti connector missing draft capabilities: {sorted(missing)}", file=sys.stderr)
    sys.exit(1)
for op, cap in found.items():
    if cap.get("read_only") is not False or cap.get("side_effect") is not True:
        print(f"{op} must be write side-effect in Axis: {cap}", file=sys.stderr)
        sys.exit(1)
    if cap.get("requires_nexus_approval") is not True:
        print(f"{op} must require Nexus in Axis: {cap}", file=sys.stderr)
        sys.exit(1)
    if cap.get("nexus_action_type") != "agent.capability.invoke":
        print(f"{op} must use Nexus action type agent.capability.invoke: {cap}", file=sys.stderr)
        sys.exit(1)
    if (cap.get("idempotency") or {}).get("required") is not True and cap.get("idempotency_mode") != "required":
        print(f"{op} must require idempotency in Axis: {cap}", file=sys.stderr)
        sys.exit(1)
PY
}

validate_ungated_block() {
  local connector_id="$1"
  local payload
  payload='{
    "connector_id": "'"${connector_id}"'",
    "operation": "ponti.workorder.draft.prepare",
    "idempotency_key": "ponti-draft-smoke-ungated",
    "payload": {
      "project_id": '"${PONTI_PROJECT_ID}"',
      "work_type": "smoke-preview",
      "workspace": {
        "project_id": '"${PONTI_PROJECT_ID}"'
      }
    }
  }'
  local response
  response="$(axis_request_allow_failure POST "/v1/connectors/execute" "${payload}")"
  local status
  local body
  status="$(printf '%s\n' "${response}" | sed -n '1p')"
  body="$(printf '%s\n' "${response}" | sed -n '2,$p')"
  STATUS="${status}" RAW_JSON="${body}" python3 - <<'PY'
import json
import os
import sys

if os.environ["STATUS"] != "403":
    print(f"expected HTTP 403 for ungated execution, got {os.environ['STATUS']}: {os.environ['RAW_JSON']}", file=sys.stderr)
    sys.exit(1)
data = json.loads(os.environ["RAW_JSON"])
if data.get("code") != "UNGATED":
    print(f"expected code=UNGATED, got {data}", file=sys.stderr)
    sys.exit(1)
if "nexus" not in (data.get("message") or "").lower():
    print(f"expected Nexus approval message, got {data}", file=sys.stderr)
    sys.exit(1)
PY
}

require_env AXIS_COMPANION_BASE_URL
require_env AXIS_COMPANION_API_KEY
require_env PONTI_BASE_URL
require_env PONTI_ORG_ID
require_env PONTI_AXIS_API_KEY
setup_axis_headers

echo "Validating Ponti published draft action capabilities directly..."
validate_ponti_direct

echo "Running Ponti onboarding against Axis..."
"$(dirname "$0")/onboard-ponti.sh"

echo "Refreshing Axis connectors..."
axis_request POST "/v1/connectors/refresh" '{}' >/dev/null

echo "Validating draft action discovery in Axis..."
validate_axis_discovery

echo "Ensuring persisted Ponti connector config in Axis..."
CONNECTOR_ID="$(ensure_ponti_connector_id)"
if [[ -z "${CONNECTOR_ID}" ]]; then
  echo "Ponti connector id is empty" >&2
  exit 1
fi

echo "Validating Axis blocks draft preview execution without Nexus..."
validate_ungated_block "${CONNECTOR_ID}"

echo "Ponti Axis draft action governance smoke complete."
