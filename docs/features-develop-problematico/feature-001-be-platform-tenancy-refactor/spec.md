# spec — feature-001 · Platform tenancy refactor (Fase 7)

- **id**: feature-001
- **slug**: be-platform-tenancy-refactor
- **nombre**: Platform tenancy refactor (Fase 7)
- **tipo**: refactor (infra/plataforma transversal)
- **repo**: Backend Go — `ponti-backend` (`/home/pablocristo/Proyectos/pablo/ponti/core`)
- **existe en FE**: NO. Solo-BE. En el `cross-repo-map` del FE figura como "sin cambios FE".
- **existe en BE**: SÍ (es la base de casi todo el BE).
- **rango fuente de verdad (diff)**: `0972e565..777e5f6a`
- **SOURCE de extracción**: `develop-problematico~1` (SHA `777e5f6a`). NUNCA usar `develop-problematico` (su tip es un restore/vacío).
- **rama destino**: `develop` (tip `003a9b8f`).

## Resumen

Migración transversal de la base de infraestructura del BE de los paquetes `github.com/devpablocristo/core/*` a `github.com/devpablocristo/platform/*` (parte de la migración new-cns3), combinada con el endurecimiento del scoping multi-tenant: se elimina el helper local `MaybeTenantScope`/`TenantWhere` y todos los repositorios pasan a usar `platform/persistence/gorm/go/tenancy.Scope(ctx, db, alias)` directamente. Se introduce un paquete nuevo `internal/shared/authz` que centraliza `Principal`, permisos y la resolución de tenant desde el contexto (`TenantFromContext`, `OptionalTenantOrStrict`, `RequireTenant`, `TenantStrictModeEnabled`).

## Objetivo

- Unificar imports de plataforma: `core/errors/go/domainerr`, `core/security/go/contextkeys`, `core/http/gin/go`, `core/http/go/httperr`, `core/databases/postgres/go`, `core/authn/go/jwks` → equivalentes en `platform/*`.
- Centralizar permisos y resolución de tenant en `internal/shared/authz`.
- Aplicar tenant-scoping consistente (`tenancy.Scope`) en ~20 repositorios de dominio.
- Endurecer middleware de auth (rechazar local-dev authz fuera de entornos local-like, validar claims issuer/audience, redactar headers sensibles) e introducir observabilidad (`platform/observability/go`: request-id, slog JSON, OTel spans, RED metrics).
- Reemplazar logging ad-hoc (`fmt.Printf`, `log.Printf`, `log.Println`) por `slog` estructurado.

## Problema que resuelve

`core/*` quedó deprecado por `platform/*` (new-cns3). Mientras conviva el import viejo, el BE no compila contra la plataforma nueva. Además el scoping multi-tenant estaba disperso en un helper local frágil (`MaybeTenantScope` no respetaba alias con `.`), sin un punto único de verdad para permisos ni para "modo strict".

## Alcance en este repo (BE)

1. **Import swap core→platform** (todos los archivos de la flist).
2. **Paquete nuevo `internal/shared/authz`** (`authz.go` + `authz_test.go`).
3. **`tenancy.Scope` en repos de dominio** (business-parameters, category, class-type, commercialization, crop, customer, dollar, field, investor, invoice, labor, lease-type, lot, manager, provider, stock, supply, work-order, work-order-draft). En `businessinsights`, `dashboard`, `report` solo cambia import + threading de `ctx` (0 `tenancy.Scope`).
4. **`internal/shared/filters/workspace.go`**: tenant-scope en `ResolveProjectIDs`, nueva `ValidateProjectAccess`, scope en `ValidateFieldBelongsToProject`; semántica strict-mode "todos los proyectos del tenant" + retorno `[]int64{0}` cuando no hay matches en modo tenant.
5. **Middlewares gin**: `RejectUnsafeLocalAuthz` (nuevo), `observability.go` (nuevo), validación claims, redacción de headers, slog.
6. **Borrados**: `internal/shared/utils/jwt_tools.go`, `internal/shared/utils/strings.go`, `internal/platform/http/middlewares/gin/require_jwt.go` (JWT casero reemplazado por JWKS de platform); `internal/platform/files/excel/excelize/{bootstrap,config,service}.go` — **OJO: este borrado pertenece a feature-013 (be-csv-export), NO a 001** (ver fuera de alcance).
7. **`godotenv.go`, `gorm/repository.go`, `server.go`**: cambio de logging a slog + remoción de conexión Cloud SQL IAM no usada.

## Alcance en el otro repo (FE)

Ninguno. Sin cambio de contrato API; el FE no se entera de este refactor. Mencionar en `cross-repo-map` del FE como "sin cambios FE".

## Fuera de alcance (NO extraer aquí)

