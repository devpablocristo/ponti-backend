# extraction-plan.md — feature-007 actor-system (BE)

- **repo**: ponti-backend (core)
- **rama base**: `develop` (tip `003a9b8f`)
- **SOURCE**: `develop-problematico~1` (SHA `777e5f6a`). NUNCA `develop-problematico` (tip restore/vacío).
- **rama sugerida**: `pr/feature-007-actor-system-be`
- **orden cross-repo**: **BE-first**, luego FE feature-007.

## PR title

`feat(be): actor system — canonical identity, /api/v1/actors, legacy sync (migr 223/226/231/234)`

## PR description (borrador)

> Introduce el modelo canónico **Actor** (personas/sociedades) como identidad única
> por tenant para los roles cliente/inversor/proveedor/responsable/arrendatario/
> contratista/facturador.
>
> - CRUD REST `/api/v1/actors` (+ archive/restore/hard-delete, roles, aliases, merge,
>   duplicate-candidates).
> - Unicidad por `(tenant_id, normalized_name)` para actores activos; normalización
>   espejada en Go (`normalizeName`) y SQL (`normalize_actor_name`).
> - Sync legacy (`legacy_sync.go`, `master_link.go`) + `legacy_actor_map`, backfill 1:1
>   sin cambiar lecturas productivas.
> - Migraciones 223 (modelo+backfill), 226 (customers.actor_id), 231 (consolidación
>   `deleted_at`), 234 (índice único + merge de duplicados).
> - Wiring: providers en `wire/actor_providers.go`; hunks en `wire/wire.go`,
>   `wire/wire_gen.go` y `cmd/api/http_server.go` para registrar el handler.
>
> Depende de: 001 platform-tenancy, 002 crudar-lifecycle, 003 multitenant-db, 004
> shared-text. Coordinar con projects (010) por las tablas `project_*`. FE feature-007
> va después.

## Pasos ordenados

1. **Pre-requisitos**: confirmar en `develop` que existen `internal/shared/text.CanonicalizeName`
   (004), `internal/shared/lifecycle` (002), `shared/models.Base`/`shared/domain.Base`,
   `shared/authz`, `shared/handlers`, `shared/repository`, y que las tablas legacy
   (`customers`, `investors`, `managers`, `providers`) y `projects` existen (010).
   Si falta algo → mergear esa dependencia antes.
2. Crear la rama desde `develop`.
3. Traer el paquete entero `internal/actor/**` + `wire/actor_providers.go` desde SOURCE.
4. Traer las 4 migraciones (up + down).
5. Aplicar los **hunks de cableado** (`wire/wire.go`, `wire/wire_gen.go`,
   `cmd/api/http_server.go`). Preferir **regenerar `wire_gen.go`** con `go run github.com/google/wire/cmd/wire ./wire/...` para evitar arrastrar generaciones de otras features.
6. `go build ./...` + `go vet ./internal/actor/...`.
7. `go test ./internal/actor/...`.
8. `git diff --check` (whitespace) y revisar que el diff no incluya código ajeno.
9. Validar migraciones up/down contra una DB con el set legacy (ver `validation.md`).
10. Abrir PR.

## Archivos enteros vs parciales

- **Enteros**: todo `internal/actor/**`, `wire/actor_providers.go`, las 8 SQL.
- **Parciales (hunks)**: `wire/wire.go` (línea `ActorHandler *actor.Handler` + sets),
  `cmd/api/http_server.go` (línea `deps.ActorHandler.Routes()`). `wire/wire_gen.go`:
  preferir regenerar; si no, hunks de `ProvideActorHandler` + asignación en struct deps.

## Migraciones / tests a incluir

- Migraciones: 223, 226, 231, 234 (up + down).
- Tests: las 4 suites del paquete. `repository_tenant_test.go` usa **sqlite in-memory**,
  corre en CI sin docker.

## Dependencias previas (deben estar en develop antes)

001 platform-tenancy → 002 crudar-lifecycle → 003 multitenant-db → 004 shared-text.
Coordinar 010 projects (tablas `project_*` que toca la migración 223).
Coordinar 023 wire-di si los hunks de wire no se incluyen acá.

## Coordinación con el otro repo

- **BE-first**. Mergear este PR, desplegar, verificar `/api/v1/actors`, luego FE feature-007.
- El FE consume `ActorResponse`, `DuplicateCandidateResponse`, `MergeImpactResponse` y
  paginación `data + page_info`. No cambiar esos JSON sin coordinar con FE.

## Comandos git SUGERIDOS (para un humano — NO ejecutados acá)

```bash
# situarse y crear rama
git -C /home/pablocristo/Proyectos/pablo/ponti/core checkout develop
git -C /home/pablocristo/Proyectos/pablo/ponti/core pull
git -C /home/pablocristo/Proyectos/pablo/ponti/core checkout -b pr/feature-007-actor-system-be

# paquete + provider (archivos enteros nuevos)
git -C .../core checkout develop-problematico~1 -- internal/actor wire/actor_providers.go

# migraciones (up + down)
git -C .../core checkout develop-problematico~1 -- \
  migrations_v4/000223_actors_safe_migration.up.sql \
  migrations_v4/000223_actors_safe_migration.down.sql \
  migrations_v4/000226_customer_actor_master_link.up.sql \
  migrations_v4/000226_customer_actor_master_link.down.sql \
  migrations_v4/000231_consolidate_actor_archived_at.up.sql \
  migrations_v4/000231_consolidate_actor_archived_at.down.sql \
  migrations_v4/000234_actor_unique_normalized_name.up.sql \
  migrations_v4/000234_actor_unique_normalized_name.down.sql

# cableado: parcial (elegir SOLO los hunks del actor)
git -C .../core restore -p --source=develop-problematico~1 -- wire/wire.go cmd/api/http_server.go
# wire_gen.go: preferible regenerar
#   go run github.com/google/wire/cmd/wire ./wire/...
# (o, si se porta a mano:)  git restore -p --source=develop-problematico~1 -- wire/wire_gen.go

# verificación
git -C .../core diff --check
go -C .../core build ./...
go -C .../core test ./internal/actor/...
```

## Qué NO traer

- Hunks de `wire/wire_gen.go` que no sean del actor (otras features).
- Otras migraciones `migrations_v4/*` fuera de las 4.
- Nada del FE.
- json-tags cleanup del dominio (eso es feature-027).

## Qué podría romperse

- `go build` falla si 001/002/003/004 no están (imports `shared/*`, `platform/*`).
- `wire` no compila si se traen hunks parciales inconsistentes de `wire_gen.go` → regenerar.
- Migración 223/226 falla si `projects`/`project_*`/`customers` no existen (orden con 010).
- Migración 234 hace merge destructivo de duplicados activos en datos existentes.

## Cómo detectar extracción incompleta

- `grep -rn "ActorHandler" wire/ cmd/api/` debe devolver 4 hits (`wire.go:49`,
  `wire_gen.go:75/340/380`, `http_server.go:148`). Si faltan, las rutas no se registran.
- Levantar el server y `curl /api/v1/actors` → 200/list. Si 404, falta `Routes()`.
- `go test ./internal/actor/...` rojo por símbolos `shared/*` faltantes → falta una dep.

## Qué validar antes del PR

Ver `validation.md` (checklist completo).

## Qué hacer después de mergear

- Aplicar migraciones en cada entorno; verificar `legacy_actor_map` poblado.
- Disparar el FE feature-007 (BE-first).
- Avisar a 010/011/018 que `actors` + `project_*` ya están disponibles.
