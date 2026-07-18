#!/usr/bin/env bash
set -euo pipefail

AXIS_COMPANION_BASE_URL="${AXIS_COMPANION_BASE_URL:-}"
AXIS_COMPANION_API_KEY="${AXIS_COMPANION_API_KEY:-}"
NEXUS_BASE_URL="${NEXUS_BASE_URL:-}"
NEXUS_ADMIN_API_KEY="${NEXUS_ADMIN_API_KEY:-${NEXUS_API_KEY:-}}"
PONTI_BASE_URL="${PONTI_BASE_URL:-}"
PONTI_ORG_ID="${PONTI_ORG_ID:-}"
PONTI_AXIS_API_KEY="${PONTI_AXIS_API_KEY:-}"
PONTI_PROJECT_ID="${PONTI_PROJECT_ID:-1}"
PONTI_CUSTOMER_ID="${PONTI_CUSTOMER_ID:-1}"
PONTI_FIELD_ID="${PONTI_FIELD_ID:-1}"
PONTI_LOT_ID="${PONTI_LOT_ID:-1}"
PONTI_CROP_ID="${PONTI_CROP_ID:-1}"
PONTI_LABOR_ID="${PONTI_LABOR_ID:-1}"
PONTI_INVESTOR_ID="${PONTI_INVESTOR_ID:-1}"
AXIS_ACTOR_ID="${AXIS_ACTOR_ID:-ponti-nexus-approved-smoke}"
NEXUS_REQUESTER_ID="${NEXUS_REQUESTER_ID:-ponti-nexus-smoke-requester}"
NEXUS_APPROVER_ID="${NEXUS_APPROVER_ID:-ponti-nexus-smoke-approver}"
AXIS_SCOPES="${AXIS_SCOPES:-companion:products:admin companion:runtime:admin companion:capabilities:admin companion:connectors:execute ponti:actions:prepare ponti:actions:draft}"
NEXUS_SCOPES="${NEXUS_SCOPES:-nexus:requests:read nexus:requests:write nexus:requests:result nexus:approvals:decide nexus:policies:admin nexus:cross_org}"
OPERATION="ponti.workorder_draft.create"
NEXUS_ACTION_TYPE="agent.capability.invoke"
NEXUS_POLICY_NAME="ponti-agent-capability-invoke-require-approval"
IDEMPOTENCY_KEY="${IDEMPOTENCY_KEY:-ponti-nexus-approved-draft-$(date +%s)}"

require_env() {
  local name="$1"
  local value="${!name:-}"
  if [[ -z "${value}" ]]; then
    echo "missing required env: ${name}" >&2
    exit 2
  fi
}

axis_headers=()

setup_headers() {
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

nexus_request() {
  local method="$1"
  local path="$2"
  local body="${3:-}"
  local idempotency="${4:-}"
  local actor="${5:-${NEXUS_REQUESTER_ID}}"
  local url="${NEXUS_BASE_URL%/}${path}"
  local headers=(
    -H "Accept: application/json"
    -H "X-API-Key: ${NEXUS_ADMIN_API_KEY}"
    -H "X-Org-ID: ${PONTI_ORG_ID}"
    -H "X-User-ID: ${actor}"
    -H "X-Auth-Scopes: ${NEXUS_SCOPES}"
  )
  if [[ -n "${idempotency}" ]]; then
    headers+=(-H "Idempotency-Key: ${idempotency}")
  fi
  if [[ -n "${body}" ]]; then
    curl -fsS -X "${method}" "${url}" \
      "${headers[@]}" \
      -H "Content-Type: application/json" \
      --data "${body}"
  else
    curl -fsS -X "${method}" "${url}" "${headers[@]}"
  fi
}

json_get() {
  local raw="$1"
  local expr="$2"
  RAW_JSON="${raw}" JSON_EXPR="${expr}" python3 - <<'PY'
import json
import os

data = json.loads(os.environ["RAW_JSON"])
value = data
for part in os.environ["JSON_EXPR"].split("."):
    if isinstance(value, dict):
        value = value.get(part)
    else:
        value = None
        break
if value is None:
    print("")
elif isinstance(value, (dict, list)):
    print(json.dumps(value, separators=(",", ":"), sort_keys=True))
else:
    print(value)
PY
}

ensure_nexus_action_type() {
  local raw
  raw="$(nexus_request GET "/v1/action-types?org_id=${PONTI_ORG_ID}")"
  local id
  id="$(RAW_JSON="${raw}" PONTI_ORG_ID="${PONTI_ORG_ID}" NEXUS_ACTION_TYPE="${NEXUS_ACTION_TYPE}" python3 - <<'PY'
import json
import os

data = json.loads(os.environ["RAW_JSON"])
items = data.get("data") or []
target = os.environ["NEXUS_ACTION_TYPE"]
org = os.environ["PONTI_ORG_ID"]
for item in items:
    if item.get("name") == target and item.get("org_id") == org:
        print(item.get("id", ""))
        raise SystemExit
print("")
PY
)"
  local body
  body="$(PONTI_ORG_ID="${PONTI_ORG_ID}" NEXUS_ACTION_TYPE="${NEXUS_ACTION_TYPE}" python3 - <<'PY'
import json
import os

print(json.dumps({
    "org_id": os.environ["PONTI_ORG_ID"],
    "name": os.environ["NEXUS_ACTION_TYPE"],
    "description": "Ponti capability invocation governed by Nexus",
    "category": "ponti",
    "risk_class": "medium",
    "schema": {},
    "reversible": True,
    "requires_break_glass": False,
}))
PY
)"
  if [[ -n "${id}" ]]; then
    nexus_request PATCH "/v1/action-types/${id}" "${body}" >/dev/null
    return
  fi
  nexus_request POST "/v1/action-types" "${body}" >/dev/null
}

