# feature-018 · data-integrity-admin · dependencies (BE)

## Depende de (este BE necesita que exista antes)

### FUERTES (sin esto NO compila)

1. **Tenancy base — `internal/shared/authz`** (features **001 be-platform-tenancy-refactor** / **003 be-multitenant-db-hardening**).
   - Símbolos usados por los 4 métodos RAW: `authz.TenantFromContext`, `authz.TenantStrictModeEnabled`, `authz.TenantMissing`.
   - Estado en develop: **AUSENTE** (`git ls-tree develop -- internal/shared/authz` no devuelve nada).
   - Tipo de dependencia: **cross-feature dura**. 018-BE no debe mergear antes que 001/003.

2. **Métodos RAW en repos compartidos** (hunks dentro de archivos que pertenecen a otras features; en el rango `0972e565..777e5f6a` pero asignados a 001/003/009/supply/lot/project):
   - `internal/report/repository.go::GetRawNetIncome` — AUSENTE en develop.
   - `internal/supply/repository.go::GetRawSupplyInvestment` — AUSENTE en develop.
   - `internal/project/repository.go::GetRawAdminCostTotal` — AUSENTE en develop.
   - `internal/lot/repository.go::GetRawLeaseExecuted` — AUSENTE en develop.
   - `internal/work-order/repository.go::GetRawDirectCost` — PRESENTE en develop (pero sin filtro tenant).
   - Estos métodos definen el lado "RAW" de cada control; las interfaces (`ports`) viven en mi `usecases.go`, pero la implementación vive en estos repos.

3. **Re-cableado wire** (`wire/data_integrity_providers.go` + `wire/wire_gen.go`):
   - Firma nueva de `ProvideDataIntegrityUseCases`: orden `dashboard, workorder, report, supply, project, lot`; **elimina** `StockRepositoryPort`; **agrega** `ProjectRepositoryPort`; los providers de report/supply/project/lot ahora reciben el **repo concreto** (`*report.ReportRepository`, `*supply.Repository`, `*project.Repository`, `*lot.Repository`) porque `GetRaw*` no se expone en las interfaces públicas.
   - Estado en develop: providers con la firma VIEJA (con `Stock`, sin `Project`). Hay que actualizar y regenerar.

### DÉBILES (ya satisfechas en develop)

4. **`internal/dashboard/usecases/domain`** — campos de `DashboardBalanceSummary` consumidos por los controles: `DirectCostsExecutedUSD`, `IncomeUSD`, `SemillasInvertidosUSD`, `AgroquimicosInvertidosUSD`, `FertilizantesInvertidosUSD`, `StructureExecutedUSD`, `RentExecutedUSD`. **Todos presentes en develop** (verificado). Sin acción.

5. **Migración de import paths `core/* → platform/*`** (parte de new-cns3). `platform/errors/go/domainerr` y `platform/security/go/contextkeys` ya se usan ampliamente en develop. Sin acción.

### INCIERTAS

6. **Tenantización de `GetRawDirectCost`** (work-order): en develop existe sin authz; en SOURCE tiene filtro tenant. ¿Entra en 018 o lo aporta la feature de tenancy? Decisión humana. No bloquea compilación (la firma del método no cambia).

## Bloquea a (qué espera a este BE)

- **FE feature-018** (`pages/admin/data-integrity`, hook `useDatabase`): espera el DTO con `check_type/severity/recommendation` y 5 checks. Recomendado **BE-first**; el FE debe mergear después.

## Cross-repo

- **Mismo `feature-018` en ambos repos.** Contrato compartido: endpoint `GET /data-integrity/costs-check?project_id=` y el shape `IntegrityReportResponse { checks: IntegrityCheckDTO[] }`.
- Cambios de contrato hacia el FE: +3 campos (`check_type`, `severity`, `recommendation`), `status` ahora `OK|ERROR` emitidos por `buildCheck` (el dominio documenta también `WARNING/SKIPPED`, hoy no emitidos).

## Archivos / tipos / config / migraciones / APIs compartidos

| recurso | compartido con | nota |
|---|---|---|
| `internal/shared/authz/*` | 001/003 y todo el repo | base de multitenancy |
| `internal/report/repository.go` | 001/003/report features | partial-hunks: solo `GetRawNetIncome` |
| `internal/supply/repository.go` | supply/tenancy | partial-hunks: solo `GetRawSupplyInvestment` |
| `internal/project/repository.go` | 010-projects/tenancy | partial-hunks: solo `GetRawAdminCostTotal` |
| `internal/lot/repository.go` | lot/tenancy/lot-metrics | partial-hunks: solo `GetRawLeaseExecuted` |
| `internal/work-order/repository.go` | work-order/tenancy | `GetRawDirectCost` ya en develop |
| `wire/data_integrity_providers.go` | wire/DI | tomar solo el bloque DataIntegrity |
| `wire/wire_gen.go` | TODAS las features | regenerar, no copiar |
| API `GET /data-integrity/costs-check` | FE 018 | contrato cross-repo |
| Migraciones | — | NINGUNA en 018 |

## Recomendación de orden

1. **001 / 003** (tenancy + `internal/shared/authz`) → develop.
2. (En el mismo PR de **018-BE**) partial-hunks de los 4 RAW + re-cableado wire + módulo entero.
3. Regenerar `wire_gen.go`.
4. **FE 018** después del BE (BE-first).
