# feature-018 · data-integrity-admin · spec (BE / ponti-backend)

- **id**: feature-018
- **slug**: data-integrity-admin
- **nombre**: Data-integrity admin
- **tipo**: feature
- **merge**: coordinado (FULL-STACK)
- **repo**: Backend Go (ponti-backend) — `/home/pablocristo/Proyectos/pablo/ponti/core`
- **existe en BE**: SI (este paquete)
- **existe en FE**: SI (mismo `feature-018`; FE `pages/admin/data-integrity` + hook `useDatabase`)
- **SOURCE de extracción**: `develop-problematico~1` = SHA `777e5f6a` (NUNCA usar `develop-problematico`: su tip es un restore/vacío)
- **rama destino**: `develop` (tip `003a9b8f`)
- **rango fuente-de-verdad (diff)**: `0972e565..777e5f6a`

## Resumen

El módulo `internal/data-integrity` expone un endpoint admin que compara, por proyecto,
los valores que el **dashboard muestra al usuario** (SSOT, vistas `v4_report.* / v4_ssot.* /
v4_calc.*`) contra un **recálculo RAW** hecho directamente sobre las tablas base (`public.*`).
Si la diferencia supera la tolerancia → `Status = ERROR`; si coincide → `OK`.

En `777e5f6a` el módulo fue **reescrito y drásticamente simplificado**: `usecases.go` pasó de
**1068 → 343 líneas**. Se eliminaron los 17 controles "encadenados" (control1..control17 que
comparaban módulo-contra-módulo a partir de un `sharedData` gigante) y se reemplazaron por
**5 controles independientes** (SSOT-dashboard vs RAW-tabla-base), ejecutados en paralelo con
`sync.WaitGroup` + cancelación al primer error.

## Objetivo

- Detectar incoherencias de datos entre lo que ve el usuario (dashboard) y la realidad de las
  tablas base, por proyecto, en una sola llamada barata.
- Reducir la complejidad y el costo del control: el handler bajó el timeout de **8 min → 30 s**
  porque cada control hace solo 2 queries cortas en vez de hidratar un `sharedData` enorme.
- Reforzar el aislamiento multitenant: cada query RAW filtra por `tenant_id` vía
  `internal/shared/authz` (modo estricto opcional).

## Problema que resuelve

El diseño viejo (17 controles) era frágil y lento: comparaba derivaciones entre sí (lotes vs
informe de campo vs informe generales vs dashboard...), arrastraba estados intermedios y
warnings hardcodeados, y necesitaba 8 minutos de timeout. El nuevo diseño compara cada métrica
SSOT contra UN recálculo RAW genuinamente independiente, sin estados intermedios.

## Alcance en ESTE repo (BE) — solo lo que está en mi flist

Archivos PROPIOS del módulo (mi flist `be-018.txt`):

- `internal/data-integrity/usecases.go` — reescritura completa (1068→343). 5 controles + `buildCheck`.
- `internal/data-integrity/handler.go` — quita `GetProtected()` de `MiddlewaresEnginePort`; timeout 8m→30s; import `core/...domainerr` → `platform/...domainerr`.
- `internal/data-integrity/handler/dto/integrity_check.go` — agrega `CheckType`, `Severity`, `Recommendation` al DTO.
- `internal/data-integrity/usecases/domain/types.go` — agrega `CheckType/Severity/Recommendation`; **quita TODAS las json tags** del dominio (purity, ver 027); `Status` ahora documenta `OK, ERROR, WARNING, SKIPPED`.
- `internal/data-integrity/handler_test.go` — **NUEVO**: 2 tests de handler (`ParsesProjectID`, `RequiresProjectID`).
- `internal/data-integrity/usecases_tenant_test.go` — **NUEVO**: verifica propagación de tenant context a los 6 repos.
- `internal/data-integrity/usecases_test.go` — reescrito a la nueva API (6 tests).
- `internal/data-integrity/usecases_mock_test.go` — mocks gomock regenerados para los 6 ports nuevos.

## Alcance que está FUERA de mi flist pero ES REQUERIDO (compartido con otras features)

El refactor **NO compila** solo con mi flist. Depende de hunks que viven en archivos
asignados a OTRAS features:

- **`internal/shared/authz/authz.go`** (`TenantFromContext`, `TenantStrictModeEnabled`, `TenantMissing`) — NO existe en `develop`. Lo aportan las features de tenancy (001/003).
- **4 métodos RAW nuevos** en repos compartidos (NO existen en `develop`):
  - `internal/report/repository.go` → `GetRawNetIncome`
  - `internal/supply/repository.go` → `GetRawSupplyInvestment`
  - `internal/project/repository.go` → `GetRawAdminCostTotal`
  - `internal/lot/repository.go` → `GetRawLeaseExecuted`
  - (`internal/work-order/repository.go` → `GetRawDirectCost` SÍ existe en `develop`, pero **sin** filtro tenant).
- **Re-cableado wire**: `wire/data_integrity_providers.go` (cambia firma de `ProvideDataIntegrityUseCases`, elimina `StockRepositoryPort`, agrega `ProjectRepositoryPort`, pasa repos concretos `*report.ReportRepository`, `*supply.Repository`, `*project.Repository`, `*lot.Repository`) y `wire/wire_gen.go`.

## Alcance en el OTRO repo (FE)