ensure_nexus_policy() {
  local raw
  raw="$(nexus_request GET "/v1/policies")"
  local id
  id="$(RAW_JSON="${raw}" PONTI_ORG_ID="${PONTI_ORG_ID}" NEXUS_POLICY_NAME="${NEXUS_POLICY_NAME}" python3 - <<'PY'
import json
import os

data = json.loads(os.environ["RAW_JSON"])
items = data.get("data") or []
target = os.environ["NEXUS_POLICY_NAME"]
org = os.environ["PONTI_ORG_ID"]
for item in items:
    if item.get("name") == target and item.get("org_id") == org:
        print(item.get("id", ""))
        raise SystemExit
print("")
PY
)"
  local body
  body="$(NEXUS_POLICY_NAME="${NEXUS_POLICY_NAME}" NEXUS_ACTION_TYPE="${NEXUS_ACTION_TYPE}" python3 - <<'PY'
import json
import os

print(json.dumps({
    "name": os.environ["NEXUS_POLICY_NAME"],
    "description": "Local smoke policy: Ponti draft capabilities require explicit approval",
    "action_type": os.environ["NEXUS_ACTION_TYPE"],
    "target_system": "ponti",
    "expression": "request.action_type == 'agent.capability.invoke' && request.target_system == 'ponti'",
    "effect": "require_approval",
    "risk_override": "medium",
    "priority": 1,
    "mode": "enforced",
    "enabled": True,
}))
PY
)"
  if [[ -n "${id}" ]]; then
    nexus_request PATCH "/v1/policies/${id}" "${body}" >/dev/null
    return
  fi
  nexus_request POST "/v1/policies" "${body}" >/dev/null
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

draft_payload() {
  PONTI_PROJECT_ID="${PONTI_PROJECT_ID}" \
  PONTI_CUSTOMER_ID="${PONTI_CUSTOMER_ID}" \
  PONTI_FIELD_ID="${PONTI_FIELD_ID}" \
  PONTI_LOT_ID="${PONTI_LOT_ID}" \
  PONTI_CROP_ID="${PONTI_CROP_ID}" \
  PONTI_LABOR_ID="${PONTI_LABOR_ID}" \
  PONTI_INVESTOR_ID="${PONTI_INVESTOR_ID}" \
  python3 - <<'PY'
import json
import os

project_id = int(os.environ["PONTI_PROJECT_ID"])
print(json.dumps({
    "project_id": project_id,
    "customer_id": int(os.environ["PONTI_CUSTOMER_ID"]),
    "field_id": int(os.environ["PONTI_FIELD_ID"]),
    "lot_id": int(os.environ["PONTI_LOT_ID"]),
    "crop_id": int(os.environ["PONTI_CROP_ID"]),
    "labor_id": int(os.environ["PONTI_LABOR_ID"]),
    "investor_id": int(os.environ["PONTI_INVESTOR_ID"]),
    "scheduled_date": "2026-07-01",
    "date": "2026-07-01",
    "contractor": "Ponti Axis smoke",
    "effective_area": 1,
    "observations": "Nexus approved draft smoke.",
    "workspace": {
        "project_id": project_id,
    },
}, separators=(",", ":"), sort_keys=True))
PY
}

