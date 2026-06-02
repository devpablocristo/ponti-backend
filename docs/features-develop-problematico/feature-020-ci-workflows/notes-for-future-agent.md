# notes-for-future-agent.md — feature-020 · CI / GitHub workflows (BE)

## Resumen corto

6 GitHub Actions del backend modificados (todos M). Aunque es feature "infra", los cambios **arrastran configuración de runtime de otras features** y por eso NO se traen tal cual de un solo PR. El propio enunciado avisa: "pueden romper deploy si se traen sin el resto".

Los 4 ejes de cambio:
1. `ci-pr.yml`: quita `CORE_GOVERNANCE_MODULE` (migración core/*→platform/*, ya en develop) + agrega **coverage gate warning-only** (umbral 25%, no falla) + sube `coverage.out`.
2. `deploy-{dev,staging,prod}.yml`: AI_SERVICE_* → **COMPANION_* + NEXUS_*** (BaseURL, JWT interno HS256, secret `companion-internal-jwt-secret-*`). DEV además: `AUTH_ENABLED=true` + `APP_ENV=dev`.
3. `audit-service-alignment.yml` + smoke de deploys: set `watched` pasa a `{BASE_MANAGER_API, COMPANION_BASE_URL, NEXUS_BASE_URL}`.
4. `reset-dev-db-from-prod.yml`: restore compatible con multi-tenant (`goto 224`, truncar `auth_*`, dedupe pg_restore list, backfill `tenant_id` con `000224_*.up.sql`, hardening por `psql` vía proxy).

## Qué está en FE y qué en BE

- **BE** (este paquete): los 6 YAML de `.github/workflows/`. Confirmado por diff.
- **FE**: feature-020-FE espejo (`.github/workflows` del FE) — NO inspeccionado aquí. Probable CI/deploy del FE y, si comparte el servicio AI, el mismo renombre de vars y/o el job de auditoría. Coordinar.

## Archivos esenciales / peligrosos / mezclados

- **Esencial y seguro**: `ci-pr.yml` (Sub-PR A) — mergeable solo, bajo riesgo.
- **Peligroso**: `deploy-prod.yml` (PROD) y el bloque `AUTH_ENABLED=true` de `deploy-dev.yml` (rompe DEV si el FE no manda auth).
- **Mezclado (partial-hunks)**: `deploy-dev.yml` combina COMPANION/NEXUS (feature-012) + AUTH (feature-008). Separar con `git restore -p`.
- **Gated por DB**: `reset-dev-db-from-prod.yml` depende de migración 224 + dir `migrations_v4/` (features 003/019).

## Decisiones ya tomadas

- **Partir en 3 sub-PRs** (A: ci-pr / B: deploys+audit / C: reset-db). No un único PR.
- Coverage gate se trae como **warning-only** (no enforcement) — coincide con el comentario del propio YAML.
- Hardening post-restore se cambió a `psql` por el proxy a propósito (el import user de `gcloud sql import` no puede otorgar privilegios sobre `fx_rates`). NO revertir a `gcloud sql import`.
- El reset (Sub-PR C) se posterga hasta tener 003/019 — ya hubo fixes previos al script de reset por numeración de migraciones (ver memoria *reset-local-db-from-prod (migraciones viejas)*).

## Dudas abiertas

- ¿La numeración `goto 224` calza con el set de migraciones que finalmente quede en develop? **Verificar antes de Sub-PR C.**
- ¿`audit-service-alignment.yml` es compartido/duplicado con el FE?
- ¿Existen en GCP las repo vars `COMPANION_BASE_URL_*`/`NEXUS_BASE_URL_*` y los secrets `companion-internal-jwt-secret-*`? (config externa, no en git).

## Qué comandos mirar primero

```bash
core=/home/pablocristo/Proyectos/pablo/ponti/core
cat /tmp/flists/be-020.txt
git -C "$core" diff 0972e565..777e5f6a -- .github/workflows/ci-pr.yml
git -C "$core" diff 0972e565..777e5f6a -- .github/workflows/deploy-dev.yml
git -C "$core" diff 0972e565..777e5f6a -- .github/workflows/reset-dev-db-from-prod.yml
# dependencias en develop:
git -C "$core" ls-tree develop cmd/config/companion.go cmd/config/auth.go
git -C "$core" ls-tree develop migrations_v4/000224_tenant_security_foundation.up.sql
git -C "$core" grep -n CORE_GOVERNANCE_MODULE develop -- .github/
```

## Errores a evitar

- NO usar `develop-problematico` (tip = restore/vacío). SOURCE = `777e5f6a` (`develop-problematico~1`).
- NO traer `deploy-*.yml` enteros si 012/008 no están en develop → deja deploy/DEV roto.
- NO activar `AUTH_ENABLED=true` en DEV sin coordinar con el FE.
- NO mergear Sub-PR C sin confirmar la migración 224.
- NO editar los YAML a mano (strings `ENV_VARS` enormes, fácil romper) — usar `git checkout <sha> -- <file>`.
- NO reintroducir `CORE_GOVERNANCE_MODULE` ni `AI_SERVICE_*`.

## Camino más seguro

1. Mergear **Sub-PR A** (`ci-pr.yml`) ya — sin dependencias bloqueantes.
2. Portear 012 (companion/nexus) y 008 (auth); crear vars/secrets GCP.
3. **Sub-PR B** con el bloque AUTH gateado a 008; coordinar ventana con FE.
4. Portear 003/019; luego **Sub-PR C**.

## PRs del otro repo que deben ir antes/después

- **Coordinar simultáneo con feature-020-FE** la ventana del renombre de vars (Sub-PR B). Config **BE-first** (las vars/secrets de Cloud Run deben existir antes de que el FE apunte a Companion/Nexus).
- El lado FE de feature-008 (auth) debe acompañar o preceder al `AUTH_ENABLED=true` de DEV.