- **Borrado excelize** (`internal/platform/files/excel/excelize/*`): es de **feature-013 be-csv-export** (commit `19f46cd5 refactor(be): replace XLSX exports with CSV`). No traerlo en 001.
- **Lógica de lifecycle/crudar** que aparece mezclada en customer/lot/etc. (`lifecycle.RunCascadeArchive`, `ArchiveUpdates`, `RestoreScopedRows`, `Cause`): es **feature-002 be-crudar-lifecycle-framework** + **feature-009 crudar-archive-surface**.
- **actor-sync** (`actorsync.EnsureCustomerFromActor`, `SyncLegacyActor`, `RefreshProjectActorMirrors`, `legacy_actor_map`): es **feature-007 actor-system**.
- **report/dashboard cleanup** (cambio `fmt.Errorf`→`domainerr.Internal`, `project_id IN ?`): solapa **feature-027** y **feature-013**.
- **lot-metrics/total_tons** y **tentative-prices**: YA PORTEADO (#117/#121/#124). No re-extraer.

## Comportamiento esperado

- BE compila e importa solo `platform/*`.
- Con contexto tenant presente, las queries de los repos quedan acotadas por `tenant_id`; sin contexto tenant (modo transición) se preserva comportamiento legacy salvo que `TenantStrictModeEnabled()` esté activo (entonces falla cerrado con `domainerr.TenantMissing()`).
- Auth: en entornos no local-like sin `cfg.Auth.Enabled` se rechaza el request (`RejectUnsafeLocalAuthz`). Claims con issuer/audience mismatch se rechazan.
- Logs en JSON estructurado con request-id; headers `Authorization`/`Cookie` redactados.

## Estado en dp~1 (777e5f6a)

Implementado y con tests (`authz_test.go`, `workspace_test.go`, `auth_hardening_test.go`). PERO el diff de los repos de dominio está **fuertemente entrelazado** con lifecycle/crudar (002/009) y actor-sync (007); no es separable con `git checkout <path>` entero salvo en los archivos puramente de plataforma.

## Criterios de aceptación

- [ ] `git grep "devpablocristo/core/" internal/` no devuelve nada en los paths de esta flist.
- [ ] `go build ./...` y `go vet ./...` OK.
- [ ] `go test ./internal/shared/authz/... ./internal/shared/filters/... ./internal/platform/http/middlewares/gin/...` verde.
- [ ] `internal/shared/authz` expone `Principal`, `PrincipalFromContext`, `RequireTenant`, `TenantFromContext`, `OptionalTenantOrStrict`, `TenantStrictModeEnabled`, `HasPermission`, `RequirePermission`, constantes `Permission*`.
- [ ] No queda referencia a `MaybeTenantScope`/`TenantWhere` locales (solo comentarios históricos).
- [ ] No se arrastra lógica de lifecycle/actor-sync/crudar en este PR.

## Endpoints / modelos / UI / DB / tests afectados

- **Endpoints**: ninguno nuevo ni cambio de contrato (refactor interno).
- **Modelos/tipos nuevos**: `authz.Principal`; constantes `authz.Permission*`. `internal/shared/filters` agrega `ValidateProjectAccess`.
- **UI**: ninguna.
- **DB/migraciones**: ninguna migración en esta flist (el scoping usa columnas `tenant_id` que ya deben existir — provistas por features 003/008).
- **Tests**: `internal/shared/authz/authz_test.go` (A), `internal/shared/filters/workspace_test.go` (A), `internal/platform/http/middlewares/gin/auth_hardening_test.go` (A).

## Dependencias

- **Intra-repo**: el scoping multi-tenant asume columnas `tenant_id` y contextkeys de tenant → **003 be-multitenant-db-hardening** y **008 identity-tenant-context**. Repos customer/lot importan `internal/shared/lifecycle` (002/009) e `internal/actor` (007) en dp~1: esos imports NO son de 001 y deben quedar fuera del PR de 001.
- **Cross-repo**: ninguna (solo-BE).
- **Plataforma**: requiere que `go.mod` resuelva `github.com/devpablocristo/platform/*` (módulos prefetch en CI — los bumps go-jose/x/net ya porteados en #124, excluir).

## Riesgos

- **Funcional**: cambio de semántica en `ResolveProjectIDs` (strict-mode = todos los del tenant; `[]int64{0}` cuando 0 matches) puede alterar listados si se mergea sin 003/008.
- **Técnico**: extracción parcial difícil — los `repository.go` mezclan 4 features. Alto riesgo de arrastrar lifecycle/actor-sync.
- **Plataforma**: si `platform/*` no está disponible en `go.mod`/replace, no compila.

## DECISIÓN recomendada

**PARTIR EN SUBFEATURES + ARREGLAR ANTES**. No es extraíble "tal cual" porque los `repository.go` de dominio están entrelazados con 002/007/009. Recomendación:
1. Extraer **whole-file** lo limpio de plataforma: `internal/shared/authz/*`, `internal/shared/db/tx_context.go`, handlers import-only, middlewares gin, `godotenv.go`, `gorm/repository.go`, `server.go`, `workspace.go`/`workspace_test.go`.
2. Para los `repository.go` de dominio: **partial-hunks manuales** tomando solo los hunks de import-swap + `tenancy.Scope`, dejando lifecycle/actor-sync para 002/007/009.
3. NO traer el borrado excelize (es 013).
