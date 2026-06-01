# implementation-status — feature-001 · Platform tenancy refactor (Fase 7)

- **estado global**: COMPLETA en dp~1 (777e5f6a), pero **NO aislada** — el código de 001 convive con 002/007/009 en los mismos archivos.
- **% completitud (de la intención de 001)**: ~95% implementado en dp~1. Lo que falta para portar es **separar** el código de 001 del de las features mezcladas, no escribir lógica nueva.

## Estado en este repo (BE)

| Sub-parte | Estado | Nota |
|---|---|---|
| Paquete `internal/shared/authz` | completa | authz.go + authz_test.go presentes y autocontenidos |
| Import swap core→platform (shared/handlers, repository, types, db) | completa | diffs triviales, whole-file |
| `tenancy.Scope` en repos de dominio | completa (mezclada) | 20 repos con scoping; entrelazado con lifecycle/actor-sync |
| `workspace.go` (ResolveProjectIDs/ValidateProjectAccess) | completa | diff 100% tenancy, cohesivo |
| Middlewares gin (observability, hardening, RejectUnsafeLocalAuthz, redacción) | completa | tests en auth_hardening_test.go |
| Borrado JWT casero (require_jwt, jwt_tools) | completa | reemplazado por JWKS platform |
| slog en godotenv/gorm/server | completa | logging estructurado |

## Estado en el otro repo (FE)

- N/A. Sin cambios FE.

## Tests

- `internal/shared/authz/authz_test.go` (A): TestTenantOwnerIsNotGlobalWildcard, TestSaaSSuperadminIsGlobalWildcard, TestTenantFromContextAllowsTransitionMode.
- `internal/shared/filters/workspace_test.go` (A): TestResolveProjectIDsScopesByTenant (sqlite in-memory).
- `internal/platform/http/middlewares/gin/auth_hardening_test.go` (A): validación issuer/audience + RejectUnsafeLocalAuthz bloquea production.
- Comentario en authz_test.go indica que los tests de `TenantWhere`/`MaybeTenantScope` se movieron a `platform/persistence/gorm/go/tenancy/tenancy_test.go` (fuera de este repo).

## Pendientes para portar (NO son bugs, son trabajo de extracción)

### BLOQUEANTE para mergear
- **Separar partial-hunks** de los 20 `repository.go`: dejar solo import-swap + `tenancy.Scope`, excluir lifecycle/actor-sync. Si un repo no compila standalone sin esos símbolos, va con su feature dueña (002/007/009), no en 001.
- **Confirmar `platform/*` en go.mod** de develop antes del PR; si falta, no compila.
- **Excluir borrado excelize** (es 013).
- **report/dashboard/businessinsights**: decidir si entran (solo import+ctx) o se posponen (solapan 027/013). Recomendado posponer dashboard/report.

### Mejora futura
- Migrar la restauración de cascada a `RunCascadeRestore` (comentario en customer/repository.go §10C) → pertenece a 002/009, no a 001.

### Deuda aceptable
- Modo transición: `OptionalTenantOrStrict` permite operar sin tenant si strict-mode off. Aceptable hasta que 008 esté en develop.

### Duda humana
- ¿`internal/shared/utils/strings.go` se puede borrar? `godotenv.go` aún importa `pkgutils`. Verificar callers de `IsNumeric`/`NormalizeString` antes de borrar (puede ser de otra feature de cleanup).
- ¿`dashboard`/`report` (0 tenancy.Scope, diffs grandes) van en 001 o en 027/013? Confianza baja; revisar con dueños de 027/013.
