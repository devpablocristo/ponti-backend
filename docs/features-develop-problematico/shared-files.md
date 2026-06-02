# shared-files.md — Archivos compartidos / peligrosos de develop-problematico (BE core)

> Análisis GLOBAL de descomposición. Fecha: **2026-05-30**.
> Repo: **ponti-backend / core** (`/home/pablocristo/Proyectos/pablo/ponti/core`).
> Rango fuente del diff: `0972e565..777e5f6a`.
> **SOURCE de extracción = `develop-problematico~1` (SHA `777e5f6a`)**, el pico, NUNCA el tip
> (el tip `develop-problematico` es un restore que vacía la rama). Destino = `develop`.
>
> Este doc lista **TODOS los archivos del repo tocados por más de una feature** y, para cada uno,
> qué parte pertenece a cada feature y cómo extraerlo. Para los `partial` se sugiere el comando
> `git restore -p --source=develop-problematico~1 -- <path>` (los comandos git son **sugerencias**;
> este doc no aplica cambios de código).
>
> Fuentes: los `file-list.md` de cada paquete (`feature-0NN-*/file-list.md`), sus secciones
> "Compartidos", y la lista conocida de archivos peligrosos del repo.

---

## Leyenda

- **whole-file**: el archivo es coherente con UNA feature aunque la toquen varias por referencia → traer entero (normalmente con `git checkout 777e5f6a -- <path>` desde su feature dueña).
- **partial-hunks**: el archivo MEZCLA hunks de >1 feature → traer solo los hunks de la feature en curso con `git restore -p --source=develop-problematico~1 -- <path>`.
- **manual-port / regenerar**: NO copiar a mano; reescribir o regenerar (caso típico: artefactos generados como `wire_gen.go`, `go.sum`, mocks).
- **do-not-extract-yet**: el hunk pertenece a una feature todavía no porteada; el build/runtime rompe si entra antes.

Convención de riesgo: **alto** = romper build/deploy/datos si se hace mal; **medio** = conflicto de merge o comportamiento; **bajo** = trivial.

---

## Tabla maestra de archivos compartidos

