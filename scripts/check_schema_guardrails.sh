#!/usr/bin/env bash
set -euo pipefail

DB_HOST="${DB_HOST:-127.0.0.1}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-}"
DB_NAME="${DB_NAME:-}"
DB_PASSWORD="${DB_PASSWORD:-}"
MIGRATIONS_DIR="${MIGRATIONS_DIR:-migrations_v4}"

if [[ -z "${DB_USER}" || -z "${DB_NAME}" || -z "${DB_PASSWORD}" ]]; then
  echo "ERROR: DB_USER/DB_NAME/DB_PASSWORD son requeridos para check_schema_guardrails.sh" >&2
  exit 1
fi

if [[ ! -d "${MIGRATIONS_DIR}" ]]; then
  echo "ERROR: MIGRATIONS_DIR no existe: ${MIGRATIONS_DIR}" >&2
  exit 1
fi

expected_version="$(
  python3 - <<'PY' "${MIGRATIONS_DIR}"
import os
import re
import sys

path = sys.argv[1]
mx = 0
for name in os.listdir(path):
    m = re.match(r"(\d+)_.*\.up\.sql$", name)
    if m:
        mx = max(mx, int(m.group(1)))
print(mx)
PY
)"

if [[ -z "${expected_version}" || "${expected_version}" == "0" ]]; then
  echo "ERROR: no se pudo resolver versión esperada de migraciones" >&2
  exit 1
fi

echo "[schema-check] expected migration version: ${expected_version}"

query_result="$(
  PGPASSWORD="${DB_PASSWORD}" psql \
    -h "${DB_HOST}" \
    -p "${DB_PORT}" \
    -U "${DB_USER}" \
    -d "${DB_NAME}" \
    -v ON_ERROR_STOP=1 \
    -At -F '|' <<'SQL'
SELECT
  COALESCE(MAX(version), 0)::text AS max_version,
  COALESCE(bool_or(dirty), false)::text AS any_dirty
FROM schema_migrations;

SELECT CASE
  WHEN position('workorder_investor_splits' in pg_get_viewdef('v4_report.labor_list'::regclass, true)) > 0
  THEN 'yes' ELSE 'no' END;

SELECT CASE
  WHEN position('workorder_investor_splits' in pg_get_viewdef('v4_calc.investor_real_contributions'::regclass, true)) > 0
  THEN 'yes' ELSE 'no' END;

SELECT CASE
  WHEN position('income_net_total_for_lot' in pg_get_viewdef('v4_calc.lot_base_costs'::regclass, true)) > 0
    OR position('rent_per_ha_for_lot' in pg_get_viewdef('v4_calc.lot_base_costs'::regclass, true)) > 0
    OR position('yield_tn_per_ha_for_lot' in pg_get_viewdef('v4_calc.lot_base_costs'::regclass, true)) > 0
  THEN 'yes' ELSE 'no' END;
SQL
)"

max_version="$(printf "%s" "${query_result}" | sed -n '1p' | cut -d'|' -f1)"
any_dirty="$(printf "%s" "${query_result}" | sed -n '1p' | cut -d'|' -f2)"
labor_uses_splits="$(printf "%s" "${query_result}" | sed -n '2p')"
irc_uses_splits="$(printf "%s" "${query_result}" | sed -n '3p')"
lot_base_costs_uses_per_lot_functions="$(printf "%s" "${query_result}" | sed -n '4p')"

if [[ "${max_version}" != "${expected_version}" ]]; then
  echo "ERROR: schema_migrations en ${DB_NAME} está en ${max_version}, esperado ${expected_version}" >&2
  exit 1
fi

if [[ "${any_dirty}" != "false" ]]; then
  echo "ERROR: schema_migrations tiene estado dirty=true en ${DB_NAME}" >&2
  exit 1
fi

if [[ "${labor_uses_splits}" != "yes" ]]; then
  echo "ERROR: vista v4_report.labor_list no usa workorder_investor_splits en ${DB_NAME}" >&2
  exit 1
fi

if [[ "${irc_uses_splits}" != "yes" ]]; then
  echo "ERROR: vista v4_calc.investor_real_contributions no usa workorder_investor_splits en ${DB_NAME}" >&2
  exit 1
fi

if [[ "${lot_base_costs_uses_per_lot_functions}" != "no" ]]; then
  echo "ERROR: vista v4_calc.lot_base_costs usa funciones por-lote lentas en ${DB_NAME}" >&2
  exit 1
fi

echo "[schema-check] OK - schema y vistas alineadas en ${DB_NAME}"
