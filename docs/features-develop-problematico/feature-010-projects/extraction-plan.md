# extraction-plan.md — feature-010 projects (BE)

- **repo:** `ponti-backend` (`/home/pablocristo/Proyectos/pablo/ponti/core`)
- **rama base:** `develop` (tip `003a9b8f`)
- **SOURCE de extracción:** `develop-problematico~1` (SHA `777e5f6a`). NUNCA `develop-problematico` (tip restore/vacío).
- **rango de verdad:** `0972e565..777e5f6a`
- **rama sugerida:** `pr/feature-010-projects-be`
- **orden cross-repo:** BE-first, luego FE.

## PR title
`feat(be): projects multi-tenant + actor-sync + lifecycle archive/hard-delete (#010)`

## PR description (borrador)
Migra `internal/project` a multi-tenant y lo integra con el actor-system y el lifecycle/archive framework:
- Tenancy scope (`tenancy.Scope`) y filtros `tenant_id` en todas las queries del grafo de project.
- Actor-sync de customer/manager/investor (`SyncLegacyActor`, `EnsureCustomerFromActor`, `EnsureLegacyEntityFromActor`, `RefreshProjectActorMirrors`) + hidratación de `actor_id` legacy en `GetProject`.
- Cascada de archive/restore centralizada en `internal/shared/lifecycle` (reemplaza la cascada artesanal, corrige el bug de pluck-after-archive que perdía `workorder_items`/`work_order_draft_items`).
- `DeleteProject` → `HardDeleteProject`: ruta `DELETE /:project_id/hard`, exige proyecto archivado y sin dependientes activos (409).
- `GetProjectByNameAndCampaignID` → `GetProjectByNameCustomerAndCampaignID`.
- DTO/modelos con `ActorID`; nombres canonicalizados con `shared/text`.
- Tests sqlite in-memory: tenant isolation, rename propagation, archived-refs guard, rutas de acción.

**Depende de (deben estar en develop antes):** bump `platform/persistence/gorm/go` (tenancy), 004 (shared-text), 008 (authz/tenant-context), 007 (actor-system), 009 (lifecycle/archive-surface). **Coordinar con FE feature-010** (cambio de path `/hard`, `actor_id`, 409s).

## Pasos ordenados
1. **Pre-requisito (NO opcional):** confirmar que `develop` ya tiene mergeados, en este orden, los PRs de: bump tenancy de plataforma, feature-004, feature-008, feature-007, feature-009. Sin esto el módulo NO compila.
   - Verificación: `git show develop:internal/actor/legacy_sync.go >/dev/null && git show develop:internal/shared/lifecycle/cascade.go >/dev/null && git show develop:internal/shared/authz/authz.go >/dev/null && git show develop:internal/shared/text/propername.go >/dev/null && grep -q 'platform/persistence/gorm/go' go.mod`
2. Crear rama desde develop.
3. Traer los 9 archivos enteros desde el SOURCE (`develop-problematico~1`).
4. `go build ./internal/project/...` y resolver imports faltantes (deberían estar todos si el paso 1 está OK).
5. `go test ./internal/project/...`.
6. `git diff --check` (whitespace/conflict markers).
7. Abrir PR contra develop con la description de arriba.

## Archivos enteros vs parciales
- **Enteros (whole-file), los 9:** `internal/project/handler.go`, `internal/project/handler/dto/project.go`, `internal/project/usecases.go`, `internal/project/repository.go`, `internal/project/repository/models/project.go`, `internal/project/repository_archived_refs_test.go`, `internal/project/repository_rename_test.go`, `internal/project/repository_tenant_test.go`, `internal/project/handler_integration_test.go`.
- **Parciales:** ninguno en este paquete.

## Migraciones / tests a incluir
- **Migraciones:** ninguna en mi flist; las aportan las features de dependencia (tenant_id, actor_id, tablas actors/legacy_actor_map, columnas archive_*). NO crear migraciones nuevas en este PR.
- **Tests:** incluir los 3 nuevos + el integration modificado (son parte del whole-file extract).

## Comandos git SUGERIDOS (para un humano; este agente NO los ejecuta)
```sh
# 0) Verificar dependencias presentes en develop (debe imprimir todo OK)
git -C ~/Proyectos/pablo/ponti/core show develop:internal/actor/legacy_sync.go >/dev/null 2>&1 && echo actor-ok
git -C ~/Proyectos/pablo/ponti/core show develop:internal/shared/lifecycle/cascade.go >/dev/null 2>&1 && echo lifecycle-ok
git -C ~/Proyectos/pablo/ponti/core show develop:internal/shared/authz/authz.go >/dev/null 2>&1 && echo authz-ok
git -C ~/Proyectos/pablo/ponti/core show develop:internal/shared/text/propername.go >/dev/null 2>&1 && echo text-ok
grep -q 'platform/persistence/gorm/go' <(git -C ~/Proyectos/pablo/ponti/core show develop:go.mod) && echo tenancy-ok

# 1) rama
git checkout develop
git checkout -b pr/feature-010-projects-be

# 2) traer los 9 archivos enteros del SOURCE
git checkout develop-problematico~1 -- \
  internal/project/handler.go \
  internal/project/handler/dto/project.go \
  internal/project/usecases.go \
  internal/project/repository.go \
  internal/project/repository/models/project.go \
  internal/project/repository_archived_refs_test.go \
  internal/project/repository_rename_test.go \
  internal/project/repository_tenant_test.go \
  internal/project/handler_integration_test.go

# 3) validar
git diff --check
go build ./internal/project/...
go test ./internal/project/...
```
> No usar `git restore -p` aquí: no hay archivos compartidos mezclados en mi flist, así que el extract es por archivo entero.

## Qué NO traer
- Nada fuera de `internal/project/`. Las dependencias (`internal/actor`, `internal/shared/{lifecycle,authz,text}`, `platform/persistence/gorm/go`, domain de customer/manager/investor con `ActorID`) vienen de SUS features, no de este PR.
- No reintroducir `lot-metrics`/`tentative-prices` (ya DONE en otros archivos del módulo).

## Qué podría romperse
- Compilación si falta cualquier dependencia (alto si se ignora el paso 1).
- FE feature-010 si se mergea BE sin coordinar el path `/hard` (mitiga: BE-first + aviso a FE).
- `internal/data-integrity` (018) si esperaba `GetRawAdminCostTotal` con otra firma — verificar que 018 no se mergeó antes con un stub.

## Cómo detectar extracción incompleta
- `go build ./internal/project/...` falla con `undefined: actorsync.*` / `lifecycle.*` / `authz.*` / `tenancy.*` / `text.CanonicalizeName` → falta una dependencia.
- `undefined: (customerdom.Customer).ActorID` → falta el domain con `ActorID` (007/011).
- `pq: column "tenant_id" does not exist` en runtime → faltan migraciones de las deps.

## Qué validar antes del PR
- `go build ./...` del repo entero (no sólo project), `go vet ./internal/project/...`, `go test ./internal/project/...`.
- Que `DELETE /:project_id/hard` esté registrada y `DELETE /:project_id` ausente.

## Qué hacer después de mergear
- Avisar/coordinar el merge de FE feature-010 (path + actor_id + 409s).
- Confirmar con 018 que `GetRawAdminCostTotal` está disponible (no duplicar).
