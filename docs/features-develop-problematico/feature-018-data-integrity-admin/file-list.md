# feature-018 · data-integrity-admin · file-list (BE)

Fuente de la lista: `/tmp/flists/be-018.txt` (formato `STATUS<TAB>path`).
Rango de diff: `0972e565..777e5f6a`. SOURCE = `develop-problematico~1` (SHA `777e5f6a`).

Leyenda extracción: `whole-file` = traer el archivo entero desde SOURCE ·
`partial-hunks` = traer solo algunas líneas (archivo compartido con otras features) ·
`manual-port` = re-aplicar a mano · `do-not-extract-yet` = no traer en este PR.

## Propios del módulo (mi flist — todos extraíbles whole-file)

| path | status | tipo | rol en la feature | extracción | motivo | riesgo | confianza |
|---|---|---|---|---|---|---|---|
| `internal/data-integrity/usecases.go` | M | lógica/casos de uso | reescritura: 5 controles SSOT-vs-RAW + `buildCheck`; define los 6 ports | whole-file | 1068→343 líneas, todo es del módulo; no se mezcla con otras features | medio (requiere RAW methods externos en runtime) | alta |
| `internal/data-integrity/handler.go` | M | handler HTTP | ruta `GET /data-integrity/costs-check`; timeout 8m→30s; quita `GetProtected()` del port; import platform/domainerr | whole-file | cambios 100% del módulo | bajo | alta |
| `internal/data-integrity/handler/dto/integrity_check.go` | M | DTO | +`check_type`,+`severity`,+`recommendation` en `IntegrityCheckDTO` | whole-file | cambio del módulo; impacta contrato con FE | bajo (BE) / medio (FE) | alta |
| `internal/data-integrity/usecases/domain/types.go` | M | dominio | +`CheckType/Severity/Recommendation`; **quita json tags** (purity, cf. 027); `Status: OK,ERROR,WARNING,SKIPPED` | whole-file | del módulo; la limpieza de tags es local a este archivo | bajo | alta |
| `internal/data-integrity/handler_test.go` | A | test | `ParsesProjectID`, `RequiresProjectID` (stub de usecases) | whole-file | nuevo, autocontenido | bajo | alta |
| `internal/data-integrity/usecases_tenant_test.go` | A | test | propaga tenant context a los 6 repos (capture mocks) | whole-file | nuevo; importa `platform/security/go/contextkeys` (ya en develop) | bajo | alta |
| `internal/data-integrity/usecases_test.go` | M | test | reescrito a la nueva API: `AllOK`, `ErrorWhenDiffExceedsTolerance`, `DiffWithinTolerance`, `RequiresProjectID`, `PropagatesRepoError`, `TestBuildCheck` | whole-file | acompaña la reescritura de usecases.go | bajo | alta |
| `internal/data-integrity/usecases_mock_test.go` | M | test (mocks) | gomock para `Dashboard/WorkOrder/Report/Supply/Project/Lot` ports | whole-file | regenerado para los 6 ports nuevos | bajo | alta |

## Compartidos (partial-hunks) — NO están en mi flist pero el módulo NO compila sin ellos

Estos archivos pertenecen a OTRAS features (por eso no están en `be-018.txt`), pero la
reescritura de `usecases.go` referencia métodos que solo existen en SOURCE. Traer SOLO los
hunks de los métodos `GetRaw*` (no el archivo entero, que arrastra cambios 001/003/009/etc.).

| path | status | tipo | rol en la feature | extracción | motivo | riesgo | confianza |
|---|---|---|---|---|---|---|---|
| `internal/report/repository.go` | M (en rango) | repo SQL | método `GetRawNetIncome` (∑ lots.tons × cc.net_price) | partial-hunks | archivo con +133/-85; solo el hunk de `GetRawNetIncome` es de 018 | alto (mezcla) | media |
| `internal/supply/repository.go` | M (en rango) | repo SQL | método `GetRawSupplyInvestment` (∑ sm.quantity × s.price, type_id∈1,2,3) | partial-hunks | archivo con +249/-136; resto es de supply/tenancy | alto | media |
| `internal/project/repository.go` | M (en rango) | repo SQL | método `GetRawAdminCostTotal` (admin_cost × Σ hectares) | partial-hunks | archivo gigante (+1642/-431); casi todo es 001/003 | alto | media |
| `internal/lot/repository.go` | M (en rango) | repo SQL | método `GetRawLeaseExecuted` (lease_type 3/4 × hectares) | partial-hunks | archivo con +367/-79; resto lot/tenancy | alto | media |
| `internal/work-order/repository.go` | M (en rango) | repo SQL | `GetRawDirectCost` (ya en develop, **sin** tenant filter) → tenantizar | partial-hunks | en develop existe sin authz; traer solo el hunk del filtro tenant si se desea paridad | medio | media |
| `wire/data_integrity_providers.go` | M (en rango) | wire/DI | nueva firma `ProvideDataIntegrityUseCases`, quita `Stock`, agrega `Project`, repos concretos | partial-hunks | archivo de wire compartido; tomar solo el bloque DataIntegrity | medio | alta |
| `wire/wire_gen.go` | M (en rango) | wire generado | cableado regenerado de data-integrity | manual-port (regenerar) | NO copiar a mano: correr `go generate ./wire/...` tras los providers | medio | alta |

## Requeridos por dependencia (otra feature debe ir antes)

| path | status | tipo | rol | extracción | motivo | riesgo | confianza |
|---|---|---|---|---|---|---|---|
| `internal/shared/authz/authz.go` | A (en rango, owner 001/003) | infra tenancy | `TenantFromContext`, `TenantStrictModeEnabled`, `TenantMissing` usados por los RAW methods | do-not-extract-yet (lo trae 001/003) | NO existe en develop; es base de tenancy | alto si falta | alta |
| `internal/shared/authz/authz_test.go` | A (owner 001/003) | test | acompaña authz | do-not-extract-yet | viene con su feature | bajo | alta |

## Dudosos

| path | nota | confianza |
|---|---|---|
| `internal/work-order/repository.go` (`GetRawDirectCost`) | en develop la versión SIN tenant ya compila con `usecases.go` nuevo (la firma del método coincide). La tenantización es deseable por consistencia, pero NO bloquea compilación. Decidir si entra en 018 o en la feature de tenancy. | media |
| json-tags removidas en `usecases/domain/types.go` | técnicamente parte de la "domain purity" (027), pero como es local al módulo conviene extraerla junto a 018 para no dejar el dominio a medias. | media |

## NO traer todavía (en este PR de 018-BE)

| path | motivo |
|---|---|
| `wire/wire.go` (alta de `ActorSet`/`ActorHandler`) | pertenece a **feature-007 actor-system**, no a 018. El bloque `DataIntegritySet` ya está registrado en develop. |
| `wire/actor_providers.go`, `wire/companion_providers.go`, etc. | otras features (007/012). |
| cualquier hunk no-`GetRaw*` de los 4 repos compartidos | pertenece a 001/003/009/supply/lot/project, NO a 018. |
