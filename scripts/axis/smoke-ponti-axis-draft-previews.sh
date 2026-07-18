#!/usr/bin/env bash
set -euo pipefail

PONTI_BASE_URL="${PONTI_BASE_URL:-}"
PONTI_ORG_ID="${PONTI_ORG_ID:-}"
PONTI_AXIS_API_KEY="${PONTI_AXIS_API_KEY:-}"
PONTI_API_KEY="${PONTI_API_KEY:-${X_API_KEY:-}}"
PONTI_AUTH_MODE="${PONTI_AUTH_MODE:-}"
PONTI_USER_ID="${PONTI_USER_ID:-ponti-draft-preview-smoke}"
PONTI_PROJECT_ID="${PONTI_PROJECT_ID:-1}"
PONTI_FIELD_ID="${PONTI_FIELD_ID:-1}"
PONTI_CAMPAIGN_ID="${PONTI_CAMPAIGN_ID:-1}"
PONTI_SUPPLY_ID="${PONTI_SUPPLY_ID:-1}"

require_env() {
  local name="$1"
  local value="${!name:-}"
  if [[ -z "${value}" ]]; then
    echo "missing required env: ${name}" >&2
    exit 2
  fi
}

setup_auth_mode() {
  if [[ -z "${PONTI_AUTH_MODE}" ]]; then
    if [[ -n "${PONTI_AXIS_API_KEY}" ]]; then
      PONTI_AUTH_MODE="axis"
    else
      PONTI_AUTH_MODE="local"
    fi
  fi
  case "${PONTI_AUTH_MODE}" in
    axis)
      require_env PONTI_AXIS_API_KEY
      ;;
    local)
      require_env PONTI_API_KEY
      ;;
    *)
      echo "PONTI_AUTH_MODE must be axis or local" >&2
      exit 2
      ;;
  esac
}

ponti_headers() {
  local headers=(
    -H "Accept: application/json"
    -H "Content-Type: application/json"
    -H "X-Tenant-Id: ${PONTI_ORG_ID}"
  )
  if [[ "${PONTI_AUTH_MODE}" == "axis" ]]; then
    headers+=(-H "Authorization: Bearer ${PONTI_AXIS_API_KEY}")
  else
    headers+=(-H "X-API-KEY: ${PONTI_API_KEY}")
    headers+=(-H "X-USER-ID: ${PONTI_USER_ID}")
  fi
  printf '%s\n' "${headers[@]}"
}

ponti_post() {
  local path="$1"
  local body="$2"
  local headers=()
  mapfile -t headers < <(ponti_headers)
  curl -fsS -X POST "${PONTI_BASE_URL%/}${path}" \
    "${headers[@]}" \
    --data "${body}"
}

validate_preview() {
  local raw="$1"
  local expected_action="$2"
  local expected_project_id="${3:-}"
  RAW_JSON="${raw}" EXPECTED_ACTION="${expected_action}" EXPECTED_PROJECT_ID="${expected_project_id}" PONTI_ORG_ID="${PONTI_ORG_ID}" python3 - <<'PY'
import json
import os
import sys

data = json.loads(os.environ["RAW_JSON"])
expected_action = os.environ["EXPECTED_ACTION"]
if data.get("status") != "preview":
    print(f"status must be preview: {data}", file=sys.stderr)
    sys.exit(1)
if data.get("action") != expected_action:
    print(f"action mismatch: {data.get('action')} != {expected_action}", file=sys.stderr)
    sys.exit(1)
if data.get("approval_required") is not True:
    print(f"approval_required must be true: {data}", file=sys.stderr)
    sys.exit(1)
if data.get("nexus_action_type") != "agent.capability.invoke":
    print(f"nexus_action_type mismatch: {data}", file=sys.stderr)
    sys.exit(1)
if data.get("side_effect_type") != "write":
    print(f"side_effect_type must be write: {data}", file=sys.stderr)
    sys.exit(1)
if data.get("preview_only") is not True:
    print(f"preview_only must be true: {data}", file=sys.stderr)
    sys.exit(1)
if data.get("write_performed") is not False:
    print(f"write_performed must be false: {data}", file=sys.stderr)
    sys.exit(1)
if data.get("execution_allowed") is not False:
    print(f"execution_allowed must be false: {data}", file=sys.stderr)
    sys.exit(1)
if data.get("execution_blocked_by") != "nexus_required":
    print(f"execution_blocked_by must be nexus_required: {data}", file=sys.stderr)
    sys.exit(1)
proposal = data.get("proposal")
if not isinstance(proposal, dict):
    print(f"proposal must be an object: {data}", file=sys.stderr)
    sys.exit(1)
if proposal.get("preview_only") is not True or proposal.get("write_performed") is not False:
    print(f"proposal must be preview-only: {proposal}", file=sys.stderr)
    sys.exit(1)
evidence = data.get("evidence")
if not isinstance(evidence, dict):
    print(f"evidence must be an object: {data}", file=sys.stderr)
    sys.exit(1)
if evidence.get("tenant_scope") != os.environ["PONTI_ORG_ID"]:
    print(f"tenant evidence mismatch: {evidence}", file=sys.stderr)
    sys.exit(1)
if evidence.get("approval_required") is not True:
    print(f"approval evidence missing: {evidence}", file=sys.stderr)
    sys.exit(1)
workspace = evidence.get("workspace")
if not isinstance(workspace, dict):
    print(f"workspace evidence must be an object: {evidence}", file=sys.stderr)
    sys.exit(1)
expected_project_id = os.environ.get("EXPECTED_PROJECT_ID", "")
if expected_project_id:
    got = workspace.get("project_id")
    if str(int(got)) != expected_project_id:
        print(f"workspace project_id mismatch: {workspace}", file=sys.stderr)
        sys.exit(1)
PY
}

