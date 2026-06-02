# validation.md — feature-005 · be-config-modularization

## Checklist pre-PR

- [ ] Rama creada desde `develop` (`pr/feature-005-be-config-modularization-be`).
- [ ] `cmd/config/ai.go` eliminado del árbol.
- [ ] `cmd/config/companion.go`, `reporting.go`, `security.go` presentes y con el contenido del SOURCE.
- [ ] `cmd/config/loadconfig.go` incluye `Companion`, `Nexus`, `Reporting`, `Security` y ya NO `AI`.
- [ ] `cmd/config/http_server.go` con `RateLimitPerMinute`, `CORSOrigins`, `CORSOriginList()` y `import "strings"`.
- [ ] `cmd/config/auth.go` con `RequireTenantHeader` y `AutoProvision` default `false`.
- [ ] `cmd/config/service.go` con campo `Env`.
- [ ] `.env.example` con bloques Companion/Nexus/Review y **sin** `DB_NAME_PROD`/`SRC_*` (feature-019).
- [ ] `git diff --check` limpio (sin whitespace errors).

## Tests sugeridos (BE)

```bash
# build del paquete y del repo entero (detecta huérfanos y referencias a config.AI)
go -C /home/pablocristo/Proyectos/pablo/ponti/core build ./cmd/config/...
go -C /home/pablocristo/Proyectos/pablo/ponti/core build ./...
go -C /home/pablocristo/Proyectos/pablo/ponti/core vet ./cmd/config/...

# (no hay tests propios en cmd/config; si se agregaran)
go -C /home/pablocristo/Proyectos/pablo/ponti/core test ./cmd/config/...
```

> FE: N/A (sin cambios FE). No correr `yarn test`/`build`/`e2e` para esta feature.

## Validación manual

1. Crear un `.env` mínimo y correr `LoadConfig()` (vía `go run ./cmd/api` si el repo lo permite, o un test ad-hoc): con `CORS_ORIGINS` vacío, `HTTPServer.CORSOriginList()` debe devolver `nil`.
2. Con `CORS_ORIGINS=https://a.com, https://b.com`, `CORSOriginList()` debe devolver `["https://a.com","https://b.com"]` (trimmed, sin vacíos).
3. `Reporting.ReadMode` default `legacy`; `IsActorsShadow()`/`IsActorsLive()` `false` por default.
4. Sin `COMPANION_BASE_URL` / `NEXUS_BASE_URL`, los structs cargan con campos vacíos (válido; el gating lo decide el consumer, no este paquete).

## Casos borde

- `CORS_ORIGINS=" , , "` → `CORSOriginList()` debe devolver `nil`/slice vacío (todos trimmean a vacío).
- `CORS_ORIGINS="x"` (un solo origen, sin coma) → `["x"]`.
- Variables no seteadas → defaults de los tags `envconfig` aplican (TTL 300, timeouts, etc.).

## Qué revisar en UI / API / DB / env

- **UI:** nada.
- **API:** nada directo (los endpoints CORS/rate-limit los enciende `cmd/api/http_server.go`, fuera de scope).
- **DB:** nada.
- **env:** `.env.example` final no debe contener `DB_NAME_PROD`, `DB_USER_PROD`, `CLOUDSQL_PROJECT_PROD`, `DB_INSTANCE_NAME_PROD`, `SRC_INSTANCE_*`, `SRC_PASS_SECRET_*` (esos son de feature-019).

## Qué validar en el otro repo

- N/A (sin cambios FE).

## Señales de incompletitud / incompatibilidad

- `go build ./...` falla con `undefined: config.AI` → quedó un consumer del struct viejo (no es de esta feature; removerlo del lote/coordinar).
- `git grep -nE "config\.AI|AI_SERVICE_URL" -- cmd internal wire` con hits en `.go` → falta limpieza.
- `.env.example` con `DB_NAME_PROD` → se arrastró el hunk de feature-019.
- Compila pero `wire` (si por error se trajo) falla → se incluyó un consumer de 012/023 que no corresponde.
