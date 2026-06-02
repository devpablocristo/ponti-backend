# extraction-plan — feature-001 · Platform tenancy refactor (Fase 7)

- **repo**: `ponti-backend` (`/home/pablocristo/Proyectos/pablo/ponti/core`)
- **rama base**: `develop` (tip `003a9b8f`)
- **SOURCE**: `develop-problematico~1` (SHA `777e5f6a`). NUNCA `develop-problematico` (tip = restore/vacío).
- **rama sugerida**: `pr/feature-001-be-platform-tenancy-refactor-be`
- **merge**: BE independiente. Sin coordinación FE (solo-BE).

## PR title

`refactor(be): migrar core→platform y centralizar tenancy/authz (Fase 7)`

## PR description (borrador)

> Migra la infraestructura del BE de `devpablocristo/core/*` a `devpablocristo/platform/*` (new-cns3) y centraliza el scoping multi-tenant.
> - Nuevo paquete `internal/shared/authz` (Principal, permisos, TenantFromContext/OptionalTenantOrStrict/RequireTenant/TenantStrictModeEnabled).
> - Reemplaza el helper local `MaybeTenantScope`/`TenantWhere` por `platform/persistence/gorm/go/tenancy.Scope` en los repos de dominio.
> - Endurece auth (RejectUnsafeLocalAuthz fuera de entornos local-like, validación issuer/audience, redacción de headers sensibles) e introduce observabilidad (request-id, slog JSON, OTel, RED metrics).
> - Reemplaza `fmt.Printf`/`log.*` por `slog` estructurado; elimina JWT casero (`require_jwt.go`, `jwt_tools.go`).
> Sin cambio de contrato API. Base del resto del BE.
> NO incluye: lifecycle/crudar (002/009), actor-sync (007), borrado XLSX (013), report-cleanup (027).

## Orden de pasos