validate_rejected_zero_delta() {
  local status
  local headers=()
  mapfile -t headers < <(ponti_headers)
  status="$(curl -sS -o /tmp/ponti-stock-adjustment-zero-delta.json -w "%{http_code}" \
    -X POST "${PONTI_BASE_URL%/}/api/v1/ai/actions/stock-adjustment/prepare" \
    "${headers[@]}" \
    --data '{
      "project_id": '"${PONTI_PROJECT_ID}"',
      "supply_id": '"${PONTI_SUPPLY_ID}"',
      "quantity_delta": 0,
      "reason": "smoke invalid delta"
    }')"
  if [[ "${status}" != "400" ]]; then
    echo "expected zero-delta validation to return 400, got ${status}" >&2
    cat /tmp/ponti-stock-adjustment-zero-delta.json >&2 || true
    exit 1
  fi
}

require_env PONTI_BASE_URL
require_env PONTI_ORG_ID
setup_auth_mode

echo "Using Ponti auth mode: ${PONTI_AUTH_MODE}"

echo "Calling ponti.insight.resolve.prepare preview..."
insight_raw="$(ponti_post "/api/v1/ai/actions/insight-resolve/prepare" '{
  "insight_id": "00000000-0000-0000-0000-000000000001",
  "resolution_note": "Smoke preview only.",
  "workspace": {
    "project_id": '"${PONTI_PROJECT_ID}"'
  }
}')"
validate_preview "${insight_raw}" "ponti.insight.resolve.prepare" "${PONTI_PROJECT_ID}"

echo "Calling ponti.workorder.draft.prepare preview..."
workorder_raw="$(ponti_post "/api/v1/ai/actions/workorder-draft/prepare" '{
  "project_id": '"${PONTI_PROJECT_ID}"',
  "field_id": '"${PONTI_FIELD_ID}"',
  "campaign_id": '"${PONTI_CAMPAIGN_ID}"',
  "work_type": "smoke",
  "scheduled_date": "2026-07-01",
  "notes": "Smoke preview only."
}')"
validate_preview "${workorder_raw}" "ponti.workorder.draft.prepare" "${PONTI_PROJECT_ID}"

echo "Calling ponti.stock_adjustment.prepare preview..."
stock_raw="$(ponti_post "/api/v1/ai/actions/stock-adjustment/prepare" '{
  "project_id": '"${PONTI_PROJECT_ID}"',
  "supply_id": '"${PONTI_SUPPLY_ID}"',
  "quantity_delta": -1,
  "reason": "Smoke preview only."
}')"
validate_preview "${stock_raw}" "ponti.stock_adjustment.prepare" "${PONTI_PROJECT_ID}"

echo "Validating stock adjustment zero-delta rejection..."
validate_rejected_zero_delta

echo "Ponti Axis draft preview smoke complete."