build_action_binding_body() {
  local connector_id="$1"
  local payload="$2"
  CONNECTOR_ID="${connector_id}" OPERATION="${OPERATION}" IDEMPOTENCY_KEY="${IDEMPOTENCY_KEY}" PAYLOAD_JSON="${payload}" python3 - <<'PY'
import json
import os

print(json.dumps({
    "connector_id": os.environ["CONNECTOR_ID"],
    "operation": os.environ["OPERATION"],
    "idempotency_key": os.environ["IDEMPOTENCY_KEY"],
    "payload": json.loads(os.environ["PAYLOAD_JSON"]),
}, separators=(",", ":"), sort_keys=True))
PY
}

validate_action_binding() {
  local raw="$1"
  local connector_id="$2"
  RAW_JSON="${raw}" CONNECTOR_ID="${connector_id}" OPERATION="${OPERATION}" IDEMPOTENCY_KEY="${IDEMPOTENCY_KEY}" PONTI_ORG_ID="${PONTI_ORG_ID}" python3 - <<'PY'
import json
import os
import sys

data = json.loads(os.environ["RAW_JSON"])
binding = data.get("action_binding")
if not isinstance(binding, dict):
    print(f"missing action_binding: {data}", file=sys.stderr)
    sys.exit(1)
if not data.get("binding_hash"):
    print(f"missing binding_hash: {data}", file=sys.stderr)
    sys.exit(1)
expected = {
    "org_id": os.environ["PONTI_ORG_ID"],
    "product_surface": "ponti",
    "connector_id": os.environ["CONNECTOR_ID"],
    "capability_id": os.environ["OPERATION"],
    "operation": os.environ["OPERATION"],
    "target_system": "ponti",
    "target_resource": os.environ["CONNECTOR_ID"],
    "idempotency_key": os.environ["IDEMPOTENCY_KEY"],
    "nexus_action_type": "agent.capability.invoke",
    "side_effect_type": "write",
}
for key, want in expected.items():
    if str(binding.get(key, "")) != want:
        print(f"binding {key} mismatch: {binding.get(key)!r} != {want!r}; binding={binding}", file=sys.stderr)
        sys.exit(1)
for key in ("schema_version", "actor_id", "actor_type", "run_id", "tool_invocation_id", "payload_hash"):
    if not str(binding.get(key, "")).strip():
        print(f"binding missing {key}: {binding}", file=sys.stderr)
        sys.exit(1)
PY
}

submit_nexus_request_body() {
  local binding_response="$1"
  local connector_id="$2"
  RAW_JSON="${binding_response}" CONNECTOR_ID="${connector_id}" OPERATION="${OPERATION}" IDEMPOTENCY_KEY="${IDEMPOTENCY_KEY}" PONTI_ORG_ID="${PONTI_ORG_ID}" NEXUS_ACTION_TYPE="${NEXUS_ACTION_TYPE}" NEXUS_REQUESTER_ID="${NEXUS_REQUESTER_ID}" python3 - <<'PY'
import json
import os

binding_response = json.loads(os.environ["RAW_JSON"])
binding = binding_response["action_binding"]
print(json.dumps({
    "idempotency_key": os.environ["IDEMPOTENCY_KEY"],
    "requester_type": "agent",
    "requester_id": os.environ["NEXUS_REQUESTER_ID"],
    "requester_name": "Ponti Axis smoke",
    "action_type": os.environ["NEXUS_ACTION_TYPE"],
    "target_system": "ponti",
    "target_resource": os.environ["CONNECTOR_ID"],
    "action_binding": binding,
    "params": {
        "org_id": os.environ["PONTI_ORG_ID"],
        "product_surface": "ponti",
        "capability_id": os.environ["OPERATION"],
        "operation": os.environ["OPERATION"],
    },
    "reason": "Validate Ponti draft action execution after explicit Nexus approval",
    "context": "local smoke",
}, separators=(",", ":"), sort_keys=True))
PY
}

