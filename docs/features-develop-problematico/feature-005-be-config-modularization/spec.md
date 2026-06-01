# spec.md — feature-005 · be-config-modularization

- **id:** feature-005
- **nombre:** Config modularization (companion/security/reporting)
- **tipo:** infra
- **repo:** Backend Go (ponti-backend) — `path=/home/pablocristo/Proyectos/pablo/ponti/core`
- **existe-en-FE/BE:** Solo-BE. En FE **no hay carpeta** para esta feature (declararla en el cross-repo-map del FE como "sin cambios FE").
- **merge:** BE independiente.
- **fuente de verdad (diff):** `0972e565..777e5f6a`
- **SOURCE de extracción:** `develop-problematico~1` (SHA `777e5f6a`). NUNCA usar `develop-problematico` (su tip es un restore/vacío).
- **rama destino:** `develop` (tip `003a9b8f`).

## Resumen

Modularización del paquete `cmd/config`: se parte la configuración en structs por dominio,
se elimina el struct legacy `AI` (servicio `ponti-ai` deprecado) y se introducen structs nuevos
(`Companion`, `Nexus`, `Reporting`, `Security`) además de campos nuevos en structs existentes
(`HTTPServer`, `Auth`, `Service`). Se acompaña con la actualización de `.env.example`.

Esta feature es **infraestructura fundacional**: la nota de la feature dice explícitamente
**"cmd/config split + .env.example. Funda 012 y 023"**. Es decir, habilita
`feature-012-ai-companion-integration` (clientes Companion/Nexus) y
`feature-023-be-wire-di` (providers de wire que leen estos structs).

## Objetivo

Dejar el paquete `cmd/config` con un struct por sub-dominio de configuración, de forma que
features posteriores (012 companion/nexus, 023 wire-di, 001/003 tenancy/security flags,
013/027 reporting read-mode) puedan cablear sus dependencias contra config tipada y estable,
sin tocar el legacy `AI`.

## Problema

- El struct `AI` (`cmd/config/ai.go`) apuntaba a `ponti-ai` (deprecado, ver `AI_SERVICE_URL`).
- Faltaban structs tipados para: clientes internos JWT (Companion/Nexus), feature flags de
  seguridad/multi-tenant (`TENANT_STRICT_MODE`, `DOMAIN_POLICIES_V2`, `AI_TENANT_SCOPE`),
  modo de lectura de reportes (`REPORTING_READ_MODE`), CORS y rate-limit del HTTP server,
  y bandera `AUTH_REQUIRE_TENANT_HEADER`.
- `.env.example` no documentaba ninguna de esas variables y aún mostraba el patrón viejo de AI.

## Alcance en este repo (BE)

Archivos del flist (`/tmp/flists/be-005.txt`):

- `cmd/config/loadconfig.go` (M) — agrega campos `Companion`, `Nexus`, `Reporting`, `Security` al `Config`; quita `AI`.
- `cmd/config/companion.go` (A) — structs `Companion` y `Nexus` (cliente HTTP + JWT HS256 interno).
- `cmd/config/reporting.go` (A) — struct `Reporting` + constantes `legacy/actors_shadow/actors_live` + helpers `IsActorsShadow()/IsActorsLive()`.
- `cmd/config/security.go` (A) — struct `Security` (`TenantStrictMode`, `DomainPoliciesV2`, `AITenantScope`).
- `cmd/config/http_server.go` (M) — agrega `RateLimitPerMinute`, `CORSOrigins` y método `CORSOriginList()`.
- `cmd/config/auth.go` (M) — agrega `RequireTenantHeader` (default `true`); cambia default de `AutoProvision` de `true` a `false`.
- `cmd/config/service.go` (M) — agrega `Env` (`APP_ENV`, default `local`).
- `cmd/config/ai.go` (D) — borra el struct `AI` legacy.
- `.env.example` (M) — **COMPARTIDO**: reemplaza bloque AI por Companion/Nexus/Review/CORS/RateLimit; agrega también bloque `DB_*_PROD` (que NO es de esta feature, ver abajo).

## Alcance en el otro repo (FE)

Sin cambios FE. Esta feature no toca el frontend. En el cross-repo-map del FE registrar
`feature-005` como "sin carpeta / sin cambios FE".

## Fuera de alcance

- **Consumidores** de estos structs: NO están en el flist y NO se extraen aquí.
  - `wire/companion_providers.go`, `wire/config_providers.go`, `wire/middleware_providers.go` → feature-023 (wire-di) / feature-012.
  - `cmd/api/http_server.go` (usa `CORSOriginList()`, `RateLimitPerMinute`, `Reporting.ReadMode`) → feature-021/012/013.
  - `internal/shared/authz/authz.go`, `internal/*/repository.go` (`TenantStrictModeEnabled()`) → feature-001/003.
  - `internal/platform/http/middlewares/gin/*` (`RequireTenantHeader`) → feature-001/008.
