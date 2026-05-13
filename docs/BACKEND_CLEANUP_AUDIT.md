# Backend Cleanup Audit

Fecha: 2026-05-12

Este documento es la base versionada para limpiar el backend de Ponti sin hacer un
refactor masivo ni cambiar contratos publicos. La regla de trabajo es avanzar por
etapas chicas, validar despues de cada etapa y no tocar calculos sensibles sin
golden master.

## Estado de linea base

Comandos verificados antes de iniciar la limpieza:

- `go test ./...`: verde
- `go build ./cmd/api ./cmd/migrate`: verde
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.11.4 run --timeout=5m`: verde

Inventario actual:

- 403 archivos Go
- 41 archivos de test
- 56 migraciones `up`
- 210 referencias a `authz.MaybeTenantScope`
- 166 usos de SQL raw/exec
- 318 menciones a `legacy`, `deprecated`, `TODO`, `FIXME` o `HACK`

Hotspots por tamano/riesgo:

- `internal/project/repository.go`
- `internal/supply/handler.go`
- `internal/data-integrity/usecases.go`
- `internal/labor/repository.go`
- `internal/actor/repository.go`
- `internal/work-order/repository.go`
- `internal/dashboard/handler/dto/dashboard.go`
- `internal/stock/repository.go`
- `internal/report/repository.go`

## Diagnostico arquitectonico

El backend es un monolito Go modular por dominio bajo `internal/`. La estructura
principal es Gin para HTTP, GORM para persistencia CRUD, SQL directo para
reportes/calculos, usecases por modulo, DTOs HTTP por handler y modelos de
persistencia por repository.

Lo que esta bien:

- Hay separacion general entre handler, usecase, repository, domain y DTOs.
- Los errores estan mayormente normalizados via `domainerr` y `sharedhandlers`.
- Los CRUDs principales ya tienen acciones explicitas `archive`, `restore` y
  `hard`.
- Hay base multi-tenant con `RequireTenant`, `TenantScope`,
  `MaybeTenantScope` y migraciones `000224`/`000225`.
- `MaybeTenantScope` conserva compatibilidad transicional, pero ya falla
  cerrado cuando `TENANT_STRICT_MODE=true` y no hay tenant en contexto.
- CI ejecuta lint, build, tests y govulncheck.
- Existen scripts de schema guardrails, tenant audit y golden master.

Problemas estructurales:

- `/me/context` esta implementado inline en `cmd/api/http_server.go`; deberia
  vivir en un modulo auth/admin con repository/usecase propio.
- `MaybeTenantScope` esta muy extendido en paths tenant-owned. Es correcto como
  transicion y ahora queda cubierto por strict mode, pero el estado final debe
  migrar los paths tenant-owned a `RequireTenant`/`TenantScope` explicitos.
- Hay handlers grandes con demasiada logica operativa, especialmente
  `internal/supply/handler.go` y export/import handlers.
- Hay endpoints `DELETE /:id` legacy con semantica inconsistente: en algunas
  entidades archivan y en otras hacen hard delete.
- Varias interfaces triviales se repiten en todos los handlers
  (`GinEnginePort`, `ConfigAPIPort`, `MiddlewaresEnginePort`) y generan ruido.
- Hay SQL raw valido para reportes, pero no siempre queda documentado por que la
  query debe ser SQL directo.

## Matriz inicial por modulo

| Modulo | Tests directos | MaybeTenantScope | Raw/Exec | Riesgo inicial |
| --- | ---: | ---: | ---: | --- |
| actor | 1 | 0 | 39 | Alto: merge/deduplicacion y SQL manual |
| admin | 1 | 0 | 1 | Medio: auth/RBAC y contexto usuario |
| business-parameters | 2 | 8 | 0 | Bajo: CRUD tenant-owned de parametros |
| businessinsights | 2 | 0 | 0 | Medio: tenant/user context para insights |
| campaign | 1 | 13 | 0 | Medio: CRUD tenant-owned sin tests directos |
| category | 2 | 5 | 0 | Bajo: catalogo tenant-owned simple |
| class-type | 2 | 6 | 0 | Bajo: catalogo tenant-owned simple |
| commercialization | 2 | 3 | 0 | Bajo: bulk tenant-owned por proyecto |
| crop | 2 | 5 | 0 | Bajo: catalogo tenant-owned simple |
| customer | 1 | 10 | 1 | Medio: CRUD base, delete legacy hard |
| dashboard | 5 | 0 | 0 | Alto: calculos sensibles; resolucion tenant cubierta, requiere golden master para cambios de formula |
| data-integrity | 4 | 0 | 0 | Alto: validaciones sensibles; contrato de contexto tenant cubierto |
| dollar | 2 | 3 | 0 | Bajo: bulk tenant-owned por proyecto |
| field | 1 | 11 | 0 | Medio: CRUD tenant-owned sin tests directos |
| investor | 1 | 11 | 0 | Medio: mirrors actors y relaciones |
| invoice | 2 | 5 | 7 | Medio: documentos y snapshots |
| labor | 2 | 15 | 8 | Alto: dos rutas semanticas y dependencias con OT |
| lease-type | 2 | 5 | 0 | Bajo: catalogo tenant-owned simple |
| lot | 4 | 25 | 4 | Alto: fechas, hectareas y hard-delete con referencias |
| manager | 1 | 11 | 0 | Medio: mirrors actors y relaciones |
| project | 2 | 13 | 32 | Alto: entidad central y relaciones pivots |
| provider | 2 | 1 | 0 | Medio: catalogo derivado tenant-scoped, no CRUDAR propio por ahora |
| report | 5 | 0 | 5 | Alto: calculos sensibles; resolucion tenant cubierta, requiere golden master para cambios de formula |
| stock | 4 | 15 | 43 | Alto: stock real/calculado y movimientos |
| supply | 7 | 48 | 4 | Alto: handler grande, imports y movimientos |
| work-order | 4 | 18 | 5 | Alto: OT, splits, facturas y stock |
| work-order-draft | 2 | 8 | 0 | Medio: publicacion digital |

Los modulos de catalogo/simple y bulk por proyecto (`category`, `class-type`,
`crop`, `lease-type`, `dollar`, `business-parameters`, `commercialization`) ya
estan alineados con las migraciones tenant-owned: modelos con `tenant_id`,
escritura desde contexto, scoping en list/get/update/delete y tests IDOR/strict
dedicados. `dollar` y `commercialization` ademas validan que el `project_id`
pertenezca al tenant antes de leer/escribir, para preservar el modelo
project-owned sin duplicar reglas de negocio.

## Codigo muerto o sospechoso

Se elimino solo codigo muerto confirmado por `rg`, wiring, tests y revision de
contratos. No se eliminaron endpoints ni DTOs publicos.

Candidatos restantes a revisar antes de eliminar:

- No quedan `//nolint:unused` en `internal`, `cmd` ni `wire`.

