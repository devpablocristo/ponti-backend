#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${1:-${BASE_URL:-http://localhost:8080}}"
API_PREFIX="${API_PREFIX:-/api/v1}"
X_API_KEY="${X_API_KEY:-}"
AUTH_BEARER_TOKEN="${AUTH_BEARER_TOKEN:-}"

if [[ -z "${BASE_URL}" ]]; then
  echo "ERROR: BASE_URL vacío." >&2
  exit 1
fi

api_url() {
  local path="$1"
  printf "%s%s%s" "${BASE_URL%/}" "${API_PREFIX}" "$path"
}

request() {
  local method="$1"
  local url="$2"
  local body="${3:-}"
  local tmp
  tmp="$(mktemp)"

  local -a args
  args=(-sS -X "$method" "$url" -H "Content-Type: application/json")
  if [[ -n "${X_API_KEY}" ]]; then
    args+=(-H "X-API-Key: ${X_API_KEY}")
  fi
  if [[ -n "${AUTH_BEARER_TOKEN}" ]]; then
    args+=(-H "Authorization: Bearer ${AUTH_BEARER_TOKEN}")
  fi
  if [[ -n "${body}" ]]; then
    args+=(-d "${body}")
  fi

  local status
  status="$(curl "${args[@]}" -o "${tmp}" -w "%{http_code}")"
  local content
  content="$(cat "${tmp}")"
  rm -f "${tmp}"
  printf "%s\n%s" "${status}" "${content}"
}

expect_status() {
  local expected="$1"
  local got="$2"
  local msg="$3"
  if [[ "${expected}" != "${got}" ]]; then
    echo "ERROR: ${msg}. Esperado ${expected}, obtenido ${got}." >&2
    exit 1
  fi
}

json_extract() {
  local body="$1"
  local code="$2"
  python3 -c "import json,sys
body=sys.stdin.read()
obj=json.loads(body)
${code}
" <<<"${body}"
}

echo "[smoke] Ping service..."
ping_status="$(curl -sS -o /dev/null -w "%{http_code}" "${BASE_URL%/}/ping")"
expect_status "200" "${ping_status}" "Ping falló"

echo "[smoke] Resolve project..."
resp="$(request GET "$(api_url "/projects?page=1&per_page=1")")"
proj_status="$(printf "%s" "${resp}" | awk 'NR==1{print $1}')"
proj_body="$(printf "%s" "${resp}" | awk 'NR>1{print}')"
expect_status "200" "${proj_status}" "No se pudieron listar proyectos"

project_id="$(json_extract "${proj_body}" '
items=obj.get("items")
if isinstance(items,list) and items:
    print(items[0].get("id",""))
    raise SystemExit
data=obj.get("data")
if isinstance(data,list) and data:
    print(data[0].get("id",""))
    raise SystemExit
if isinstance(data,dict):
    rows=data.get("data")
    if isinstance(rows,list) and rows:
        print(rows[0].get("id",""))
        raise SystemExit
print("")
')"

if [[ -z "${project_id}" ]]; then
  echo "ERROR: No hay proyectos para correr smoke test." >&2
  exit 1
fi

echo "[smoke] Resolve project detail and investors..."
resp="$(request GET "$(api_url "/projects/${project_id}")")"
detail_status="$(printf "%s" "${resp}" | awk 'NR==1{print $1}')"
detail_body="$(printf "%s" "${resp}" | awk 'NR>1{print}')"
expect_status "200" "${detail_status}" "No se pudo obtener detalle de proyecto"

