# file-list — feature-001 · Platform tenancy refactor (Fase 7)

Flist autoritativa: `/tmp/flists/be-001.txt` (57 paths). Diff de verdad: `0972e565..777e5f6a`.
Leyenda extracción: `whole-file` = traer archivo completo; `partial-hunks` = solo algunos hunks (archivo compartido con otra feature); `manual-port` = reescribir a mano; `do-not-extract-yet` = no traer en este PR.

## Propios de feature-001 (whole-file)

| path | status | tipo | rol en la feature | extracción | motivo | riesgo | confianza |
|---|---|---|---|---|---|---|---|
| internal/shared/authz/authz.go | A | go/pkg | núcleo: Principal, permisos, TenantFromContext/OptionalTenantOrStrict/RequireTenant/TenantStrictModeEnabled | whole-file | archivo nuevo, base de todo el scoping | bajo | alta |
| internal/shared/authz/authz_test.go | A | go/test | tests del pkg authz | whole-file | nuevo | bajo | alta |
| internal/shared/db/tx_context.go | M | go/infra | solo import core→platform (`databases/postgres/go`) | whole-file | diff trivial 1 línea | bajo | alta |
| internal/shared/repository/errors.go | M | go/infra | import core→platform (domainerr + postgres) | whole-file | import-only | bajo | alta |
| internal/shared/repository/validation.go | M | go/infra | import core→platform (domainerr) | whole-file | import-only | bajo | alta |
| internal/shared/types/errors_compat.go | M | go/infra | import core→platform (domainerr) | whole-file | import-only | bajo | alta |
| internal/shared/handlers/auth.go | M | go/handler | import swap + remueve `ParseUserID` deprecado | whole-file | import-only + drop deprecado | bajo | alta |
| internal/shared/handlers/bind.go | M | go/handler | import core→platform (http/gin/go) | whole-file | import-only | bajo | alta |
| internal/shared/handlers/errors.go | M | go/handler | import swap + observability logging del error | whole-file | import + log | bajo | alta |
| internal/shared/handlers/pagination.go | M | go/handler | import core→platform | whole-file | import-only | bajo | alta |
| internal/shared/handlers/params_compat.go | M | go/handler | import core→platform (domainerr) | whole-file | import-only | bajo | alta |
| internal/shared/handlers/query.go | M | go/handler | import core→platform | whole-file | import-only | bajo | alta |
| internal/shared/handlers/responses.go | M | go/handler | import core→platform | whole-file | import-only | bajo | alta |
| internal/shared/handlers/workspace_filters.go | M | go/handler | import core→platform (domainerr + http/gin/go) | whole-file | import-only | bajo | alta |
| internal/platform/config/godotenv/godotenv.go | M | go/config | fmt.Printf → slog | whole-file | logging | bajo | alta |
| internal/platform/persistence/gorm/repository.go | M | go/infra | import swap + log→slog + remueve Cloud SQL IAM connector no usado | whole-file | infra/logging | medio | alta |

## Middlewares gin (whole-file, propios)

| path | status | tipo | rol | extracción | motivo | riesgo | confianza |
|---|---|---|---|---|---|---|---|
| internal/platform/http/middlewares/gin/observability.go | A | go/mw | nuevo: request-id, slog JSON, OTel spans, RED metrics (platform/observability/go) | whole-file | nuevo | bajo | alta |
| internal/platform/http/middlewares/gin/auth_hardening_test.go | A | go/test | tests de validateIdentityClaims + RejectUnsafeLocalAuthz | whole-file | nuevo | bajo | alta |
| internal/platform/http/middlewares/gin/middleares.go | M | go/mw | swap + RejectUnsafeLocalAuthz en no-local, drop `protected` | whole-file | mw | medio | alta |
| internal/platform/http/middlewares/gin/require_identity_platform_authz.go | M | go/mw | swap core→platform (jwks, httperr, observability) + validación claims | whole-file | mw seguridad | medio | alta |
| internal/platform/http/middlewares/gin/require_credentials.go | M | go/mw | import core→platform | whole-file | import-only | bajo | alta |
| internal/platform/http/middlewares/gin/local_dev_authz.go | M | go/mw | swap + tenant header handling + observability | whole-file | mw | medio | alta |
| internal/platform/http/middlewares/gin/error_handling.go | M | go/mw | swap + log de error con observability | whole-file | mw | bajo | alta |
| internal/platform/http/middlewares/gin/require_user_id_header.go | M | go/mw | import core→platform | whole-file | import-only | bajo | alta |
| internal/platform/http/middlewares/gin/request_and_response_logger.go | M | go/mw | redacción de headers sensibles | whole-file | seguridad logs | bajo | alta |
| internal/platform/http/servers/gin/server.go | M | go/infra | log→slog en shutdown | whole-file | logging | bajo | alta |

## Compartidos (partial-hunks) — workspace/filters

| path | status | tipo | rol | extracción | motivo | riesgo | confianza |
|---|---|---|---|---|---|---|---|
| internal/shared/filters/workspace.go | M | go/infra | tenancy.Scope + ValidateProjectAccess + strict-mode | whole-file (de 001) | el diff es 100% tenancy; cohesivo con 001 | medio | alta |
| internal/shared/filters/workspace_test.go | A | go/test | tests de scoping por tenant | whole-file | nuevo, prueba 001 | bajo | alta |
| internal/shared/models/base.go | M | go/model | import contextkeys core→platform | partial-hunks | COMPARTIDO con 007/008 (Base/ActorFromContext); en este diff solo cambia import → traer ese hunk | bajo | alta |

## Compartidos (partial-hunks) — repositorios de dominio

