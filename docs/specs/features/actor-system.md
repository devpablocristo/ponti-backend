# spec — actor-system (feature-007)

> **Spec definitivo** re-baselineado contra `develop` real (tip `19b96dc4`). Fuente: `777e5f6a`
> (= `develop-problematico~1`). NO se implementa acá; es el contrato de qué traer y cómo.

- **id / slug:** feature-007 · `actor-system`
- **tipo:** feature FULL-STACK (este spec cubre el **BE**; el FE es el feature-007 del repo web)
- **fuente:** `777e5f6a` · **destino:** `develop` (`19b96dc4`)
- **orden en el cluster tenants/users:** **3º** (usuarios/identidad; después de 001+003, antes de 008)

---

## 1. Propósito

Introducir el modelo canónico **Actor** como identidad única por tenant (personas/sociedades) para todos
los roles del dominio (cliente, inversor, proveedor, responsable, arrendatario, contratista, facturador),
con CRUD REST `/api/v1/actors`, merge de duplicados y sync legacy 1:1.

## 2. Estado vs `develop` (diff real re-baselineado)

Verificado contra develop actual — la feature está al **0%**:
- `git ls-tree develop -- internal/actor/` → **vacío**. El paquete no existe.
- `git grep "ActorHandler" develop -- wire/ cmd/` → **0**. No hay cableado.
- Migraciones 223/226/231/234 → **ausentes** en develop. (develop tiene un `000231` propio, ver §4.)

**⚠️ Las deps de compilación de 007 TAMPOCO están en develop** (verificado):

| símbolo / paquete | feature dueña | ¿en develop? |
|---|---|---|
| `internal/shared/text.CanonicalizeName` | **004** shared-text-propername | **NO** |
| `internal/shared/lifecycle.*` (soft-delete `deleted_at`) | **002** crudar-lifecycle | **NO** |
| `internal/shared/authz.*` | **001** platform-tenancy | **NO** (lo trae 001) |
| `internal/shared/models.Base` / `shared/handlers` / `shared/repository` | 001/002/023 | parcial |

⇒ **007 no compila standalone sobre este develop.** Necesita 001, **002** y **004** antes (002 y 004
**no** están en el scope "tenants y users" que elegiste — son prerequisitos a recuperar aparte).
Las tablas legacy (`projects` migr 020, `investors` 070, customers, etc.) **sí** existen en develop.

## 3. Alcance / archivos

Flist: `/tmp/flists/be-007.txt` (24 paths, todos `A`/nuevos). Tres bloques:

### 3A. Paquete `internal/actor/**` + provider — whole-file (nuevos)
```
internal/actor/SPEC.md                          (reglas de unicidad + tests SDD — leer primero)
internal/actor/handler.go                       (12 rutas REST + Routes())
internal/actor/handler/dto/{actor,duplicate}.go (contrato JSON con el FE)
internal/actor/usecases.go
internal/actor/usecases/domain/actor.go         (kinds/roles, Validate, IsArchived)
internal/actor/repository.go                    (tenancy, normalizeName, MergeActors, ListDuplicateCandidates)
internal/actor/repository/models/actor.go       (mapeo GORM, deleted_at post-231)
internal/actor/legacy_sync.go                   (backfill legacy_actor_map + actors + project_* desde legacy)
internal/actor/master_link.go                   (EnsureCustomerFromActor; toca customers/projects)
wire/actor_providers.go                         (ActorSet)
internal/actor/{handler_test,usecases_test,repository_tenant_test}.go
internal/actor/usecases/domain/actor_test.go    (repository_tenant_test usa sqlite in-memory, sin docker)
```
> `legacy_sync.go`/`master_link.go` son archivos nuevos (whole-file) pero su **SQL referencia tablas de
> otras features** (customers/projects/project_*). El riesgo no es el diff sino la coexistencia de tablas.

### 3B. Cableado — partial-hunks (FUERA del flist, imprescindibles)
Sin estos el paquete compila pero **no enruta** (`/api/v1/actors` → 404). Confirmado en `777e5f6a`:
`wire/wire.go:49`, `wire/wire_gen.go:75/340/380`, `cmd/api/http_server.go:148`.
```
wire/wire.go               (hunk: ActorHandler *actor.Handler en deps)   → git restore -p
wire/wire_gen.go           (ProvideActorHandler + asignación)            → PREFERIR REGENERAR
cmd/api/http_server.go     (deps.ActorHandler.Routes())                  → git restore -p
```
Regenerar wire: `go run github.com/google/wire/cmd/wire ./wire/...` (no copiar hunks generados a mano).
`grep -rn "ActorHandler" wire/ cmd/api/` debe dar **4 hits**.

### 3C. Migraciones — whole-file (up + down siempre juntos)
```
000223_actors_safe_migration            (crea modelo actor + tablas project_* + función normalize_actor_name + backfill 1:1)
000226_customer_actor_master_link        (customers.actor_id, projects.customer_actor_id, FK NOT VALID, idx no-único)
000231_consolidate_actor_archived_at     (archived_at → deleted_at en actors/roles/aliases)
000234_actor_unique_normalized_name      (índice único parcial (tenant_id, normalized_name) + MERGE DESTRUCTIVO de duplicados activos)
```

### EXCLUIR
- json-tags cleanup del dominio → 027. Cualquier migración fuera de las 4. Todo el FE.

## 4. Migraciones — riesgo de numeración (BLOQUEANTE, re-baselineado)

