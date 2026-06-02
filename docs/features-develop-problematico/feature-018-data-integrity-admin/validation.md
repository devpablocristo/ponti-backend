# feature-018 · data-integrity-admin · validation (BE)

## Checklist pre-PR (BE)

- [ ] `internal/shared/authz` presente en develop (`git -C <repo> ls-tree -r --name-only develop -- internal/shared/authz` lista `authz.go`).
- [ ] Los 8 archivos de mi flist traídos whole-file desde `develop-problematico~1`.
- [ ] Hunks RAW traídos (partial) en `report/supply/project/lot/repository.go`:
      `grep -rn "GetRawNetIncome\|GetRawSupplyInvestment\|GetRawAdminCostTotal\|GetRawLeaseExecuted" internal/` → 4 funcs presentes.
- [ ] `wire/data_integrity_providers.go` re-cableado: sin `StockRepositoryPort`, con `ProjectRepositoryPort`.
      `grep -n "StockRepositoryPort\|ProjectRepositoryPort" wire/data_integrity_providers.go`
- [ ] `wire/wire_gen.go` regenerado: `cd wire && go generate ./...` no deja diff sucio.
- [ ] `grep -rn "devpablocristo/core/errors" internal/data-integrity` → 0 (todo es platform).
- [ ] `git diff --check` sin marcadores de conflicto / whitespace roto.

## Comandos de build/test (BE)

```bash
cd <repo>
go build ./...                              # gate real: compila el repo entero
go vet ./internal/data-integrity/...
go test ./internal/data-integrity/...       # handler + usecases + tenant + mocks
# wire idempotente:
cd wire && go generate ./... && git diff --stat wire/wire_gen.go   # debe quedar vacío
```

Tests esperados (nombres reales en SOURCE):
- `TestHandler_CheckCostsCoherence_ParsesProjectID`
- `TestHandler_CheckCostsCoherence_RequiresProjectID`
- `TestCheckCostsCoherence_AllOK`
- `TestCheckCostsCoherence_ErrorWhenDiffExceedsTolerance`
- `TestCheckCostsCoherence_DiffWithinTolerance`
- `TestCheckCostsCoherence_RequiresProjectID`
- `TestCheckCostsCoherence_PropagatesRepoError`
- `TestBuildCheck`
- `TestCheckCostsCoherence_PropagatesTenantContext`

## Validación manual (API)

```bash
# 200 con 5 checks:
curl -s "<base>/data-integrity/costs-check?project_id=<real>" | jq '.checks | length'   # = 5
curl -s "<base>/data-integrity/costs-check?project_id=<real>" | jq '.checks[0] | keys'   # incluye check_type, severity, recommendation, status, difference_a, tolerance

# 400 sin project_id:
curl -s -o /dev/null -w "%{http_code}\n" "<base>/data-integrity/costs-check"             # = 400
```

## Casos borde

- `project_id` ausente / vacío / `0` / negativo → 400 (`ParseOptionalInt64Query` + guardia `*ProjectID <= 0`).
- Proyecto sin dashboard (summary nil) → 500 `dashboard summary is empty for project`.
- Diferencia exactamente = tolerancia (1 USD) → `OK` (la regla es `> tolerance`).
- Tenant ausente con modo estricto activado → los RAW devuelven `TenantMissing` → control falla con error 500 controlado.
- Datos cruzados de otro tenant en control 1 (si `GetRawDirectCost` no está tenantizado) → posible falso ERROR/OK.

## Qué revisar en UI / API / DB / env

- **API**: shape `IntegrityReportResponse { checks: [...] }`; cada item con `system_value`, `recalc_a_value`, `difference_a` formateados (string decimal vía `formatDecimal`).
- **DB**: vistas SSOT `v4_report.dashboard_management_balance` / `v4_ssot.*` desplegadas; tablas base `public.workorders, workorder_items, labors, supplies, lots, fields, crop_commercializations, supply_movements, categories, projects` con `deleted_at` y `tenant_id`.
- **env**: flag/condición de `TenantStrictModeEnabled` configurada como en el resto del repo.

## Qué validar en el OTRO repo (FE)

- `pages/admin/data-integrity` renderiza 5 checks y muestra `severity`/`recommendation`.
- `useDatabase` mapea por `control_number` y no asume cantidad fija de controles.
- Build FE (`yarn build`), tests (`yarn test`), y si hay e2e de la página admin, correrlos.
- Confirmar que el FE no exige `WARNING/SKIPPED` (el BE no los emite hoy).

## Señales de incompletitud / incompatibilidad

- `undefined: authz.TenantFromContext` → falta tenancy (001/003).
- `r.GetRaw*** undefined` → faltan hunks RAW.
- `wire`: provider faltante/sobrante o `cannot use ... as dataintegrity.XxxRepositoryPort` → providers/wire_gen desincronizados.
- FE muestra columnas vacías de `severity/recommendation` → BE viejo aún desplegado (mergeá BE-first).