Decision: cualquier eliminacion adicional requiere confirmar ausencia de
callers, impacto publico y validacion completa.

## Codigo duplicado e inconsistencias

Duplicacion detectada:

- Interfaces de handler repetidas por modulo para engine/config/middlewares.
- Parseo manual de IDs y parametros en handlers, aunque ya existen helpers en
  `internal/shared/handlers`.
- Patron CRUD repetido con pequenas diferencias de status codes y nombres.
- Logica de archive/restore/hard-delete repetida en repositories.
- Reglas de dependencia hard-delete repetidas por entidad.
- Export/import mezclado en handlers con parseo de archivos, response building
  y reglas de dominio.

Inconsistencias visibles:

- `DELETE /:id`:
  - `lot`: legacy hacia archive.
  - `customer`, `project`, `work-order`, `supply`: legacy hacia hard delete.
  - `supply movement`: delete operativo, no equivalente a archive.
  - `labor`: conviven rutas de proyecto y OT, con hard delete explicito solo en
    el grupo de OT.
- Algunos modulos usan `sharedhandlers.RespondNoContent`, otros responden JSON
  manual.
- Algunos listados devuelven envoltorios `{success,data}` y otros payloads
  directos.
- Algunos repositorios validan dependencias antes de hard delete; otros delegan
  mas en constraints o checks parciales.

## Riesgos criticos

