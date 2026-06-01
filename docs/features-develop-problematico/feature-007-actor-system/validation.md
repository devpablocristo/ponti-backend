# validation.md — feature-007 actor-system (BE)

## Checklist pre-PR

- [ ] Dependencias en `develop`: existen `internal/shared/text.CanonicalizeName` (004),
      `internal/shared/lifecycle` (002), `internal/shared/models.Base` /
      `internal/shared/domain.Base`, `internal/shared/authz`, `internal/shared/handlers`,
      `internal/shared/repository`, y los paquetes `platform/*` en `go.mod`.
- [ ] Tablas legacy presentes: `customers`, `investors`, `managers`, `providers`; y
      `projects` (010) si se va a aplicar la migración 223/226.
- [ ] Paquete `internal/actor/**` + `wire/actor_providers.go` portados enteros.
- [ ] 4 migraciones up + down portadas.
- [ ] **Hunks de cableado** aplicados: `wire/wire.go` (`ActorHandler`), `wire/wire_gen.go`
      (`ProvideActorHandler` + asignación), `cmd/api/http_server.go`
      (`deps.ActorHandler.Routes()`). Verificar:
      `grep -rn "ActorHandler" wire/ cmd/api/` → 4 hits.
- [ ] `wire_gen.go` regenerado (preferido) con `go run github.com/google/wire/cmd/wire ./wire/...`.
- [ ] `go build ./...` ok.
- [ ] `go vet ./internal/actor/...` ok.
- [ ] `git diff --check` sin warnings de whitespace.
- [ ] El diff del PR no incluye código ajeno (otras features de `wire_gen.go`).

## Tests sugeridos (BE)

```bash
go -C /home/pablocristo/Proyectos/pablo/ponti/core test ./internal/actor/...
go -C /home/pablocristo/Proyectos/pablo/ponti/core test ./internal/actor/usecases/domain/...
go -C /home/pablocristo/Proyectos/pablo/ponti/core build ./...
```

Suites esperadas verdes:
- `usecases/domain/actor_test.go` — `TestIsValidKind`, `TestIsValidRole`,
  `TestActorValidate`, `TestActorIsArchived`.
- `usecases_test.go` — delegación create/list/get/update/archive/restore/harddelete/role/alias/merge.
- `handler_test.go` — parseo de filtros/paginación, `archived` forzado, rutas
  archive/restore/hard-delete, merge usa actor autenticado.
- `repository_tenant_test.go` (sqlite in-memory) — `TestActorRepositoryTenantIsolation`,
  `TestCreateActorDuplicateNormalizedNameReturnsConflict`,
  `TestUpdateActorDuplicateNormalizedNameReturnsConflict`.

## Validación de migraciones

```bash
# aplicar y revertir sobre una DB con el set legacy (NO ejecutar en prod sin staging)
# up:   223 -> 226 -> 231 -> 234
# down: 234 -> 231 -> 226 -> 223
```

- [ ] 223 crea `actors`, perfiles, `actor_roles/aliases/identifiers/relationships/merge_log`,
      `legacy_actor_map`, `project_responsibles`, `project_investor_allocations`,
      `project_admin_cost_allocations`, `field_lease_participants`, función
      `normalize_actor_name`; backfill 1:1 desde legacy sin tocar lecturas productivas.
- [ ] 226 agrega `customers.actor_id` (FK NOT VALID), `projects.customer_actor_id`,
      índices, y el índice no-único `idx_actors_tenant_normalized_name`.
- [ ] 231 backfillea `archived_at`→`deleted_at` y dropea `archived_at` en
      actors/roles/aliases; los modelos GORM ya esperan `deleted_at`.
- [ ] 234 crea el índice **único** parcial `(tenant_id, normalized_name) WHERE deleted_at
      IS NULL AND merged_into_actor_id IS NULL` y consolida duplicados activos.
      Revisar `actor_merge_log` tras aplicar.
- [ ] Cada `.down.sql` revierte sin error.

## Manual / API

```bash
# tras levantar el server (requiere cableado portado)
curl -s "$BASE/api/v1/actors" | jq             # 200 + {data, page_info}
curl -s "$BASE/api/v1/actors/archived" | jq    # solo archivados
curl -s "$BASE/api/v1/actors/duplicate-candidates" | jq
curl -X POST "$BASE/api/v1/actors" -d '{"display_name":"Juan Perez","actor_kind":"natural_person"}'  # 201
# crear otro con mismo nombre normalizado -> 409 "actor already exists"
```

## Casos borde

- Crear actor con `display_name` que normaliza igual a otro activo → 409.
- Editar actor para tomar el nombre de otro activo → 409; mantener el propio → OK.
- Mismo nombre normalizado en otro `tenant_id` → permitido.
- Nombre de un actor archivado (`deleted_at` set) o fusionado (`merged_into_actor_id`)
  → no bloquea crear uno nuevo.
- `role` inválido en filtro de listado → 400 (`invalid actor role`).
- Merge con `confirm=false` → preview (impacto sin ejecutar); `confirm=true` → ejecuta.
- `actor_kind` fuera de `{natural_person, organization, other, unknown}` → validación falla.

## Qué revisar en UI / API / DB / env

- **API**: shape de `ActorResponse` (incluye `archived_at`, `normalized_name`, `roles`,
  `aliases`, `identifiers`, `person_profile`, `organization_profile`), `ListActorsResponse`
  (`data` + `page_info`), `DuplicateCandidateResponse`, `MergeImpactResponse`.
- **DB**: índice único parcial existe; función `normalize_actor_name` definida;
  `customers.actor_id` y `projects.customer_actor_id` poblados donde corresponde.
- **env**: confirmar que el flag de sync legacy (`actorSyncDisabled`) está en el estado
  esperado por entorno (revisar cómo se setea antes de aplicar el backfill).

## Qué validar en el otro repo (FE feature-007)

- `useActors` apunta a `/api/v1/actors` y mapea los campos del DTO sin renombrar.
- BFF `api/src/routes/actors.ts` proxyea los 12 endpoints.
- Páginas `master-data/actors` (lista/detalle/duplicados/merge) contra el BE desplegado.
- Mergear FE **después** del BE.

## Señales de incompletitud / incompatibilidad

- `curl /api/v1/actors` → 404 ⇒ falta cableado.
- build falla por `shared/*`/`platform/*` ⇒ falta una dependencia (001-004).
- migración 223 aborta por tabla `projects`/`customers` inexistente ⇒ falta 010 / set legacy.
- FE muestra error de red en master-data/actors ⇒ BE no desplegado o ruta no registrada.
