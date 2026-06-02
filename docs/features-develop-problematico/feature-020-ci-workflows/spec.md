# spec.md — feature-020 · CI / GitHub workflows (Backend)

- **id**: feature-020
- **slug**: ci-workflows
- **nombre**: CI / GitHub workflows
- **tipo**: infra
- **repo**: Backend Go (ponti-backend) — `/home/pablocristo/Proyectos/pablo/ponti/core`
- **existe-en-FE**: SÍ (FULL-STACK, mismo feature-020 en `.github/workflows` del repo FE)
- **existe-en-BE**: SÍ (este paquete)
- **merge**: por repo
- **rango fuente-de-verdad**: `0972e565..777e5f6a`
- **SOURCE REF de extracción**: `develop-problematico~1` (SHA `777e5f6a`). NUNCA usar `develop-problematico` (tip = restore/vacío).
- **rama destino**: `develop` (tip `003a9b8f`).

## Resumen

Cambios en los 6 GitHub Actions workflows del backend: el pipeline de CI de PRs (`ci-pr.yml`) y los pipelines de deploy a Cloud Run (`deploy-dev.yml`, `deploy-staging.yml`, `deploy-prod.yml`), más el job de auditoría de alineación de servicios (`audit-service-alignment.yml`) y el job de reset de la DB de DEV desde PROD (`reset-dev-db-from-prod.yml`).

Los cambios NO son auto-contenidos de infra: arrastran configuración de runtime de otras features (renombre AI_SERVICE → COMPANION/NEXUS, activación de AUTH_ENABLED, esquema multi-tenant migración 224) y por eso pueden romper el deploy si se traen aislados.

## Objetivo