validate_nexus_pending() {
  local raw="$1"
  local binding_hash="$2"
  RAW_JSON="${raw}" BINDING_HASH="${binding_hash}" python3 - <<'PY'
import json
import os
import sys

data = json.loads(os.environ["RAW_JSON"])
if data.get("decision") != "require_approval" or data.get("status") != "pending_approval":
    print(f"expected pending approval, got {data}", file=sys.stderr)
    sys.exit(1)
if data.get("binding_hash") != os.environ["BINDING_HASH"]:
    print(f"Nexus binding hash mismatch: {data}", file=sys.stderr)
    sys.exit(1)
approval = data.get("approval")
if not isinstance(approval, dict) or not approval.get("id"):
    print(f"expected approval payload: {data}", file=sys.stderr)
    sys.exit(1)
if not data.get("request_id"):
    print(f"expected request_id: {data}", file=sys.stderr)
    sys.exit(1)
PY
}

execute_connector_body() {
  local connector_id="$1"
  local payload="$2"
  local request_id="$3"
  CONNECTOR_ID="${connector_id}" OPERATION="${OPERATION}" IDEMPOTENCY_KEY="${IDEMPOTENCY_KEY}" PAYLOAD_JSON="${payload}" REQUEST_ID="${request_id}" python3 - <<'PY'
import json
import os

print(json.dumps({
    "connector_id": os.environ["CONNECTOR_ID"],
    "operation": os.environ["OPERATION"],
    "idempotency_key": os.environ["IDEMPOTENCY_KEY"],
    "nexus_request_id": os.environ["REQUEST_ID"],
    "payload": json.loads(os.environ["PAYLOAD_JSON"]),
}, separators=(",", ":"), sort_keys=True))
PY
}

validate_axis_execution() {
  local raw="$1"
  local request_id="$2"
  RAW_JSON="${raw}" REQUEST_ID="${request_id}" OPERATION="${OPERATION}" PONTI_ORG_ID="${PONTI_ORG_ID}" python3 - <<'PY'
import json
import os
import sys

data = json.loads(os.environ["RAW_JSON"])
if data.get("status") != "success":
    print(f"execution must succeed: {data}", file=sys.stderr)
    sys.exit(1)
if data.get("operation") != os.environ["OPERATION"]:
    print(f"operation mismatch: {data}", file=sys.stderr)
    sys.exit(1)
if data.get("org_id") != os.environ["PONTI_ORG_ID"]:
    print(f"org_id mismatch: {data}", file=sys.stderr)
    sys.exit(1)
result = data.get("result") or {}
if result.get("status") != "draft" or result.get("action") != os.environ["OPERATION"]:
    print(f"Ponti result must be draft for operation: {result}", file=sys.stderr)
    sys.exit(1)
checks = {
    "write_performed": True,
}
for key, want in checks.items():
    if result.get(key) is not want:
        print(f"result {key} mismatch: {result}", file=sys.stderr)
        sys.exit(1)
if result.get("execution_status") != "draft_created":
    print(f"expected execution_status=draft_created: {result}", file=sys.stderr)
    sys.exit(1)
if not result.get("draft_id"):
    print(f"expected draft_id in result: {result}", file=sys.stderr)
    sys.exit(1)
if result.get("nexus_request_id") != os.environ["REQUEST_ID"]:
    print(f"result nexus_request_id mismatch: {result}", file=sys.stderr)
    sys.exit(1)
evidence = data.get("evidence") or {}
if evidence.get("nexus_request_id") != os.environ["REQUEST_ID"]:
    print(f"execution evidence missing nexus_request_id: {evidence}", file=sys.stderr)
    sys.exit(1)
action_binding = evidence.get("action_binding") or {}
if action_binding.get("operation") != os.environ["OPERATION"] or action_binding.get("product_surface") != "ponti":
    print(f"execution evidence has invalid action_binding: {evidence}", file=sys.stderr)
    sys.exit(1)
PY
}

validate_nexus_request_status() {
  local raw="$1"
  local expected="$2"
  RAW_JSON="${raw}" EXPECTED_STATUS="${expected}" python3 - <<'PY'
import json
import os
import sys

data = json.loads(os.environ["RAW_JSON"])
if data.get("status") != os.environ["EXPECTED_STATUS"]:
    print(f"expected Nexus status {os.environ['EXPECTED_STATUS']}, got {data}", file=sys.stderr)
    sys.exit(1)
PY
}

