# extraction-plan.md — feature-005 · be-config-modularization

- **repo:** ponti-backend (`/home/pablocristo/Proyectos/pablo/ponti/core`)
- **rama base:** `develop` (tip `003a9b8f`)
- **SOURCE:** `develop-problematico~1` (SHA `777e5f6a`). NUNCA `develop-problematico` (tip = restore/vacío).
- **rango fuente-de-verdad:** `0972e565..777e5f6a`
- **rama sugerida:** `pr/feature-005-be-config-modularization-be`

## PR title

`feat(be): modularizar cmd/config (companion/nexus/reporting/security) y limpiar AI legacy`

## PR description (sugerida)

```
Modulariza el paquete cmd/config en structs por dominio, base para la integración
Companion/Nexus (feature-012) y el wiring de DI (feature-023).

Cambios:
- Elimina struct legacy AI (cmd/config/ai.go); ponti-ai deprecado.
- Agrega structs Companion y Nexus (cliente HTTP + JWT HS256 interno).
- Agrega struct Reporting (REPORTING_READ_MODE: legacy/actors_shadow/actors_live).
- Agrega struct Security (TENANT_STRICT_MODE, DOMAIN_POLICIES_V2, AI_TENANT_SCOPE).
- HTTPServer: RateLimitPerMinute, CORSOrigins + helper CORSOriginList().
- Auth: RequireTenantHeader (default true); AutoProvision default true -> false.
- Service: campo Env (APP_ENV).
- loadconfig.go: agrega los nuevos sub-configs al Config; quita AI.
- .env.example: documenta vars Companion/Nexus/Review (sin el bloque DB_*_PROD de feature-019).

Nota: los consumidores de estos campos llegan en features posteriores (012/023/001/021).
Hasta entonces son tipos exportados sin uso (no rompen build).
```

## Pasos ordenados

1. Crear rama desde `develop`.
2. Traer **whole-file** los archivos de `cmd/config` (incl. el delete de `ai.go`).
3. Traer **partial-hunks** de `loadconfig.go` (solo el bloque de campos del `Config`).
4. Traer **partial-hunks** de `.env.example` (solo bloques Companion/Nexus/Review; NO el bloque `DB_*_PROD`).
5. Compilar el paquete y el repo entero.
6. Verificar que no queden referencias a `config.AI`.
7. PR contra `develop`.

## Archivos enteros vs parciales

- **Enteros:** `cmd/config/companion.go`, `reporting.go`, `security.go`, `service.go`, `auth.go`, `http_server.go`, y el delete `ai.go`.
- **Parciales:** `cmd/config/loadconfig.go` (agregador) y `.env.example` (mezcla con 019/012/021).

## Migraciones / tests a incluir

- Migraciones: ninguna.
- Tests: ninguno propio. NO traer `*_tenant_test.go` ni `auth_hardening_test.go` (pertenecen a features 001/003/008).

## Dependencias previas

Ninguna. Es leaf fundacional.

## Coordinación con el otro repo

**Solo-BE.** No hay PR de FE asociado. Orden: indistinto respecto del FE. Respecto del BE,
esta feature **debe ir ANTES** que 012 (companion) y 023 (wire-di), porque las funda.

## Comandos git SUGERIDOS (para un humano — NO ejecutar aquí)

```bash
# 0) posicionarse
git -C /home/pablocristo/Proyectos/pablo/ponti/core checkout develop
git -C /home/pablocristo/Proyectos/pablo/ponti/core checkout -b pr/feature-005-be-config-modularization-be

# 1) archivos enteros (incluye creados y modificados de hunk único)
git -C /home/pablocristo/Proyectos/pablo/ponti/core checkout develop-problematico~1 -- \
  cmd/config/companion.go \
  cmd/config/reporting.go \
  cmd/config/security.go \
  cmd/config/service.go \
  cmd/config/auth.go \
  cmd/config/http_server.go

# 2) borrar el legacy AI (replicar el delete del source)
git -C /home/pablocristo/Proyectos/pablo/ponti/core rm cmd/config/ai.go

# 3) loadconfig.go: archivo agregador. Es seguro traerlo entero PORQUE en el rango
#    solo cambió el bloque de campos del Config (verificar diff antes). Si hubiera
#    otros hunks de otras features en develop, usar restore -p.
git -C /home/pablocristo/Proyectos/pablo/ponti/core diff develop..develop-problematico~1 -- cmd/config/loadconfig.go
git -C /home/pablocristo/Proyectos/pablo/ponti/core checkout develop-problematico~1 -- cmd/config/loadconfig.go
#   (alternativa quirúrgica)
#   git -C ... restore -p --source=develop-problematico~1 -- cmd/config/loadconfig.go

# 4) .env.example: SOLO los hunks de config; rechazar el bloque DB_*_PROD (feature-019)
git -C /home/pablocristo/Proyectos/pablo/ponti/core restore -p --source=develop-problematico~1 -- .env.example
#   En el prompt interactivo: aceptar (y) los hunks COMPANION_*/NEXUS_*/REVIEW_* y el header;
#   rechazar (n) el hunk "# PROD data source for local DB reset" (DB_NAME_PROD ... SRC_PASS_SECRET_NAME).

# 5) sanidad de whitespace y build
git -C /home/pablocristo/Proyectos/pablo/ponti/core diff --check
go -C /home/pablocristo/Proyectos/pablo/ponti/core build ./cmd/config/...
go -C /home/pablocristo/Proyectos/pablo/ponti/core build ./...
go -C /home/pablocristo/Proyectos/pablo/ponti/core vet ./cmd/config/...
```

## Qué NO traer

- `wire/*`, `cmd/api/http_server.go`, `internal/shared/authz/*`, middlewares gin, repositories — son consumidores (features 012/023/001/003/021).
- El bloque `DB_*_PROD` de `.env.example` (feature-019) y `scripts/db/reset-local-db-from-prod.sh`.
- Cualquier `*_tenant_test.go` / `auth_hardening_test.go`.

## Qué podría romperse

- **El repo entero** si en `develop` ya hubiera código que referencia `config.AI` (no debería; verificar con `git grep config.AI`). Si existe, esta feature debe coordinarse con quien lo introdujo (improbable en `develop`).
- Si por error se trae `loadconfig.go` con hunks de otras features, el build fallará por structs no definidos (Companion/Nexus están aquí, pero otros sub-configs futuros no).

## Cómo detectar extracción incompleta

- `go build ./...` falla con "undefined: config.AI" → quedó un consumer; no es de esta feature, removerlo del lote o coordinar.
- `git grep -nE "config\.AI|AI_SERVICE_URL" -- cmd internal wire` devuelve hits en código (no docs) → falta limpiar.
- `.env.example` contiene `DB_NAME_PROD` → se arrastró el hunk de feature-019.

## Qué validar antes del PR

- `go build ./...` y `go vet ./cmd/config/...` en verde.
- `git diff --check` sin warnings de whitespace.
- `.env.example` final solo con vars de config (sin `DB_*_PROD`).
- No quedó `cmd/config/ai.go`.

## Qué hacer después de mergear

- Habilitar el merge de feature-012 (companion/nexus) y feature-023 (wire-di) que consumen estos structs.
- Avisar en 001/008 que `Auth.AutoProvision` ahora es `false` por default y `RequireTenantHeader` es `true`.
