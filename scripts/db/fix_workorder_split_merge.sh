#!/usr/bin/env bash
set -euo pipefail

# Fusiona workorders duplicadas del split viejo:
# - BASE_NUMBER=23 PROJECT_ID=123 ./scripts/db/fix_workorder_split_merge.sh
#
# Reglas:
# - Mantiene 1 sola OT: la canónica es la que tiene number=BASE_NUMBER (si hay varias, la de menor id).
# - Suma effective_area e items.total_used de todas las OTs del grupo.
# - Recalcula final_dose = total_used / total_effective_area.
# - Crea workorder_investor_splits en la OT canónica usando % por effective_area.
# - Soft-delete de las OTs duplicadas (ej: 23-2, 23-3, etc) y sus items.

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
ENV_FILE="${ROOT_DIR}/.env"

if [[ ! -f "${ENV_FILE}" ]]; then
  echo "ERROR: No existe ${ENV_FILE}"
  exit 1
fi

set -a
source "${ENV_FILE}"
set +a

BASE_NUMBER="${BASE_NUMBER:-}"
PROJECT_ID="${PROJECT_ID:-}"

if [[ -z "${BASE_NUMBER}" || -z "${PROJECT_ID}" ]]; then
  echo "Uso: BASE_NUMBER=23 PROJECT_ID=123 bash ./scripts/db/fix_workorder_split_merge.sh"
  exit 1
fi

# Validaciones básicas para evitar inyección accidental en queries inline.
if ! [[ "${PROJECT_ID}" =~ ^[0-9]+$ ]]; then
  echo "ERROR: PROJECT_ID inválido (${PROJECT_ID}). Debe ser numérico."
  exit 1
fi
if ! [[ "${BASE_NUMBER}" =~ ^[A-Za-z0-9][A-Za-z0-9._-]*$ ]]; then
  echo "ERROR: BASE_NUMBER inválido (${BASE_NUMBER}). Use solo letras/números y . _ -"
  exit 1
fi

SQL_FILE="$(mktemp)"
cat > "${SQL_FILE}" <<'SQL'
BEGIN;

-- Parametros esperados:
--  :project_id
--  :base_number

-- 1) Validaciones se hacen en bash (psql no soporta RAISE EXCEPTION en SQL plano)

-- 2) Normalizar observaciones canónicas (quitar sufijos " | Split xx%")
UPDATE public.workorders w
SET observations = regexp_replace(COALESCE(w.observations, ''), '\\s*\\|\\s*Split\\s+[0-9]+(\\.[0-9]+)?%\\s*$', '', 'i'),
    updated_at = now()
WHERE w.id = (
  SELECT id
  FROM public.workorders
  WHERE deleted_at IS NULL
    AND project_id = :project_id
    AND number = :'base_number'
  ORDER BY id ASC
  LIMIT 1
);

-- 3) Recalcular items agregados
--    - total_used: suma por supply
--    - final_dose: total_used / total_area (para mantener consistencia)
WITH candidates AS (
  SELECT id
  FROM public.workorders
  WHERE deleted_at IS NULL
    AND project_id = :project_id
    AND (number = :'base_number' OR number LIKE (:'base_number' || '-%'))
),
canon AS (
  SELECT id
  FROM public.workorders
  WHERE deleted_at IS NULL
    AND project_id = :project_id
    AND number = :'base_number'
  ORDER BY id ASC
  LIMIT 1
),
total AS (
  SELECT COALESCE(SUM(effective_area), 0)::numeric AS total_area
  FROM public.workorders
  WHERE id IN (SELECT id FROM candidates)
),
agg_items AS (
  SELECT
    wi.supply_id,
    COALESCE(SUM(wi.total_used), 0)::numeric AS total_used
  FROM public.workorder_items wi
  WHERE wi.deleted_at IS NULL
    AND wi.workorder_id IN (SELECT id FROM candidates)
  GROUP BY wi.supply_id
),
del_old AS (
  -- soft-delete items previos canónicos
  UPDATE public.workorder_items
  SET deleted_at = now()
  WHERE deleted_at IS NULL
    AND workorder_id = (SELECT id FROM canon)
  RETURNING 1
)
INSERT INTO public.workorder_items (workorder_id, supply_id, total_used, final_dose)
SELECT
  (SELECT id FROM canon) AS workorder_id,
  ai.supply_id,
  ai.total_used,
  CASE
    WHEN (SELECT total_area FROM total) > 0 THEN ai.total_used / (SELECT total_area FROM total)
    ELSE 0
  END AS final_dose
FROM agg_items ai;

