# dependencies.md — feature-005 · be-config-modularization

## Resumen

Feature **leaf fundacional**. No depende de ninguna otra feature, pero **funda** a varias.
Nota de la feature: "Funda 012 y 023".

## Depende de

- **Nada (feature-level).** No requiere ninguna feature previa.
- **Dependencia técnica ya presente en `develop`:** import `internal/platform/config/godotenv` (loader `.env`). Existe, no se modifica.
- Libs externas ya en `go.mod` de develop: `github.com/kelseyhightower/envconfig`, `github.com/go-playground/validator/v10`. No se toca `go.mod`.

## Bloquea a (esta feature debe ir ANTES)

| feature | por qué |
|---------|---------|
| **012 ai-companion-integration** [BEFE] | consume `config.Companion` y `config.Nexus` (clientes axis). Sin estos structs, los providers no compilan. |
| **023 be-wire-di** [BE] | `wire/companion_providers.go`, `wire/config_providers.go` proveen `*config.Companion`/`*config.Nexus` desde `*config.Config`. |
| **001 be-platform-tenancy-refactor** / **003 be-multitenant-db-hardening** [BE] | usan `Security.TenantStrictMode` vía `authz.TenantStrictModeEnabled()` y `Auth.RequireTenantHeader`. |
| **021 build-and-deploy-config** [BEFE] | `cmd/api/http_server.go` usa `CORSOriginList()` y `RateLimitPerMinute`. |
| **013 be-csv-export** / **027 be-cleanup-domain-purity** [BE] | `Reporting.ReadMode` (legacy/actors_shadow/actors_live) gobierna lectura de reportes. |

## Fuerza de las dependencias

- **Fuertes (compilación):** 012 y 023 NO compilan sin estos structs. Orden obligatorio 005 → {012, 023}.
- **Débiles (comportamiento):** 001/003/008/021/013 funcionan a nivel build aunque 005 no esté (usan sus propios accessors), pero el **default** correcto de los flags vive aquí.
- **Inciertas:** el split exacto de `.env.example` entre 005/012/019/021 — confianza media; verificar hunks antes de aplicar.

## Intra-repo: archivos / tipos / config / migraciones / APIs compartidos

- **Archivo compartido (agregador):** `cmd/config/loadconfig.go` — varias features agregan sub-configs aquí. En el rango de 005 solo cambió el bloque de campos del `Config`.
- **Archivo compartido (env doc):** `.env.example` — hunks de 005 + 012 + 019 (DB_*_PROD) + 021 (deploy). Tratar con partial-hunks.
- **Tipos exportados consumidos fuera:** `config.Companion`, `config.Nexus`, `config.Reporting`, `config.Security`, `HTTPServer.CORSOriginList`, `Auth.RequireTenantHeader`, `Service.Env`, consts `ReportingReadMode*`.
- **Migraciones:** ninguna.
- **APIs:** ninguna API HTTP propia.

## Cross-repo

- **Ninguna.** Solo-BE. En FE: "sin cambios". No hay orden a coordinar con el FE.

## Recomendación de orden

1. **feature-005 (esta)** — primero entre las BE de config.
2. Luego **012** y **023** (dependen fuerte).
3. Luego/en paralelo **001/003/008/013/021** (dependen débil de los defaults).
4. **feature-019** maneja su propio hunk `DB_*_PROD` de `.env.example`; coordinar para evitar doble-edición del mismo archivo (mergear primero el que toque, rebasear el otro).
