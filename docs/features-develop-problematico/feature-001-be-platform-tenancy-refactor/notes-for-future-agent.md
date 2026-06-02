# notes-for-future-agent — feature-001 · Platform tenancy refactor (Fase 7)

## Resumen corto

Refactor transversal del BE: migra imports `devpablocristo/core/*` → `devpablocristo/platform/*` (new-cns3), introduce `internal/shared/authz` (Principal + permisos + resolución de tenant) y reemplaza el helper local `MaybeTenantScope`/`TenantWhere` por `platform/persistence/gorm/go/tenancy.Scope(ctx, db, alias)` en ~20 repos de dominio. Suma endurecimiento de auth y observabilidad (slog/OTel/request-id). Sin cambio de contrato API. Es la **base de casi todo el BE**.

## Qué está en FE y en BE

- **FE**: nada. Solo-BE. Marcar "sin cambios FE" en el cross-repo-map.
- **BE**: todo (57 archivos en `/tmp/flists/be-001.txt`).

## Archivos esenciales (traer enteros, bajo riesgo)

- `internal/shared/authz/authz.go` + `authz_test.go` — núcleo.
- `internal/shared/filters/workspace.go` + `workspace_test.go` — scoping de proyectos.
- `internal/platform/http/middlewares/gin/*` (observability.go y auth_hardening_test.go son nuevos).
- handlers/repository/types/db: import-swap whole-file.

## Archivos PELIGROSOS (mezclados — partial-hunks obligatorio)

- Los 20 `internal/*/repository.go` con `tenancy.Scope`: mezclan **001** con **002 lifecycle/crudar**, **007 actor-sync**, **009 archive-surface**. El más contaminado: `customer/repository.go` (EnsureCustomerFromActor + RunCascadeArchive) y `lot/repository.go` (Archive/Restore lifecycle). Tomar SOLO los hunks de import + `tenancy.Scope`.
- `internal/report/repository.go` y `internal/dashboard/repository.go`: 0 `tenancy.Scope`, diffs grandes (220 y 145 líneas) que son report-cleanup de **027/013**. Recomendación: NO traerlos en 001 (o solo el import-swap).
- `internal/platform/files/excel/excelize/*` (3 borrados): pertenecen a **feature-013 be-csv-export** (commit 19f46cd5). NO traer en 001.
- `internal/shared/models/base.go`: compartido con 007/008; en 001 solo el hunk del import contextkeys.

## Decisiones ya tomadas

- `MaybeTenantScope`/`TenantWhere` locales se eliminan; sus tests se movieron a `platform/.../tenancy/tenancy_test.go` (fuera de este repo).
- JWT casero (`require_jwt.go`, `jwt_tools.go`) se borra; auth pasa por JWKS de platform (`RequireIdentityPlatformAuthz`).
- Logging ad-hoc → slog estructurado.
- DECISIÓN de extracción: **partir + arreglar antes**. Whole-file lo de plataforma; partial-hunks los repos de dominio; NO traer excelize/lifecycle/actor-sync/report-cleanup.

## Dudas abiertas (para humano)

1. ¿`dashboard`/`report` van en 001 o se posponen a 027/013? (confianza baja — diffs grandes sin scoping).
2. ¿Se puede borrar `internal/shared/utils/strings.go`? `godotenv.go` aún importa `pkgutils`. Verificar `IsNumeric`/`NormalizeString` callers.
3. ¿`platform/*` ya resuelve en `go.mod` de develop? Pre-requisito duro.

## Comandos a mirar primero

```bash
cat /tmp/flists/be-001.txt
git -C /home/pablocristo/Proyectos/pablo/ponti/core diff 0972e565..777e5f6a -- internal/shared/filters/workspace.go internal/shared/db/tx_context.go
git -C ... show 777e5f6a:internal/shared/authz/authz.go | head -200
git -C ... grep -n "tenancy.Scope" 777e5f6a -- internal/customer/repository.go   # ver mezcla
git -C ... log --oneline 0972e565..777e5f6a -- internal/platform/files/excel/    # confirma que excelize es 013
```

## Errores a evitar

- NO usar `develop-problematico` (tip vacío/restore). Usar `develop-problematico~1` = `777e5f6a`.
- NO traer el `repository.go` entero de customer/lot/etc.: arrastra lifecycle + actor-sync.
- NO traer el borrado excelize (es 013).
- NO tocar go.mod/go.sum si platform/* ya está en develop (los bumps son #124).
- NO ejecutar comandos git que muten (checkout/restore/commit) — son sugerencias para el humano.

## Camino más seguro

1. Confirmar `platform/*` en go.mod.
2. PR 1: solo los whole-file de plataforma (authz, handlers, middlewares, workspace, godotenv, gorm, server) + borrados JWT. Compila standalone, bajo riesgo.
3. PR 2 (o mismo PR): partial-hunks de los repos de dominio (solo tenancy.Scope), validando `go build` repo por repo.
4. Posponer dashboard/report a sus features dueñas.

## Qué PR del otro repo va antes/después

- Ninguno del FE. Solo-BE.
- Orden BE interno recomendado: **001 → 003 → 008 → 002/007/009 → resto**.
