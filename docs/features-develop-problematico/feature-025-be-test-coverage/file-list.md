# file-list.md — feature-025 · BE test coverage sweep

Fuente de verdad: `cat /tmp/flists/be-025.txt` (45 paths). Diff: `0972e565..777e5f6a`. SOURCE = 777e5f6a.

Leyenda extracción: **whole-file** (copiar entero) · **partial-hunks** · **manual-port** · **do-not-extract-yet**.

Todos los archivos son **white-box** (`package <modulo>`, no `<modulo>_test`): acceden a símbolos no
exportados de producción. Por eso cada test SOLO compila si su feature productora ya está en la rama base.

## Propios (todos los de esta feature son propios — son tests nuevos)

### Familia A — `handler_test.go` (tests de handlers HTTP / rutas)

| path | status | tipo | rol | extracción | motivo | riesgo | confianza |
|---|---|---|---|---|---|---|---|
| internal/business-parameters/handler_test.go | A | test handler | rutas + actor propagation (`bparams`) | whole-file | usa `GetArchivedParameters` (002/009) | medio | alta |
| internal/businessinsights/handler_test.go | A | test handler | rutas (`businessinsights`) | whole-file | depende de handler 001 | bajo | alta |
| internal/category/handler_test.go | A | test handler | rutas (`category`) | whole-file | — | bajo | alta |
| internal/class-type/handler_test.go | A | test handler | rutas (`classtype`) | whole-file | — | bajo | alta |
| internal/commercialization/handler_test.go | A | test handler | rutas (`commercialization`) | whole-file | — | bajo | alta |
| internal/crop/handler_test.go | A | test handler | rutas (`crop`) | whole-file | — | bajo | alta |
| internal/dashboard/handler_test.go | A | test handler | rutas (`dashboard`) | whole-file | — | bajo | alta |
| internal/dollar/handler_test.go | A | test handler | rutas (`dollar`) | whole-file | — | bajo | alta |
| internal/invoice/handler_test.go | A | test handler | rutas (`invoice`) | whole-file | — | bajo | alta |
| internal/provider/handler_test.go | A | test handler | rutas (`provider`) | whole-file | — | bajo | alta |
| internal/report/handler_test.go | A | test handler | rutas (`report`) | whole-file | — | bajo | alta |
| internal/work-order/handler_test.go | **M** | test handler | rutas archive/restore/hard (`workorder`) | whole-file | renombra stub `DeleteWorkOrderByID`→`HardDeleteWorkOrder`, agrega `ListArchivedWorkOrders`, prueba `/archive`,`/restore`,`/hard` | **ALTO** | alta |

### Familia B — `repository_tenant_test.go` (aislamiento multi-tenant, sqlite in-memory)

| path | status | tipo | rol | extracción | motivo | riesgo | confianza |
|---|---|---|---|---|---|---|---|
| internal/business-parameters/repository_tenant_test.go | A | test repo tenant | aislamiento + strict mode | whole-file | usa `contextkeys.OrgID` (001), `NewRepository`, `ListAll` | medio | alta |
| internal/businessinsights/repository_tenant_test.go | A | test repo tenant | aislamiento | whole-file | dep 001/003 | medio | alta |
| internal/category/repository_tenant_test.go | A | test repo tenant | aislamiento | whole-file | dep 001/003 | medio | alta |
| internal/class-type/repository_tenant_test.go | A | test repo tenant | aislamiento | whole-file | dep 001/003 | medio | alta |
| internal/commercialization/repository_tenant_test.go | A | test repo tenant | aislamiento | whole-file | dep 001/003 | medio | alta |
| internal/crop/repository_tenant_test.go | A | test repo tenant | aislamiento | whole-file | dep 001/003 | medio | alta |
| internal/customer/repository_tenant_test.go | A | test repo tenant | aislamiento | whole-file | dep 001/003 | medio | alta |
| internal/dashboard/repository_tenant_test.go | A | test repo tenant | aislamiento | whole-file | dep 001/003 | medio | alta |
| internal/dollar/repository_tenant_test.go | A | test repo tenant | aislamiento | whole-file | dep 001/003 | medio | alta |
| internal/field/repository_tenant_test.go | A | test repo tenant | aislamiento | whole-file | dep 001/003 | medio | alta |
| internal/investor/repository_tenant_test.go | A | test repo tenant | aislamiento | whole-file | dep 001/003 | medio | alta |
| internal/invoice/repository_tenant_test.go | A | test repo tenant | aislamiento | whole-file | dep 001/003 | medio | alta |
| internal/labor/repository_tenant_test.go | A | test repo tenant | aislamiento | whole-file | dep 001/003 | medio | alta |
| internal/lease-type/repository_tenant_test.go | A | test repo tenant | aislamiento | whole-file | dep 001/003 | medio | alta |
| internal/lot/repository_tenant_test.go | A | test repo tenant | aislamiento | whole-file | dep 001/003 (lot-metrics ya porteado) | medio | alta |
| internal/manager/repository_tenant_test.go | A | test repo tenant | aislamiento | whole-file | dep 001/003 | medio | alta |
| internal/provider/repository_tenant_test.go | A | test repo tenant | aislamiento | whole-file | dep 001/003 | medio | alta |
| internal/report/repository_tenant_test.go | A | test repo tenant | aislamiento | whole-file | dep 001/003 | medio | alta |
| internal/stock/repository_tenant_test.go | A | test repo tenant | aislamiento | whole-file | dep 001/003 | medio | alta |
| internal/supply/repository_tenant_test.go | A | test repo tenant | aislamiento | whole-file | dep 001/003 | medio | alta |
| internal/work-order-draft/repository_tenant_test.go | A | test repo tenant | aislamiento (`workorderdraft`) | whole-file | dep 001/003 | medio | alta |
| internal/work-order/repository_tenant_test.go | A | test repo tenant | aislamiento (`workorder`) | whole-file | dep 001/003 | medio | alta |

