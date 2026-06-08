#!/usr/bin/env bash
set -euo pipefail

AXIS_COMPANION_BASE_URL="${AXIS_COMPANION_BASE_URL:-}"
AXIS_COMPANION_API_KEY="${AXIS_COMPANION_API_KEY:-}"
PONTI_BASE_URL="${PONTI_BASE_URL:-}"
PONTI_ORG_ID="${PONTI_ORG_ID:-}"
PONTI_AXIS_API_KEY="${PONTI_AXIS_API_KEY:-}"
PONTI_API_KEY_SECRET_REF="${PONTI_API_KEY_SECRET_REF:-env:PONTI_API_KEY}"
AXIS_ACTOR_ID="${AXIS_ACTOR_ID:-ponti-onboarding}"
AXIS_ADMIN_SCOPES="${AXIS_ADMIN_SCOPES:-companion:products:admin companion:runtime:admin companion:capabilities:admin companion:connectors:execute}"

require_env() {
  local name="$1"
  local value="${!name:-}"
  if [[ -z "${value}" ]]; then
    echo "missing required env: ${name}" >&2
    exit 2
  fi
}

axis_request() {
  local method="$1"
  local path="$2"
  local body="${3:-}"
  local url="${AXIS_COMPANION_BASE_URL%/}${path}"
  if [[ -n "${body}" ]]; then
    curl -fsS -X "${method}" "${url}" \
      -H "Content-Type: application/json" \
      -H "Accept: application/json" \
      -H "X-API-Key: ${AXIS_COMPANION_API_KEY}" \
      -H "X-Org-ID: ${PONTI_ORG_ID}" \
      -H "X-User-ID: ${AXIS_ACTOR_ID}" \
      -H "X-On-Behalf-Of: ${AXIS_ACTOR_ID}" \
      -H "X-Product-Surface: ponti" \
      -H "X-Auth-Scopes: ${AXIS_ADMIN_SCOPES}" \
      --data "${body}"
  else
    curl -fsS -X "${method}" "${url}" \
      -H "Accept: application/json" \
      -H "X-API-Key: ${AXIS_COMPANION_API_KEY}" \
      -H "X-Org-ID: ${PONTI_ORG_ID}" \
      -H "X-User-ID: ${AXIS_ACTOR_ID}" \
      -H "X-On-Behalf-Of: ${AXIS_ACTOR_ID}" \
      -H "X-Product-Surface: ponti" \
      -H "X-Auth-Scopes: ${AXIS_ADMIN_SCOPES}"
  fi
}

validate_ponti_capabilities() {
  local url="${PONTI_BASE_URL%/}/api/v1/capabilities"
  local raw
  local headers=(-H "Accept: application/json" -H "X-Tenant-Id: ${PONTI_ORG_ID}")
  if [[ -n "${PONTI_AXIS_API_KEY}" ]]; then
    headers+=(-H "Authorization: Bearer ${PONTI_AXIS_API_KEY}")
  fi
  raw="$(curl -fsS "${url}" "${headers[@]}")"
  RAW_JSON="${raw}" python3 - <<'PY'
import json
import os
import sys

data = json.loads(os.environ["RAW_JSON"])
items = data.get("items")
if not isinstance(items, list) or not items:
    print("Ponti capabilities response must contain non-empty items[]", file=sys.stderr)
    sys.exit(1)
if not any(item.get("id") == "ponti.insights" for item in items if isinstance(item, dict)):
    print("Ponti capabilities response must include ponti.insights", file=sys.stderr)
    sys.exit(1)
PY
}

require_env AXIS_COMPANION_BASE_URL
require_env AXIS_COMPANION_API_KEY
require_env PONTI_BASE_URL
require_env PONTI_ORG_ID

echo "Validating Ponti capabilities..."
validate_ponti_capabilities

echo "Registering Axis product ponti..."
axis_request PUT "/v1/products/ponti" '{
  "display_name": "Ponti",
  "status": "active",
  "metadata": {
    "owner": "ponti",
    "capabilities_url": "'"${PONTI_BASE_URL%/}"'/api/v1/capabilities"
  }
}' >/dev/null

echo "Creating Axis installation for ${PONTI_ORG_ID} + ponti..."
axis_request PUT "/v1/product-installations/ponti?org_id=${PONTI_ORG_ID}" '{
  "external_tenant_id": "'"${PONTI_ORG_ID}"'",
  "base_url": "'"${PONTI_BASE_URL%/}"'",
  "auth_mode": "api_key_ref",
  "secret_ref": "'"${PONTI_API_KEY_SECRET_REF}"'",
  "enabled": true,
  "config": {
    "capabilities_path": "/api/v1/capabilities"
  }
}' >/dev/null

echo "Resolving active installation..."
axis_request GET "/v1/product-installations/ponti/resolve?org_id=${PONTI_ORG_ID}" >/dev/null

echo "Refreshing Axis connectors..."
axis_request POST "/v1/connectors/refresh" '{}' >/dev/null

echo "Ponti Axis onboarding complete."