report_result_body() {
  local execution_raw="$1"
  RAW_JSON="${execution_raw}" python3 - <<'PY'
import json
import os

execution = json.loads(os.environ["RAW_JSON"])
print(json.dumps({
    "result_id": execution.get("id", ""),
    "success": execution.get("status") == "success",
    "duration_ms": int(execution.get("duration_ms") or 0),
    "result": {
        "connector_execution_id": execution.get("id", ""),
        "operation": execution.get("operation", ""),
        "draft_id": (execution.get("result") or {}).get("draft_id"),
        "execution_status": (execution.get("result") or {}).get("execution_status"),
        "write_performed": (execution.get("result") or {}).get("write_performed") is True,
    },
}, separators=(",", ":"), sort_keys=True))
PY
}

require_env AXIS_COMPANION_BASE_URL
require_env AXIS_COMPANION_API_KEY
require_env NEXUS_BASE_URL
require_env NEXUS_ADMIN_API_KEY
require_env PONTI_BASE_URL
require_env PONTI_ORG_ID
require_env PONTI_AXIS_API_KEY
setup_headers

echo "Running Ponti onboarding against Axis..."
"$(dirname "$0")/onboard-ponti.sh"

echo "Refreshing Axis connectors..."
axis_request POST "/v1/connectors/refresh" '{}' >/dev/null

echo "Ensuring Nexus action type and approval policy..."
ensure_nexus_action_type
ensure_nexus_policy

echo "Ensuring persisted Ponti connector config in Axis..."
CONNECTOR_ID="$(ensure_ponti_connector_id)"
if [[ -z "${CONNECTOR_ID}" ]]; then
  echo "Ponti connector id is empty" >&2
  exit 1
fi

PAYLOAD="$(draft_payload)"
BINDING_BODY="$(build_action_binding_body "${CONNECTOR_ID}" "${PAYLOAD}")"

echo "Building Companion action binding for ${OPERATION}..."
BINDING_RAW="$(axis_request POST "/v1/connectors/action-binding" "${BINDING_BODY}")"
validate_action_binding "${BINDING_RAW}" "${CONNECTOR_ID}"
BINDING_HASH="$(json_get "${BINDING_RAW}" "binding_hash")"

echo "Submitting governed request to Nexus..."
SUBMIT_BODY="$(submit_nexus_request_body "${BINDING_RAW}" "${CONNECTOR_ID}")"
SUBMIT_RAW="$(nexus_request POST "/v1/requests" "${SUBMIT_BODY}" "${IDEMPOTENCY_KEY}")"
validate_nexus_pending "${SUBMIT_RAW}" "${BINDING_HASH}"
REQUEST_ID="$(json_get "${SUBMIT_RAW}" "request_id")"
APPROVAL_ID="$(json_get "${SUBMIT_RAW}" "approval.id")"

echo "Approving Nexus request ${REQUEST_ID}..."
nexus_request POST "/v1/approvals/${APPROVAL_ID}/approve" '{"note":"approved by Ponti Axis smoke"}' "" "${NEXUS_APPROVER_ID}" >/dev/null
APPROVED_RAW="$(nexus_request GET "/v1/requests/${REQUEST_ID}")"
validate_nexus_request_status "${APPROVED_RAW}" "approved"

echo "Executing approved draft capability through Axis connector..."
EXEC_BODY="$(execute_connector_body "${CONNECTOR_ID}" "${PAYLOAD}" "${REQUEST_ID}")"
EXEC_RAW="$(axis_request POST "/v1/connectors/execute" "${EXEC_BODY}")"
validate_axis_execution "${EXEC_RAW}" "${REQUEST_ID}"

echo "Reporting execution result back to Nexus..."
RESULT_BODY="$(report_result_body "${EXEC_RAW}")"
nexus_request POST "/v1/requests/${REQUEST_ID}/result" "${RESULT_BODY}" "${IDEMPOTENCY_KEY}-result" >/dev/null
EXECUTED_RAW="$(nexus_request GET "/v1/requests/${REQUEST_ID}")"
validate_nexus_request_status "${EXECUTED_RAW}" "executed"

echo "Ponti Axis Nexus-approved draft smoke complete."