Estos `repository.go` mezclan **001 (import-swap + tenancy.Scope)** con **002 lifecycle/crudar**, **007 actor-sync**, **009 archive-surface**. Traer SOLO los hunks de import-swap + `tenancy.Scope`. Número entre paréntesis = ocurrencias de `tenancy.Scope` en dp~1.

| path | status | tipo | rol | extracción | motivo | riesgo | confianza |
|---|---|---|---|---|---|---|---|
| internal/lot/repository.go (28) | M | go/repo | tenancy.Scope masivo | partial-hunks | mezcla lifecycle Archive/Restore (002/009) + assertLotReferencesActive | alto | media |
| internal/supply/repository.go (27) | M | go/repo | tenancy.Scope masivo | partial-hunks | mezcla lifecycle; comentario MaybeTenantScope en repository_movement.go (no en flist) | alto | media |
| internal/work-order/repository.go (22) | M | go/repo | tenancy.Scope | partial-hunks | mezcla lifecycle/actor-sync | alto | media |
| internal/work-order-draft/repository.go (16) | M | go/repo | tenancy.Scope | partial-hunks | idem | alto | media |
| internal/stock/repository.go (15) | M | go/repo | tenancy.Scope | partial-hunks | mezcla lifecycle | alto | media |
| internal/field/repository.go (14) | M | go/repo | tenancy.Scope | partial-hunks | mezcla lifecycle/field_investors | alto | media |
| internal/provider/repository.go (13) | M | go/repo | tenancy.Scope | partial-hunks | posible actor-sync (provider→actor) | alto | media |
| internal/business-parameters/repository.go (12) | M | go/repo | tenancy.Scope | partial-hunks | revisar lifecycle | medio | media |
| internal/category/repository.go (12) | M | go/repo | tenancy.Scope | partial-hunks | revisar lifecycle | medio | media |
| internal/crop/repository.go (12) | M | go/repo | tenancy.Scope | partial-hunks | revisar lifecycle | medio | media |
| internal/lease-type/repository.go (12) | M | go/repo | tenancy.Scope | partial-hunks | revisar lifecycle | medio | media |
| internal/manager/repository.go (12) | M | go/repo | tenancy.Scope | partial-hunks | posible actor-sync (manager→actor) | alto | media |
| internal/investor/repository.go (11) | M | go/repo | tenancy.Scope | partial-hunks | posible actor-sync (investor→actor) | alto | media |
| internal/customer/repository.go (10) | M | go/repo | tenancy.Scope | partial-hunks | mezcla pesada: EnsureCustomerFromActor (007) + RunCascadeArchive (002/009) | alto | media |
| internal/invoice/repository.go (5) | M | go/repo | tenancy.Scope | partial-hunks | revisar lifecycle | medio | media |
| internal/dollar/repository.go (4) | M | go/repo | tenancy.Scope | partial-hunks | revisar | medio | media |
| internal/commercialization/repository.go (3) | M | go/repo | tenancy.Scope | partial-hunks | revisar | medio | media |
| internal/class-type/repository.go (1) | M | go/repo | tenancy.Scope | partial-hunks | mínimo | bajo | media |

## Dudosos (import-only puro o solapan otras features)

| path | status | tipo | rol | extracción | motivo | riesgo | confianza |
|---|---|---|---|---|---|---|---|
| internal/businessinsights/repository.go | M | go/repo | import swap + ctx threading (0 tenancy.Scope) | partial-hunks | solo import + ctx; revisar si arrastra report-cleanup | medio | media |
| internal/dashboard/repository.go | M | go/repo | import swap + ctx (0 tenancy.Scope, 145 líneas) | partial-hunks | diff grande sin scoping → solapa 027/013 cleanup; revisar | alto | baja |
| internal/report/repository.go | M | go/repo | import swap + ctx + authz (0 tenancy.Scope, 220 líneas) | partial-hunks | mezcla report-cleanup (`fmt.Errorf`→domainerr, `project_id IN ?`) de 027/013 | alto | baja |

## Borrados — clasificación

| path | status | tipo | rol | extracción | motivo | riesgo | confianza |
|---|---|---|---|---|---|---|---|
| internal/shared/utils/jwt_tools.go | D | go/util | JWT casero reemplazado por JWKS platform | whole-file (borrado de 001) | obsoleto por require_jwt drop | medio | alta |
| internal/shared/utils/strings.go | D | go/util | helpers string deprecados | do-not-extract-yet | revisar: ¿callers vivos? `godotenv.go` importa `pkgutils` aún | medio | media |
| internal/platform/http/middlewares/gin/require_jwt.go | D | go/mw | RequireJWT casero | whole-file (borrado de 001) | reemplazado por RequireIdentityPlatformAuthz | medio | alta |
| internal/platform/files/excel/excelize/bootstrap.go | D | go/files | export XLSX | do-not-extract-yet | **pertenece a feature-013 be-csv-export** (commit 19f46cd5) | alto | alta |
| internal/platform/files/excel/excelize/config.go | D | go/files | export XLSX | do-not-extract-yet | feature-013 | alto | alta |
| internal/platform/files/excel/excelize/service.go | D | go/files | export XLSX | do-not-extract-yet | feature-013 | alto | alta |

## NO traer todavía (resumen)

- `internal/platform/files/excel/excelize/*` → feature-013.
- Hunks lifecycle/crudar de cualquier `repository.go` → feature-002 / 009.
- Hunks actor-sync (`actorsync.*`, `legacy_actor_map`) → feature-007.
- Hunks report-cleanup (`fmt.Errorf`→domainerr, `project_id IN ?`) en report/dashboard → feature-027 / 013.
- `internal/shared/utils/strings.go` borrado: verificar callers antes de borrar.
