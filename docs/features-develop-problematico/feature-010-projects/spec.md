# spec.md — feature-010 projects (Backend / ponti-core)

## Identidad
- **id:** feature-010
- **slug:** projects
- **nombre:** Projects feature
- **tipo:** feature
- **repo (este paquete):** Backend Go — `ponti-backend` (path local `/home/pablocristo/Proyectos/pablo/ponti/core`)
- **existe-en-FE/BE:** FULL-STACK. Existe en AMBOS repos con el mismo `feature-010`.
  - **BE (este):** project-archive-entidades-bridge + scope/creator (tenancy, actor-sync, lifecycle cascade, hard-delete con bloqueo).
  - **FE:** `pages/admin/projects` + BFF `projects.ts` (no documentado aquí; ver paquete del repo FE).
- **merge:** BE-first, luego FE.
- **SOURCE de extracción:** `develop-problematico~1` (SHA `777e5f6a`). NUNCA usar `develop-problematico` (tip = restore/vacío).
- **Rama destino:** `develop` (tip `003a9b8f`).
- **Rango fuente-de-verdad (diff):** `0972e565..777e5f6a`.

## Resumen
El módulo `internal/project` deja de ser un CRUD single-tenant con cascada de archive/delete escrita a mano y pasa a:
1. **Tenancy:** todas las tablas del grafo de project (`projects`, `customers`, `campaigns`, `managers`, `investors`, `project_managers`, `project_investors`, `admin_cost_investors`, `field_investors`, `lots`, …) llevan `tenant_id` y todas las queries pasan por `tenancy.Scope(ctx, db, "<tabla>")` / filtros `WHERE tenant_id = ?` cuando hay tenant en contexto.
2. **Actor-system bridge:** customer/manager/investor exponen `ActorID *int64`. Los helpers `ensure*` y `ensure*ForUpdate` sincronizan cada entidad legacy con la tabla `actors` vía `internal/actor` (`SyncLegacyActor`, `EnsureCustomerFromActor`, `EnsureLegacyEntityFromActor`, `RefreshProjectActorMirrors`), y `GetProject` hidrata `ActorID` desde `legacy_actor_map` (`hydrateLegacyActorIDs`).
3. **Lifecycle / archive cascade centralizado:** `ArchiveProject` y `RestoreProject` dejan de hacer `UPDATE ... SET deleted_at` tabla por tabla y delegan en `internal/shared/lifecycle` (`CreateArchiveBatch`, `RunCascadeArchive`, `RestoreScopedRows`, `ArchiveUpdates`, `RestoreUpdates`, `ApplyCauseScope`, `ReadRowState`, `CauseFromRow/Batch`).
4. **Hard-delete con bloqueo (renombre de contrato HTTP):** `DeleteProject` → `HardDeleteProject`. La ruta `DELETE /:project_id` pasa a `DELETE /:project_id/hard`. Bloquea con `409 Conflict` si quedan dependientes activos o si el proyecto no está archivado primero.
5. **Validación de referencias activas:** `assertProjectReferencesActive` (vía `lifecycle.RequireAllActive`) impide crear/actualizar un proyecto que referencia customer/campaign/manager/investor/field/lot/crop/actor archivados.
6. **Name canonicalization:** el DTO usa `internal/shared/text.CanonicalizeName(...)` en lugar de `strings.TrimSpace`, salvo el código de campaña que queda solo trimmed.

## Objetivo
Hacer que Project sea multi-tenant y consistente con el actor-system y el lifecycle/archive framework, eliminando la cascada artesanal (con un bug conocido de pluck-after-archive que se saltaba `workorder_items` / `work_order_draft_items`) y unificando creación/relink de entidades del grafo bajo actor-sync.

