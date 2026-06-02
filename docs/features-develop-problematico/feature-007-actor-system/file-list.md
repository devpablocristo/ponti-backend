# file-list.md — feature-007 actor-system (BE)

Flist autoritativo: `/tmp/flists/be-007.txt` (24 paths, todos `A` = created en `0972e565..777e5f6a`).
Todos los archivos del paquete son **NUEVOS** (created), por lo que la extracción por
defecto es `whole-file`. Los archivos compartidos NO están en mi flist y se documentan
en la sección "Compartidos (fuera del flist)".

Leyenda extracción: whole-file / partial-hunks / manual-port / do-not-extract-yet.

## Propios (núcleo de la feature — whole-file)

| path | status | tipo | rol en la feature | extracción | motivo | riesgo | confianza |
|---|---|---|---|---|---|---|---|
| `internal/actor/SPEC.md` | A | doc/SDD | spec de unicidad + tests SDD | whole-file | doc del módulo, sin código | bajo | alta |
| `internal/actor/handler.go` | A | http handler | 12 rutas REST + `Routes()` | whole-file | define endpoints | bajo | alta |
| `internal/actor/handler/dto/actor.go` | A | dto | request/response + `ToDomain/FromDomain` | whole-file | contrato API | bajo | alta |
| `internal/actor/handler/dto/duplicate.go` | A | dto | duplicate-candidate + merge-impact | whole-file | contrato API | bajo | alta |
| `internal/actor/usecases.go` | A | usecases | port + delegación a repo | whole-file | capa app | bajo | alta |
| `internal/actor/usecases/domain/actor.go` | A | domain | `Actor`, kinds/roles, `Validate` | whole-file | dominio puro | bajo | alta |
| `internal/actor/repository.go` | A | repository | persistencia GORM + tenancy + merge + duplicates | whole-file | `normalizeName`/`requestTenantID`/`MergeActors`/`ListDuplicateCandidates` | medio | alta |
| `internal/actor/repository/models/actor.go` | A | gorm models | mapeo tablas + `ToDomain/FromDomain` | whole-file | esquema GORM | bajo | alta |
| `wire/actor_providers.go` | A | wire DI | `ActorSet` (providers del módulo) | whole-file | provider propio del módulo | bajo | alta |

## Compartidos (partial-hunks) — sync legacy y cableado

| path | status | tipo | rol en la feature | extracción | motivo | riesgo | confianza |
|---|---|---|---|---|---|---|---|
| `internal/actor/legacy_sync.go` | A | sync legacy | backfill `legacy_actor_map` + `actors` + `project_*` desde customers/investors/managers/providers | whole-file (archivo) pero COMPARTIDO en intención | toca tablas de 010/011/018; revisar que existan | **alto** | media |
| `internal/actor/master_link.go` | A | sync legacy | `EnsureCustomerFromActor` y resolución de actor por nombre; toca `customers`/`projects` | whole-file (archivo) pero COMPARTIDO en intención | acopla con dominio customer (010) | **alto** | media |

> Nota: estos dos son archivos NUEVOS y propios del paquete `actor`, así que se copian
> enteros. Se marcan COMPARTIDOS porque su lógica SQL referencia tablas que pertenecen a
> otras features; el riesgo no es el diff del archivo sino la coexistencia de tablas.

## Compartidos (FUERA del flist — partial-hunks obligatorios para que funcione)

Estos NO aparecen en `/tmp/flists/be-007.txt` pero el módulo es inalcanzable sin ellos.
Confirmado en `777e5f6a`: `wire/wire.go:49`, `wire/wire_gen.go:75/340/380`,
`cmd/api/http_server.go:148`.

| path | status | tipo | rol en la feature | extracción | motivo | riesgo | confianza |
|---|---|---|---|---|---|---|---|
| `wire/wire.go` | M | wire DI | declara `ActorHandler *actor.Handler` en deps | partial-hunks | hunk del actor; resto pertenece a otras features | **alto** | alta |
| `wire/wire_gen.go` | M | wire DI (generado) | `ProvideActorHandler(...)` + asignación en struct | partial-hunks (o regenerar con `wire`) | generado; mejor regenerar | **alto** | alta |
| `cmd/api/http_server.go` | M | bootstrap http | `deps.ActorHandler.Routes()` | partial-hunks | una línea entre muchos handlers | **alto** | alta |

## Requeridos por dependencia (NO traer acá — ya deben existir)

| path/símbolo | feature | rol | acción |
|---|---|---|---|
| `internal/shared/text.CanonicalizeName` | 004 | normaliza display name | debe existir antes |
| `internal/shared/lifecycle.*` | 002 | CRUDAR soft-delete | debe existir antes |
| `internal/shared/models.Base` / `shared/domain.Base` | 002/003 | `deleted_at`/auditoría | debe existir antes |
| `internal/shared/repository.*`, `internal/shared/handlers.*` | 001/023 | helpers http+repo | debe existir antes |
| `internal/shared/authz.*` | 003 | scopes/tenant | debe existir antes |
| `github.com/devpablocristo/platform/...` | 001/005 | platform libs | debe existir (go.mod) |