Mismo problema que 003, agravado por la **colisión dura en 231**:
- develop ocupa `000231` con **`lot_yield_over_total_hectares`** (de #125/#127); el source usa `000231`
  para **`consolidate_actor_archived_at`**. **No se puede traer el 231 del source tal cual.**
- 223/226 (< 231) tampoco se aplican: `golang-migrate` está en 231.
- ⇒ Renumerar las 4 a **> 231** preservando el orden `223<226<231<234` y respetando que **224/225 (003)
  van antes** que 226/231/234 (que dependen de `tenant_id`). Mapeo sugerido (coordinado con 003 — §4 de
  [multitenant-db-hardening.md](multitenant-db-hardening.md)):

  | source | → nuevo nº sugerido | nota |
  |--------|---------------------|------|
  | 223 actors_safe_migration | 232 | crea modelo actor (después de 003 = 233/234) — **ojo**: si 003 va a 233/234, 223 debe ir DESPUÉS porque actors referencia `auth_tenants`/`tenant_id`. Revisar: 223 quizás → 235 |
  | 226 customer_actor_master_link | 235→ | depende de 224 (tenant_id) y de actors (223) |
  | 231 consolidate_actor_archived_at | 238→ | |
  | 234 actor_unique_normalized_name | 241→ | merge destructivo, va último |

  > **Decisión de orden abierta:** 223 (actors) requiere que `tenant_id`/`auth_tenants` existan (003) →
  > 003 debería numerarse **antes** que 223. El mapeo del spec de 003 (003=233/234, 223=232) **viola** eso.
  > Hay que fijar el orden global del bloque al implementar: lo más seguro es **003 primero (232/233),
  > luego 223 actors (234), luego 226/231/234**. Esta es una decisión de orquestación para el humano.

## 5. Dependencias

- **Compile (DURO, faltan en develop):** 001 (`shared/authz`, platform en go.mod), **002**
  (`shared/lifecycle`, `shared/models.Base`), **004** (`shared/text.CanonicalizeName`). Sin las tres, 007
  no compila. → recuperar/portar 002 y 004 **además** de 001/003.
- **Runtime/DB:** 003 (columnas `tenant_id` + `auth_tenants`); tablas legacy + `projects` (010) para el
  backfill de 223/226 (projects = migr 020, ya en develop).
- **Cableado:** 023 (wire) — o incluir los 3 hunks acá (recomendado).
- **007 desbloquea:** 008 (/me usa el modelo de identidad), 010 (projects usa `project_*`), 011, 018.
- **Cross-repo:** FE feature-007 (`useActors`, `master-data/actors`, BFF `actors.ts`) — **BE-first**.

## 6. Plan de implementación (pasos, sin ejecutar acá)

1. **Asegurar deps en develop:** 001, **002**, **004** (compile), 003 (runtime/DB). Si no están, portarlas
   antes (002 y 004 quedan fuera del scope tenants/users — avisar al humano).
2. Fijar la **renumeración global** del bloque de migraciones (§4) — coordinar con 003.
3. `git checkout 777e5f6a -- internal/actor wire/actor_providers.go` + las 8 SQL; `git mv` a la numeración nueva.
4. Cableado: `git restore -p` de `wire/wire.go` + `cmd/api/http_server.go`; **regenerar** `wire_gen.go`.
5. `go build ./...` + `go vet ./internal/actor/...` + `go test ./internal/actor/...`.
6. Validar migraciones up/down en DB con set legacy (revisar `actor_merge_log` tras 234).
7. PR BE → desplegar → `curl /api/v1/actors` → recién entonces el FE feature-007.

## 7. Validación

- `go build ./...` (falla por `shared/*`/`platform/*` ⇒ falta 001/002/004).
- `go test ./internal/actor/...` — suites: `actor_test` (dominio: `TestIsValidKind/IsValidRole/ActorValidate/
  ActorIsArchived`), `usecases_test`, `handler_test`, `repository_tenant_test` (sqlite in-memory:
  `TestActorRepositoryTenantIsolation`, `TestCreateActorDuplicateNormalizedNameReturnsConflict`,
  `TestUpdateActorDuplicateNormalizedNameReturnsConflict`).
- `grep -rn "ActorHandler" wire/ cmd/api/` → 4 hits; server up + `curl /api/v1/actors` → 200 `{data,page_info}`.
- Migraciones up `223→226→231→234` y down `234→231→226→223` sin error (sobre los nº renumerados).
- SDD: duplicado por normalized_name → 409; tomar nombre de otro activo → 409; mantener el propio → OK;
  mismo nombre en otro tenant → OK; nombre de actor archivado/fusionado no bloquea.

## 8. Riesgos y decisiones pendientes

- **Deps de compilación 002 + 004 fuera de scope (ALTO):** elegiste "tenants y users" (001/003/007/008),
  pero 007 **no compila** sin 002 (lifecycle) y 004 (text). Decisión: recuperarlas/portarlas también, o
  aceptar que 007 espera a que entren.
- **Colisión/numeración de migraciones (ALTO):** 231 colisiona; orden 003-vs-223 a fijar (§4).
- **Migración 234 = merge DESTRUCTIVO** de duplicados activos (conserva menor id, marca el resto
  `merged_into_actor_id`+`deleted_at`, loguea en `actor_merge_log`). Revisar en staging; confirmar con
  humano si aplica a prod. Su `.down.sql` **no deshace el merge** (solo índices).
- **Cableado fuera del flist (ALTO):** portar solo el flist deja la feature "muerta" (404). Incluir los 3
  hunks; regenerar `wire_gen.go`.
- **223 crea tablas `project_*` que "pertenecen" a 010:** coordinar para no duplicar `CREATE TABLE` con el
  PR de projects.
- **normalizeName (Go) ≡ normalize_actor_name (SQL):** trim→lower→unaccent→colapsar espacios. Si se cambia
  una, cambiar la otra.
- **FE-first prohibido:** mergear el FE feature-007 antes del BE → 404. Respetar BE-first.