## Problema que resuelve
- Cascada de archive/restore/delete duplicada e inconsistente entre `ArchiveProject`, `RestoreProject` y `DeleteProject`, con un bug: el pluck de IDs corría después de archivar, perdiendo hijos de workorders.
- Falta de aislamiento por tenant en todas las queries de project.
- Managers/investors/customers que no se reflejaban en la tabla `actors`, rompiendo el guard de nombres duplicados del editor FE (recibía `actor_id = nil`).
- Hard-delete que ocultaba estado operacional activo en lugar de bloquear.

## Alcance EN ESTE REPO (BE)
9 archivos, todos bajo `internal/project/`:
- `handler.go` — ruteo + renombre `DeleteProject`→`HardDeleteProject`, ruta `/hard`, helper `runProjectIDAction`, baja de `GetProtected()` del port de middlewares.
- `handler/dto/project.go` — `ActorID` en DTOs Customer/Manager/Investor/AdminCostInvestor/(field) Investor; `CanonicalizeName`; errores `domainerr.Validation` en vez de `fmt.Errorf`.
- `repository.go` — el grueso (2075 líneas de diff): tenancy scope, actor-sync, lifecycle cascade, hard-delete con bloqueo, `ensure*ForUpdate`, `assertProjectReferencesActive`, `GetProjectByNameCustomerAndCampaignID`, `hydrateLegacyActorIDs`, `GetRawAdminCostTotal`.
- `repository/models/project.go` — `TenantID uuid.UUID` y `ActorID *int64` en los modelos GORM; baja del `unique` en `Name`; mapeo `ToDomain` con `ActorID`.
- `usecases.go` — port renombrado (`GetProjectByNameCustomerAndCampaignID`, `HardDeleteProject`).
- Tests nuevos: `repository_archived_refs_test.go`, `repository_rename_test.go`, `repository_tenant_test.go` (sqlite in-memory).
- Test modificado: `handler_integration_test.go` (rutas archive/restore/hard, payload con `actor_id`).

## Alcance EN EL OTRO REPO (FE)
- `pages/admin/projects` (UI de proyectos) y BFF `projects.ts`.
- Debe consumir: nuevo path `DELETE /api/v1/projects/:id/hard` (ya NO `DELETE /:id`), campo `actor_id` en customer/manager/investor, y los `409 Conflict` de bloqueo (hard-delete con dependientes activos, "project must be archived before hard delete", "project parent customer is archived", "customer already exists").

## Fuera de alcance (de este paquete)
- Implementación de `internal/actor` (feature 007), `internal/shared/lifecycle` (009), `internal/shared/authz` (008), `internal/shared/text/propername` (004), y el bump de `platform/persistence/gorm/go` (tenancy). Son DEPENDENCIAS, no se extraen aquí.
- Los modelos de dominio `customer/manager/investor` con `ActorID` (los toca feature 007/011, no están en mi flist).
- `GetRawAdminCostTotal` se agrega aquí pero su consumidor está en `internal/data-integrity` (feature 018) — no se incluye su wiring.

## Comportamiento esperado
- Crear/actualizar project escribe `tenant_id` en todo el grafo y sincroniza actores.
- Archivar usa batch + cascade del lifecycle; restaurar reactiva el grafo por `cause` (batch o columnas archive_*).
- `DELETE /:id/hard` exige proyecto archivado (409 si no) y sin dependientes activos (409 con conteo y label, p.ej. "project has 3 active field(s); archive or hard-delete them first").
- `GetProject` devuelve `actor_id` hidratado para legacy.

## Estado en dp~1 (777e5f6a)
Completo y con tests propios (sqlite in-memory). Compila en el árbol completo de `develop-problematico~1` porque ahí ya están actor/lifecycle/authz/text/tenancy. **NO compila aislado sobre `develop`** (faltan esas dependencias).

## Criterios de aceptación
- `go build ./internal/project/...` y `go test ./internal/project/...` verdes **una vez mergeadas las dependencias** 004/007/008/009 y el bump de `platform/persistence/gorm/go`.
- Ruta `DELETE /api/v1/projects/:id/hard` registrada; vieja `DELETE /:id` ausente.
- Tests `TestProjectActionRoutesCallExplicitUseCases`, `TestUpdateProjectPropagatesRenameToLegacyTables`, y los de archived-refs / tenant pasando.

