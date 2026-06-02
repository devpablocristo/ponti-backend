# notes-for-future-agent.md — feature-007 actor-system (BE)

## Resumen corto

Feature grande, FULL-STACK. BE introduce el modelo canónico **Actor** (identidad única
por tenant para cliente/inversor/proveedor/responsable/arrendatario/contratista/
facturador), CRUD REST `/api/v1/actors`, sync legacy con `legacy_actor_map`, y 4
migraciones (223/226/231/234). Todo el código del paquete está en `internal/actor/**`
(SOURCE `777e5f6a` = `develop-problematico~1`). El cableado vive en archivos compartidos
**fuera del flist**.

## Qué está en FE y en BE

- **BE (este repo)**: `internal/actor/**` + `wire/actor_providers.go` + migraciones
  223/226/231/234 + (hunks fuera de flist) `wire/wire.go`, `wire/wire_gen.go`,
  `cmd/api/http_server.go`.
- **FE (otro repo, feature-007)**: `useActors`, `master-data/actors`, BFF
  `api/src/routes/actors.ts`. **BE-first**: mergear y desplegar BE antes que FE.

## Archivos esenciales

- `internal/actor/handler.go` — las 12 rutas (`Routes()`). Mirar primero para los endpoints.
- `internal/actor/handler/dto/actor.go` + `duplicate.go` — el contrato JSON con el FE.
- `internal/actor/usecases/domain/actor.go` — kinds/roles válidos, `Validate`, `IsArchived`.
- `internal/actor/repository.go` — tenancy (`requestTenantID`), unicidad,
  `normalizeName` (debe espejar la función SQL), `MergeActors`, `ListDuplicateCandidates`.
- `internal/actor/SPEC.md` — reglas de unicidad y tests SDD (leerlo entero, es la fuente).

## Archivos peligrosos / mezclados

- `internal/actor/legacy_sync.go` y `internal/actor/master_link.go`: SQL crudo contra
  `customers`, `projects`, `project_responsibles`, `project_investor_allocations`,
  `project_admin_cost_allocations`, `legacy_actor_map`. Acoplan con features 010/011/018.
- `migrations_v4/000223_*.up.sql`: crea tablas que "pertenecen" a projects (010) y
  backfillea desde legacy. Coordinar para no duplicar `CREATE TABLE` con el PR de 010.
- `migrations_v4/000234_*.up.sql`: **merge destructivo** de duplicados activos. Revisar en
  staging y `actor_merge_log` antes de prod.
- **FUERA del flist pero imprescindibles**: `wire/wire.go` (línea 49 `ActorHandler`),
  `wire/wire_gen.go` (75/340/380), `cmd/api/http_server.go` (148 `deps.ActorHandler.Routes()`).
  Sin estos, la feature compila pero no enruta.

## Decisiones ya tomadas

- Lifecycle unificado en `deleted_at` (migración 231 retiró `archived_at` de actors/roles/
  aliases). El DTO sigue exponiendo `archived_at` por compat con FE.
- Normalización del nombre duplicada a propósito en Go y SQL; cambiar las dos juntas.
- `ActorRole`/`ActorAlias` no embeben `sharedmodels.Base` (sin updated_at/created_by) —
  decisión documentada en `repository/models/actor.go`.
- Test de integración usa **sqlite in-memory** (no testcontainers): corre en CI sin docker.

## Dudas abiertas

- ¿`projects` y las tablas `project_*` ya existen en `develop`, o se crean acá (223)?
  En SOURCE las crea 223; coordinar con el extractor de 010.
- ¿La migración 234 (merge destructivo) aplica también a prod o solo dev? El SPEC la
  describe como consolidación dev. Confirmar con humano.
- ¿Incluir los hunks de wire en este PR o mergear 023 antes? Recomendado incluirlos acá
  porque sin ellos la feature no funciona.

## Qué comandos mirar primero

```bash
cat /tmp/flists/be-007.txt
git -C .../core show 777e5f6a:internal/actor/SPEC.md
git -C .../core show 777e5f6a:internal/actor/handler.go | head -90      # rutas
git -C .../core grep -n "ActorHandler" 777e5f6a -- wire/ cmd/api/        # cableado (fuera del flist)
git -C .../core show 777e5f6a:migrations_v4/000223_actors_safe_migration.up.sql | head -200
```

## Errores a evitar

- NO portar solo el flist: queda la feature "muerta" (404). Incluir los 3 hunks de wire.
- NO copiar `wire_gen.go` parcial a mano: regenerarlo (`go run github.com/google/wire/cmd/wire ./wire/...`).
- NO aplicar migraciones sin verificar que el set legacy + `projects`/`customers` existen.
- NO mergear el FE antes del BE.
- NO confundir `develop-problematico` (tip vacío/restore) con `develop-problematico~1`
  (`777e5f6a`, la fuente real).
- NO tocar otras migraciones ni código de otras features al portar.

## Camino más seguro

1. Asegurar deps 001/002/003/004 (y 010 si projects no está) en `develop`.
2. Rama `pr/feature-007-actor-system-be`.
3. Portar paquete entero + migraciones + provider + 3 hunks de wire (regenerar wire_gen).
4. `go build ./...`, `go test ./internal/actor/...`, validar migraciones en staging.
5. PR, merge, desplegar, `curl /api/v1/actors`.
6. Recién entonces, el FE feature-007.

## Qué PR del otro repo va antes/después

- **Antes**: ninguno del FE (es BE-first).
- **Después**: FE feature-007 (useActors / master-data/actors / BFF actors.ts).
- En este repo, **antes**: 001/002/003/004 (y 010 projects para las migraciones); 023 si
  no se incluyen los hunks de wire acá. **Después**: 010/011/018 se apoyan en `actors` y
  en los endpoints de duplicados/merge.