-- 4) Crear splits en workorder_investor_splits (por effective_area ORIGINAL)
-- Importante: esto debe correr ANTES de actualizar el effective_area canónico,
-- porque si no el canónico queda con el total y distorsiona los porcentajes.
WITH candidates AS (
  SELECT id, investor_id, effective_area
  FROM public.workorders
  WHERE deleted_at IS NULL
    AND project_id = :project_id
    AND (number = :'base_number' OR number LIKE (:'base_number' || '-%'))
),
canon AS (
  SELECT id
  FROM public.workorders
  WHERE deleted_at IS NULL
    AND project_id = :project_id
    AND number = :'base_number'
  ORDER BY id ASC
  LIMIT 1
),
total AS (
  SELECT COALESCE(SUM(effective_area), 0)::numeric AS total_area
  FROM candidates
),
agg AS (
  SELECT
    investor_id,
    COALESCE(SUM(effective_area), 0)::numeric AS area
  FROM candidates
  GROUP BY investor_id
),
del_splits AS (
  -- borrar splits previos canónicos
  UPDATE public.workorder_investor_splits
  SET deleted_at = now(), updated_at = now()
  WHERE deleted_at IS NULL
    AND workorder_id = (SELECT id FROM canon)
  RETURNING 1
)
INSERT INTO public.workorder_investor_splits (workorder_id, investor_id, percentage)
SELECT
  (SELECT id FROM canon) AS workorder_id,
  a.investor_id,
  CASE
    WHEN (SELECT total_area FROM total) > 0 THEN (a.area / (SELECT total_area FROM total)) * 100
    ELSE 0
  END AS percentage
FROM agg a
WHERE a.investor_id IS NOT NULL;

-- 5) Actualizar effective_area canónica (suma de todas)
WITH candidates AS (
  SELECT id
  FROM public.workorders
  WHERE deleted_at IS NULL
    AND project_id = :project_id
    AND (number = :'base_number' OR number LIKE (:'base_number' || '-%'))
),
canon AS (
  SELECT id
  FROM public.workorders
  WHERE deleted_at IS NULL
    AND project_id = :project_id
    AND number = :'base_number'
  ORDER BY id ASC
  LIMIT 1
),
total AS (
  SELECT COALESCE(SUM(effective_area), 0)::numeric AS total_area
  FROM public.workorders
  WHERE id IN (SELECT id FROM candidates)
)
UPDATE public.workorders
SET effective_area = (SELECT total_area FROM total),
    updated_at = now()
WHERE id = (SELECT id FROM canon);

-- 6) Soft-delete duplicadas (dejar viva solo la canónica)
WITH canon AS (
  SELECT id
  FROM public.workorders
  WHERE deleted_at IS NULL
    AND project_id = :project_id
    AND number = :'base_number'
  ORDER BY id ASC
  LIMIT 1
),
dups AS (
  SELECT id
  FROM public.workorders
  WHERE deleted_at IS NULL
    AND project_id = :project_id
    AND number LIKE (:'base_number' || '-%')
)
UPDATE public.workorder_items wi
SET deleted_at = now()
WHERE wi.deleted_at IS NULL
  AND wi.workorder_id IN (SELECT id FROM dups);

WITH dups AS (
  SELECT id
  FROM public.workorders
  WHERE deleted_at IS NULL
    AND project_id = :project_id
    AND number LIKE (:'base_number' || '-%')
)
UPDATE public.workorders
SET deleted_at = now(),
    updated_at = now(),
    observations = CASE
      WHEN observations IS NULL OR observations = '' THEN 'Merged into ' || :'base_number'
      ELSE observations || ' | Merged into ' || :'base_number'
    END
WHERE id IN (SELECT id FROM dups);

COMMIT;
SQL

echo "Merging split workorders for project_id=${PROJECT_ID} base_number=${BASE_NUMBER} ..."

# Validaciones previas (evita ejecutar SQL a medias)
CANON_ID="$(
  docker exec -i ponti-backend-ponti-db-1 psql -U "${DB_USER}" -d "${DB_NAME}" -At \
    -c "select id from public.workorders where deleted_at is null and project_id=${PROJECT_ID} and number='${BASE_NUMBER}' order by id asc limit 1;"
)"
if [[ -z "${CANON_ID}" ]]; then
  echo "ERROR: No existe OT canónica number=${BASE_NUMBER} en project_id=${PROJECT_ID}"
  exit 1
fi

COUNT="$(
  docker exec -i ponti-backend-ponti-db-1 psql -U "${DB_USER}" -d "${DB_NAME}" -At \
    -c "select count(*) from public.workorders where deleted_at is null and project_id=${PROJECT_ID} and (number='${BASE_NUMBER}' or number like '${BASE_NUMBER}-%');"
)"
if [[ "${COUNT}" -le 1 ]]; then
  echo "ERROR: No hay nada para mergear (encontradas ${COUNT} OTs)"
  exit 1
fi

TOTAL_AREA="$(
  docker exec -i ponti-backend-ponti-db-1 psql -U "${DB_USER}" -d "${DB_NAME}" -At \
    -c "select coalesce(sum(effective_area),0)::numeric from public.workorders where deleted_at is null and project_id=${PROJECT_ID} and (number='${BASE_NUMBER}' or number like '${BASE_NUMBER}-%');"
)"
python - <<PY
import sys
try:
    v=float("${TOTAL_AREA}")
except Exception:
    print("ERROR: total_area inválida (${TOTAL_AREA})")
    sys.exit(1)
if v<=0:
    print("ERROR: total_area <= 0 (${TOTAL_AREA})")
    sys.exit(1)
PY

docker exec -i ponti-backend-ponti-db-1 psql \
  -U "${DB_USER}" \
  -d "${DB_NAME}" \
  -v ON_ERROR_STOP=1 \
  -v "project_id=${PROJECT_ID}" \
  -v "base_number=${BASE_NUMBER}" \
  -f - < "${SQL_FILE}"

rm -f "${SQL_FILE}"

echo "OK"