- `pages/admin/data-integrity` (página admin del reporte de integridad).
- hook `useDatabase` (consume el endpoint `GET /data-integrity/costs-check?project_id=...`).
- El FE debe tolerar el nuevo shape de DTO: 3 campos extra (`check_type`, `severity`,
  `recommendation`) y `status` que ahora puede ser `OK | ERROR` (el dominio documenta también
  `WARNING/SKIPPED`, pero el `buildCheck` actual solo emite `OK/ERROR`).
- Detalle del FE en el paquete `feature-018` del repo FE (coordinar).

## Fuera de alcance (NO extraer aquí)

- **tentative-prices** (#121/#124): YA DONE. No toca ninguno de mis archivos (verificado: 0 referencias en el diff de `internal/data-integrity/`).
- **lot-metrics / total_tons** (#117/#121/#124): YA DONE.
- La limpieza de json-tags del dominio BE en general pertenece a **feature-027** (be-cleanup-domain-purity). Aquí solo aparece la limpieza puntual de `usecases/domain/types.go` de data-integrity.
- Migraciones / `.sql`: **ninguna** en el scope de data-integrity (verificado).
- El alta del módulo `actor`/`ActorSet` que aparece en `wire/wire.go` pertenece a **feature-007 (actor-system)**, no a 018.

## Comportamiento esperado

- `GET {APIBaseURL}/data-integrity/costs-check?project_id=<int>`
- Sin `project_id` → `400` (`missing required query param: project_id`).
- Con `project_id` válido → `200` con `{ "checks": [ ...5 IntegrityCheckDTO... ] }`.
- Cada check trae SSOT vs RAW, `difference_a`, `status` (`OK`/`ERROR`), `tolerance` (1 USD).
- Ruta montada en `r.Group(base, mws.GetValidation()...)` (grupo "public", sin `GetProtected`).

## Estado en `develop-problematico~1` (777e5f6a)

- Módulo `internal/data-integrity` **completo y coherente** internamente (compila contra el árbol de 777e5f6a, que ya tiene authz + RAW methods + wire nuevo).
- Tests del módulo presentes y verdes en ese árbol.
- **Sobre `develop` actual NO compila** sin traer las dependencias compartidas (authz + 4 RAW + wire).

## Criterios de aceptación

1. `go build ./...` y `go vet ./internal/data-integrity/...` OK tras traer dependencias compartidas.
2. `go test ./internal/data-integrity/...` verde (8 archivos de test, incl. tenant test).
3. `GET /data-integrity/costs-check?project_id=X` devuelve 5 checks; sin `project_id` → 400.
4. `wire` regenerado sin diff sucio (`go generate ./wire/...` idempotente).
5. FE `pages/admin/data-integrity` renderiza con el nuevo DTO (coordinar con repo FE).

## Endpoints / modelos / UI / DB / tests afectados

- **Endpoint**: `GET /data-integrity/costs-check` (query `project_id`). Sin cambio de ruta; cambia el shape de respuesta.
- **DTO** (`handler/dto`): `IntegrityCheckDTO` +`check_type`,+`severity`,+`recommendation`.
- **Dominio** (`usecases/domain`): `IntegrityCheck` +`CheckType/Severity/Recommendation`, sin json tags; `IntegrityReport` sin json tag.
- **Ports** (interfaces en `usecases.go`): `DashboardRepositoryPort.GetDashboard`, `WorkOrderRepositoryPort.GetRawDirectCost`, `ReportRepositoryPort.GetRawNetIncome`, `SupplyRepositoryPort.GetRawSupplyInvestment`, `ProjectRepositoryPort.GetRawAdminCostTotal`, `LotRepositoryPort.GetRawLeaseExecuted`. **Eliminado** `StockRepositoryPort`.
- **UI** (FE): página admin data-integrity + `useDatabase`.
- **DB**: sin migraciones; las queries RAW leen `public.workorders/workorder_items/labors/supplies/lots/fields/crop_commercializations/supply_movements/categories/projects` y dependen de las vistas SSOT `v4_*`.
- **Tests**: `handler_test.go`, `usecases_test.go`, `usecases_tenant_test.go`, `usecases_mock_test.go`.

## Dependencias

- **Intra-repo (fuertes)**: tenancy → `internal/shared/authz` (features 001/003); RAW methods en `report/supply/project/lot/work-order/repository.go`; re-cableado `wire/`.
- **Intra-repo (débil)**: `internal/dashboard/usecases/domain` (campos `DashboardBalanceSummary.*`) — ya presentes en `develop`.
- **Cross-repo (coordinado)**: FE `feature-018` debe alinearse con el nuevo DTO. BE-first recomendado.

## Riesgos

- **Funcional**: si el FE espera el shape viejo (sin `check_type/severity/recommendation`, 9/17 controles) puede romper render o lógica de severidad.
- **Técnico (alto)**: extracción incompleta → no compila por faltar `authz` + 4 RAW + wire (archivos fuera de mi flist).
- **Compartido**: `report/supply/project/lot/repository.go` y `wire/*` traen mucho cambio NO-018; requieren **partial-hunks**, no whole-file.

## DECISIÓN recomendada

**Arreglar antes de extraer (no extraer tal cual).** El módulo en sí es portable como
whole-file, pero su PR es inviable sin sus dependencias compartidas. Orden sugerido:
1) mergear primero la base de tenancy (`internal/shared/authz`) — features 001/003;
2) en el MISMO PR de 018-BE, traer por **partial-hunks** los 4 métodos RAW (+ tenantizar
`GetRawDirectCost`) y el re-cableado wire;
3) regenerar wire; 4) coordinar el merge con el FE (BE-first).