- Cross-tenant accidental si un path tenant-owned queda con `MaybeTenantScope`
  sin tenant en contexto o con SQL raw sin filtro. Mitigacion: migrar modulo por
  modulo a `RequireTenant`/`TenantScope` y agregar tests IDOR.
- Cambios en `dashboard`, `report` o `data-integrity` pueden alterar calculos.
  Mitigacion: golden master antes/despues y diff numerico `0`.
- Hard delete legacy puede borrar datos donde la UI cree estar archivando.
  Mitigacion: mantener rutas por compatibilidad, pero documentar y mover UI a
  endpoints explicitos.
- Logs/middleware pueden exponer headers sensibles si se agrega debugging sin
  redaccion. Mitigacion: centralizar redaccion y tests de logging.

## Riesgos medios

- Handlers grandes son dificiles de testear y favorecen validaciones duplicadas.
- SQL raw sin comentario operacional complica mantenibilidad y tuning.
- Interfaces por paquete pueden ocultar dependencias reales y hacer mas dificil
  refactorizar.
- Migraciones acumuladas contienen reconciliaciones y compat layers; no deben
  compactarse sin estrategia de baseline.

## Quick wins seguros

- Alinear `make lint` con CI. Hecho: el target usa golangci-lint v2.11.4 via
  `go run` y no depende de un `.golangci.yml` inexistente.
- Crear este reporte de auditoria versionado. Hecho.
- Agregar tests de comportamiento para rutas legacy antes de modificar
  semantica.
- Extraer `/me/context` fuera de `cmd/api/http_server.go` sin cambiar response.
- Crear helpers compartidos para legacy delete wrappers y parseo de ID donde ya
  se repite el mismo patron.

## Plan incremental

### Etapa 0 - Auditoria

- Mantener este documento actualizado.
- Completar matriz por modulo con handler/usecase/repo/dto/model/migrations/tests.
- Registrar rutas legacy y consumidores FE/BFF antes de remover o cambiar nada.

### Etapa 1 - Cleanup mecanico

- Eliminar imports, comentarios obsoletos y helpers internos muertos solo si
  lint/tests y wiring lo confirman.
- No eliminar endpoints ni DTOs publicos.
- Normalizar targets de Makefile y comandos de validacion.

### Etapa 2 - Consistencia CRUD

- Normalizar `Create`, `List`, `Get`, `Update`, `Archive`, `Restore`,
  `HardDelete` por entidad.
- Mantener rutas legacy pero redirigir internamente a metodos explicitos.
- Agregar tests para status code y efecto real de archive/hard-delete.

### Etapa 3 - Boundaries

- Extraer logica pesada de handlers hacia usecases/services locales.
- Mover `/me/context` a modulo auth/admin.
- Reducir interfaces triviales cuando no agregan inversion real ni facilitan
  tests.

### Etapa 4 - Datos y seguridad

- Migrar repos tenant-owned desde `MaybeTenantScope` a scope obligatorio.
- Revisar cada SQL raw con tenant, parametros y comentario de intencion.
- Agregar tests IDOR por modulo.

### Etapa 5 - Reportes y calculos

- Congelar golden master antes de tocar reportes.
- Refactorizar solo si el diff numerico sigue en `0`.

### Etapa 6 - Cierre

- Ejecutar test/build/lint/migraciones.
- Actualizar deuda restante con owner, riesgo y decision.
- Dejar CI verde.

## Cambios implementados en esta etapa

- `Makefile`: `make lint` ahora ejecuta el mismo golangci-lint que CI
  (`v2.11.4`) sin depender de un archivo de config inexistente.
- `/me/context`: la ruta salio del bootstrap `cmd/api/http_server.go` y ahora
  vive en `internal/admin`, preservando path, middleware y payload.
- `internal/admin`: agregado test de contrato para `/me/context` con usuario,
  tenant activo, rol y permisos. El test detecto y corrigio el mapeo GORM de
  `idp_sub` / `idp_email`. Invites y cambios de membership usan timestamps
  unicos por operacion.
- `internal/campaign`: agregado test de contrato para archive, restore y hard
  delete; se redujo duplicacion de parseo/respuesta en el handler con helper
  local. Archive/restore usan timestamp unico por operacion. No se cambio el
  shape del listado legacy.
