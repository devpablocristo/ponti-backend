# dependencies.md — feature-020 · CI / GitHub workflows (BE)

## Depende-de (esta feature necesita que estas otras estén en develop)

| dependencia | fuerza | tipo | qué comparte/necesita | verificación |
|---|---|---|---|---|
| **feature-012 ai-companion-integration** (BE) | **FUERTE** | código/config runtime | El binario debe leer `COMPANION_*`/`NEXUS_*` que los `deploy-*.yml` inyectan. Structs en `cmd/config/companion.go` (`Companion`, `Nexus`), cliente `internal/axis/client.go`. | `git -C <core> ls-tree develop cmd/config/companion.go internal/axis/client.go` |
| **feature-008 identity-tenant-context** (BEFE) | **FUERTE** (solo bloque AUTH de DEV) | config runtime | `deploy-dev.yml` pasa `AUTH_ENABLED=true` + `APP_ENV=dev`. Requiere `cmd/config/auth.go` + middlewares gin (`internal/platform/http/middlewares/gin/*`). | `git -C <core> ls-tree develop cmd/config/auth.go` |
| **feature-003 be-multitenant-db-hardening** (BE) | **FUERTE** (solo reset-db) | migración/DB | `reset-dev-db-from-prod.yml` hace `goto 224` y aplica `migrations_v4/000224_tenant_security_foundation.up.sql`; trunca `auth_*` y backfillea `tenant_id`. | `git -C <core> ls-tree develop migrations_v4/000224_tenant_security_foundation.up.sql` |
| **feature-019 be-local-tooling-db-scripts** (BE) | **FUERTE** (solo reset-db) | scripts DB | Lógica de reset DEV-from-PROD; `scripts/db/hardening_post_restore.sql` referenciado como fallback del hardening. Ver nota de memoria *reset-local-db-from-prod (migraciones viejas)*. | `git -C <core> ls-tree develop scripts/db/hardening_post_restore.sql` |
| **migración platform/* + #124 (deps)** | **DÉBIL** | infra/deps | Remoción de `CORE_GOVERNANCE_MODULE` en `ci-pr.yml`. Ya porteado (go-jose, x/net) y core/* deprecado por platform/*. | `git -C <core> grep -n CORE_GOVERNANCE_MODULE develop` (debería estar ausente) |
| **feature-025 be-test-coverage** (BE) | **DÉBIL** | tests | El coverage gate mide `go test ./...`. Es **warning-only** (umbral 25%, no bloquea) → no es bloqueante, solo mejora la señal. | — |

## Bloquea-a (otras features que esperan a esta)

| feature | fuerza | por qué |
|---|---|---|
| **feature-021 build-and-deploy-config** (BEFE) | **DÉBIL/INCIERTA** | Si 021 toca Dockerfile/Cloud Run config, conviene alinear con el env de COMPANION/NEXUS de estos deploys. Verificar solapamiento de paths. |
| Cualquier feature que dependa de deploy a DEV con auth | DÉBIL | El cambio `AUTH_ENABLED=true` en DEV afecta cómo se prueban features en ese ambiente. |

## Cross-repo (FE)

| relación | fuerza | detalle |
|---|---|---|
| **feature-020-FE** (mismo feature, `.github/workflows` del FE) | **FUERTE (coordinación)** | Si comparten renombre AI_SERVICE→COMPANION/NEXUS o el job `audit-service-alignment`, deben mergearse en la misma ventana. El `AUTH_ENABLED=true` de DEV exige que el FE de DEV mande credenciales. |
| **feature-012 (lado FE)** | media | Si el FE consume Companion/Nexus, sus URLs deben estar configuradas antes de que el BE rote vars. |

## Artefactos / tipos / config / migraciones / APIs compartidos

- **Env vars de Cloud Run** (no en git; en GitHub repo vars + GCP Secret Manager):
  - `COMPANION_BASE_URL_{DEV,STG,PROD}`, `NEXUS_BASE_URL_{DEV,STG,PROD}` (repo vars)
  - secrets `companion-internal-jwt-secret-{dev,stg,prod}` (Secret Manager)
  - se eliminan: `ai-service-keys-{dev,...}`, `AI_SERVICE_URL`, `AI_SERVICE_TIMEOUT_MS`
- **Migración**: `migrations_v4/000224_tenant_security_foundation.up.sql` (additiva) + dir `migrations_v4/`.
- **Scripts**: `scripts/db/hardening_post_restore.sql` (fallback del hardening en reset-db).
- **Tablas DB referenciadas en reset-db**: `public.users`, `auth_memberships`, `auth_role_permissions`, `auth_tenants`, `auth_roles`, `auth_permissions`, `fx_rates`.
- **Config Go**: `cmd/config/companion.go` (`Companion`/`Nexus`), `cmd/config/auth.go` (`AUTH_ENABLED`/`APP_ENV`).

## Recomendación de orden

1. (ya) platform/* + #124 → habilita `ci-pr.yml`.
2. **Sub-PR A** (`ci-pr.yml`) — sin esperar a nadie más.
3. feature-012 → luego feature-008 (DEV).
4. **Sub-PR B** (deploys + audit) — con vars/secrets GCP listos; gatear bloque AUTH a 008.
5. features 003 + 019.
6. **Sub-PR C** (`reset-dev-db-from-prod.yml`).
7. Coordinar con **feature-020-FE** en las mismas ventanas que B (renombre vars) — config BE-first.
