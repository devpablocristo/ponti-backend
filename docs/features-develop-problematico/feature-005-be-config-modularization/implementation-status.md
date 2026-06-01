# implementation-status.md — feature-005 · be-config-modularization

## Estado global

- **Estado:** COMPLETA (a nivel del paquete `cmd/config`).
- **% completitud:** ~100% del alcance propio (config structs + `.env.example`).
- **Naturaleza:** infra fundacional; el "valor" se realiza cuando llegan los consumers (012/023/...).

## Estado en este repo (BE)

- Structs nuevos presentes y bien tipados en SOURCE (`777e5f6a`):
  - `cmd/config/companion.go` → `Companion`, `Nexus` (JWT HS256 interno, defaults razonables).
  - `cmd/config/reporting.go` → `Reporting` + consts + helpers.
  - `cmd/config/security.go` → `Security` (3 feature flags).
- Campos agregados a structs existentes: `HTTPServer.{RateLimitPerMinute,CORSOrigins}` + `CORSOriginList()`, `Auth.RequireTenantHeader`, `Service.Env`.
- `cmd/config/ai.go` eliminado; `Config.AI` removido de `loadconfig.go`.
- Sin referencias residuales a `config.AI` en código BE en SOURCE (solo una mención textual en un `.md` de docs, no compila).
- El paquete es autocontenido: compila sin depender de features posteriores.

## Estado en el otro repo (FE)

- **N/A.** Sin cambios FE. Registrar "sin carpeta / sin cambios" en el cross-repo-map del FE.

## Tests

- **Propios:** ninguno en el flist.
- **Externos que dependen de estas vars (fuera de scope, NO traer aquí):**
  - `internal/.../repository_tenant_test.go` (varios) usan `t.Setenv("TENANT_STRICT_MODE","true")`.
  - `internal/platform/http/middlewares/gin/auth_hardening_test.go` usa `RequireTenantHeader: true`.
  - Estos validan a los **consumers** (001/003/008), no a este paquete.

## Pendientes

- Confirmar que `CORS_ORIGINS` y `HTTP_RATE_LIMIT_PER_MINUTE` queden documentadas en `.env.example` (en el diff de 005 no aparecen explícitas; los campos sí existen en `http_server.go`). Mejora opcional, no bloqueante.

## Clasificación de items

### BLOQUEANTE para mergear
- Aplicar `.env.example` con **partial-hunks** para NO arrastrar el bloque `DB_*_PROD` (feature-019). Si se arrastra, hay solapamiento con 019.
- `go build ./...` debe pasar tras la extracción (verificar que `develop` no tenga consumers de `config.AI`).

### Mejora futura
- Documentar `CORS_ORIGINS` / `HTTP_RATE_LIMIT_PER_MINUTE` en `.env.example`.
- Considerar `validate` tags en `Companion.BaseURL`/`JWTSecret` si 012 los vuelve obligatorios (hoy son opcionales por diseño).

### Deuda aceptable
- Structs nuevos sin consumidores hasta que aterricen 012/023/021/001 (tipos exportados sin uso; no rompen build ni vet por defecto).

### Duda humana
- ¿El cambio de default `Auth.AutoProvision` true→false debe ir en esta feature de "config" o coordinarse con 001/008 (que cambian el flujo de provisioning)? Está en el mismo hunk del struct; recomendado mantenerlo aquí y comunicarlo.
