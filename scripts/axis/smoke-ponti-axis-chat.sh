#!/usr/bin/env bash
set -euo pipefail

PONTI_BASE_URL="${PONTI_BASE_URL:-}"
PONTI_API_KEY="${PONTI_API_KEY:-}"
PONTI_ORG_ID="${PONTI_ORG_ID:-}"
PONTI_PROJECT_ID="${PONTI_PROJECT_ID:-}"
PONTI_USER_ID="${PONTI_USER_ID:-ponti-axis-chat-smoke}"

require_env() {
  local name="$1"
  local value="${!name:-}"
  if [[ -z "${value}" ]]; then
    echo "missing required env: ${name}" >&2
    exit 2
  fi
}

post_chat() {
  curl -fsS -X POST "${PONTI_BASE_URL%/}/api/v1/ai/chat" \
    -H "Accept: application/json" \
    -H "Content-Type: application/json" \
    -H "X-API-KEY: ${PONTI_API_KEY}" \
    -H "X-Tenant-Id: ${PONTI_ORG_ID}" \
    -H "X-USER-ID: ${PONTI_USER_ID}" \
    -H "X-PROJECT-ID: ${PONTI_PROJECT_ID}" \
    --data '{
      "message": "Resumime los insights operativos activos de Ponti.",
      "route_hint": "dashboard",
      "preferred_language": "es",
      "workspace": {
        "project_id": '"${PONTI_PROJECT_ID}"'
      }
    }'
}

post_stream() {
  curl -fsS -N -X POST "${PONTI_BASE_URL%/}/api/v1/ai/chat/stream" \
    -H "Accept: text/event-stream" \
    -H "Content-Type: application/json" \
    -H "X-API-KEY: ${PONTI_API_KEY}" \
    -H "X-Tenant-Id: ${PONTI_ORG_ID}" \
    -H "X-USER-ID: ${PONTI_USER_ID}" \
    -H "X-PROJECT-ID: ${PONTI_PROJECT_ID}" \
    --data '{
      "message": "Dame un resumen corto de los insights.",
      "preferred_language": "es",
      "workspace": {
        "project_id": '"${PONTI_PROJECT_ID}"'
      }
    }'
}

validate_chat_response() {
  local raw="$1"
  RAW_JSON="${raw}" python3 - <<'PY'
import json
import os
import sys

data = json.loads(os.environ["RAW_JSON"])
required = ["request_id", "output_kind", "chat_id", "reply", "blocks", "tool_calls", "pending_confirmations", "routed_agent", "routing_source"]
missing = [k for k in required if k not in data]
if missing:
    print(f"chat response missing fields: {missing}\n{data}", file=sys.stderr)
    sys.exit(1)
if data["output_kind"] != "chat_reply":
    print(f"output_kind must be chat_reply: {data}", file=sys.stderr)
    sys.exit(1)
if data["routing_source"] != "axis":
    print(f"routing_source must be axis: {data}", file=sys.stderr)
    sys.exit(1)
if not isinstance(data["reply"], str) or not data["reply"].strip():
    print(f"reply must be non-empty: {data}", file=sys.stderr)
    sys.exit(1)
if not isinstance(data["blocks"], list):
    print(f"blocks must be a list: {data}", file=sys.stderr)
    sys.exit(1)
PY
}

validate_stream_response() {
  local raw="$1"
  RAW_SSE="${raw}" python3 - <<'PY'
import os
import sys

sse = os.environ["RAW_SSE"]
required = ["event: start", "event: text", "event: done", '"routing_source":"axis"']
missing = [part for part in required if part not in sse]
if missing:
    print(f"SSE response missing {missing}:\n{sse}", file=sys.stderr)
    sys.exit(1)
PY
}

require_env PONTI_BASE_URL
require_env PONTI_API_KEY
require_env PONTI_ORG_ID
require_env PONTI_PROJECT_ID

echo "Calling Ponti chat through Axis provider..."
chat_raw="$(post_chat)"
validate_chat_response "${chat_raw}"

echo "Calling Ponti chat stream through Axis provider..."
stream_raw="$(post_stream)"
validate_stream_response "${stream_raw}"

echo "Ponti Axis chat smoke complete."
