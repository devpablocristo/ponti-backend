# file-list.md — feature-005 · be-config-modularization

Flist autoritativo: `/tmp/flists/be-005.txt`. Rango diff: `0972e565..777e5f6a`. SOURCE: `develop-problematico~1` (777e5f6a).

Status: A=created · M=modified · D=deleted.

## Propios (whole-file)

| path | status | tipo | rol en la feature | extracción | motivo | riesgo | confianza |
|------|--------|------|-------------------|------------|--------|--------|-----------|
| `cmd/config/companion.go` | A | Go struct config | structs `Companion` + `Nexus` (cliente HTTP/JWT interno) | whole-file | archivo nuevo, 100% de esta feature | bajo | alta |
| `cmd/config/reporting.go` | A | Go struct config | struct `Reporting` + consts `legacy/actors_shadow/actors_live` + helpers `IsActorsShadow/IsActorsLive` | whole-file | archivo nuevo, 100% de esta feature | bajo | alta |
| `cmd/config/security.go` | A | Go struct config | struct `Security` (`TenantStrictMode`, `DomainPoliciesV2`, `AITenantScope`) | whole-file | archivo nuevo, 100% de esta feature | bajo | alta |
| `cmd/config/service.go` | M | Go struct config | agrega campo `Env` (`APP_ENV`, default `local`) | whole-file | hunk único, sin mezcla | bajo | alta |
| `cmd/config/auth.go` | M | Go struct config | agrega `RequireTenantHeader`; cambia default `AutoProvision` true→false | whole-file | hunk único; ojo cambio de default (ver risks) | medio | alta |
| `cmd/config/http_server.go` | M | Go struct config | agrega `RateLimitPerMinute`, `CORSOrigins`, método `CORSOriginList()`, `import "strings"` | whole-file | hunk único, sin mezcla | bajo | alta |
| `cmd/config/ai.go` | D | Go struct config (legacy) | elimina struct `AI` (ponti-ai deprecado) | whole-file (delete) | sin referencias `config.AI` en código BE al source | bajo | alta |

## Compartidos (partial-hunks)

| path | status | tipo | rol en la feature | extracción | motivo | riesgo | confianza |
|------|--------|------|-------------------|------------|--------|--------|-----------|
| `cmd/config/loadconfig.go` | M | Go agregador de config | quita `AI`; agrega `Companion`, `Nexus`, `Reporting`, `Security` al `Config` | partial-hunks | archivo **agregador** que toca varias features; aquí solo el hunk del bloque de campos | medio | alta |
| `.env.example` | M | env doc | reemplaza bloque AI por Companion/Nexus/Review/CORS/RateLimit; **también** trae bloque `DB_*_PROD` (feat-019) | partial-hunks | hunks de **varias** features (005, 012, 019, 021); traer SOLO los de config | alto | alta |

### Detalle de hunks de `.env.example` (para `git restore -p`)

- **SÍ (feature-005/012):** bloque AI→Companion/Nexus (`COMPANION_*`, `NEXUS_*`), `REVIEW_*`, y el cambio del header de uso (`# editar .env ...`). CORS/RateLimit no aparecen literalmente en `.env.example` pero las vars `CORS_ORIGINS`/`HTTP_RATE_LIMIT_PER_MINUTE` pueden no estar documentadas — verificar y, si se quiere, agregarlas (opcional).
- **NO en esta feature (feature-019):** bloque `# PROD data source for local DB reset` (`DB_NAME_PROD`, `DB_USER_PROD`, `CLOUDSQL_PROJECT_PROD`, `DB_INSTANCE_NAME_PROD`, `SRC_INSTANCE_PROJECT/REGION/NAME`, `SRC_PASS_SECRET_PROJECT/NAME`). Lo consume `scripts/db/reset-local-db-from-prod.sh`.

## Requeridos por dependencia

Ninguno. La feature es leaf; importa `internal/platform/config/godotenv` que **ya existe en develop** y no se toca.

## Dudosos

| path | status | nota |
|------|--------|------|
| `.env.example` | M | dudoso por mezcla con feature-019 (DB_*_PROD). Resuelto vía partial-hunks. Confianza media en el split exacto: revisar `git diff` antes de aplicar. |
| `cmd/config/auth.go` | M | el flip `AutoProvision` true→false es un cambio de comportamiento; whole-file está bien pero documentarlo (risks.md). |

## NO traer todavía (do-not-extract-yet) — consumidores fuera de scope

Estos NO están en el flist; se listan para que el extractor NO los arrastre y sepa quién los porta:

| path | extracción | dueño probable |
|------|-----------|----------------|
| `wire/companion_providers.go` | do-not-extract-yet | feature-023 / 012 |
| `wire/config_providers.go` | do-not-extract-yet | feature-023 |
| `wire/middleware_providers.go` | do-not-extract-yet | feature-023 / 001 |
| `cmd/api/http_server.go` | do-not-extract-yet | feature-021 / 012 / 013 |
| `internal/shared/authz/authz.go` | do-not-extract-yet | feature-001 / 003 |
| `internal/platform/http/middlewares/gin/*` | do-not-extract-yet | feature-001 / 008 |
| `internal/*/repository.go` (`TenantStrictModeEnabled()`) | do-not-extract-yet | feature-001 / 003 |
| `scripts/db/reset-local-db-from-prod.sh` (+ `DB_*_PROD` en .env.example) | do-not-extract-yet | feature-019 |