1. **Pre-requisito**: confirmar que `develop` ya resuelve `platform/*` en `go.mod`/`go.sum`. Si no, este PR debe ir DESPUÉS de que la base de módulos platform esté disponible (los bumps go-jose/x/net de #124 ya están en develop).
2. Crear rama desde `develop`.
3. Traer **whole-file** los archivos limpios de plataforma (sección A abajo).
4. Traer **whole-file** los borrados de JWT casero (require_jwt.go, jwt_tools.go) y workspace.go/workspace_test.go.
5. Para los `repository.go` de dominio: **partial-hunks** — solo import-swap + `tenancy.Scope`. Verificar que NO entren `lifecycle.*`, `actorsync.*`, `legacy_actor_map`.
6. `base.go`: solo el hunk del import contextkeys.
7. `go build ./... && go vet ./...`. Resolver imports faltantes (si un repo quedó importando `lifecycle`/`actor` por hunks parciales, NO los traigas: el `tenancy.Scope` debe quedar standalone; si no compila standalone, ese repo va con su feature dueña).
8. `go test` de los paquetes con tests nuevos.
9. `git diff --check` (whitespace) y revisar que no quedó `devpablocristo/core/`.

### Sección A — archivos enteros (whole-file)

```
internal/shared/authz/authz.go
internal/shared/authz/authz_test.go
internal/shared/db/tx_context.go
internal/shared/repository/errors.go
internal/shared/repository/validation.go
internal/shared/types/errors_compat.go
internal/shared/handlers/auth.go
internal/shared/handlers/bind.go
internal/shared/handlers/errors.go
internal/shared/handlers/pagination.go
internal/shared/handlers/params_compat.go
internal/shared/handlers/query.go
internal/shared/handlers/responses.go
internal/shared/handlers/workspace_filters.go
internal/shared/filters/workspace.go
internal/shared/filters/workspace_test.go
internal/platform/config/godotenv/godotenv.go
internal/platform/persistence/gorm/repository.go
internal/platform/http/servers/gin/server.go
internal/platform/http/middlewares/gin/observability.go
internal/platform/http/middlewares/gin/auth_hardening_test.go
internal/platform/http/middlewares/gin/middleares.go
internal/platform/http/middlewares/gin/require_identity_platform_authz.go
internal/platform/http/middlewares/gin/require_credentials.go
internal/platform/http/middlewares/gin/local_dev_authz.go
internal/platform/http/middlewares/gin/error_handling.go
internal/platform/http/middlewares/gin/require_user_id_header.go
internal/platform/http/middlewares/gin/request_and_response_logger.go
```

Borrados (whole-file):
```
internal/platform/http/middlewares/gin/require_jwt.go
internal/shared/utils/jwt_tools.go
```

### Sección B — partial-hunks (solo import-swap + tenancy.Scope)

`base.go` + los 20 `repository.go` con `tenancy.Scope`. Revisar `businessinsights`/`dashboard`/`report` (probablemente NO traer: solapan 013/027).

## Comandos git SUGERIDOS (para un humano — NO ejecutar aquí)

```bash
# 0. partir de develop
git checkout develop
git checkout -b pr/feature-001-be-platform-tenancy-refactor-be

# 1. archivos enteros (Sección A)
git checkout develop-problematico~1 -- \
  internal/shared/authz/authz.go internal/shared/authz/authz_test.go \
  internal/shared/db/tx_context.go \
  internal/shared/repository/errors.go internal/shared/repository/validation.go \
  internal/shared/types/errors_compat.go \
  internal/shared/handlers/auth.go internal/shared/handlers/bind.go \
  internal/shared/handlers/errors.go internal/shared/handlers/pagination.go \
  internal/shared/handlers/params_compat.go internal/shared/handlers/query.go \
  internal/shared/handlers/responses.go internal/shared/handlers/workspace_filters.go \
  internal/shared/filters/workspace.go internal/shared/filters/workspace_test.go \
  internal/platform/config/godotenv/godotenv.go \
  internal/platform/persistence/gorm/repository.go \
  internal/platform/http/servers/gin/server.go \
  internal/platform/http/middlewares/gin/observability.go \
  internal/platform/http/middlewares/gin/auth_hardening_test.go \
  internal/platform/http/middlewares/gin/middleares.go \
  internal/platform/http/middlewares/gin/require_identity_platform_authz.go \
  internal/platform/http/middlewares/gin/require_credentials.go \
  internal/platform/http/middlewares/gin/local_dev_authz.go \
  internal/platform/http/middlewares/gin/error_handling.go \
  internal/platform/http/middlewares/gin/require_user_id_header.go \
  internal/platform/http/middlewares/gin/request_and_response_logger.go

# 2. borrados de JWT casero
git rm internal/platform/http/middlewares/gin/require_jwt.go internal/shared/utils/jwt_tools.go

# 3. partial-hunks de base.go + repos de dominio (elegir SOLO hunks de import + tenancy.Scope)
git restore -p --source=develop-problematico~1 -- \
  internal/shared/models/base.go \
  internal/lot/repository.go internal/supply/repository.go \
  internal/work-order/repository.go internal/work-order-draft/repository.go \
  internal/stock/repository.go internal/field/repository.go \
  internal/provider/repository.go internal/business-parameters/repository.go \
  internal/category/repository.go internal/crop/repository.go \
  internal/lease-type/repository.go internal/manager/repository.go \
  internal/investor/repository.go internal/customer/repository.go \
  internal/invoice/repository.go internal/dollar/repository.go \
  internal/commercialization/repository.go internal/class-type/repository.go

# 4. validar
git diff --check
git grep -n "devpablocristo/core/" internal/    # debe quedar vacío en los paths tocados
go build ./... && go vet ./...
go test ./internal/shared/authz/... ./internal/shared/filters/... ./internal/platform/http/middlewares/gin/...
```

> Nota sobre `git restore -p`: si un hunk de `tenancy.Scope` está pegado a uno de `lifecycle`/`actorsync` (mismo hunk), hay que `s`plit o `e`ditar el hunk para excluir lo de 002/007/009. Si no se puede separar limpio, ese repo NO va en 001 — dejarlo para su feature dueña.

## Migraciones / tests a incluir

- Migraciones: ninguna (las columnas `tenant_id` vienen de 003/008).
- Tests: `authz_test.go`, `workspace_test.go`, `auth_hardening_test.go` (los tres nuevos, whole-file).

## Dependencias previas

- `platform/*` disponible en `go.mod` (new-cns3). Verificar antes de empezar.
- Idealmente 003 (db hardening) y 008 (identity-tenant-context) ya en develop para que el scoping tenga sentido y columnas; si no, el modo transición preserva comportamiento (no rompe build, pero el scoping no acota nada).

## Coordinación con el otro repo

Ninguna. Solo-BE. FE = "sin cambios".

## Qué NO traer

- `internal/platform/files/excel/excelize/*` (013).
- Hunks `lifecycle.*`, `actorsync.*`, `legacy_actor_map`, `RunCascadeArchive`, `ArchiveUpdates`, `RestoreScopedRows`, `Cause`, `EnsureCustomerFromActor`, `SyncLegacyActor`, `RefreshProjectActorMirrors`.
- report/dashboard cleanup (`fmt.Errorf`→domainerr, `project_id IN ?`).
- Borrado de `strings.go` si tiene callers vivos.

## Qué podría romperse

- Si un repo queda con import `lifecycle`/`actor` sin usar (por hunks parciales mal cortados) → no compila. Solución: cortar mejor o dejar el repo para su feature.
- Si `platform/*` no resuelve → no compila.
- Cambio de semántica `ResolveProjectIDs` puede alterar listados si se mergea sin 003/008.

## Cómo detectar extracción incompleta

- `git grep "devpablocristo/core/"` en paths tocados → debe ser 0.
- `go build ./...` falla por símbolos `lifecycle`/`actorsync` no traídos → esos hunks no eran de 001.
- `goimports`/`go vet` reporta imports sin usar.

## Qué validar antes del PR

- Build + vet + los 3 paquetes de test verdes.
- Diff no incluye excelize, lifecycle, actor-sync, report-cleanup.
- No quedan referencias a `MaybeTenantScope`/`TenantWhere`/`RequireJWT` locales.

## Qué hacer después de mergear

- Desbloquea 002, 003, 007, 008, 009 (todos asumen authz + tenancy.Scope + platform imports).
- Verificar que features posteriores que tocan los mismos repos rebasen limpio.