- `internal/customer`: agregado test de contrato para las acciones por ID:
  legacy `DELETE /:id`, hard delete explicito, archive y restore. Tambien se
  redujo duplicacion de parseo/respuesta en el handler con un helper local.
  Restore usa timestamp unico para fila legacy y actor mirror.
- `internal/customer/repository_harddelete_test.go`: actualizado el contrato
  documentado del test de integracion; el comportamiento actual bloquea hard
  delete si el cliente tiene proyectos, no borra en cascada.
- `internal/field`: agregado test de contrato para hard delete, archive y
  restore; se redujo duplicacion de parseo/respuesta en el handler con helper
  local. Archive/restore usan timestamp unico por operacion.
- `internal/investor`: agregado test de contrato para hard delete, archive y
  restore; se redujo duplicacion de parseo/respuesta en el handler con helper
  local. Restore usa timestamp unico para fila legacy y actor mirror.
- `internal/lot`: agregado test de contrato para legacy delete, archive, restore
  y hard delete; se preservo la semantica legacy de `DELETE /:id` como alias de
  archive. Tambien se redujo duplicacion de parseo/respuesta en el handler con
  helper local. Archive/restore usan timestamp unico por operacion.
- `internal/lot/repository`: `GetMetrics` y `ListLots` resuelven proyectos via
  `sharedfilters.ResolveProjectIDs` antes de consultar vistas de reporte, por lo
  que `field_id` directo ya no saltea tenant. `lot_dates` se lee por
  `tenant_id` cuando hay contexto y ambos paths fallan cerrado sin tenant en
  `TENANT_STRICT_MODE=true`.
- `internal/manager`: agregado test de contrato para hard delete, archive y
  restore; se redujo duplicacion de parseo/respuesta en el handler con helper
  local equivalente al de `customer`. Restore usa timestamp unico para fila
  legacy y actor mirror.
- `internal/provider`: agregado test de contrato para listado y uso de
  `sharedhandlers.RespondOK` sin cambiar payload/status. Queda documentado que
  provider todavia no implementa CRUDAR completo.
- Decision de producto/arquitectura: `provider` no se eleva a CRUDAR propio en
  esta etapa. Se mantiene como catalogo derivado de movimientos/importaciones.
  Si luego necesita administracion completa, debe resolverse como vista/rol
  `proveedor` dentro de `actors`, no como mundo paralelo duplicado.
- `internal/project`: agregado test de contrato de rutas reales para archive,
  restore, hard delete explicito y legacy `DELETE /:id`. Se preservo que el
  legacy delete llama al alias legacy de hard delete. El handler solo recibio
  helper local de parseo/respuesta; no se tocaron repository, relaciones ni
  calculos.
- `internal/project/repository`: reducido el bloque repetido de archive/restore
  de dependientes por `project_id` a helpers locales con las mismas queries y
  mensajes de error. Tambien se usa un timestamp unico por operacion. No se
  cambio la semantica publica ni se tocaron calculos. Las mutaciones de
  relaciones del editor (`project_managers`, `project_investors`,
  `admin_cost_investors`, `field_investors`, fields y lots) y las cascadas de
  archive/restore ahora aplican `tenant_id` cuando existe en contexto,
  preservando compatibilidad temporal sin tenant.
- `internal/work-order`: agregado test de contrato de rutas reales para archive,
  restore, hard delete explicito y legacy `DELETE /:id`. Se preservo que el
  legacy delete llama al alias legacy de hard delete. El handler solo recibio
  helper local de parseo/respuesta; no se tocaron repository, imports, exports,
  splits, facturas, stock ni calculos.
- `internal/work-order/repository`: removido preload innecesario de `Items` en
  archive/restore, ya que esas operaciones solo validan la cabecera. Tambien se
  usa timestamp unico por operacion. No se tocaron splits, facturas ni calculos.
- `internal/work-order/repository`: `GetMetrics`, `GetRawDirectCost` y
  `GetHarvestAreaSnapshot` fallan cerrado sin tenant cuando
  `TENANT_STRICT_MODE=true`. Los raw joins de costo directo y metricas por
  insumo ahora matchean `tenant_id` entre ordenes, items, insumos y labores.
  No se cambiaron formulas; solo se acoto el universo de datos.
