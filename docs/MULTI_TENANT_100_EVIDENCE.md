# Multi-Tenant 100 Evidence

Fecha: 2026-05-12

Este documento define que se puede afirmar sin mentir sobre el alcance
multi-tenant enterprise de Ponti.

## Claim permitido hoy

El alcance multi-tenant enterprise definido esta implementado y validado en
modo local/controlado contra dump fresco de dev: hardening de auth, helpers de
tenant, BFF/UI tenant foundation, actors/AI foundation, migraciones base,
scoping tenant en CRUDs operativos, scoping tenant en reportes analiticos clave,
auditoria SQL, golden masters numericos verdes y suites estrictas con
`TENANT_STRICT_MODE=true`, `BFF_REQUIRE_TENANT=true` y AI tenant scope cubierto
por tests de backend.

## Claim no permitido todavia

No decir "produccion enterprise cerrada" hasta desplegar y observar en un
entorno objetivo con las flags strict activas:

- `TENANT_STRICT_MODE=true`
- `BFF_REQUIRE_TENANT=true`
- `AI_TENANT_SCOPE=true`
- validacion runtime de esas flags con trafico/smoke real del entorno

## Evidencia local verde

Backend:

- `go test ./...`: verde.
- `TENANT_STRICT_MODE=true go test ./...`: verde.
- `go build ./cmd/api ./cmd/migrate`: verde.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.11.4 run --timeout=5m`:
  verde, 0 issues.
- Dump fresco de dev restaurado localmente el 2026-05-12 desde
  `new_ponti_db_dev` con migracion hasta `000224`, restore data-only, backfill
  tenant, sync actors y migracion final a `000225`.
- DB restaurada: `schema_migrations = 225`, `dirty = false`.
- Conteos post-restore: `customers=6`, `projects=20`, `lots=79`,
  `workorders=482`, `actors=183`.
- `scripts/db/tenant_isolation_audit.sql` contra dump dev restaurado:
  verde, 0 hallazgos.
- `scripts/db/actors_golden_master.sql` contra dump dev restaurado:
  verde, 27 checks con `diff = 0`.
- `scripts/db/multi_tenant_golden_master.sql` contra dump dev restaurado:
  verde, snapshot de control y 4 diffs tenant-safe vs legacy en `0`.

BFF:

- `npm run lint`
- `BFF_REQUIRE_TENANT=true npm test`
- `npm run build`

UI:

- `yarn typecheck`
- `yarn lint` con 0 errores y warnings existentes
- `yarn build`
- `PLAYWRIGHT_SKIP_WEBSERVER=1 PLAYWRIGHT_BASE_URL=http://127.0.0.1:5173 yarn test:e2e`:
  verde, 4/4 tests sobre dump dev restaurado.

## Cambios cubiertos en esta tanda

- `tenant_id` agregado a modelos operativos que ya estan en migraciones 224/225:
  - campaigns
  - fields / field investors
  - lots / lot dates
  - labors
  - supplies
  - supply movements
  - stocks
  - workorders / workorder items / workorder investor splits
  - invoices
  - work order drafts