## Endpoints afectados
| método | path (nuevo) | handler | cambio |
|---|---|---|---|
| DELETE | `/api/v1/projects/:project_id/hard` | `HardDeleteProject` | **renombrado** desde `DELETE /:project_id` (`DeleteProject`) |
| POST | `/api/v1/projects/:project_id/archive` | `ArchiveProject` | refactor interno (lifecycle) |
| POST | `/api/v1/projects/:project_id/restore` | `RestoreProject` | refactor interno (lifecycle) |
| GET/PUT/POST | resto de `/projects/*` | sin cambio de path | tenancy scope interno |

## Modelos / DTOs / tipos
- DTO (`handler/dto/project.go`): `Customer/Manager/Investor/AdminCostInvestor` + field `Investor` ganan `ActorID *int64 json:"actor_id,omitempty"`.
- GORM (`repository/models/project.go`): `Project.TenantID uuid.UUID`; `Manager/Customer/Campaign/Investor/ProjectInvestor/AdminCostInvestor` ganan `TenantID`; `Manager/Investor` ganan `ActorID *int64` (`gorm:"-"` salvo Customer que es columna real `actor_id`); se quita `unique` de `Name`.
- Port (`usecases.go` / `handler.go`): `GetProjectByNameAndCampaignID`→`GetProjectByNameCustomerAndCampaignID(ctx, name, customerID, campaignID)`; `DeleteProject`→`HardDeleteProject`.

## UI afectada
N/A en este repo (ver FE feature-010).

## DB / migraciones
- **NO hay archivos de migración en mi flist.** Las columnas `tenant_id` (en todas las tablas del grafo) y `actor_id` (en `customers`) + tablas `actors`, `legacy_actor_map`, columnas `archive_*` las aportan las migraciones de features 001/003/007/009. Sin esas migraciones el código falla en runtime (columnas inexistentes).

## Tests afectados
- Nuevos: `repository_archived_refs_test.go` (208), `repository_rename_test.go` (478), `repository_tenant_test.go` (895). Usan `gorm sqlite :memory:` y `platform/security/go/contextkeys`.
- Modificado: `handler_integration_test.go` (+94).

## Dependencias
- **Intra-repo (fuertes/bloqueantes):** 004 (shared-text/propername), 007 (actor-system → `internal/actor`), 008 (identity-tenant-context → `internal/shared/authz` + `contextkeys`), 009 (crudar-archive-surface → `internal/shared/lifecycle`).
- **Plataforma:** bump de `github.com/devpablocristo/platform/persistence/gorm/go` (paquete `tenancy`) en `go.mod`/`go.sum` (NO presente en develop).
- **Bloquea-a:** 011 (campaign-dto-projectid, comparte DTOs/domain de project), 018 (data-integrity usa `GetRawAdminCostTotal`), y el FE feature-010.
- **Cross-repo:** FE feature-010 consume el contrato (path `/hard`, `actor_id`, 409s).

## Riesgos
- **Funcional:** cambio de path DELETE rompe FE si no se coordina (BE-first ayuda, pero el FE debe portar el cambio de path).
- **Técnico:** no compila aislado (dependencias faltantes en develop). Alto riesgo de "extracción incompleta".
- **Datos:** requiere migraciones de tenancy/actor/archive presentes; en una DB sin ellas el módulo rompe en runtime.

## DECISIÓN recomendada
**Arreglar antes / postergar dentro del tren de dependencias.** No extraer feature-010 BE sola. Mergear primero (en orden): plataforma tenancy bump → 004 → 008 → 007 → 009; y recién entonces abrir el PR de feature-010 BE. Extraer los 9 archivos como `whole-file` (es un snapshot coherente del módulo a `777e5f6a`), no por hunks. Luego FE feature-010.