- El **bloque `DB_*_PROD`** de `.env.example` (`DB_NAME_PROD`, `SRC_INSTANCE_*`, `SRC_PASS_SECRET_*`, etc.) pertenece a **feature-019 (be-local-tooling-db-scripts)**: lo consume `scripts/db/reset-local-db-from-prod.sh`. No traerlo en esta feature salvo coordinación.

## Comportamiento esperado

- `config.LoadConfig()` arranca leyendo `.env` (si existe) + env vars, hace `envconfig.Process` y valida.
- Con defaults: `Companion.BaseURL`/`Nexus.BaseURL` vacíos → clientes opcionales (el binary arranca igual; el gating Nexus se saltea — pero ESO lo decide el consumer en 012/023, no este paquete).
- `HTTPServer.CORSOriginList()` devuelve `nil` si `CORS_ORIGINS` vacío, o slice trimmed coma-separado.
- `Reporting` default `legacy`; helpers para shadow/live.
- `Auth.AutoProvision` ahora **default `false`** (cambio de comportamiento) y `RequireTenantHeader` default `true`.

## Estado en dp~1 (SHA 777e5f6a)

Completo y coherente a nivel paquete `cmd/config`: compila como unidad (structs puros + helpers).
`AI` fue eliminado y no quedan referencias a `config.AI` en código BE en el source
(única mención residual a `AI_SERVICE_URL` es en un `.md` de docs de investigación, no compila).

## Criterios de aceptación

1. `go build ./cmd/config/...` y `go vet ./cmd/config/...` pasan.
2. No existe `cmd/config/ai.go` ni referencias a `config.AI` en el árbol portado.
3. `Config` incluye `Companion`, `Nexus`, `Reporting`, `Security`; `HTTPServer` expone `CORSOriginList()`.
4. `.env.example` documenta las variables Companion/Nexus/Review/CORS/RateLimit/Security (sin arrastrar el bloque `DB_*_PROD` de 019, salvo decisión explícita).
5. El repo entero **sigue compilando** (`go build ./...`): ver riesgo de campos huérfanos abajo.

## Endpoints / modelos / UI / DB / tests afectados

- **Endpoints:** ninguno directo. Indirectamente habilita CORS/rate-limit en `cmd/api/http_server.go` y el health/info que expone `Reporting.ReadMode` (consumer fuera de scope).
- **Modelos/DTOs/tipos:** structs de config `Companion`, `Nexus`, `Reporting`, `Security`; campos nuevos en `HTTPServer`, `Auth`, `Service`. Constantes `ReportingReadMode*`.
- **UI:** ninguna.
- **DB / migraciones:** ninguna.
- **Tests:** ninguno propio en el flist. Tests que dependen de estas vars existen fuera de scope (`*_tenant_test.go` usan `TENANT_STRICT_MODE`; `auth_hardening_test.go` usa `RequireTenantHeader`).

## Dependencias

- **Intra-repo:** ninguna dependencia previa. Es hoja fundacional. Importa `internal/platform/config/godotenv` (ya existe en develop, no se toca).
- **Cross-repo:** ninguna.
- **Bloquea a:** feature-012 (companion/nexus), feature-023 (wire-di), y campos consumidos por 001/003/013/021.

## Riesgos

- **Funcional:** cambio de default `AutoProvision` `true`→`false` y nuevo `RequireTenantHeader=true` por default endurecen auth. Si se mergea SOLO este paquete sin los consumidores, no aplica (nadie lee los campos todavía); pero al llegar 001/008 cambia el comportamiento de provisioning/tenant header.
- **Técnico (campos huérfanos):** los structs nuevos quedan **sin consumidores** hasta que lleguen 012/023/021/001. Eso NO rompe el build (son tipos exportados sin uso), pero `go vet`/linters estrictos podrían quejarse de structs sin uso si hubiera unused-checks a nivel paquete (no es el caso por defecto en Go).
- **Extracción parcial de `.env.example`:** si se copia el archivo entero se arrastra el bloque `DB_*_PROD` (feature-019), generando solapamiento. Usar partial-hunks.

## DECISIÓN recomendada

**EXTRAER TAL CUAL** (es leaf fundacional, bajo riesgo), con **una sola precisión de partial-hunk en `.env.example`** para no arrastrar el bloque `DB_*_PROD` de feature-019. El resto de `cmd/config/*` se trae como whole-file. No requiere arreglos previos ni partición en subfeatures.
