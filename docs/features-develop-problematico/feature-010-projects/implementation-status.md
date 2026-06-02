# implementation-status.md — feature-010 projects (BE)

## Estado general
**Completa** en el SOURCE (`develop-problematico~1` / `777e5f6a`). El código está terminado, con tests propios, y compila/pasa dentro del árbol completo del SOURCE. **No es extraíble en aislamiento** sobre `develop` por dependencias faltantes.

- **% completitud (a nivel código del módulo):** ~100% en el SOURCE.
- **% extraíble-hoy sobre develop sin tocar deps:** ~0% (no compila hasta mergear 004/007/008/009 + bump tenancy).

## Estado en este repo (BE)
- `handler.go`: completo. Renombre de contrato (`/hard`), helper `runProjectIDAction`, baja de `GetProtected()`.
- `handler/dto/project.go`: completo. `ActorID` + `CanonicalizeName` + `domainerr.Validation`.
- `repository.go`: completo y extenso (2075 líneas de diff). Tenancy + actor-sync + lifecycle + hard-delete-bloqueo + `assertProjectReferencesActive` + `hydrateLegacyActorIDs` + `GetRawAdminCostTotal`.
- `repository/models/project.go`: completo. `TenantID`/`ActorID`, baja de `unique`.
- `usecases.go`: completo. Port renombrado.

## Estado en el otro repo (FE)
Desconocido desde este paquete. El FE feature-010 (`pages/admin/projects` + BFF `projects.ts`) debe adaptar: path `/hard`, `actor_id`, manejo de 409s. **Validar con el paquete FE-010.**

## Tests
- `repository_tenant_test.go` (sqlite in-memory, 895 líneas): aislamiento por tenant en create/update/archive/restore.
- `repository_rename_test.go` (478): `TestUpdateProjectPropagatesRenameToLegacyTables` — rename de customer por ID propaga; managers/investors no se renombran salvo su propio editor.
- `repository_archived_refs_test.go` (208): `assertProjectReferencesActive` bloquea refs archivadas (ID activo=1, archivado=99 por tabla).
- `handler_integration_test.go` (+94): `TestProjectActionRoutesCallExplicitUseCases` (archive/restore/hard → usecase correcto, 204) y payload de update con `actor_id`.
- **Cobertura del happy-path y de los guards.** Los tests dependen de `platform/security/go/contextkeys` y de sqlite; corren sin DB real.

## Pendientes
- Ninguno a nivel de código del módulo.
- A nivel extracción: traer dependencias primero.

## Bugs / observaciones
- El propio diff documenta que la cascada vieja tenía un **bug** (pluck-after-archive perdía `workorder_items`/`work_order_draft_items`); la versión nueva lo corrige vía `lifecycle.RunCascadeArchive`. → mejora, no bug nuevo.
- `GetRawAdminCostTotal` vive aquí pero su consumidor es 018 (data-integrity). Riesgo de doble-ownership.

## Clasificación

### BLOQUEANTE para mergear
- Dependencias 004/007/008/009 + bump `platform/persistence/gorm/go` deben estar en develop ANTES. Sin esto el PR no compila.
- Coordinación BE-first con FE-010 (path `/hard`).

### Mejora futura
- Tests de integración contra Postgres real (hoy sqlite in-memory) para validar las migraciones tenant_id/archive_*.

### Deuda aceptable
- `GetRawAdminCostTotal` compartido con 018 — aceptable si se documenta y no se duplica.
- Algunos helpers `ensure*` y `ensure*ForUpdate` tienen lógica duplicada (sync legacy) — refactor opcional, no bloqueante.

### Duda humana
- ¿018 ya se mergeó/extrajo con un stub de `GetRawAdminCostTotal`? Si sí, evitar conflicto de firma.
- ¿El FE-010 ya espera `/hard` o todavía `DELETE /:id`?
