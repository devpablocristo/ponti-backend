#!/usr/bin/env bash
# lint-tenant-leaks.sh — Salvaguarda CI contra regresiones del refactor de
# multi-tenancy (Fase 7 del plan). Falla si encuentra:
#
#   1. String literal `"tenant context required"` fuera de platform/errors.
#      Debe ir vía `domainerr.TenantMissing()`.
#   2. Llamadas a `authz.MaybeTenantScope` (eliminado en Fase 7).
#      Usar `tenancy.Scope` de platform.
#   3. Llamadas a `authz.TenantScope` / `authz.TenantWhere` (eliminados).
#   4. Patrón `db.Where("tenant_id = ?", ...)` o `db.Where("org_id = ?", ...)`
#      sin pasar por `tenancy.ScopeWithColumn` — bypass del strict mode.
#
# Salida: 0 si todo limpio, 1 si hay hits.
#
# Uso:
#   ./scripts/lint-tenant-leaks.sh
#   CI: agregar a Makefile target `lint` o pipeline step.

set -uo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT" || exit 2

fail=0
note() { printf '\033[31m✗\033[0m %s\n' "$*" >&2; fail=1; }
pass() { printf '\033[32m✓\033[0m %s\n' "$*"; }

echo "lint-tenant-leaks: scanning $ROOT/internal..."

# 1. Mensaje canónico no debe aparecer como string literal en código (excepto domainerr)
if grep -rn '"tenant context required"' internal/ 2>/dev/null | grep -v "_test.go"; then
    note 'string literal "tenant context required" en internal/ — usar domainerr.TenantMissing()'
else
    pass 'string "tenant context required" no aparece inline'
fi

# 2/3. Bridge functions eliminadas en Fase 7
for pattern in 'authz\.MaybeTenantScope' 'authz\.TenantScope\b' 'authz\.TenantWhere\b'; do
    if grep -rnE "$pattern" internal/ 2>/dev/null | grep -v "_test.go"; then
        note "$pattern todavía en uso — migrar a tenancy.Scope / tenancy.Where"
    else
        pass "$pattern eliminado"
    fi
done

# 4. Inline tenant filters sin pasar por platform tenancy
if grep -rnE 'db\.Where\("(tenant_id|org_id) = \?",' internal/ 2>/dev/null | grep -v "_test.go" | grep -v "tenancy"; then
    note 'WHERE tenant_id/org_id inline detectado — usar tenancy.Scope / tenancy.ScopeWithColumn'
else
    pass 'no WHERE inline tenant_id/org_id en repos productivos'
fi

# 5. Sanity: tenancy.Scope sí debe estar importado en repositories
imported=$(grep -lrE 'platform/persistence/gorm/go/tenancy' internal/ 2>/dev/null | wc -l)
echo "tenancy package importado en $imported archivos"

if [ "$fail" -eq 0 ]; then
    echo ""
    pass "lint-tenant-leaks: TODO LIMPIO"
    exit 0
else
    echo ""
    note "lint-tenant-leaks: encontradas regresiones"
    exit 1
fi
