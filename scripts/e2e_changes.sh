#!/usr/bin/env bash
# E2E tests para los cambios: AI dummies, data-integrity cache, timeout
# Uso: ./scripts/e2e_changes.sh [BASE_URL]
# Default: http://localhost:8080

set -euo pipefail

BASE="${1:-http://localhost:8080}"
API_KEY="${X_API_KEY:-abc123secreta}"
USER_ID="${X_USER_ID:-1}"

headers=(-H "X-API-KEY: $API_KEY" -H "X-User-Id: $USER_ID" -H "X-Project-ID: 20")

echo "=== E2E: Cambios (AI dummies, data-integrity) ==="
echo "Base: $BASE"
echo ""

failed=0

# 1. Ping
echo "[1/6] GET /ping"
if ! curl -sf "$BASE/ping" >/dev/null; then
  echo "  FAIL: /ping no responde"
  ((failed++))
else
  echo "  OK"
fi

# 2. project_id
project_id="${PROJECT_ID:-20}"
echo "[2/6] project_id=$project_id"

# 3. AI insights/summary (dummy)
echo "[3/6] GET /api/v1/ai/insights/summary (dummy)"
resp=$(curl -s -w "\n%{http_code}" "$BASE/api/v1/ai/insights/summary" "${headers[@]}" 2>/dev/null || true)
code=$(echo "$resp" | tail -1)
body=$(echo "$resp" | sed '$d')
if [[ "$code" != "200" ]]; then
  echo "  FAIL: expected 200, got $code"
  ((failed++))
elif ! echo "$body" | grep -q "new_count_total"; then
  echo "  FAIL: response no tiene new_count_total (no es dummy)"
  ((failed++))
else
  echo "  OK (200, dummy)"
fi

# 4. AI insights by entity (dummy)
echo "[4/6] GET /api/v1/ai/insights/lot/1 (dummy)"
resp=$(curl -s -w "\n%{http_code}" "$BASE/api/v1/ai/insights/lot/1" "${headers[@]}" 2>/dev/null || true)
code=$(echo "$resp" | tail -1)
if [[ "$code" != "200" ]]; then
  echo "  FAIL: expected 200, got $code"
  ((failed++))
else
  echo "  OK (200, dummy)"
fi

# 5. Data-integrity costs-check (optimizado con cache)
echo "[5/6] GET /api/v1/data-integrity/costs-check?project_id=$project_id"
start=$SECONDS
resp=$(curl -s -w "\n%{http_code}" "$BASE/api/v1/data-integrity/costs-check?project_id=$project_id" "${headers[@]}" 2>/dev/null || true)
elapsed=$((SECONDS - start))
code=$(echo "$resp" | tail -1)
body=$(echo "$resp" | sed '$d')
if [[ "$code" != "200" ]]; then
  echo "  FAIL: expected 200, got $code"
  ((failed++))
elif ! echo "$body" | grep -q "checks"; then
  echo "  FAIL: response no tiene checks"
  ((failed++))
elif [[ $elapsed -gt 60 ]]; then
  echo "  WARN: tardó ${elapsed}s (esperado <60s con cache)"
  echo "  OK (200)"
else
  echo "  OK (200, ${elapsed}s)"
fi

# 6. Lots (smoke)
echo "[6/6] GET /api/v1/lots?project_id=$project_id"
resp=$(curl -s -w "\n%{http_code}" "$BASE/api/v1/lots?project_id=$project_id&page=1&page_size=5" "${headers[@]}" 2>/dev/null || true)
code=$(echo "$resp" | tail -1)
if [[ "$code" != "200" ]]; then
  echo "  FAIL: expected 200, got $code"
  ((failed++))
else
  echo "  OK (200)"
fi

echo ""
if [[ $failed -gt 0 ]]; then
  echo "=== FAIL: $failed tests fallaron ==="
  exit 1
fi
echo "=== OK: todos los tests pasaron ==="
exit 0