- `internal/work-order-draft`: agregado test de contrato para `DELETE /:id`.
  Se preservo que el delete de draft es hard delete operativo y que el usecase
  mantiene la regla de no borrar drafts publicados. El handler solo recibio
  helper local de parseo/respuesta; no se tocaron publicacion, PDF ni repository.
- `internal/supply`: agregado test de contrato de rutas reales para archive,
  restore, hard delete explicito y legacy `DELETE /:id` de insumos. Tambien se
  agrego test de contrato para archive, restore, hard delete y legacy delete de
  `supply-movements`, incluyendo los alias `stock-movements`. El handler solo
  recibio helpers locales de parseo/respuesta y un helper comun para parsear
  actor, proyecto y body bulk de movimientos. No se tocaron imports, exports,
  creacion de movimientos, stock, repository ni calculos.
- `internal/supply/repository`: reducido preload repetido de `Category` y `Type`
  a helper local `withSupplyLookups`; archive/restore usan timestamp unico por
  operacion. No se tocaron movimientos, stock ni calculos.
- `internal/supply/repository_movement`: reducido preload repetido de movimiento
  (`Supply`, `Type`, `Category`, `Investor`, `Provider`) a helper local. No se
  tocaron reglas de creacion, actualizacion, delete ni sincronizacion de stock.
  El join de providers contra `legacy_actor_map` ahora tambien matchea
  `tenant_id`, igual que `internal/provider`, para evitar actor mirror cruzado.
  Los joins de metadata de movimientos y origenes de insumos ahora matchean
  proveedores, proyectos, clientes y campanas por tenant cuando corresponde.
- `internal/stock`: se redujo duplicacion de parseo de `project_id` y `stock_id`
  en el handler con helpers locales, preservando el mensaje de negocio para
  stock invalido. No se tocaron calculos, repository, cierre de periodos ni
  exportacion.
- `internal/stock/repository`: reducido preload repetido de relaciones de stock
  (`Project`, `Supply`, `Type`, `Category`, `Investor`, `SupplyMovements`) a
  helpers locales. No se tocaron formulas de stock, consumo, cierre ni
  exportacion.
- `internal/labor`: agregado test de contrato para las rutas de delete por
  proyecto, delete global, archive, restore y hard delete. Se redujo duplicacion
  de parseo/respuesta en el handler con helper local. Se preservo que las rutas
  legacy `DELETE` llaman a `DeleteLabor`; no se tocaron exports, metricas,
  repository ni calculos de labores/OT.
- `internal/labor/repository`: eliminado `ListGroupLaborOld`, metodo sin
  callers marcado como TODO de eliminacion. Evita mantener una segunda version
  paralela de calculo de labores. El metodo activo `ListGroupLabor` no se cambio.
  Archive/restore usan timestamp unico por operacion.
- `internal/invoice`: agregado test de contrato para parseo de `work_order_id`,
  `investor_id`, `project_id` y paginacion. Se unifico el target operativo
  `work_order_id + investor_id` en helper local del handler. No se agrego CRUDAR
  porque factura es documento operativo por OT/inversor; no se tocaron snapshots
  ni repository.
- `internal/invoice/repository`: reducida duplicacion de validacion
  `work_order_id + investor_id` y error not found operativo; `Update` usa un
  timestamp unico por operacion. No se tocaron snapshots, SQL de pertenencia de
  inversor, creacion ni semantica de delete.
- `internal/actor`: agregado test de contrato para filtros de activos,
  archivados, archive, restore, hard delete y merge usando el actor autenticado.
  Se redujo duplicacion de parseo/respuesta en acciones por `actor_id` con
  helper local. No se tocaron SQL de deduplicacion, merge, legacy sync ni
  repository.
- `internal/actor/repository`: reducido el chequeo duplicado de existencia de
  actor activo en `AddRole` y `AddAlias` con helper local. `RestoreActor` usa un
  timestamp unico por operacion. No se tocaron SQL de deduplicacion, merge ni
  legacy sync.