- Scoping transicional con `authz.MaybeTenantScope` en CRUD/list/archive/restore/hard-delete de esos modulos.
- `sharedfilters.ResolveProjectIDs` ahora respeta tenant cuando existe contexto y permite "todos los proyectos del tenant" sin caer en "todos los proyectos de la base".
- `dashboard` resuelve siempre los proyectos via `sharedfilters.ResolveProjectIDs`; un `project_id` directo ya no saltea validacion tenant.
- `report` y `data-integrity` propagan `context.Context` hasta las consultas analiticas para heredar tenant.
- `actors` ya no resuelve `default` en requests normales: list/get/update/archive/restore/hard-delete/merge/deduplicacion usan el tenant del contexto.
- `project` escribe `tenant_id` en proyectos, pivots y relaciones nuevas; sus list/get/update/archive/restore/hard-delete quedan tenant-scoped en modo transicional.
- `scripts/db/tenant_isolation_audit.sql` cubre mas relaciones cross-tenant y constraints globales peligrosos.
- `scripts/db/tenant_isolation_audit.sql` ahora emite filas detectables para checks opcionales, no solo notices.
- `scripts/db/multi_tenant_golden_master.sql` deja un snapshot numerico reproducible y falla si los diffs tenant-safe vs legacy no son `0`.
- `scripts/db/actors_golden_master.sql` compara legacy vs actor mirrors y cobertura 1:1 con diff `0` local.
- Tests agregados: `internal/shared/filters` verifica que `ResolveProjectIDs` falle cerrado frente a `project_id`/`field_id` de otro tenant.
- Tests IDOR de repositorio agregados para `customer`, `campaign`, `field`, `investor`, `manager`, `provider`, `lot`, `project`, `work-order`, `work-order-draft`, `labor`, `invoice`, `supply`, `supply_movements`, `stock` y `actors`.
- `project` ahora suma hectareas con joins tenant-safe entre `projects`, `fields` y `lots`; el test cubre un caso corrupto donde un field/lot de otro tenant apunta al proyecto actual y no debe contaminar superficie.
- `dashboard` tiene test de resolucion de proyectos tenant-scoped y fallback de aportes endurecido para unir `project_investors` con `investors` por `tenant_id`.
- `report` tiene test de resolucion/info de proyecto tenant-scoped y `getProjectInfo` valida `tenant_id` aun despues de resolver el proyecto.
- `data-integrity` tiene test de contrato para asegurar que preserva el `tenant_id` del contexto y pasa el `project_id` a todas las fuentes que compara.
- `provider` ya no tiene bypass SQLite sin tenant en tests; el repo lista proveedores con `MaybeTenantScope` tambien en ese camino.
- `authz.MaybeTenantScope` mantiene compatibilidad transicional cuando `TENANT_STRICT_MODE=false`, pero falla cerrado si falta tenant cuando `TENANT_STRICT_MODE=true`.
- Tests agregados: `internal/shared/authz` cubre strict/transition mode de `MaybeTenantScope`; `internal/customer` cubre un repositorio real rechazando listados sin tenant en strict mode.
- `sharedfilters.ResolveProjectIDs` y `sharedfilters.ValidateProjectAccess` fallan cerrado sin tenant cuando `TENANT_STRICT_MODE=true`; las tablas project-owned sin tenant historico propio (`project_dollar_values`, `crop_commercializations`) validan pertenencia del proyecto antes de leer/escribir.
- Modelos y repositorios de catalogos tenant-owned ya migrados por DB (`crops`, `categories`, `types`, `lease_types`, `business_parameters`, `project_dollar_values`, `crop_commercializations`) ahora incluyen `tenant_id`, escriben el tenant del contexto y filtran list/get/update/delete por tenant.
- Tests IDOR/strict agregados para catalogos y parametros chicos: `crop`, `category`, `class-type`, `lease-type`, `business-parameters`, `dollar` y `commercialization`.
- `labor.ListAllGroupLabor` ya no consulta toda la vista de reporte cuando hay tenant activo: resuelve proyectos por tenant y filtra por esos IDs; en strict mode sin tenant falla cerrado.
- `work-order` endurece metricas y raw helpers: `GetMetrics`, `GetRawDirectCost` y `GetHarvestAreaSnapshot` fallan cerrado sin tenant en strict mode, y los raw joins criticos matchean `tenant_id` entre workorders, items, supplies y labors.
- `lot` endurece metricas/listados sobre vistas: `GetMetrics` y `ListLots` resuelven proyectos por tenant antes de aplicar `project_id`/`field_id`; `lot_dates` se lee con `tenant_id` cuando hay contexto; strict mode sin tenant falla cerrado.
- `stock`, `supply` e `invoice` endurecen helpers raw residuales: consumo de stock, origenes de insumos/movimientos, metadata de destino e invoice-investor validation fallan cerrado sin tenant en strict mode y conservan compatibilidad transicional con `TENANT_STRICT_MODE=false`.
- `work-order-draft` endurece los helpers raw de numeros ocupados/publicados para fallar cerrado sin tenant en strict mode.
- `project.ensureCrop` ahora respeta tenant tanto por ID explicito como por nombre, evitando reutilizar cultivos de otro tenant en creacion/edicion de proyectos; tiene test de aislamiento del helper.
- `businessinsights` ya no devuelve falso exito al resolver/reabrir candidatos de otro tenant: esas acciones devuelven not found y el test cubre list/read/resolve/reopen/notify cross-tenant.
- `scripts/db/reset-local-db-from-prod.sh` permite restaurar dumps legacy de dev
  migrando primero hasta una version objetivo (`MIGRATE_TARGET_VERSION=224`),
  reejecutando backfill tenant post-restore, sync actors y migraciones finales.
  Esto evita que dumps sin `tenant_id` fallen contra constraints `NOT NULL`.
- E2E auth helper usa `/me/context` para persistir el tenant restaurado antes
  de navegar, evitando que las pruebas dependan de un tenant hardcodeado.
- Local-dev auth respeta `AUTH_REQUIRE_TENANT_HEADER=true`: rutas
  tenant-scoped sin `X-Tenant-Id` o con tenant invalido fallan cerrado aun en
  modo local; `/me/context` conserva acceso implicito para bootstrap.
- `internal/ai` valida `ai.use`, `X-PROJECT-ID` y pertenencia
  `project_id + tenant_id` cuando AI tenant scope esta activo; los tests cubren
  falta de permiso, falta de proyecto, proyecto de otro tenant y proyecto valido.
- Smoke runtime local con backend levantado en strict:
  - `/api/v1/healthz`: `200`.
  - `GET /api/v1/customers` sin `X-Tenant-Id`: `403 tenant header required`.
  - `GET /api/v1/customers` con tenant invalido: `403 invalid tenant header`.
  - `GET /api/v1/customers` con tenant valido: `200`.
  - `GET /api/v1/me/context` sin tenant: `200` para bootstrap.
  - `POST /api/v1/ai/chat` sin tenant: `403 tenant header required`.
  - `POST /api/v1/ai/chat` con tenant pero sin `ai.use`: `403 insufficient permissions`.
- Logger HTTP redacted headers sensibles con guion y underscore:
  `X-API-Key`, `X_API_KEY`, `X_api_key`, `X-Tenant-Id`, tokens, cookies y
  service keys.
- BFF/UI validations siguen verdes.

## Gates operativos restantes

- Activar `TENANT_STRICT_MODE=true`, `BFF_REQUIRE_TENANT=true` y
  `AI_TENANT_SCOPE=true` en el entorno objetivo y ejecutar smoke/runtime real.
- Ejecutar pruebas especificas contra Ponti AI real cuando el servicio AI este
  desplegado con memoria indexada por `tenant_id + project_id + user_id`.
- Revisar si quedan SQL raw residuales permitidos por compatibilidad
  transicional y documentar excepciones antes de remover `MaybeTenantScope`.

## Frase final permitida hoy

Se completo 100% del alcance local/CI-verificable definido: auth, tenant
isolation, RBAC base, BFF, UI, AI backend scope, auditoria DB y golden master,
sin activar cambios que alteren calculos.

## Frase final permitida despues de cerrar gates operativos

Se completo y se desplego 100% del alcance enterprise definido: auth, tenant
isolation, RBAC, BFF, UI, AI, auditoria DB y golden master, con flags strict
activas en el entorno objetivo y sin cambios numericos en calculos.