## Tests (whole-file)

| path | status | tipo | infra | extracción | confianza |
|---|---|---|---|---|---|
| `internal/actor/handler_test.go` | A | unit (mocks gin) | sin DB | whole-file | alta |
| `internal/actor/usecases_test.go` | A | unit (mock repo) | sin DB | whole-file | alta |
| `internal/actor/usecases/domain/actor_test.go` | A | unit dominio | sin DB | whole-file | alta |
| `internal/actor/repository_tenant_test.go` | A | integración liviana | **sqlite in-memory** (no docker) | whole-file | alta |

## Migraciones (whole-file, up + down)

| path | status | rol | riesgo | confianza |
|---|---|---|---|---|
| `migrations_v4/000223_actors_safe_migration.up.sql` / `.down.sql` | A | crea modelo actor + tablas + backfill legacy 1:1 | **alto** (toca legacy + projects) | media |
| `migrations_v4/000226_customer_actor_master_link.up.sql` / `.down.sql` | A | `customers.actor_id` + `projects.customer_actor_id` + índice único parcial | alto (depende de 010 projects) | media |
| `migrations_v4/000231_consolidate_actor_archived_at.up.sql` / `.down.sql` | A | consolida `archived_at`→`deleted_at` en actors/roles/aliases | medio | alta |
| `migrations_v4/000234_actor_unique_normalized_name.up.sql` / `.down.sql` | A | índice único parcial + merge de duplicados activos | **alto** (merge destructivo en datos) | media |

## Dudosos

- `legacy_sync.go` y `master_link.go`: ¿el dominio de `customers`/`projects` que
  referencian ya está en `develop`? Verificar antes (ver `validation.md`). Confianza media.

## NO traer todavía

- Cualquier hunk de `wire/wire_gen.go` ajeno al actor (lo generado de otras features).
- No tocar otras migraciones `migrations_v4/*` fuera de las 4 listadas.
- Nada del FE (va en el paquete feature-007 del otro repo).

## Inventario adicional (completitud)

Estos 4 paths son archivos `.down.sql` reales (`A` = created en `0972e565..777e5f6a`,
SOURCE `777e5f6a`) que ya se mencionaban como par up/down en la sección "Migraciones"
pero faltaban como filas propias. Verificado en el diff: cada `.down.sql` es el inverso
de su `.up.sql` (mismo número de migración), envuelto en `BEGIN;`/`COMMIT;` y todo con
`IF EXISTS`/`IF NOT EXISTS`. La extracción es `whole-file` y SIEMPRE junto a su `.up`
correspondiente (un `.down` huérfano deja el rollback roto).

| path | status | tipo | rol | extracción | motivo | riesgo | confianza |
|---|---|---|---|---|---|---|---|
| `migrations_v4/000223_actors_safe_migration.down.sql` | A | rollback migración | revierte 000223: DROP de la vista `actor_migration_coverage`, FKs `fk_*_actor`, índices `idx_*_actor_id` y columnas `*_actor_id`/`*_name_snapshot` en invoices/labors/supply_movements/stocks/workorders/workorder_investor_splits/projects | whole-file (traer con su `.up`) | inverso 1:1 del up; toca tablas legacy + projects | **alto** (rollback sobre tablas de otras features) | alta |
| `migrations_v4/000226_customer_actor_master_link.down.sql` | A | rollback migración | revierte 000226: DROP de índices (`ux_customers_tenant_actor_id`, `idx_customers_actor_id`, `idx_actors_tenant_normalized_name`), FK `fk_customers_actor` y columna `customers.actor_id` | whole-file (traer con su `.up`) | inverso 1:1 del up; depende de la tabla `customers` (010) | alto | alta |
| `migrations_v4/000231_consolidate_actor_archived_at.down.sql` | A | rollback migración | revierte 000231: re-crea `archived_at` y copia `deleted_at`→`archived_at` en actors/actor_roles/actor_aliases, restaura índices `idx_*_archived_at` y dropea `deleted_at` en roles/aliases (donde no existía antes) | whole-file (traer con su `.up`) | inverso 1:1 del up; rollback con backfill de columnas | medio | alta |
| `migrations_v4/000234_actor_unique_normalized_name.down.sql` | A | rollback migración | revierte 000234: DROP del índice único parcial `ux_actors_tenant_normalized_name_active` y re-crea el índice no-único `idx_actors_tenant_normalized_name` (no revierte el merge de duplicados del up) | whole-file (traer con su `.up`) | inverso parcial del up: recupera índices pero NO deshace el merge destructivo | medio | media |