### Familia C — `repository_archived_refs_test.go` (integridad vs entidades archivadas)

| path | status | tipo | rol | extracción | motivo | riesgo | confianza |
|---|---|---|---|---|---|---|---|
| internal/customer/repository_archived_refs_test.go | A | test integridad | `assertCustomerReferencesActive` | whole-file | dep 009 (función no exportada) | medio | alta |
| internal/field/repository_archived_refs_test.go | A | test integridad | `assertFieldReferencesActive` | whole-file | dep 009 | medio | alta |
| internal/investor/repository_archived_refs_test.go | A | test integridad | `assertInvestorReferencesActive` | whole-file | dep 009 | medio | alta |
| internal/labor/repository_archived_refs_test.go | A | test integridad | `assertLaborReferencesActive` | whole-file | dep 009 | medio | alta |
| internal/lot/repository_archived_refs_test.go | A | test integridad | `assertLotReferencesActive` | whole-file | dep 009 | medio | alta |
| internal/manager/repository_archived_refs_test.go | A | test integridad | `assertManagerReferencesActive` | whole-file | dep 009 | medio | alta |
| internal/supply/repository_archived_refs_test.go | A | test integridad | `assertSupplyReferencesActive` | whole-file | dep 009 | medio | alta |
| internal/supply/repository_movement_archived_refs_test.go | A | test integridad | `assertSupplyMovementReferencesActive` (tablas projects/supplies/investors/providers/actors) | whole-file | dep 009 | medio | alta |
| internal/work-order-draft/repository_archived_refs_test.go | A | test integridad (`workorderdraft`) | `assert...ReferencesActive` | whole-file | dep 009 | medio | alta |
| internal/work-order/repository_archived_refs_test.go | A | test integridad (`workorder`) | `assert...ReferencesActive` | whole-file | dep 009 | medio | alta |

## Compartidos (partial-hunks)

Ninguno. No hay tocados los archivos compartidos típicos del repo (`wire/*`, `cmd/api/*`, `go.mod`, `go.sum`,
`Makefile`, `internal/shared/**`). Cada test vive aislado en el paquete de su módulo.

## Requeridos por dependencia (NO están en esta flist — deben venir de otras features)

- Producción de **001** (contextkeys/tenant en repositories) — sin esto, los 23 `repository_tenant_test.go` no compilan.
- Producción de **002** (`HardDeleteWorkOrder`, archive/restore en handler/usecases) — sin esto, `work-order/handler_test.go` (M) y los handler_test que usan archive no compilan.
- Producción de **009** (`assert*ReferencesActive`, `GetArchivedParameters`, `ListArchivedWorkOrders`) — sin esto, los 10 `repository_archived_refs_test.go` no compilan.

## Dudosos

Ninguno. La lista es homogénea (solo tests) y verificada contra el código de producción del SOURCE.

## NO traer todavía (depende de orden, no de contenido)

- **Todo el paquete** si su feature productora no está mergeada aún. Concretamente:
  - No mergear ningún `repository_tenant_test.go` antes de 001/003.
  - No mergear ningún `repository_archived_refs_test.go` ni `GetArchivedParameters`-relacionado antes de 009.
  - No mergear `work-order/handler_test.go` (M) antes de 002 (renombre `DeleteWorkOrderByID`→`HardDeleteWorkOrder`).
