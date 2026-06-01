# notes-for-future-agent.md — feature-005 · be-config-modularization

## Resumen corto

Split/limpieza del paquete `cmd/config` del backend. Borra el struct legacy `AI` (ponti-ai
deprecado) y agrega structs por dominio: `Companion`, `Nexus`, `Reporting`, `Security`, más
campos nuevos en `HTTPServer` (CORS + rate-limit), `Auth` (RequireTenantHeader, flip AutoProvision)
y `Service` (Env). Acompaña con `.env.example`. Es **infra fundacional**: "Funda 012 y 023".

## Qué está en FE y en BE

- **BE:** todo (este es el repo). 8 paths en `/tmp/flists/be-005.txt`.
- **FE:** nada. Sin carpeta. Registrar "sin cambios FE" en el cross-repo-map del FE.

## Archivos esenciales

- `cmd/config/companion.go` (A) — `Companion` + `Nexus`.
- `cmd/config/reporting.go` (A) — `Reporting` + consts/helpers.
- `cmd/config/security.go` (A) — `Security` (3 flags).
- `cmd/config/loadconfig.go` (M) — agregador: suma los nuevos sub-configs, quita `AI`.

## Archivos peligrosos / mezclados (partial-hunks)

- **`.env.example`** — MEZCLA hunks de 005 + 012 + 019 + 021. El bloque
  `# PROD data source for local DB reset` (`DB_NAME_PROD`, `SRC_*`, `SRC_PASS_SECRET_*`) es de
  **feature-019**, NO traerlo aquí. Usar `git restore -p`.
- **`cmd/config/loadconfig.go`** — agregador compartido; en este rango solo cambió el bloque de
  campos del `Config`, así que whole-file es seguro, pero verificar `git diff develop..SOURCE`.

## Decisiones ya tomadas

- Extraer tal cual (leaf, bajo riesgo).
- `.env.example` con partial-hunks (excluir bloque DB_*_PROD de 019).
- Mantener el flip `AutoProvision` true→false dentro de esta feature (mismo hunk), comunicándolo a 001/008.
- NO traer consumers (wire/, cmd/api/http_server.go, middlewares gin, authz, repositories): son de 012/023/001/003/021.

## Dudas abiertas

- ¿`CORS_ORIGINS` / `HTTP_RATE_LIMIT_PER_MINUTE` deben documentarse en `.env.example`? Los campos existen en `http_server.go` pero pueden no aparecer en el diff de `.env.example`. Mejora opcional.
- Límite exacto del split de `.env.example` entre 005/012/019/021 — confianza media; revisar hunks.

## Qué comandos mirar primero

```bash
cat /tmp/flists/be-005.txt
git -C /home/pablocristo/Proyectos/pablo/ponti/core diff 0972e565..777e5f6a -- cmd/config .env.example
git -C /home/pablocristo/Proyectos/pablo/ponti/core show 777e5f6a:cmd/config/companion.go
git -C /home/pablocristo/Proyectos/pablo/ponti/core grep -nE "cfg\.(Companion|Nexus|Reporting|Security)|config\.AI" 777e5f6a -- cmd internal wire
```

## Errores a evitar

- NO usar `develop-problematico` (tip = restore/vacío). Usar `develop-problematico~1` (`777e5f6a`).
- NO arrastrar el bloque `DB_*_PROD` de `.env.example` (feature-019).
- NO traer wire/, cmd/api/http_server.go ni los `*_tenant_test.go`/`auth_hardening_test.go`.
- NO mergear 012/023 antes que esta (no compilarían sin `config.Companion`/`config.Nexus`).

## Camino más seguro

1. Branch desde `develop`.
2. `checkout SOURCE -- cmd/config/{companion,reporting,security,service,auth,http_server}.go`.
3. `git rm cmd/config/ai.go`.
4. `checkout SOURCE -- cmd/config/loadconfig.go` (tras verificar diff).
5. `git restore -p --source=SOURCE -- .env.example` (aceptar solo config; rechazar DB_*_PROD).
6. `go build ./...` + `go vet ./cmd/config/...` + `git diff --check`.
7. PR contra `develop`.

## Qué PR del otro repo debe ir antes/después

- **Antes:** nada (FE no participa; BE: ninguna feature previa).
- **Después (BE):** feature-012 (companion/nexus) y feature-023 (wire-di) consumen estos structs; deben mergear DESPUÉS. Coordinar con feature-019 para no doble-editar `.env.example`.
