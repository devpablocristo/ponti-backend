#!/usr/bin/env bash
set -euo pipefail

# === Cargar variables desde .env (y variantes) ===
load_env_file() {
  local f="$1"
  if [ -f "$f" ]; then
    # Exporta automáticamente todas las variables definidas en el archivo
    set -o allexport
    # shellcheck disable=SC1091
    . "$f"
    set +o allexport
    echo "→ Cargado: $f"
  fi
}

# Archivos a intentar (de más específico a más general)
# ej: .env.local.dev, .env.local, .env.dev, .env
DEPLOY_PLATFORM="${DEPLOY_PLATFORM:-local}"
DEPLOY_ENV="${DEPLOY_ENV:-dev}"

load_env_file ".env.${DEPLOY_PLATFORM}.${DEPLOY_ENV}"
load_env_file ".env.${DEPLOY_PLATFORM}"
load_env_file ".env.${DEPLOY_ENV}"
load_env_file ".env"


########################################
# CONFIG
########################################
API_BASE="${API_BASE:-http://localhost:8080/api/v1}"
API_KEY="${API_KEY:-abc123secreta}"
USER_ID="${USER_ID:-123}"

# Postgres (expuesto por docker-compose)
PGHOST="${PGHOST:-localhost}"
PGPORT="${PGPORT:-5432}"
PGDATABASE="${PGDATABASE:-ponti_api_db}"
PGUSER="${PGUSER:-postgres}"
PGPASSWORD="${PGPASSWORD:-postgres}"

export PGPASSWORD

# Nombres de demo (no cambies salvo que quieras duplicar fixtures)
DEMO_USER_ID=123
DEMO_USER_EMAIL="demo@ponti.com"
PROJECT_NAME="Construcción Torre Norte"
FIELD_A="Campo A"
FIELD_B="Campo B"
LOT_A1="Parcela A1"
LOT_A2="Parcela A2"
LOT_B1="Parcela B1"
LABOR_SOW_NAME="Siembra Maíz"
LABOR_HARVEST_NAME="Cosecha Maíz"
SUPPLY_NAME="Semilla Maíz DK 72-10"
INVESTOR_A="Fondo Capital Innovador"
INVESTOR_B="Grupo Inversor del Sur"
CUSTOMER_NAME="Inmobiliaria Buenos Aires S.A."
CAMPAIGN_NAME="Campaña Loteo 2025"

check_bin() {
  command -v "$1" >/dev/null 2>&1 || { echo "❌ Falta $1"; exit 1; }
}
check_bin curl
check_bin jq
check_bin psql

sql() {
  psql -h "$PGHOST" -p "$PGPORT" -U "$PGUSER" -d "$PGDATABASE" -v ON_ERROR_STOP=1 -qAtX -c "$1"
}

json_post() {
  local url="$1"
  shift
  curl -sS -H "X-API-KEY: $API_KEY" -H "X-USER-ID: $USER_ID" -H "Content-Type: application/json" -X POST "$url" -d "$@" 
}

json_put() {
  local url="$1"
  shift
  curl -sS -H "X-API-KEY: $API_KEY" -H "X-USER-ID: $USER_ID" -H "Content-Type: application/json" -X PUT "$url" -d "$@" 
}

echo "== 1) Usuario demo"
sql "
INSERT INTO users (id, email, username, password, token_hash, refresh_tokens, id_rol, is_verified, active, created_by, updated_by, created_at, updated_at)
VALUES ($DEMO_USER_ID, '$DEMO_USER_EMAIL', 'demo_user', 'demo_password', 'demo_token_hash', '{}'::text[], 1, TRUE, TRUE, 1, 1, now(), now())
ON CONFLICT (id) DO NOTHING;
"

