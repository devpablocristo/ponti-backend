# dependencies.md — feature-007 actor-system (BE)

## Depende-de

### Fuertes (sin esto NO compila / NO migra)

| dep | feature | qué provee | evidencia (en mi flist) |
|---|---|---|---|
| platform tenancy | 001 | `requestTenantID(ctx)` via OrgID; `platform/*` libs | `repository.go` imports `platform/...`, `requestTenantID` (repo.go:639) |
| CRUDAR lifecycle | 002 | `internal/shared/lifecycle`, soft-delete `deleted_at`, `shared/models.Base` | import `shared/lifecycle` (1x), `sharedmodels.Base` en models/actor.go |
| multitenant DB hardening | 003 | `tenant_id uuid` FK → `auth_tenants`, `shared/authz` | migración 223 `REFERENCES auth_tenants`, import `shared/authz` (2x) |
| shared text propername | 004 | `internal/shared/text.CanonicalizeName` | import en dto/actor.go y master_link.go (`text`/`sharedtext`) |
| shared handlers/repository | 001/023 | `shared/handlers` (BindJSON/Respond*), `shared/repository` | imports en handler.go y repository.go |

### Fuertes (de cableado, fuera del flist)

| dep | feature | qué provee | evidencia |
|---|---|---|---|
| wire DI | 023 | `wire/wire.go` declara `ActorHandler`, `wire/wire_gen.go` inyecta | `wire.go:49`, `wire_gen.go:75/340/380` (en 777e5f6a) |
| http bootstrap | (shared/cmd) | `cmd/api/http_server.go` llama `deps.ActorHandler.Routes()` | `http_server.go:148` |

### Débiles / inciertas

| dep | feature | nota | confianza |
|---|---|---|---|
| projects | 010 | migración 223 crea/backfillea `project_responsibles`, `project_investor_allocations`, `project_admin_cost_allocations`, y 226 setea `projects.customer_actor_id`. Si `projects` no existe en `develop`, falla. | media |
| campaign dto projectId | 011 | comparte tablas de projects/allocations | baja |
| data-integrity-admin | 018 | la UI de duplicate-candidates / merge se apoya en estos endpoints | baja |

## Bloquea-a

| feature | por qué |
|---|---|
| 010 projects | depende de `actors` y de las tablas `project_*` declaradas por la migración 223 |
| 011 campaign-dto-projectid | usa `customer_actor_id`/allocations |
| 018 data-integrity-admin | consume `/actors/duplicate-candidates` y `/actors/merge` |
| FE feature-007 | consume todos los endpoints (BE-first) |

## Cross-repo

- FE feature-007: `useActors`, `master-data/actors`, BFF `api/src/routes/actors.ts`.
  Consume contratos JSON definidos en `dto/actor.go` y `dto/duplicate.go`. **Coordinar:
  el FE debe mergearse después del BE.** Cualquier cambio de shape (p. ej. `archived_at`
  vs `deleted_at`: el DTO expone `archived_at` por compat aunque la DB use `deleted_at`)
  debe coordinarse.

## Archivos / tipos / config / migraciones / APIs compartidos

- **Tablas creadas que sirven a otras features**: `project_responsibles`,
  `project_investor_allocations`, `project_admin_cost_allocations`,
  `field_lease_participants`, `legacy_actor_map` (223).
- **Columnas agregadas a tablas de otras features**: `customers.actor_id`,
  `projects.customer_actor_id` (226).
- **Símbolos compartidos consumidos**: `shared/text.CanonicalizeName`,
  `shared/lifecycle`, `shared/models.Base`, `shared/domain.Base`, `shared/authz`,
  `shared/handlers`, `shared/repository`, `shared/types.PageInfo`.
- **Archivos compartidos a patchear (fuera del flist)**: `wire/wire.go`,
  `wire/wire_gen.go`, `cmd/api/http_server.go`.
- **Función SQL compartida**: `public.normalize_actor_name` (debe coincidir con
  `normalizeName` Go).

## Recomendación de orden

1. 001 → 002 → 003 → 004 (base shared/platform/tenancy).
2. (010 projects si las tablas `project_*`/`projects` no están en develop).
3. **feature-007 BE** (este paquete + hunks de wire; o mergear 023 antes).
4. FE feature-007.
5. Después: 018 (data-integrity admin), 011 (campaign dto).