readarray -t ids < <(json_extract "${detail_body}" '
data=obj.get("data",obj)
investors=data.get("investors") or []
if len(investors)<2:
    print("")
    print("")
else:
    print(investors[0].get("id",""))
    print(investors[1].get("id",""))
')
investor_a="${ids[0]:-}"
investor_b="${ids[1]:-}"
if [[ -z "${investor_a}" || -z "${investor_b}" ]]; then
  echo "ERROR: El proyecto ${project_id} no tiene al menos 2 inversores." >&2
  exit 1
fi

readarray -t lot_data < <(json_extract "${detail_body}" '
data=obj.get("data",obj)
fields=data.get("fields") or []
for f in fields:
    lots=f.get("lots") or []
    for lot in lots:
        lot_id=lot.get("id")
        crop_id=lot.get("current_crop_id")
        if lot_id and crop_id:
            print(f.get("id",""))
            print(lot_id)
            print(crop_id)
            raise SystemExit
print("")
print("")
print("")
')
field_id="${lot_data[0]:-}"
lot_id="${lot_data[1]:-}"
crop_id="${lot_data[2]:-}"
if [[ -z "${field_id}" || -z "${lot_id}" || -z "${crop_id}" ]]; then
  echo "ERROR: El proyecto ${project_id} no tiene campo/lote/cultivo actual válidos." >&2
  exit 1
fi

echo "[smoke] Resolve labor..."
resp="$(request GET "$(api_url "/projects/${project_id}/labors")")"
labor_status="$(printf "%s" "${resp}" | awk 'NR==1{print $1}')"
labor_body="$(printf "%s" "${resp}" | awk 'NR>1{print}')"
expect_status "200" "${labor_status}" "No se pudieron listar labores"

readarray -t labor_data < <(json_extract "${labor_body}" '
data=obj.get("data",obj)
rows=data.get("data") if isinstance(data,dict) else data
if not isinstance(rows,list):
    rows=[]
if rows:
    first=rows[0]
    print(first.get("id",""))
    print(first.get("contractor_name",""))
else:
    print("")
    print("")
')
labor_id="${labor_data[0]:-}"
contractor="${labor_data[1]:-}"
if [[ -z "${labor_id}" ]]; then
  echo "ERROR: El proyecto ${project_id} no tiene labores para prueba." >&2
  exit 1
fi

order_number="SMOKE-$(date +%s)"
today="$(date +%F)"
payload="$(PROJECT_ID="${project_id}" FIELD_ID="${field_id}" LOT_ID="${lot_id}" CROP_ID="${crop_id}" LABOR_ID="${labor_id}" CONTRACTOR="${contractor}" INVESTOR_A="${investor_a}" INVESTOR_B="${investor_b}" ORDER_NUMBER="${order_number}" TODAY="${today}" python3 - <<'PY'
import json
import os
print(json.dumps({
  "number": os.environ["ORDER_NUMBER"],
  "project_id": int(os.environ["PROJECT_ID"]),
  "field_id": int(os.environ["FIELD_ID"]),
  "lot_id": int(os.environ["LOT_ID"]),
  "crop_id": int(os.environ["CROP_ID"]),
  "labor_id": int(os.environ["LABOR_ID"]),
  "contractor": os.environ.get("CONTRACTOR", ""),
  "observations": "Smoke deploy validation",
  "date": f'{os.environ["TODAY"]}T00:00:00Z',
  "investor_id": int(os.environ["INVESTOR_A"]),
  "effective_area": 1,
  "items": [],
  "investor_splits": [
    {"investor_id": int(os.environ["INVESTOR_A"]), "percentage": 60},
    {"investor_id": int(os.environ["INVESTOR_B"]), "percentage": 40}
  ]
}))
PY
)"

echo "[smoke] Create split work-order..."
resp="$(request POST "$(api_url "/work-orders")" "${payload}")"
create_status="$(printf "%s" "${resp}" | awk 'NR==1{print $1}')"
create_body="$(printf "%s" "${resp}" | awk 'NR>1{print}')"
if [[ "${create_status}" != "201" ]]; then
  echo "${create_body}" >&2
fi
expect_status "201" "${create_status}" "No se pudo crear OT con divisor"

workorder_id="$(json_extract "${create_body}" '
if isinstance(obj.get("id"), int):
    print(obj.get("id"))
elif isinstance(obj.get("number"), int):
    print(obj.get("number"))
elif isinstance(obj.get("data"), dict) and isinstance(obj["data"].get("id"), int):
    print(obj["data"].get("id"))
else:
    print("")
')"
if [[ -z "${workorder_id}" ]]; then
  echo "ERROR: No se pudo resolver ID de la OT creada." >&2
  exit 1
fi

echo "[smoke] Validate split persisted..."
resp="$(request GET "$(api_url "/work-orders/${workorder_id}")")"
wo_status="$(printf "%s" "${resp}" | awk 'NR==1{print $1}')"
wo_body="$(printf "%s" "${resp}" | awk 'NR>1{print}')"
expect_status "200" "${wo_status}" "No se pudo recuperar OT creada"

json_extract "${wo_body}" '
row=obj.get("data",obj)
splits=row.get("investor_splits") or []
if len(splits) < 2:
    raise SystemExit("OT sin investor_splits persistidos")
total=sum(float(s.get("percentage",0) or 0) for s in splits)
if abs(total-100) > 0.001:
    raise SystemExit(f"Porcentajes inválidos: {total}")
print("ok")
' >/dev/null

echo "[smoke] Validate investor-contribution report endpoint..."
resp="$(request GET "$(api_url "/reports/investor-contribution?project_id=${project_id}")")"
report_status="$(printf "%s" "${resp}" | awk 'NR==1{print $1}')"
expect_status "200" "${report_status}" "Reporte de aporte por inversor no disponible"

echo "[smoke] Cleanup work-order..."
resp="$(request DELETE "$(api_url "/work-orders/${workorder_id}")")"
delete_status="$(printf "%s" "${resp}" | awk 'NR==1{print $1}')"
if [[ "${delete_status}" != "204" && "${delete_status}" != "200" ]]; then
  echo "WARN: No se pudo borrar la OT de smoke (status ${delete_status})." >&2
fi

echo "[smoke] OK - división de aportes validada."