echo "== 2) Crear proyecto + campos + lotes (vía API)"
PROJECT_PAYLOAD=$(jq -n --arg name "$PROJECT_NAME" \
  --arg customer "$CUSTOMER_NAME" \
  --arg campaign "$CAMPAIGN_NAME" \
  --arg invA "$INVESTOR_A" \
  --arg invB "$INVESTOR_B" \
  --arg fieldA "$FIELD_A" --arg fieldB "$FIELD_B" \
  --arg lotA1 "$LOT_A1" --arg lotA2 "$LOT_A2" --arg lotB1 "$LOT_B1" \
'{
  name: $name,
  admin_cost: 15000,
  customer: { id: 4, name: $customer, type: "tipo 1A" },
  campaign: { id: 1, name: $campaign },
  managers: [{id:1, name:"María López"}, {id:1, name:"Juan Pérez"}],
  investors: [{id:1, name:$invA, percentage:50}, {id:1, name:$invB, percentage:50}],
  fields: [
    {
      name: $fieldA,
      lease_type_id: 1,
      lease_type_value: 100,
      lots: [
        { name: $lotA1, hectares: 2.5, previous_crop_id: 1, current_crop_id: 2, season: "Invierno 2025" },
        { name: $lotA2, hectares: 3.0, previous_crop_id: 1, current_crop_id: 2, season: "Verano 2025" }
      ]
    },
    {
      name: $fieldB,
      lease_type_id: 2,
      lease_type_percent: 15,
      lots: [
        { name: $lotB1, hectares: 1.2, previous_crop_id: 1, current_crop_id: 2, season: "Otoño 2025" }
      ]
    }
  ]
}')

# Intentamos crear; si ya existe, seguimos
json_post "$API_BASE/projects" "$PROJECT_PAYLOAD" >/dev/null || true

# Resolvemos IDs por DB (idempotente y seguro)
PROJECT_ID=$(sql "SELECT id FROM projects WHERE name = '$PROJECT_NAME' AND deleted_at IS NULL LIMIT 1;")
[ -n "$PROJECT_ID" ] || { echo "❌ No se pudo resolver PROJECT_ID"; exit 1; }

FIELD_A_ID=$(sql "SELECT id FROM fields WHERE name = '$FIELD_A' AND project_id = $PROJECT_ID AND deleted_at IS NULL LIMIT 1;")
FIELD_B_ID=$(sql "SELECT id FROM fields WHERE name = '$FIELD_B' AND project_id = $PROJECT_ID AND deleted_at IS NULL LIMIT 1;")
LOT_A1_ID=$(sql "SELECT l.id FROM lots l JOIN fields f ON f.id=l.field_id WHERE l.name='$LOT_A1' AND f.project_id=$PROJECT_ID AND l.deleted_at IS NULL LIMIT 1;")
LOT_A2_ID=$(sql "SELECT l.id FROM lots l JOIN fields f ON f.id=l.field_id WHERE l.name='$LOT_A2' AND f.project_id=$PROJECT_ID AND l.deleted_at IS NULL LIMIT 1;")
LOT_B1_ID=$(sql "SELECT l.id FROM lots l JOIN fields f ON f.id=l.field_id WHERE l.name='$LOT_B1' AND f.project_id=$PROJECT_ID AND l.deleted_at IS NULL LIMIT 1;")

echo "   PROJECT_ID=$PROJECT_ID  FIELD_A_ID=$FIELD_A_ID  FIELD_B_ID=$FIELD_B_ID"
echo "   LOTS: A1=$LOT_A1_ID A2=$LOT_A2_ID B1=$LOT_B1_ID"