- `internal/actor/legacy_sync`: los refresh de mirrors legacy ahora matchean
  `legacy_actor_map` tambien por `tenant_id` en proyectos, responsables,
  inversores, movimientos, ordenes, facturas y stock. Esto evita asociaciones de
  actor cruzadas cuando distintos tenants comparten IDs legacy. Tambien se
  corrigio la propagacion incremental de `SyncLegacyActor` para que los updates
  por `customer_id`, `investor_id`, `provider_id` o `manager_id` queden
  acotados al `tenant_id` del actor espejo. `RefreshProjectActorMirrors` ahora
  resuelve el tenant real del proyecto antes de limpiar/recrear responsibles,
  allocations y field lease participants; las tablas mirror no tienen
  `tenant_id` propio, por lo que el scope se aplica via `projects`/`fields`.
- `internal/actor/repository`: las agrupaciones de posibles duplicados por
  identificador, alias, razon social y nombre comercial ahora nacen acotadas al
  tenant activo. Antes el resultado final se cruzaba contra actores del tenant,
  pero el grupo podia formarse globalmente y producir falsos positivos entre
  empresas.
- `internal/report/repository`: `categories` quedo documentada y tratada como
  catalogo global segun schema; se removio `MaybeTenantScope` para evitar
  consultas a `c.tenant_id`, columna inexistente. En `getProjectInfo`, los joins
  a `customers` y `campaigns` ahora matchean `tenant_id` contra `projects`.
- `internal/project/repository`: los calculos de superficie en listados activos
  y archivados ahora matchean `tenant_id` en los joins `projects -> fields ->
  lots`. No cambia numeros con datos sanos, pero evita que una relacion corrupta
  cross-tenant infle hectareas.
- `internal/labor/repository`: `ListByWorkOrder` valida primero que la orden
  pertenezca al tenant activo antes de consultar la vista de reporte. No se
  tocaron formulas ni vistas de costo.
- `internal/labor/repository`: `ListAllGroupLabor` conserva compatibilidad
  transicional sin tenant, pero con tenant activo filtra por los proyectos
  resueltos del tenant y en `TENANT_STRICT_MODE=true` falla cerrado si falta
  tenant. No se tocaron formulas ni vistas de costo.
- `internal/invoice/repository`: `ListByProjectID` une `workorders` con
  `invoices` tambien por `tenant_id`, evitando listados con relaciones cruzadas.
- `internal/stock/repository`: los helpers raw que calculan consumo por stock
  (`loadConsumedByStockKey`, `loadAllConsumedByStockKey`) fallan cerrado sin
  tenant en `TENANT_STRICT_MODE=true` y matchean `tenant_id` entre ordenes e
  items. No se tocaron formulas de consumo.
- `internal/supply/repository` y `internal/supply/repository_movement`: los
  helpers raw de origenes de insumos/movimientos y metadata de proyecto destino
  fallan cerrado sin tenant en `TENANT_STRICT_MODE=true`. Con tenant activo
  siguen filtrando por `tenant_id`; sin strict conservan compatibilidad
  transicional.
- `internal/invoice/repository`: `InvestorBelongsToWorkOrder` falla cerrado sin
  tenant en `TENANT_STRICT_MODE=true`; con tenant activo mantiene los checks por
  `workorder_investor_splits`/`workorders` tenant-scoped.
- `internal/work-order-draft/repository`: el listado de borradores une
  `projects` y `fields` por `tenant_id` y excluye relaciones archivadas. Los
  raw SQL de numeros ocupados ya tenian variante tenant-scoped y se conservaron.
- Tests IDOR de repositorio agregados para `customer`, `campaign`, `field`,
  `investor`, `manager`, `lot`, `project`, `work-order`, `work-order-draft`,
  `labor` e `invoice`. Cubren
  list/get/update/archive/restore/hard-delete o acciones equivalentes contra
  IDs de otro tenant y verifican que no haya mutaciones laterales. El test de
  `project` agrega ademas un caso de field/lot corrupto de otro tenant apuntando
  al proyecto activo para validar superficie tenant-safe.
  El test de `work-order` cubre tambien la accion especifica
  `UpdateInvestorPaymentStatus` para asegurar que un split de otro tenant no se
  pueda marcar como pagado.
  El test de `labor` cubre ademas los chequeos de duplicado por nombre y el
  conteo de ordenes por labor contra IDs de otro tenant.
