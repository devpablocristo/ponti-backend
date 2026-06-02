# implementation-status.md — feature-007 actor-system (BE)

## Estado general

**Completa** a nivel de paquete (en SOURCE `777e5f6a`), con la salvedad de que el
**cableado** (`wire/*`, `cmd/api/http_server.go`) vive en archivos compartidos que NO
están en mi flist y deben portarse aparte para que la feature sea alcanzable.

- **% completitud (este repo)**: ~95% (código + migraciones + tests completos; falta
  asegurar el porte de los 3 hunks de cableado).
- **Estado en este repo (BE)**: implementado, con tests verdes esperables.
- **Estado en el otro repo (FE)**: existe como feature-007 (useActors + master-data/actors
  + BFF). No evaluado acá; va después (BE-first).

## Detalle por capa

| capa | archivo | estado | nota |
|---|---|---|---|
| handler | `handler.go` | completo | 12 rutas, parseo de filtros/paginación |
| dto | `handler/dto/actor.go`, `duplicate.go` | completo | request/response + mappers |
| usecases | `usecases.go` | completo | delegación a repo (capa fina) |
| domain | `usecases/domain/actor.go` | completo | kinds/roles, `Validate`, `IsArchived` |
| repository | `repository.go` | completo | tenancy, unicidad, merge, duplicates, normalizeName |
| models | `repository/models/actor.go` | completo | mapeo a `deleted_at` (post-231) |
| sync legacy | `legacy_sync.go`, `master_link.go` | completo | backfill legacy→actor; toca tablas de 010 |
| wire | `wire/actor_providers.go` | completo | `ActorSet` |
| migraciones | 223/226/231/234 (up+down) | completas | ver bloqueantes |

## Tests

| suite | tipo | infra | estado esperado |
|---|---|---|---|
| `usecases/domain/actor_test.go` | unit dominio | ninguna | verde |
| `usecases_test.go` | unit (mock repo) | ninguna | verde |
| `handler_test.go` | unit (gin httptest) | ninguna | verde |
| `repository_tenant_test.go` | integración liviana | sqlite in-memory | verde (no docker) — cubre aislamiento de tenant + conflicto duplicado en create/update |

Cobertura SDD (del SPEC.md): duplicado por normalized_name, edición tomando nombre de
otro activo, edición manteniendo el propio, mismo nombre en otro tenant, nombre de
actor archivado/fusionado. Tests presentes en `repository_tenant_test.go` y dominio.

## Pendientes / bugs por categoría

### BLOQUEANTE para mergear

1. **Portar los 3 hunks de cableado** (`wire/wire.go`, `wire/wire_gen.go`,
   `cmd/api/http_server.go`). Sin esto el paquete compila pero las rutas no se registran
   y `*actor.Handler` no se inyecta. Preferir regenerar `wire_gen.go`.
2. **Verificar dependencias en develop**: `shared/text.CanonicalizeName` (004),
   `shared/lifecycle` (002), `shared/models.Base` (002/003), `shared/authz`,
   `shared/handlers`, `shared/repository`, `platform/*`. Si falta una, no compila.
3. **Orden de migración con projects (010)**: 223 backfillea `project_*` y 226 setea
   `projects.customer_actor_id`. Confirmar que `projects`/`customers` existen en el
   entorno antes de aplicar; si no, la migración aborta.

### Mejora futura

- `legacy_sync.go`/`master_link.go` acoplan SQL crudo a tablas de otras features; un
  refactor a interfaces de dominio reduciría el acoplamiento (no bloqueante).
- El DTO expone `archived_at` por compat aunque la DB use `deleted_at` (decisión de
  compat con FE); revisar si en una limpieza futura se unifica el nombre con FE.

### Deuda aceptable

- `ActorRole`/`ActorAlias` no embeben `sharedmodels.Base` a propósito (sin
  `updated_at`/`created_by`); documentado en models/actor.go.

### Duda humana

- ¿La migración 234 (merge destructivo de duplicados activos) es segura en los datos de
  cada entorno productivo, o solo en dev? El SPEC la describe como consolidación dev;
  confirmar con un humano antes de aplicar en prod. (Ver `risks.md`.)
- ¿Las tablas `project_*` se crean acá (223) o en 010? En SOURCE las crea 223; coordinar
  con el extractor de 010 para evitar duplicar `CREATE TABLE` entre features.