echo "== 3) Crear insumo (vía API)"
SUPPLY_PAYLOAD=$(jq -n --arg name "$SUPPLY_NAME" --argjson project_id "$PROJECT_ID" '
{ name:$name, project_id:$project_id, type_id:1, category_id:1, unit_id:1, price:1.65 }')
json_post "$API_BASE/supplies" "$SUPPLY_PAYLOAD" >/dev/null || true

SUPPLY_ID=$(sql "SELECT id FROM supplies WHERE name = '$SUPPLY_NAME' AND project_id=$PROJECT_ID AND deleted_at IS NULL LIMIT 1;")
echo "   SUPPLY_ID=$SUPPLY_ID"

echo "== 4) Crear labores (usando IDs REALES de categorías)"
CAT_SIEMBRA_ID=$(sql "SELECT id FROM categories WHERE name='Siembra' AND deleted_at IS NULL LIMIT 1;")
CAT_COSECHA_ID=$(sql "SELECT id FROM categories WHERE name='Cosecha' AND deleted_at IS NULL LIMIT 1;")
[ -n "$CAT_SIEMBRA_ID" ] || { echo "❌ No existe categoría 'Siembra'"; exit 1; }
[ -n "$CAT_COSECHA_ID" ] || { echo "❌ No existe categoría 'Cosecha'"; exit 1; }

LABORS_PAYLOAD=$(jq -n --arg sow "$LABOR_SOW_NAME" --arg harv "$LABOR_HARVEST_NAME" \
  --argjson catS "$CAT_SIEMBRA_ID" --argjson catH "$CAT_COSECHA_ID" \
'{
  labors: [
    { name: $sow, price: 25.50, category_id: $catS, contractor_name: "Juan Pérez", description: "Siembra directa de maíz para calcular sowed_area" },
    { name: $harv, price: 45.00, category_id: $catH, contractor_name: "María García", description: "Cosecha mecánica de maíz para calcular harvested_area" }
  ]
}')
json_post "$API_BASE/projects/$PROJECT_ID/labors" "$LABORS_PAYLOAD" >/dev/null || true

LABOR_SOW_ID=$(sql "SELECT id FROM labors WHERE name='$LABOR_SOW_NAME' AND project_id=$PROJECT_ID AND deleted_at IS NULL LIMIT 1;")
LABOR_HARVEST_ID=$(sql "SELECT id FROM labors WHERE name='$LABOR_HARVEST_NAME' AND project_id=$PROJECT_ID AND deleted_at IS NULL LIMIT 1;")
echo "   LABOR_SOW_ID=$LABOR_SOW_ID  LABOR_HARVEST_ID=$LABOR_HARVEST_ID"

echo "== 5) Crear workorders (fechas y áreas)"
CROP_MAIZ_ID=$(sql "SELECT id FROM crops WHERE name='Maíz' AND deleted_at IS NULL LIMIT 1;")
INVESTOR_ID=$(sql "SELECT id FROM investors WHERE name='$INVESTOR_A' AND deleted_at IS NULL LIMIT 1;")
[ -n "$CROP_MAIZ_ID" ] || { echo '❌ No existe crop "Maíz"'; exit 1; }
[ -n "$INVESTOR_ID" ] || { echo "❌ No se encontró inversor '$INVESTOR_A'"; exit 1; }

create_wo() {
  local number="$1"; local field_id="$2"; local lot_id="$3"
  local labor_id="$4"; local date="$5"; local area="$6"
  json_post "$API_BASE/workorders" "$(jq -n \
    --arg num "$number" \
    --argjson project_id "$PROJECT_ID" \
    --argjson field_id "$field_id" \
    --argjson lot_id "$lot_id" \
    --argjson crop_id "$CROP_MAIZ_ID" \
    --argjson labor_id "$labor_id" \
    --argjson investor_id "$INVESTOR_ID" \
    --arg date "$date" \
    --argjson eff "$area" \
    '{number:$num, project_id:$project_id, field_id:$field_id, lot_id:$lot_id, crop_id:$crop_id, labor_id:$labor_id, investor_id:$investor_id, date:$date, effective_area:$eff}')" >/dev/null || true
}

# A1: Siembra Jun / Cosecha Dic
create_wo "WO-SOWING-A1-001" "$FIELD_A_ID" "$LOT_A1_ID" "$LABOR_SOW_ID"     "2025-06-15T00:00:00Z" 2.5
create_wo "WO-HARVEST-A1-001" "$FIELD_A_ID" "$LOT_A1_ID" "$LABOR_HARVEST_ID" "2025-12-15T00:00:00Z" 2.5

# A2: Siembra Dic / Cosecha Jun
create_wo "WO-SOWING-A2-001" "$FIELD_A_ID" "$LOT_A2_ID" "$LABOR_SOW_ID"     "2025-12-15T00:00:00Z" 3.0
create_wo "WO-HARVEST-A2-001" "$FIELD_A_ID" "$LOT_A2_ID" "$LABOR_HARVEST_ID" "2025-06-15T00:00:00Z" 3.0

# B1: Siembra Mar / Cosecha Sep
create_wo "WO-SOWING-B1-001" "$FIELD_B_ID" "$LOT_B1_ID" "$LABOR_SOW_ID"     "2025-03-15T00:00:00Z" 1.2
create_wo "WO-HARVEST-B1-001" "$FIELD_B_ID" "$LOT_B1_ID" "$LABOR_HARVEST_ID" "2025-09-15T00:00:00Z" 1.2

echo "== 6) Actualizar toneladas (rendimiento)"
json_put "$API_BASE/lots/$LOT_A1_ID/tons" '{"tons":"20.0"}' >/dev/null || true
json_put "$API_BASE/lots/$LOT_A2_ID/tons" '{"tons":"24.0"}' >/dev/null || true
json_put "$API_BASE/lots/$LOT_B1_ID/tons" '{"tons":"9.6"}'  >/dev/null || true

echo "== 7) Comercialización (precios realistas)"
COMM_PAYLOAD=$(jq -n --argjson project_id "$PROJECT_ID" '
{
  values: [
    { crop_id: 2, board_price: 600.00, freight_cost: 50.00, commercial_cost: 100.00, net_price: 450.00 },
    { crop_id: 1, board_price: 800.00, freight_cost: 50.00, commercial_cost: 100.00, net_price: 650.00 }
  ]
}')
json_post "$API_BASE/projects/$PROJECT_ID/commercializations" "$COMM_PAYLOAD" >/dev/null || true

echo "== 8) Fechas (lot_dates) vía SQL"
# Idempotente: borra si ya existe la secuencia 1 y re-inserta
sql "DELETE FROM lot_dates WHERE (lot_id, sequence) IN (($LOT_A1_ID,1),($LOT_A2_ID,1),($LOT_B1_ID,1));" || true
sql "INSERT INTO lot_dates (lot_id, sowing_date, harvest_date, sequence, created_by, updated_by, created_at, updated_at) VALUES 
($LOT_A1_ID, '2025-06-15 00:00:00', '2025-12-15 00:00:00', 1, $DEMO_USER_ID, $DEMO_USER_ID, now(), now()),
($LOT_A2_ID, '2025-12-15 00:00:00', '2025-06-15 00:00:00', 1, $DEMO_USER_ID, $DEMO_USER_ID, now(), now()),
($LOT_B1_ID, '2025-03-15 00:00:00', '2025-09-15 00:00:00', 1, $DEMO_USER_ID, $DEMO_USER_ID, now(), now())
ON CONFLICT DO NOTHING;"

echo "== 9) Verificaciones"
echo "→ Workorders del proyecto:"
curl -sS -H "X-API-KEY: $API_KEY" -H "X-USER-ID: $USER_ID" "$API_BASE/workorders?project_id=$PROJECT_ID" | jq '.[0:5]' || true
echo "→ Lotes del proyecto:"
curl -sS -H "X-API-KEY: $API_KEY" -H "X-USER-ID: $USER_ID" "$API_BASE/lots?project_id=$PROJECT_ID" | jq '.[0:5]' || true
echo "→ Comercialización:"
curl -sS -H "X-API-KEY: $API_KEY" -H "X-USER-ID: $USER_ID" "$API_BASE/projects/$PROJECT_ID/commercializations" | jq '.' || true

echo "✅ Datos de prueba cargados para proyecto: $PROJECT_NAME (ID=$PROJECT_ID)"