- `scripts/db/tenant_isolation_audit.sql`: la auditoria ahora devuelve filas
  detectables tambien para checks opcionales como `workorder_supply_items`, en
  lugar de emitir solo `NOTICE`. Se agregaron checks de actors, legacy mirrors y
  relaciones `*_actor_id` para detectar cruces de tenant antes de strict mode.
- `scripts/db/multi_tenant_golden_master.sql`: las consultas de superficie e
  invoices ahora validan joins por `tenant_id` entre proyectos/campos/lotes y
  ordenes/facturas. Ademas agrega una seccion de diffs tenant-safe vs legacy
  para superficie, facturas, movimientos de insumos y ordenes; si algun diff no
  es `0`, el script falla.
- `scripts/db/reset-local-db-from-prod.sh`: el flujo de restore soporta dumps
  legacy sin `tenant_id` poblado usando `MIGRATE_TARGET_VERSION=224`,
  `POST_RESTORE_TENANT_BACKFILL=1`, `ACTORS_BACKFILL_SYNC=1` y
  `RUN_FINAL_MIGRATIONS=1`. Esto permite restaurar dev, backfillear tenant,
  sincronizar actors y recien despues aplicar constraints `000225`.
- Modulos chicos/globales completados con cobertura de contrato HTTP:
  `category`, `class-type`, `crop`, `lease-type`, `business-parameters`,
  `dollar`, `commercialization` y `businessinsights`. Los tests cubren parseo
  de IDs, paginacion, actor/tenant cuando aplica, create/update/delete o bulk
  por proyecto. No se cambio semantica publica ni persistencia.
- `project`: el helper interno `ensureCrop` ya respeta `tenant_id` por ID y por
  nombre, con test de aislamiento para evitar reusar cultivos de otro tenant al
  crear/editar proyectos.
- `businessinsights`: acciones manuales sobre candidatos (`resolve`, `reopen`,
  `read`, `notify`) quedan cubiertas por test IDOR; `resolve` y `reopen` ya no
  devuelven exito silencioso cuando el candidato pertenece a otro tenant.
- Modulos sensibles de calculo completados solo a nivel contrato HTTP:
  `dashboard`, `report` y `data-integrity`. Los tests cubren parseo de filtros
  de workspace, tipos de reporte y validaciones requeridas. No se tocaron
  repositories, SQL, DTOs matematicos, formulas ni reglas de calculo; cualquier
  cambio futuro en esos puntos requiere golden master antes/despues.
- `docs/BACKEND_CLEANUP_AUDIT.md`: auditoria inicial, riesgos, quick wins y plan
  incremental versionado.
- Auth legacy: eliminado `internal/platform/http/middlewares/gin/require_jwt.go`
  y `internal/shared/utils/jwt_tools.go`. No estaban cableados ni llamados; el
  auth real sigue siendo Identity Platform o local-dev auth controlado. `go mod
  tidy` dejo `github.com/golang-jwt/jwt/v5` como dependencia indirecta de
  `core/authn/go/jwks`, no como uso propio del backend.
- Local-dev auth: cuando `AUTH_REQUIRE_TENANT_HEADER=true`, el middleware local
  ahora tambien falla cerrado para rutas tenant-scoped sin `X-Tenant-Id` o con
  tenant invalido. El fallback a tenant `default` queda solo para compatibilidad
  local en rutas que permiten tenant implicito, como `/me/context`.
- `internal/ai`: el chequeo de pertenencia `project_id + tenant_id` dejo de
  depender de SQL especifico de PostgreSQL y queda cubierto por tests unitarios
  contra SQLite. Los tests validan permiso `ai.use`, header de proyecto,
  rechazo de proyecto cross-tenant y aceptacion de proyecto del tenant activo.
- Logger HTTP: el redactor de headers ahora normaliza underscores antes de
  evaluar sensibilidad, por lo que tambien tapa variantes como `X_API_KEY` y
  `X_api_key`; se agrego test para evitar regresion.