Alinear los workflows con el estado post-refactor del backend:
1. **ci-pr.yml**: quitar la dependencia del módulo legacy `CORE_GOVERNANCE_MODULE` (parte de la migración core/* → platform/*), simplificar `go mod download`, y agregar **gate de cobertura warning-only** (umbral 25%, no bloquea) + subida de artefacto `coverage.out`.
2. **deploy-*.yml**: reemplazar el bloque de env del servicio AI legacy (`AI_SERVICE_URL`, `AI_SERVICE_TIMEOUT_MS`, `SECRET_AI_SERVICE_KEY_NAME`, secret `ai-service-keys-*`) por el bloque **COMPANION + NEXUS** (BaseURL, JWT interno HS256, timeouts, retries, secret `companion-internal-jwt-secret-*`). En DEV además: `APP_ENV=dev` y `AUTH_ENABLED=true` (antes `false`).
3. **audit-service-alignment.yml** y los smoke de deploy: cambiar el set `watched` de variables de `{BASE_MANAGER_API, AI_SERVICE_URL}` a `{BASE_MANAGER_API, COMPANION_BASE_URL, NEXUS_BASE_URL}`.
4. **reset-dev-db-from-prod.yml**: hacer el restore compatible con el set de migraciones nuevo/multi-tenant: aplicar migraciones hasta `goto 224` (esquema additivo) antes del restore data-only, truncar tablas de auth para evitar colisión con seeds (`users.legacy_id=1`), dedupe de `TABLE DATA`/`SEQUENCE SET` en la lista de pg_restore, backfill de `tenant_id` + aplicar `000224_tenant_security_foundation.up.sql` y el resto de migraciones `up`, y correr el hardening post-restore vía `psql` por el Cloud SQL Proxy (como app user) en vez de `gcloud sql import`.

## Problema que resuelve

- CI seguía descargando un módulo de governance que ya no existe tras pasar a `platform/*` → fallaba o era ruido (`go mod download "$CORE_GOVERNANCE_MODULE"`).
- Los deploys inyectaban env del servicio AI viejo (`AI_SERVICE_*`) que el binario nuevo ya no lee; el binario nuevo lee `COMPANION_*`/`NEXUS_*` (ver `cmd/config/companion.go`). Sin este cambio, Companion/Nexus quedan sin configurar en Cloud Run.
- DEV corría sin auth (`AUTH_ENABLED=false`); el refactor identity/tenant requiere `AUTH_ENABLED=true` + `APP_ENV=dev`.
- El reset de DB asumía un `migrate ... up` lineal sobre esquema limpio; con el esquema multi-tenant (migración 224 additiva + validación estricta posterior) y datos PROD legacy con `tenant_id` NULL, el restore se rompía (FKs NOT VALID, seeds duplicados, secuencias). (Coincide con la nota de memoria *reset-local-db-from-prod (migraciones viejas)*.)

## Alcance en este repo (BE)

Exclusivamente los 6 YAML bajo `.github/workflows/`. No toca código Go, deps ni migraciones (esos artefactos los aportan OTRAS features; aquí solo se referencian desde el pipeline).

## Alcance en el otro repo (FE)

`.github/workflows` del FE (no inspeccionado en este paquete; coordinar con feature-020-FE). Probable: CI de PR (lint/typecheck/test/build), deploy a hosting/Cloud Run y, si comparte servicio, el mismo renombre de variables de entorno y/o `audit-service-alignment`. El job de auditoría podría ser compartido entre repos o duplicado.

## Fuera de alcance

- El código que LEE las nuevas env (`cmd/config/companion.go`, `cmd/config/auth.go`, `internal/axis/client.go`) → features 012 / 008.
- Las migraciones `migrations_v4/000224_*` y el directorio `migrations_v4/` → features 003 / 019.
- Los tests reales que alimentan el coverage gate → feature 025.
- La eliminación del módulo de governance / bumps de deps → migración platform / feature 021 / #124 (ya porteado go-jose, x/net).

## Comportamiento esperado

- **PR a develop**: corren `lint`, `build`, `tests-with-coverage` (genera `coverage.out`, lo sube como artefacto 7 días, imprime total y warning si <25% — NO falla), `security-scan` (govulncheck, no bloquea).
- **Deploy DEV/STG/PROD**: el contenedor recibe `COMPANION_*` y `NEXUS_*`; DEV con `AUTH_ENABLED=true` y `APP_ENV=dev`; smoke verifica que `COMPANION_BASE_URL`/`NEXUS_BASE_URL` no apunten a URLs taggeadas (`---`).
- **Reset DEV DB**: input opcional `reason`; aplica migraciones hasta 224, restore data-only deduplicado, backfill tenant, migraciones finales, sync de secuencias, hardening por psql, smoke opcional.

## Estado en dp~1 (777e5f6a)

Los 6 archivos están en su estado final/coherente en el SOURCE REF. Es el merge `#120 quick-fix`. Sintaxis YAML válida en apariencia. NO ejecutados/validados en este entorno (no se corre Actions acá).

## Criterios de aceptación

1. `ci-pr.yml` no referencia `CORE_GOVERNANCE_MODULE` y agrega los 3 steps de coverage (Tests with coverage / Upload artifact / Coverage gate warns).
2. Los 3 `deploy-*.yml` inyectan `COMPANION_*`/`NEXUS_*` y montan el secret `COMPANION_INTERNAL_JWT_SECRET`; ya no inyectan `AI_SERVICE_*` ni montan `ai-service-keys-*`.
3. DEV: `APP_ENV=dev` y `AUTH_ENABLED=true` presentes en `ENV_VARS`.
4. `watched = {BASE_MANAGER_API, COMPANION_BASE_URL, NEXUS_BASE_URL}` en audit + 3 smoke de deploy.
5. `reset-dev-db-from-prod.yml` usa `goto 224`, trunca tablas auth, dedupe de lista, aplica `000224_*.up.sql` + `up`, hardening por psql.
6. **No** introducir referencias a paths/migraciones/secrets que no existan en `develop` al momento del merge (ver dependencies.md).

## Endpoints / Modelos / UI / DB / Tests afectados

- **Endpoints/UI**: ninguno directo (infra). Indirecto: smoke tests pegan a la URL del servicio desplegado.
- **Modelos/DTOs**: ninguno en estos archivos. (Referencian struct `config.Companion`/`config.Nexus` en `cmd/config/companion.go`.)
- **DB/Migraciones**: `reset-dev-db-from-prod.yml` referencia `migrations_v4/` y `migrations_v4/000224_tenant_security_foundation.up.sql`, y tablas `public.users`, `auth_memberships`, `auth_role_permissions`, `auth_tenants`, `auth_roles`, `auth_permissions`, `fx_rates`. Estos NO se crean aquí.
- **Tests**: `ci-pr.yml` corre `go test ./... -coverprofile=...`; el gate depende de que existan tests (feature 025).

## Dependencias

- **Intra-repo (BE)**:
  - feature-012 (ai-companion-integration): aporta `cmd/config/companion.go`/`nexus`, `internal/axis/client.go` que LEEN `COMPANION_*`/`NEXUS_*`. **FUERTE**.
  - feature-008 (identity-tenant-context): `AUTH_ENABLED=true`, `APP_ENV` — `cmd/config/auth.go`, middlewares gin. **FUERTE para DEV**.
  - feature-003 (be-multitenant-db-hardening) + feature-019 (be-local-tooling-db-scripts): migración `000224`, dir `migrations_v4/`, lógica de reset. **FUERTE para reset-db**.
  - feature-025 (be-test-coverage): tests que el gate mide. **DÉBIL** (gate es warning-only).
  - migración platform / deps (#124 ya porteado): remoción de `CORE_GOVERNANCE_MODULE`. **DÉBIL**.
- **Cross-repo (FE)**: feature-020-FE (mismo feature). Coordinar variables compartidas y job de auditoría.

## Riesgos

- **Funcional**: deploy DEV pasa a `AUTH_ENABLED=true` → si el FE/clientes no mandan auth, rompe el ambiente DEV.
- **Técnico**: deploy referencia secret `companion-internal-jwt-secret-{dev,stg,prod}` y vars `COMPANION_BASE_URL_*`/`NEXUS_BASE_URL_*` que deben existir en GitHub/GCP Secret Manager; si faltan, el deploy falla en `gcloud run deploy`.
- **DB**: reset referencia `migrations_v4/000224_*.up.sql`; si en `develop` no está esa numeración/dir, el step de reset rompe (no afecta deploy normal, sí el workflow manual).
- **Cross-repo**: mergear solo BE con renombre de vars mientras FE/infra aún usa `AI_SERVICE_*` puede dejar variables huérfanas en algunos ambientes.

## DECISIÓN recomendada

**Partir en sub-PRs + ordenar por dependencia (no extraer tal cual de una).**

- `ci-pr.yml` (remoción de governance + coverage gate warning-only): **extraíble casi tal cual**, bajo riesgo. Puede ir primero. El coverage gate no bloquea aunque no haya tests.
- `audit-service-alignment.yml` (solo set `watched` + quitar línea en blanco final): bajo riesgo, pero **acoplado al renombre de vars**; idealmente junto con los deploy.
- `deploy-*.yml` (COMPANION/NEXUS + AUTH_ENABLED + APP_ENV): **arreglar antes / coordinar**. Solo mergear cuando estén porteadas 012 (config Companion/Nexus) y 008 (auth), y cuando existan las vars/secrets en GCP. De lo contrario el deploy queda roto.
- `reset-dev-db-from-prod.yml` (goto 224 + backfill tenant): **postergar hasta** que estén 003/019 (migración 224 + dir migrations_v4) en develop. Es un workflow manual, no bloquea CI ni deploy automático.

Ver `extraction-plan.md` para el desglose por sub-PR y orden cross-repo.