| Archivo | Features relacionadas | Tipo de mezcla | Estrategia de extracción | Riesgo | Notas |
|---|---|---|---|---|---|
| `wire/wire_gen.go` | 007, 008, 012, 013, 018, 023 | DI generado: cablea TODOS los módulos | **regenerar** (`wire ./wire` / `go generate ./wire`) tras portar los providers | **alto** | NUNCA editar/cherry-pick a mano. Es el archivo más mezclado del repo. |
| `wire/wire.go` | 007, 023 (013/018 quitan providers Excel/registran sets) | DI manual: declara `Dependencies` + `*Set` | **partial-hunks** | alto | Hunk de `ActorHandler` (007); el resto de `*Set` queda al ir aterrizando 012/013/018. |
| `cmd/api/main.go` | 002, 023, 024(solo doc-ref) | bootstrap: observability + lifecycle metrics | **partial-hunks** | medio | 83+/7-. Observability (023) + `lifecycle.RegisterMetrics(...)` (002). 024 solo lo menciona. |
| `cmd/api/http_server.go` | 005, 007, 023 | router/middlewares | **partial-hunks** | alto | 33+/7-. CORS/rate-limit/metrics (023) + `ActorHandler.Routes()` (007) + `reporting.read_mode` (005). |
| `cmd/config/loadconfig.go` | 005, 012(vía 005) | agregador de config | **partial-hunks** | medio | 4+/1-. Quita `AI`; agrega `Companion`/`Nexus`/`Reporting`/`Security`. Todo es 005. |
| `.env.example` | 005, 012, 019, 021 | doc de env | **partial-hunks** | alto | 44+/9-. AI→Companion/Nexus/Review (005/012) + bloque `DB_*_PROD` (019). |
| `Makefile` | 019, 020, 021, 023, 024 | targets de tooling/build | **partial-hunks** | alto | 31+/62-. db/stack (019), `lint` golangci v2.11.4 (020/022), `cmd`→`cmd/api` (021/023), `openapi` (024), renames `core/*`→`platform/*` (001). |
| `go.mod` | 002, 013, 021, +bumps DONE(#124) | grafo de deps | **manual-port** (`go mod tidy`) | alto | +obs/gorm/prometheus/otel (002/001/005), -excelize (013), -cloudsqlconn. Bumps go-jose/x/net YA en develop (#124) → excluir. |
| `go.sum` | 002, 013, 021, +bumps DONE(#124) | checksums | **manual-port** (derivado de go.mod) | alto | Nunca portar a mano; se regenera. |
| `.gitignore` | 019, 021 | ignores | **partial-hunks** | medio | Solo el hunk `scripts/db/schema.snapshot.sql` (019/021). NO revertir `go.work`/`/api` (vigente en develop). |
| `internal/shared/models/base.go` | 001, (002/007/008 por referencia) | modelo base | **partial-hunks** | bajo | 1+/1-. En este rango el ÚNICO cambio es el import contextkeys core→platform (001). |
| `internal/supply/repository_movement.go` | 002, 009, 013 | repo (560 líneas) | **do-not-extract-yet** (va con 002) | alto | f009=4 (Archive/Restore movement); el grueso es lifecycle (002) + csvexport (013). |
| `internal/supply/mocks/mock_repository.go` | 009, 013 | mock generado | **regenerar** mock | alto | Renombres HardDelete*/Archive*/Restore* (009) + `MockExporterAdapterPort` (013). |
| `wire/admin_providers.go` | 008, 018, 023 | DI providers admin | **partial-hunks** | medio | `ProvideAdminRepository`/`ProvideAdminUseCases` (008/admin de 018). |
| `wire/ai_providers.go` | 012, 008, 023 | DI providers AI | **manual-port / partial-hunks** | alto | `ai.Client`→`axis.CompanionClient`+adapter (012) + tenant scope AI (008). |
| `wire/config_providers.go` | 005, 012, 023 | DI providers config | **partial-hunks** | medio | `ProvideConfigAI`→`ProvideConfigCompanion` (012) + config 005. |
| `wire/middleware_providers.go` | 008, 001, 023 | DI providers mw | **partial-hunks** | medio | Quita `GetProtected()`; agrega `RequireTenantHeader`/`Environment` (008). |
| `wire/data_integrity_providers.go` | 018, 023 | DI providers DI-admin | **partial-hunks** | alto | Recablea use-cases (+project +supply, -stock), repos concretos. Firma debe coincidir con `dataintegrity.NewUseCases` (018). |
| `wire/lot_providers.go` | 013, 023 | DI providers export | **partial-hunks (casi whole)** | medio | Excel→CSV (`NewCSVExporter`/`ProvideLotExporterPort`). Hunk limpio a 013. |
| `wire/supply_providers.go` | 013, 023 | DI providers export | **partial-hunks (casi whole)** | medio | Excel→CSV. Hunk limpio a 013. |
| `wire/stock_providers.go` | 013, 023 | DI providers export | **partial-hunks (casi whole)** | medio | Excel→CSV. Hunk limpio a 013. |
| `wire/work_order_providers.go` | 013, 023 | DI providers export | **partial-hunks (casi whole)** | medio | Excel→CSV (`ProvideWorkOrderExporterPort`). |
| `wire/labor_providers.go` | 013, 023 | DI providers export | **partial-hunks (casi whole)** | medio | Excel→CSV. |
| `internal/lot/repository.go` | 001, 002, 009, 013(noise), 018 | repo de dominio | **partial-hunks** | alto | tenancy.Scope×28 (001) + Archive/Restore (002/009) + `GetRawLeaseExecuted` (018). lot-metrics ya DONE. |
| `internal/supply/repository.go` | 001, 002, 009, 013, 018 | repo de dominio | **partial-hunks** | alto | tenancy.Scope×27 (001) + lifecycle (002/009) + `GetRawSupplyInvestment` (018) + csv (013). |
| `internal/work-order/repository.go` | 001, 002, 009, 018 | repo de dominio | **partial-hunks** | alto | tenancy.Scope×22 (001) + lifecycle/actor-sync (002/009) + `GetRawDirectCost` tenantizado (018, opcional). |
| `internal/project/repository.go` | 001, 002, 007, 009, 010, 018 | repo de dominio (gigante) | **whole-file (de 010)** | alto | +1642/-431. Como archivo es de 010 (entero); arrastra tenancy(001)/actor-sync(007)/lifecycle(009)/`GetRawAdminCostTotal`(018). |
| `internal/report/repository.go` | 001, 013/027(noise), 018 | repo SQL report | **partial-hunks** | alto | +133/-85. `GetRawNetIncome` (018) + import-swap/authz (001) + report-cleanup (027/013). 0 tenancy.Scope. |
| `internal/customer/repository.go` | 001, 002, 007, 009 | repo de dominio | **partial-hunks** | alto | tenancy.Scope×10 (001) + `EnsureCustomerFromActor` (007) + `RunCascadeArchive`/hard-delete (002/009). |
| `internal/field/repository.go` | 001, 002, 009 | repo de dominio | **partial-hunks** | alto | tenancy.Scope×14 (001) + lifecycle + field_investors. |
| `internal/work-order-draft/repository.go` | 001, 002, 009 | repo de dominio | **partial-hunks** | alto | tenancy.Scope×16 (001) + lifecycle/actor-sync. |
| `internal/stock/repository.go` | 001, 002, 009, 013 | repo de dominio | **partial-hunks** | alto | tenancy.Scope×15 (001) + lifecycle + filtros/csv. |
| `internal/provider/repository.go` | 001, 002, 007, 009 | repo de dominio | **partial-hunks** | alto | tenancy.Scope×13 (001) + posible actor-sync (provider→actor). |
| `internal/manager/repository.go` | 001, 002, 007, 009 | repo de dominio | **partial-hunks** | alto | tenancy.Scope×12 (001) + posible actor-sync (manager→actor). |
| `internal/investor/repository.go` | 001, 002, 007, 009 | repo de dominio | **partial-hunks** | alto | tenancy.Scope×11 (001) + posible actor-sync (investor→actor). |
| `internal/business-parameters/repository.go` | 001, 002, 009 | repo de dominio | **partial-hunks** | medio | tenancy.Scope×12 (001) + lifecycle. |
| `internal/category/repository.go` | 001, 002, 009 | repo de dominio | **partial-hunks** | medio | tenancy.Scope×12 (001) + lifecycle. |
| `internal/crop/repository.go` | 001, 002, 009 | repo de dominio | **partial-hunks** | medio | tenancy.Scope×12 (001) + lifecycle. |
| `internal/lease-type/repository.go` | 001, 002, 009 | repo de dominio | **partial-hunks** | medio | tenancy.Scope×12 (001) + lifecycle. |
| `internal/invoice/repository.go` | 001, 002, 009 | repo de dominio | **partial-hunks** | medio | tenancy.Scope×5 (001) + lifecycle. `DeleteInvoice` (clave compuesta) NO es CRUDAR. |
| `internal/dollar/repository.go` | 001, 002, 009 | repo de dominio | **partial-hunks** | medio | tenancy.Scope×4 (001) + lifecycle. |
| `internal/commercialization/repository.go` | 001, 002, 009 | repo de dominio | **partial-hunks** | medio | tenancy.Scope×3 (001) + lifecycle. |
| `internal/class-type/repository.go` | 001, 002, 009 | repo de dominio | **partial-hunks** | bajo | tenancy.Scope×1 (001) + lifecycle. |
| `internal/businessinsights/repository.go` | 001, (027/013 dudoso) | repo SQL | **partial-hunks** | medio | import-swap + ctx threading; 0 tenancy.Scope. Revisar si arrastra report-cleanup. |
| `internal/dashboard/repository.go` | 001, 013/027(dudoso) | repo SQL (145 líneas) | **partial-hunks** | alto | import-swap + ctx; 0 tenancy.Scope. Diff grande sin scoping → posible solape 027/013. |
| `internal/lot/usecases.go` | 009, 013, lot-metrics(DONE) | usecases | **partial-hunks** | alto | Archive/Restore/ListArchived (009) + rename excel→exporter (013) + `GetMetrics`/`UpdateLotTons` (DONE→excluir). |
| `internal/lot/handler.go` | 009, 013, lot-metrics(DONE) | handler | **partial-hunks** | alto | rutas CRUDAR (009) + export CSV (013) + `UpdateLotTons`/`RespondNoContent` (DONE/otras). |
| `internal/labor/usecases.go` | 009, 013 | usecases | **partial-hunks** | alto | HardDelete/Archive/Restore (009) + rename + `ExportTable` (013). |
| `internal/labor/handler.go` | 009, 013 | handler | **partial-hunks** | alto | dual-group CRUDAR (009) + 3 endpoints export→CSV (013). |
| `internal/stock/usecases.go` | 009(noise), 013 | usecases | **partial-hunks** | alto | `u.exporter.Export` (013) + filtros/dominio. |
| `internal/stock/handler.go` | 009(noise), 013 | handler | **partial-hunks** | alto | export→CSV (013) + platform/csv. |
| `internal/supply/usecases.go` | 009, 013 | usecases | **partial-hunks** | alto | HardDelete/Archive/Restore/ListArchived (009) + `ExportSupplies` (013). |
| `internal/supply/usecases_movement.go` | 009, 013 | usecases | **partial-hunks** | medio | Archive/Restore/HardDelete movement (009) + `ExportSupplyMovements` (013). |
| `internal/supply/handler.go` | 009, 013 | handler | **partial-hunks** | alto | CRUDAR supplies/movements/stock-movements (009) + rutas `/export*`→CSV (013). |
| `internal/work-order/usecases.go` | 009, 013 | usecases | **partial-hunks** | medio | HardDelete + archive/restore (009) + `u.exporter.Export` (013). |
| `internal/work-order/handler.go` | 009, 013, 025(test aparte) | handler | **partial-hunks** | medio | rutas CRUDAR (009) + `/export`→CSV (013). |
| `internal/lot/handler_export_test.go` | 009, 013 | test | **partial-hunks** | medio | mezcla export(013) + CRUDAR(009). |
| `internal/labor/handler_update_labor_test.go` | 009, 013 | test | **partial-hunks** | medio | ajuste rutas labor (009) + stub `ExportTable` (013). |
| `internal/supply/handler_update_supply_test.go` | 009, 013 | test | **partial-hunks** | medio | CRUDAR (009) + stubs export (013). |
| `internal/work-order/handler_test.go` | 013, 025 | test | **partial-hunks** | alto | ajuste export (013); en 025 figura como M (rename `DeleteWorkOrderByID`→`HardDelete`, +`/archive`,`/restore`,`/hard`). |
| `internal/customer/repository_harddelete_test.go` | 009, 025 | test | **partial-hunks** | medio | rename Delete→HardDelete; toca repo (002). |
| `internal/supply/repository_delete_test.go` | 009 | test | **partial-hunks** | bajo | f009=1 ajuste naming. |
| `internal/supply/repository_movement_delete_test.go` | 009 | test | **partial-hunks** | medio | movement delete/archive; toca repo (002). |
| `README.md` | 024, 019, 021, 001 | doc | **partial-hunks** | alto | renames `frontend`→`web`, `core/*`→`platform/*`, `staging-db-2-local-db`→`reset-local-db-from-prod`. Decidir si van con 024 o 019/021. |
| `docs/README.md` | 024, 019/021 | doc | **manual-port** | medio | 1 línea: `make migrate-up`→`make db-migrate-up` (rename de target = 019/021). |
| `docs/ARCHITECTURE.md` | 024, 001(observabilidad) | doc | **partial-hunks** | medio | grueso del diff es contenido nuevo de 024; verificar que develop no haya tocado el TLDR. |

---

## Detalle por archivo

### `wire/wire_gen.go` — 007 / 008 / 012 / 013 / 018 / 023
- **Mezcla**: archivo GENERADO por `wire`. Materializa el grafo completo: `ProvideActorHandler` (007), cableado `adminRepository→adminUseCases→adminHandler` (008), nueva firma `ProvideAIHandler` con CompanionAdapter (012), quita `...ExcelService`/`XLSXEnginePort` (013), recablea data-integrity (018), y firma `runHTTPServer`/observability (023).
- **Estrategia**: **regenerar**, no cherry-pickear. Tras portar los `*_providers.go` de la feature en curso, correr `wire ./wire` (o `go generate ./wire`).
- **Riesgo**: alto — parchear a mano rompe todo el build.
- Sugerencia (NO copiar el contenido): `git restore -p` está **desaconsejado** aquí; usar regeneración.

### `wire/wire.go` — 007 (+ 013/018 quitan/registran sets) / 023
- **Mezcla**: declara `Dependencies` y los `*Set`. Hunk de 007 = `ActorHandler *actor.Handler` + `ActorSet`. 013 quita providers Excel del set; 018 ya tiene `DataIntegritySet` registrado en develop.
- **Estrategia**: **partial-hunks** — traer el hunk del actor con 007; el resto al aterrizar cada feature dueña.
- Sugerencia: `git restore -p --source=develop-problematico~1 -- wire/wire.go`

### `cmd/api/main.go` — 002 / 023 (024 solo lo referencia)
- **Mezcla**: 83+/7-. Bloque observability (slog + metrics + tracer + swagger) y cambio de firma `runHTTPServer` = **023**. Línea `lifecycle.RegisterMetrics(metrics.Registry(),"ponti_backend")` = **002**. 024 (docs/OBSERVABILITY.md) solo lo menciona, no lo toca.
- **Estrategia**: **partial-hunks**. Traer la línea de `lifecycle.RegisterMetrics` con 002; el resto del bloque observability con 023.
- Sugerencia: `git restore -p --source=develop-problematico~1 -- cmd/api/main.go`

### `cmd/api/http_server.go` — 005 / 007 / 023
- **Mezcla**: 33+/7-. CORS / rate-limit / `/observability/metrics` / middlewares globales = **023**. `deps.ActorHandler.Routes()` = **007** (una línea entre muchos handlers). `reporting.read_mode` = **005** (sección Reporting de config).
- **Estrategia**: **partial-hunks**. La línea `ActorHandler.Routes()` va con 007; CORS/rate-limit con 023.
- Sugerencia: `git restore -p --source=develop-problematico~1 -- cmd/api/http_server.go`

### `cmd/config/loadconfig.go` — 005 (+ 012 vía 005)
- **Mezcla**: 4+/1-. Quita campo `AI`; agrega `Companion`, `Nexus`, `Reporting`, `Security` al struct `Config`. Todo el hunk pertenece a **005** (es su agregador). 012 lo necesita como prerequisito pero no lo edita.
- **Estrategia**: **partial-hunks** (el archivo agrega varias secciones; aquí el bloque de campos de config). Si 005 va entero primero, se puede traer este hunk con 005 sin partir.
- Sugerencia: `git restore -p --source=develop-problematico~1 -- cmd/config/loadconfig.go`

### `.env.example` — 005 / 012 / 019 / 021
- **Mezcla**: 44+/9-.
  - **SÍ 005/012**: bloque AI→`COMPANION_*`/`NEXUS_*`, `REVIEW_*`, header de uso.
  - **NO aquí, es 019**: bloque `# PROD data source for local DB reset` (`DB_NAME_PROD`, `DB_USER_PROD`, `CLOUDSQL_PROJECT_PROD`, `DB_INSTANCE_NAME_PROD`, `SRC_INSTANCE_*`, `SRC_PASS_SECRET_*`) → lo consume `scripts/db/reset-local-db-from-prod.sh`.
  - `CORS_ORIGINS`/`HTTP_RATE_LIMIT_PER_MINUTE` (021/023) pueden no estar documentadas; opcional.
- **Estrategia**: **partial-hunks** por feature.
- Sugerencia: `git restore -p --source=develop-problematico~1 -- .env.example`

### `Makefile` — 019 / 020 / 021 / 023 / 024 (+ 001 renames)
- **Mezcla**: 31+/62-.
  - **019**: `reset-local-db-from-prod` (reemplaza `db-staging-to-local`/`db-reset-from-staging`), `actors-backfill-sync` (reemplaza `staging-db-2-dev-db`), `up/down-ponti-local`, borrado de targets GCP peligrosos (`db-force-reset-gcp`, `db-gcp-reset-and-load-local`), `.PHONY`.
  - **024**: target `openapi:` (swag).
  - **020/022**: `lint` → `golangci-lint v2.11.4`.
  - **021/023**: `bin-build`/`run` `cmd/`→`cmd/api`.
  - **005**: borrado de `select-ponti-stg-local`/`dev-local`, `seed`, `seed-dashboard`.
  - **001**: reemplazos de texto `core/*`→`platform/*` en warnings.
- **Estrategia**: **partial-hunks** por feature.
- Sugerencia: `git restore -p --source=develop-problematico~1 -- Makefile`

### `go.mod` / `go.sum` — 002 / 013 / 021 (+ bumps DONE #124)
- **Mezcla**: agrega `prometheus/client_golang`, `platform/observability/go`, `platform/persistence/gorm/go`, otel exporters (motivados por 001/002/005); quita `xuri/excelize/v2` (013) y `cloudsqlconn`. Los bumps `go-jose/v4` y `x/net` YA están en develop (#124) → **excluir**.
- **Estrategia**: **manual-port** — NUNCA cherry-pickear líneas a mano. Regenerar con `go mod tidy` al portar el código que las motiva. `go.sum` se deriva.
- **Riesgo**: alto (build).

### `.gitignore` — 019 / 021
- **Mezcla**: 3+/6-. Hunk útil = agregar `scripts/db/schema.snapshot.sql` (019, artefacto generado de 6589 líneas). 021: NO portar el revert de `go.work`/`go.work.sum`/`/api`/`scripts/db/*.env` (siguen vigentes en develop como tooling local).
- **Estrategia**: **partial-hunks** — solo `schema.snapshot.sql`.
- Sugerencia: `git restore -p --source=develop-problematico~1 -- .gitignore`

### `internal/shared/models/base.go` — 001 (+ 002/007/008 por referencia)
- **Mezcla**: 1+/1-. En ESTE rango el único cambio es el import `contextkeys` core→platform (**001**). Es "compartido" en intención porque `Base`/`ActorFromContext` lo consumen 002/007/008, pero el diff del archivo es 100% 001.
- **Estrategia**: **partial-hunks** (de hecho un solo hunk de import). Traer con 001.
- Sugerencia: `git restore -p --source=develop-problematico~1 -- internal/shared/models/base.go`

### `internal/supply/repository_movement.go` — 002 / 009 / 013
- **Mezcla**: 560 líneas; f009=4 (hunks Archive/Restore movement). El grueso es lifecycle (002) + csvexport/import (013).
- **Estrategia**: **do-not-extract-yet** — viaja con la implementación de repo de **002**; los hunks 009 dependen de que 002 esté en develop.

### `internal/supply/mocks/mock_repository.go` — 009 / 013
- **Mezcla**: renombres `HardDelete*`/`Archive*`/`Restore*` (009) + `MockExporterAdapterPort` con `ExportSupplies`/`ExportSupplyMovements` (013).
- **Estrategia**: **regenerar** el mock (mejor que partial-hunks) una vez que las interfaces de 009 y 013 estén en el repo.

### wire `*_providers.go` (admin / ai / config / middleware / data_integrity)
- **`wire/admin_providers.go`** — 008 (`ProvideAdminRepository`/`ProvideAdminUseCases`, firma `ProvideAdminHandler`) / 018(admin) / 023. `partial-hunks`.
- **`wire/ai_providers.go`** — 012 (`ai.Client`→`axis.CompanionClient`+adapter, +Nexus) / 008 (tenant scope AI) / 023. En develop aún wirea `ai.Client` legacy → **manual-port** (reescribir) o partial-hunks gated por 012.
- **`wire/config_providers.go`** — 012 (`ProvideConfigAI`→`ProvideConfigCompanion`) / 005 / 023. `partial-hunks`.
- **`wire/middleware_providers.go`** — 008 (quita `GetProtected()`, agrega `RequireTenantHeader`/`Environment`) / 001 / 023. `partial-hunks`.
- **`wire/data_integrity_providers.go`** — 018 (recablea use-cases: +project +supply, -stock; repos concretos) / 023. La firma DEBE coincidir EXACTO con `dataintegrity.NewUseCases` de 018. `partial-hunks`.
- Sugerencia genérica: `git restore -p --source=develop-problematico~1 -- wire/<provider>.go`

### wire `*_providers.go` de export (lot / supply / stock / work_order / labor) — 013 / 023
- **Mezcla**: cada uno quita providers Excel y agrega el provider CSV (`NewCSVExporter`, `ProvideLotExporterPort`, `ProvideWorkOrderExporterPort`, etc.). El hunk es **limpio y dedicado a 013**.
- **Estrategia**: **partial-hunks (casi whole)** — prácticamente todo el diff del archivo es de 013.
- Sugerencia: `git restore -p --source=develop-problematico~1 -- wire/lot_providers.go wire/supply_providers.go wire/stock_providers.go wire/work_order_providers.go wire/labor_providers.go`

### `internal/<dominio>/repository.go` (los ~23 repos de dominio) — 001 / 002 / 007 / 009 / 013 / 018
- **Mezcla**: este es el patrón más recurrente del repo. Cada `repository.go` apila:
  - **001**: import-swap core→platform + `tenancy.Scope(ctx, db, "<tabla>")` (masivo; ver conteos por archivo en la tabla maestra).
  - **002 / 009**: helpers `Archive*`/`Restore*`/`HardDelete*` + `RunCascadeArchive`/`assert*ReferencesActive` (la implementación real de Archive/Restore/HardDelete la lleva 002; la superficie CRUDAR es 009).
  - **007**: actor-sync (`actorsync.*`, `legacy_actor_map`, `EnsureCustomerFromActor`) en customer/provider/manager/investor/project.
  - **018**: métodos `GetRaw*` (`GetRawNetIncome` en report, `GetRawSupplyInvestment` en supply, `GetRawAdminCostTotal` en project, `GetRawLeaseExecuted` en lot, `GetRawDirectCost` en work-order).
  - **013**: csvexport/filtros en stock/supply/lot.
- **Estrategia**: **partial-hunks** por feature, EXCEPTO `internal/project/repository.go` que se trae **whole-file con 010** (es coherente como archivo del módulo project, aunque internamente mezcla todo).
- **Orden**: portar 001 (tenancy) + 002 (lifecycle) antes; luego 007/009/018 agregan sus hunks. Sin las deps, no compila (faltan `internal/actor`, `internal/shared/lifecycle`, `internal/shared/authz`).
- Sugerencia (por entidad): `git restore -p --source=develop-problematico~1 -- internal/lot/repository.go` (repetir por dominio).

### `internal/<dominio>/handler.go` y `usecases.go` (lot / labor / stock / supply / work-order) — 009 / 013 (+ lot-metrics DONE)
- **Mezcla**: superficie CRUDAR (rutas `/:id/archive`, `/:id/restore`, `DELETE /:id/hard`, `GET /archived`, `runXIDAction`, `HardDeleteX`) = **009**; rename `excel`→`exporter` y endpoints export que pasan de XLSX a CSV (Content-Type `text/csv`, `c.Data`) = **013**. En lot, además hunks `UpdateLotTons`/`total_tons`/`GetMetrics` que YA están DONE (#117/#121/#124) → **excluir**.
- **Estrategia**: **partial-hunks**. Separar CRUDAR(009) de export(013) y excluir lot-metrics(DONE).
- Sugerencia: `git restore -p --source=develop-problematico~1 -- internal/lot/handler.go internal/lot/usecases.go`

### Tests compartidos (009 / 013 / 025)
- `internal/work-order/handler_test.go` — **013** (ajuste export) + **025** (M: rename `DeleteWorkOrderByID`→`HardDeleteWorkOrder`, `+ListArchivedWorkOrders`, rutas `/archive`,`/restore`,`/hard`). Riesgo alto. `partial-hunks`.
- `internal/lot/handler_export_test.go` — 009 (CRUDAR) + 013 (stub `"csv"`, Content-Type). `partial-hunks`.
- `internal/labor/handler_update_labor_test.go` — 009 (rutas) + 013 (`ExportTable`). `partial-hunks`.
- `internal/supply/handler_update_supply_test.go` — 009 (CRUDAR) + 013 (`ExportTableSupplies`/`ExportSupplyMovementsByProjectID`). `partial-hunks`.
- `internal/customer/repository_harddelete_test.go` — 009 (rename Delete→HardDelete) toca repo de 002; figura en 025. `partial-hunks`.
- **Estrategia**: cada test sigue a la última de sus features dueñas; no mergear antes de que la producción correspondiente (002/009/013) esté en develop (los tests son white-box `package <modulo>`).

### Docs compartidos (024 / 019 / 021 / 001)
- `README.md` — renames de tooling (`frontend`→`web`, `core/*`→`platform/*`, `staging-db-2-local-db`→`reset-local-db-from-prod`, `ponti-frontend/api`→`web/api`). Territorio de 019(db scripts)/021(deploy)/001(platform). En develop el README aún dice `staging-db-2-local-db`. **partial-hunks**; si 019/021 ya renombró, descartar esos hunks con `git restore -p`.
- `docs/README.md` — 1 línea `make migrate-up`→`make db-migrate-up` (rename de target Makefile = 019/021). **manual-port**.
- `docs/ARCHITECTURE.md` — grueso del diff es contenido nuevo de 024 (layout, observabilidad, lifecycle, errores). Verificar que develop no haya tocado el TLDR. **partial-hunks**.
- Sugerencia: `git restore -p --source=develop-problematico~1 -- README.md docs/README.md docs/ARCHITECTURE.md`

---

## "Compartidos lógicos" (NO comparten hunks, pero coordinan)

Archivos/recursos NUEVOS (un solo dueño) cuya LÓGICA referencia tablas/símbolos de otras features.
No requieren partial-hunks, pero el orden de merge importa.

| Recurso | Dueño | Coordina con | Nota |
|---|---|---|---|
| `internal/actor/legacy_sync.go`, `internal/actor/master_link.go` | 007 | 010, 011, 018 | Backfill SQL referencia tablas `customers`/`projects`/`investors`/`managers`/`providers`. Riesgo en coexistencia de tablas, no en el diff. |
| `migrations_v4/000224_tenant_security_foundation.*` | 003 | 008 (crea `tenant_invites`, rol `tenant_owner`), 001 (seed roles/permisos) | Seed protegido por `ON CONFLICT DO NOTHING` (idempotente). 008 lo necesita en runtime. |
| `migrations_v4/000226_customer_actor_master_link.*` | 007 | 003 (usa `tenant_id`), 010 (usa `projects`) | Depende de 224. |
| `migrations_v4/000227/228` (CRUDAR archive metadata) | 002 | 003 (deleted_at/FK), 009 | ORDEN: quedan bajo 229/230 ya en develop (#117/#121). Ver risks de 002. |
| `migrations_v4/000231/234` (actor archived/unique) | 007 | 003, 027, 009 | Referencian `tenant_id`; 234 hace merge destructivo de duplicados. |
| `cmd/config/companion.go`, `security.go`, `reporting.go` | 005 | 012 (Companion/Nexus/AITenantScope), 023 | `companion_providers.go` (012) no compila sin estos. |
| `internal/shared/authz/authz.go` | 001 (con 003) | 007, 008, 012, 018 | Base de tenancy; muchos consumidores. NO existe en develop → bloqueante de compilación de las demás. |
| `internal/shared/lifecycle/*` | 002 | 009, 010, 023 (`cmd/archive-cleanup`) | Funda toda la superficie CRUDAR. |
| `migrations_v4/000223_actors_safe_migration.*` | 007 | 003, 010 | Hueco previo en develop (salta 222→229). |

---

## Borrados que tocan a >1 feature

| Archivo (D) | Dueño | Coordina | Nota |
|---|---|---|---|
| `internal/platform/files/excel/excelize/{bootstrap,config,service}.go` | 013 | 001 (lo lista como do-not-extract-yet) | Engine XLSX de plataforma; borrar SOLO tras quitar el wire Excel (013). |
| `internal/*/excel-service.go` + `internal/*/excel/*` (labor/lot/stock/supply/work-order) | 013 | 009 (handlers) | Reemplazados por `csv-service.go`. |
| `internal/shared/utils/jwt_tools.go`, `require_jwt.go` | 001 | 007/008/027 (la nota de 027 los menciona) | JWT casero reemplazado por JWKS platform. Verificar callers vivos antes de borrar. |
| `cmd/config/ai.go` | 005 | 012 | Struct `AI` legacy (ponti-ai deprecado). |
| `internal/ai/client.go` | 012 | 005 | Cliente ponti-ai legacy; cutover a CompanionAdapter. |
| `scripts/db/schema.snapshot.sql`, `repair_stocks_investor_granularity.sql` | 019 | 021 (.gitignore) | Artefactos/one-shots; replicar el `git rm`. |

---

## Notas de seguridad de extracción

1. **Nunca** usar el tip `develop-problematico` como SOURCE: es un restore vacío. Siempre `develop-problematico~1` (`777e5f6a`).
2. **Generados** (`wire_gen.go`, `go.sum`, mocks `mock_repository.go`): **regenerar**, no `restore -p`.
3. **Orden de plataforma**: 001 (tenancy/authz) + 005 (config) + 002 (lifecycle) son fundacionales; casi todos los `partial-hunks` de repos/handlers dependen de que estén en develop, o no compilan.
4. **DONE — excluir de los hunks**: lot-metrics/`total_tons`/`GetMetrics`, tentative-prices (#117/#121/#124), bumps `go-jose`/`x/net` (#124). Aparecen mezclados en `lot/*`, `go.mod`/`go.sum` y data-integrity; no re-introducirlos.
5. Los comandos `git restore -p` de este doc son **sugerencias**; revisar cada hunk antes de stage.