- `internal/shared/handlers`: eliminado alias deprecado `ParseUserID`; no tenia
  callers reales y todos los handlers activos usan `ParseActor`.
- Middleware surface: eliminado `Middlewares.GetProtected()` y el campo
  `protected` porque no habia callers reales y siempre devolvia un slice vacio.
  Los handlers siguen usando `GetValidation()`, donde viven API key + auth real.
- Codigo muerto confirmado: eliminados helpers sin callers
  `DashboardModelMapper.investorContributionsToDomain`, `decToString` de
  exportacion Excel de ordenes, y el camino no cableado
  `connectWithConnectorIAMAuthN`. Tambien se elimino `internal/shared/utils/strings.go`,
  que solo contenia helpers string sin callers internos. Se removio la dependencia
  directa a `cloud.google.com/go/cloudsqlconn` con `go mod tidy`.

## Validacion ejecutada en esta etapa

- `make lint`: verde, 0 issues.
- `go test ./...`: verde.
- `TENANT_STRICT_MODE=true go test ./...`: verde.
- `go build ./cmd/api ./cmd/migrate`: verde.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.11.4 run --timeout=5m`:
  verde, 0 issues.
- Dump fresco de dev restaurado localmente desde `new_ponti_db_dev` usando el
  flujo `224 -> restore -> tenant backfill -> actors sync -> 225`.
- DB post-restore: `schema_migrations = 225`, `dirty = false`, con
  `customers=6`, `projects=20`, `lots=79`, `workorders=482`, `actors=183`.
- `scripts/db/tenant_isolation_audit.sql` contra dump dev restaurado:
  verde, 0 hallazgos.
- `scripts/db/actors_golden_master.sql` contra dump dev restaurado:
  verde, 27 checks con `diff = 0`.
- `scripts/db/multi_tenant_golden_master.sql` contra dump dev restaurado:
  ejecuta correctamente, genera snapshot de control por tenant, proyecto, actor
  legacy y dominio, y valida 4 diffs tenant-safe vs legacy en `0`.
- BFF: `npm run lint && BFF_REQUIRE_TENANT=true npm test && npm run build`:
  verde.
- UI: `yarn typecheck && yarn lint && yarn build`: verde.
- UI E2E: `PLAYWRIGHT_SKIP_WEBSERVER=1 PLAYWRIGHT_BASE_URL=http://127.0.0.1:5173 yarn test:e2e`:
  verde, 4/4 tests sobre dump dev restaurado.
- Runtime smoke local con backend strict:
  - `APP_ENV=local AUTH_ENABLED=false AUTH_REQUIRE_TENANT_HEADER=true AUTH_AUTO_PROVISION_MEMBERSHIP=false TENANT_STRICT_MODE=true HTTP_SERVER_PORT=18080 go run ./cmd/api`.
  - `/api/v1/healthz`: `200`.
  - `GET /api/v1/customers` sin tenant: `403`.
  - `GET /api/v1/customers` con tenant invalido: `403`.
  - `GET /api/v1/customers` con tenant valido: `200`.
  - `GET /api/v1/me/context` sin tenant: `200`.
  - `POST /api/v1/ai/chat` sin tenant: `403`.
  - `POST /api/v1/ai/chat` con tenant y sin `ai.use`: `403`.

## Deuda restante

- Falta profundizar la matriz modulo por modulo con rutas, DTOs, modelos,
  migraciones y consumidores FE/BFF.
- Falta separar `/me/context` en usecase/repository propio; la primera extraccion
  ya saco la logica del bootstrap HTTP.
- Falta homogeneizar semantica de rutas `DELETE /:id`.
- Falta convertir `MaybeTenantScope` transicional a strict en todos los paths
  tenant-owned.
- Falta golden master antes de tocar SQL, repositories, formulas o DTOs
  matematicos de `dashboard`, `report` y `data-integrity` en cambios futuros.
- Falta validacion runtime en entorno objetivo con `TENANT_STRICT_MODE=true`,
  `BFF_REQUIRE_TENANT=true` y `AI_TENANT_SCOPE=true` antes de declarar
  produccion enterprise cerrada. Las suites locales/controladas con strict mode
  ya estan verdes.
