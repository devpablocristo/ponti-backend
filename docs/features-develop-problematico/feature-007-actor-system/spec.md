# spec.md — feature-007 actor-system (Backend Go)

- **id**: feature-007
- **slug**: actor-system
- **nombre**: Actor system
- **tipo**: feature (FULL-STACK)
- **repo**: ponti-backend (core) — `/home/pablocristo/Proyectos/pablo/ponti/core`
- **existe-en-FE**: SÍ (useActors + master-data/actors + BFF `api/src/routes/actors.ts`) — paquete feature-007 del repo FE
- **existe-en-BE**: SÍ (este paquete)
- **rama destino**: `develop` (tip `003a9b8f`)
- **SOURCE de extracción**: `develop-problematico~1` (SHA `777e5f6a`). NUNCA usar `develop-problematico` (tip = restore/vacío).
- **rango diff fuente-de-verdad**: `0972e565..777e5f6a`

## Resumen

Introduce el modelo canónico **Actor** como identidad única (personas y sociedades)
para todos los roles del dominio: cliente, inversor, proveedor, responsable,
arrendatario, contratista, facturador. Reemplaza la dispersión de identidades en
tablas legacy (`customers`, `investors`, `managers`, `providers`, contratistas en
`workorders`/`labors`, compañías en `invoices`) por una sola tabla `actors` con
perfiles (persona/organización), aliases, identificadores, roles y un sistema de
detección/merge de duplicados.

Expone un CRUD REST completo bajo `/api/v1/actors` (+ acciones de lifecycle CRUDAR
y merge), un mecanismo de **sincronización legacy** (`legacy_sync.go` +
`master_link.go`) que mantiene `legacy_actor_map` y backfillea referencias
operativas (`project_responsibles`, `project_investor_allocations`, etc.), y 4
migraciones (`000223`, `000226`, `000231`, `000234`).

## Objetivo

- Tener una sola fuente de verdad de identidad por `tenant_id`.
- Garantizar unicidad por `(tenant_id, normalized_name)` para actores activos
  (no archivados ni fusionados), independiente de kind/roles/perfiles/contacto.
- Permitir migración progresiva: el modelo nuevo coexiste con las tablas legacy
  vía `legacy_actor_map` y backfill 1:1 (la migración 223 declara explícitamente
  que **no cambia lecturas productivas ni fórmulas existentes**).

## Problema

El dominio tenía la identidad de una misma persona/empresa repetida en múltiples
tablas (un "Juan Pérez" cliente y proveedor eran filas distintas, sin vínculo).
Esto impedía deduplicar, reportar por entidad y migrar a Projects/allocations
basados en actor.

## Alcance en este repo (BE)

- Paquete `internal/actor/**` (handler, dto, usecases, domain, repository, models,
  sync legacy, master-link, tests).
- Provider de wire `wire/actor_providers.go` (`ActorSet`).
- Migraciones `migrations_v4/000223, 000226, 000231, 000234` (up + down).
- **Dependencia de cableado fuera del flist** (ver riesgos): el módulo solo es
  alcanzable si se patchean también `wire/wire.go`, `wire/wire_gen.go` y
  `cmd/api/http_server.go` (línea `deps.ActorHandler.Routes()`). Esos archivos NO
  están en mi flist y pertenecen a feature-023 / shared.

## Alcance en el otro repo (FE)

- Hook `useActors`, páginas `master-data/actors`, ruta BFF `api/src/routes/actors.ts`.
- Consume los contratos definidos acá (`/api/v1/actors`, DTOs `ActorResponse`,
  `DuplicateCandidateResponse`, `MergeImpactResponse`).
- Debe mergearse DESPUÉS del BE (merge = BE-first).

## Fuera de alcance

- No se modifican lecturas productivas legacy ni fórmulas (la 223 lo dice explícito).
- No se elimina ninguna tabla legacy (customers/investors/managers/providers
  siguen existiendo; solo se les agrega `actor_id` en 226 para customers).
- La limpieza de json-tags del dominio (reports-dark-mode BE) NO es de esta feature
  (va en feature-027).

## Comportamiento esperado

Endpoints bajo `APIBaseURL()+"/actors"` (grupo con middlewares de validación):

| Método | Ruta | Handler | Acción |
|---|---|---|---|
| POST | `/actors` | `CreateActor` | crea actor (201 + id) |
| GET | `/actors` | `ListActors` | lista activos; query `status,role,q,page,per_page` |
| GET | `/actors/archived` | `ListArchivedActors` | fuerza `status=archived` |
| GET | `/actors/duplicate-candidates` | `ListDuplicateCandidates` | grupos candidatos a merge |
| POST | `/actors/merge` | `MergeActors` | merge (preview si `confirm=false`) |
| GET | `/actors/:actor_id` | `GetActor` | detalle |
| PUT | `/actors/:actor_id` | `UpdateActor` | actualiza (204) |
| POST | `/actors/:actor_id/archive` | `ArchiveActor` | soft-delete (204) |
| POST | `/actors/:actor_id/restore` | `RestoreActor` | restaura (204) |
| DELETE | `/actors/:actor_id/hard` | `HardDeleteActor` | borrado físico (204) |
| POST | `/actors/:actor_id/roles` | `AddRole` | agrega rol |
| POST | `/actors/:actor_id/aliases` | `AddAlias` | agrega alias |

