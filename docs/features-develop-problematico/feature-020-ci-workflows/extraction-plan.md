# extraction-plan.md — feature-020 · CI / GitHub workflows (BE)

- **repo**: ponti-backend (`/home/pablocristo/Proyectos/pablo/ponti/core`)
- **rama base**: `develop` (tip `003a9b8f`)
- **SOURCE**: `develop-problematico~1` (SHA `777e5f6a`). NUNCA `develop-problematico` (tip = restore/vacío).
- **rama sugerida**: `pr/feature-020-ci-workflows-be`
- **estrategia**: NO un solo PR. Partir en 3 sub-PRs por nivel de riesgo/dependencia.

> Todos los comandos `git` de abajo son **SUGERENCIAS para un humano**. Este paquete NO ejecuta nada que mute el repo.

---

## Sub-PR A — CI de PR (bajo riesgo, sin dependencias bloqueantes)

**Archivos**: `.github/workflows/ci-pr.yml`

**PR title**: `ci(be): quitar CORE_GOVERNANCE_MODULE y agregar coverage gate (warning-only)`

**PR description**:
> Alinea el CI de PR con la migración a platform/*: elimina la descarga del módulo legacy `CORE_GOVERNANCE_MODULE` y simplifica `go mod download`. Agrega un gate de cobertura *warning-only* (umbral 25%, no bloquea PRs) que genera y sube `coverage.out` como artefacto (7 días). Baseline tentativo ≈30% post-refactor; subir threshold cuando haya más tests (feature-025).

**Pasos**:
1. `git checkout develop && git pull`
2. `git checkout -b pr/feature-020-ci-workflows-ci`
3. `git checkout 777e5f6a -- .github/workflows/ci-pr.yml`  *(archivo entero: el único cambio “externo” es la remoción de governance, que ya está resuelta en develop vía #124/platform — verificar que develop no dependa de esa env en otro workflow)*
4. `git diff --check`  *(detectar trailing whitespace / conflict markers)*
5. Verificar en develop que ningún otro workflow consuma `CORE_GOVERNANCE_MODULE`: `git -C <core> grep -n CORE_GOVERNANCE_MODULE develop -- .github/`
6. Commit + PR a develop.

**Qué NO traer**: nada extra; este archivo es self-contained.

---

## Sub-PR B — Deploys COMPANION/NEXUS + auditoría (gated por features 012 y 008)

**Archivos**: `.github/workflows/deploy-dev.yml`, `deploy-staging.yml`, `deploy-prod.yml`, `audit-service-alignment.yml`

**Precondiciones (DEBEN estar en develop ANTES de mergear)**:
- feature-012 porteada: `cmd/config/companion.go` (structs `Companion`/`Nexus`), `internal/axis/client.go`. Verificar: `git -C <core> ls-tree develop cmd/config/companion.go`.
- feature-008 porteada (para el bloque AUTH de DEV): `cmd/config/auth.go`, middlewares gin que respetan `AUTH_ENABLED`.
- En GitHub repo vars / GCP Secret Manager existen: `COMPANION_BASE_URL_{DEV,STG,PROD}`, `NEXUS_BASE_URL_{DEV,STG,PROD}`, secrets `companion-internal-jwt-secret-{dev,stg,prod}`. **Esto es config externa, no de git** — coordinar con quien administra GCP.

**PR title**: `ci(be): migrar deploys de AI_SERVICE a COMPANION/NEXUS + auth en DEV`

**PR description**:
> Reemplaza el bloque de env del servicio AI legacy (`AI_SERVICE_URL`, `AI_SERVICE_TIMEOUT_MS`, secret `ai-service-keys-*`) por la config de Companion + Nexus (BaseURL, JWT interno HS256 issuer/audience/ttl, timeout, retries; secret `companion-internal-jwt-secret-*`). DEV además pasa a `APP_ENV=dev` y `AUTH_ENABLED=true`. Actualiza el set `watched` de auditoría/smoke a `{BASE_MANAGER_API, COMPANION_BASE_URL, NEXUS_BASE_URL}`. Requiere features 012/008 y las vars/secrets correspondientes en GCP.

**Pasos**:
1. `git checkout develop && git checkout -b pr/feature-020-ci-workflows-deploys`
2. Traer enteros (el diff de cada deploy es 100% del feature, salvo el bloque AUTH que conviene revisar):
   - `git checkout 777e5f6a -- .github/workflows/deploy-staging.yml .github/workflows/deploy-prod.yml .github/workflows/audit-service-alignment.yml`
3. Para `deploy-dev.yml`, traer por hunks para poder **gatear el bloque AUTH** si feature-008 aún no está lista:
   - `git restore -p --source=777e5f6a -- .github/workflows/deploy-dev.yml`
   - Aceptar hunks COMPANION/NEXUS y `watched`; **rechazar** el hunk que pone `AUTH_ENABLED=true` + `APP_ENV=dev` si auth no está porteado (dejar `AUTH_ENABLED=false` hasta entonces).
4. `git diff --check`
5. Validar YAML: `python -c "import yaml,sys; [yaml.safe_load(open(f)) for f in ['.github/workflows/deploy-dev.yml','.github/workflows/deploy-staging.yml','.github/workflows/deploy-prod.yml','.github/workflows/audit-service-alignment.yml']]"` o `actionlint .github/workflows/*.yml`.
6. Commit + PR a develop. **No mergear** sin confirmar vars/secrets en GCP.

**Qué NO traer aquí**: `reset-dev-db-from-prod.yml` (va en Sub-PR C).

---

## Sub-PR C — Reset DEV DB compatible con multi-tenant (gated por features 003/019)

**Archivos**: `.github/workflows/reset-dev-db-from-prod.yml`

**Precondiciones**:
- En develop existen el dir `migrations_v4/` y `migrations_v4/000224_tenant_security_foundation.up.sql`. Verificar: `git -C <core> ls-tree develop migrations_v4/000224_tenant_security_foundation.up.sql`. Si NO existe → **postergar este sub-PR**.
- Features 003 (multitenant-db-hardening) y 019 (be-local-tooling-db-scripts) porteadas.

**PR title**: `ci(be): reset-dev-db restore-compatible con esquema multi-tenant (goto 224 + backfill tenant)`

**PR description**:
> Hace el reset de DEV-from-PROD compatible con el set de migraciones multi-tenant: aplica migraciones hasta `goto 224` (esquema additivo) antes del restore data-only; trunca explícitamente tablas de auth (users/auth_*) para no chocar con seeds (`users.legacy_id=1`); deduplica entradas `TABLE DATA`/`SEQUENCE SET` en la lista de pg_restore; backfillea `tenant_id` aplicando `000224_*.up.sql` y luego `up` (validación estricta); corre el hardening post-restore vía `psql` por el Cloud SQL Proxy como app user (en vez de `gcloud sql import`, que corre como import user y no puede otorgar privilegios sobre tablas app como `fx_rates`). Agrega input opcional `reason`.

**Pasos**:
1. `git checkout develop && git checkout -b pr/feature-020-ci-workflows-resetdb`
2. `git checkout 777e5f6a -- .github/workflows/reset-dev-db-from-prod.yml`
3. `git diff --check`
4. Confirmar que `migrations_v4/000224_tenant_security_foundation.up.sql` existe en el árbol; si no, abortar.
5. Commit + PR. **Es workflow manual**, no bloquea CI/deploy: bajo apuro pero alto riesgo si se corre con esquema/datos incompatibles.

---

## Archivos enteros vs parciales (resumen)

- **enteros**: `ci-pr.yml`, `deploy-staging.yml`, `deploy-prod.yml`, `audit-service-alignment.yml`, `reset-dev-db-from-prod.yml`.
- **parcial (restore -p)**: `deploy-dev.yml` — separar el bloque AUTH del bloque COMPANION/NEXUS.

## Migraciones / tests a incluir

- Ninguna migración ni test se crea en este paquete. Solo se *referencian* (`000224_*`, tests para coverage). Asegurar que están en develop antes de los sub-PRs B/C.

## Dependencias previas (orden)

1. platform/* (governance removido) + #124 — **ya en develop**.
2. feature-012 (companion/nexus) — antes de Sub-PR B.
3. feature-008 (auth) — antes del bloque AUTH de DEV en Sub-PR B.
4. features 003 + 019 (migración 224, migrations_v4) — antes de Sub-PR C.
5. feature-025 (tests) — opcional, mejora el coverage gate (no bloquea).

## Coordinación con el otro repo (FE)

- **Orden recomendado: coordinado, BE config-first**. Las vars/secrets de Cloud Run (BE) deben existir antes de que el FE apunte a Companion/Nexus.
- Si `audit-service-alignment.yml` también vive en FE con el mismo renombre, mergear ambos lados en la misma ventana para no dejar el set `watched` inconsistente.
- El cambio `AUTH_ENABLED=true` en DEV impacta al FE: coordinar que el FE de DEV mande credenciales ANTES de mergear Sub-PR B (bloque AUTH).

## Comandos git SUGERIDOS (solo lectura + traer paths)

```bash
core=/home/pablocristo/Proyectos/pablo/ponti/core
git -C "$core" checkout develop
git -C "$core" checkout -b pr/feature-020-ci-workflows-ci
git -C "$core" checkout 777e5f6a -- .github/workflows/ci-pr.yml
git -C "$core" diff --check
# parcial para deploy-dev (gatear AUTH):
git -C "$core" restore -p --source=777e5f6a -- .github/workflows/deploy-dev.yml
# inspección:
git -C "$core" show 777e5f6a:.github/workflows/deploy-prod.yml | head -200
git -C "$core" ls-tree develop migrations_v4/000224_tenant_security_foundation.up.sql
```

## Qué NO traer

- No tocar `cmd/config/*`, `internal/axis/*`, `migrations_v4/*` desde este paquete (pertenecen a 012/008/003/019).
- No reactivar `CORE_GOVERNANCE_MODULE`.
- No traer `develop-problematico` (tip).

## Qué podría romperse

- **Deploy** si faltan vars/secrets COMPANION/NEXUS en GCP → `gcloud run deploy` falla.
- **Ambiente DEV** si `AUTH_ENABLED=true` y los clientes no mandan auth.
- **Reset DB** si `migrations_v4/000224_*` no está o la numeración difiere.
- **CI** prácticamente no se rompe (coverage warning-only; govulncheck no bloquea).

## Cómo detectar extracción incompleta

- `git -C <core> diff develop..pr/feature-020-ci-workflows-* -- .github/workflows/` debe coincidir hunk-a-hunk con `git diff 0972e565..777e5f6a` para los archivos del sub-PR.
- `grep -R "AI_SERVICE_URL\|CORE_GOVERNANCE_MODULE\|ai-service-keys" .github/workflows/` debe quedar **vacío** tras Sub-PR A+B.
- `grep -R "COMPANION_BASE_URL\|NEXUS_BASE_URL" .github/workflows/` debe aparecer en los 3 deploys + audit.

## Qué validar antes del PR

- `actionlint .github/workflows/*.yml` (o `yamllint`).
- Confirmar existencia de vars/secrets en GCP (Sub-PR B).
- Confirmar migración 224 / dir migrations_v4 (Sub-PR C).

## Qué hacer después de mergear

- Disparar manualmente `Reset DEV DB from PROD` (con `reason`) para validar el flujo nuevo en un momento seguro.
- Observar el primer deploy DEV: que el contenedor levante con COMPANION/NEXUS y AUTH.
- Confirmar que el artefacto `coverage-report` aparece en el primer PR posterior.