- Unicidad: `CreateActor`/`UpdateActor` devuelven `domainerr.Conflict("actor already exists")`
  cuando la DB rechaza el índice único parcial.
- Normalización del nombre: trim → lower → unaccent → colapsar espacios. **Idéntica
  en Go (`normalizeName`, repository.go:1094) y en SQL (`public.normalize_actor_name`,
  migración 223)**. Si se cambia una, hay que cambiar la otra.
- Tenant: `requestTenantID(ctx)` extrae `OrgID` del contexto (multitenant, feature-003).

## Estado en dp~1 (777e5f6a)

- Código del paquete: **completo y autocontenido** (handler→usecases→repository→models→domain).
- Tests: 4 archivos, incluyen unit (mocks), domain puro y un test de aislamiento de
  tenant + conflicto de duplicado con **sqlite in-memory** (no requiere docker).
- Migraciones: las 4 con up + down.
- Wiring (`wire.go`/`wire_gen.go`/`http_server.go`): **presente en el árbol fuente pero
  en archivos fuera del flist** (shared). Sin esos hunks, el paquete compila pero las
  rutas no se registran y el handler no se inyecta.

## Criterios de aceptación

1. `go build ./...` y `go vet ./internal/actor/...` ok tras el porte (incluye hunks de wire).
2. `go test ./internal/actor/...` verde (4 suites).
3. Las 4 migraciones aplican y revierten limpio sobre una DB con el set legacy.
4. SDD del SPEC.md: duplicado por normalized_name → conflicto; editar tomando nombre
   de otro activo → conflicto; mantener su propio nombre → OK; mismo nombre en otro
   tenant → OK; nombre de actor archivado/fusionado no bloquea.
5. `/api/v1/actors` responde (verificable solo si se portan los hunks de wiring).

## Artefactos afectados

- **Endpoints**: 12 rutas REST (tabla arriba).
- **Modelos/DTOs**: `dto.ActorRequest/ActorResponse`, `AliasRequest/Response`,
  `IdentifierRequest/Response`, `PersonProfileRequest/Response`,
  `OrganizationProfileRequest/Response`, `RoleRequest`, `MergeRequest`,
  `DuplicateCandidateResponse`, `MergeImpactResponse`, `ListActorsResponse`.
- **Domain**: `domain.Actor` (+ `ActorAlias/Identifier/PersonProfile/OrganizationProfile`),
  `ListFilters`, `MergeRequest/MergeImpact`, `DuplicateCandidate/DuplicateActor`,
  `ValidKinds/ValidRoles`, `IsValidKind/IsValidRole/Validate/IsArchived`.
- **DB tables** (creadas por 223): `actors`, `actor_person_profiles`,
  `actor_organization_profiles`, `actor_identifiers`, `actor_roles`,
  `actor_aliases`, `actor_relationships`, `actor_merge_log`, `legacy_actor_map`,
  `project_responsibles`, `project_investor_allocations`,
  `project_admin_cost_allocations`, `field_lease_participants`; función
  `normalize_actor_name`. 226 agrega `customers.actor_id` y `projects.customer_actor_id`.
- **Tests**: `handler_test.go`, `usecases_test.go`, `repository_tenant_test.go`,
  `usecases/domain/actor_test.go`.

## Dependencias

- **Intra-repo (fuertes)**: 001 (platform/tenancy — `requestTenantID`/OrgID), 002
  (CRUDAR lifecycle — `internal/shared/lifecycle`, `deleted_at` soft-delete), 003
  (multitenant DB hardening — `tenant_id` FK a `auth_tenants`), 004 (shared text
  propername — `internal/shared/text.CanonicalizeName`).
- **Intra-repo (de cableado, fuera del flist)**: 023 (wire DI — `wire/wire.go`,
  `wire/wire_gen.go`), y `cmd/api/http_server.go`.
- **Cross-repo**: FE feature-007 consume los contratos (BE-first).
- **Bloquea**: 010 (projects — usa `project_responsibles`/`project_investor_allocations`
  creadas acá), 011 (campaign dto projectId), 018 (data-integrity-admin — duplicate
  candidates UI). Ver `dependencies.md`.

## Riesgos

- **Funcional**: la sync legacy (`legacy_sync.go`, `master_link.go`) toca tablas de
  otras features (`customers`, `projects`, `project_*`). Si esas tablas no existen aún
  en `develop`, el backfill de la migración 223/226 falla.
- **Técnico**: el wiring vive en archivos compartidos fuera del flist; portar solo el
  flist deja la feature "muerta" (compila pero no enruta).
- **Migración**: 234 consolida duplicados activos (merge destructivo en datos dev) —
  ver `risks.md`.

## DECISIÓN recomendada

**Extraer tal cual el paquete + migraciones, PERO el PR es incompleto sin los hunks de
cableado** (`wire/wire.go`, `wire/wire_gen.go`, `cmd/api/http_server.go`). Tratar esos
hunks como **partial-hunks coordinados con feature-023**: o se incluyen acá (recomendado,
porque sin ellos la feature no funciona) o se mergea 023 antes. Verificar primero que
las tablas legacy y `projects`/`customers` que toca la migración 223 ya existen en
`develop` (dependen de 010). Orden: dependencias 001/002/003/004 → (010 si projects no
está) → este BE → FE feature-007.
