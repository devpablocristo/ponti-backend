# validation.md — feature-020 · CI / GitHub workflows (BE)

> `<core>` = `/home/pablocristo/Proyectos/pablo/ponti/core`

## Checklist pre-PR (común)

- [ ] Rama base = `develop` (no `develop-problematico`).
- [ ] Archivos traídos desde SOURCE `777e5f6a` con `git checkout 777e5f6a -- <path>` (copia exacta), no editados a mano.
- [ ] `git diff develop..<rama> -- .github/workflows/<file>` coincide hunk-a-hunk con `git diff 0972e565..777e5f6a -- .github/workflows/<file>`.
- [ ] `git diff --check` sin trailing whitespace / conflict markers.
- [ ] Lint de workflows: `actionlint .github/workflows/*.yml` (o `yamllint .github/workflows/`).
- [ ] Parse YAML: `python -c "import yaml; yaml.safe_load(open('.github/workflows/ci-pr.yml'))"` (repetir por archivo).

## Sub-PR A — ci-pr.yml

- [ ] `grep -n CORE_GOVERNANCE_MODULE .github/workflows/ci-pr.yml` → vacío.
- [ ] Existen los 3 steps nuevos: `Tests with coverage`, `Upload coverage artifact`, `Coverage gate (warns, no fail por ahora)`.
- [ ] `grep -n "threshold=25" .github/workflows/ci-pr.yml` presente.
- [ ] Ningún otro workflow en develop usa `CORE_GOVERNANCE_MODULE`: `git -C <core> grep -n CORE_GOVERNANCE_MODULE develop -- .github/`.
- [ ] (opcional, local) `go test ./... -coverprofile=coverage.out -covermode=atomic && go tool cover -func=coverage.out | tail -1` compila y produce un total.

## Sub-PR B — deploy-*.yml + audit

- [ ] `grep -RIn "AI_SERVICE_URL\|AI_SERVICE_TIMEOUT_MS\|ai-service-keys\|SECRET_AI_SERVICE_KEY_NAME" .github/workflows/` → vacío.
- [ ] `grep -RIn "COMPANION_BASE_URL\|NEXUS_BASE_URL" .github/workflows/` aparece en `deploy-dev.yml`, `deploy-staging.yml`, `deploy-prod.yml`, `audit-service-alignment.yml`.
- [ ] Cada deploy monta `COMPANION_INTERNAL_JWT_SECRET=...:latest` en `--set-secrets`.
- [ ] `watched = {"BASE_MANAGER_API", "COMPANION_BASE_URL", "NEXUS_BASE_URL"}` en audit y en los 3 smoke de deploy.
- [ ] DEV: `APP_ENV=dev` y `AUTH_ENABLED=true` en `ENV_VARS` **solo si feature-008 lista**; si no, dejar `AUTH_ENABLED=false`.
- [ ] **Config GCP** (fuera de git, confirmar manualmente):
  - [ ] repo vars `COMPANION_BASE_URL_{DEV,STG,PROD}`, `NEXUS_BASE_URL_{DEV,STG,PROD}` definidas.
  - [ ] secrets `companion-internal-jwt-secret-{dev,stg,prod}` existen en Secret Manager.
- [ ] Código que lee las vars en develop: `git -C <core> ls-tree develop cmd/config/companion.go`.

## Sub-PR C — reset-dev-db-from-prod.yml

- [ ] `git -C <core> ls-tree develop migrations_v4/000224_tenant_security_foundation.up.sql` existe; si no, postergar.
- [ ] `grep -n "goto 224" .github/workflows/reset-dev-db-from-prod.yml` presente.
- [ ] El step de hardening usa `psql ... -f "$hardening_sql"` y `set -euo pipefail` (ya no `gcloud sql import` con `set +e`).
- [ ] Input `reason` presente en `workflow_dispatch.inputs`.
- [ ] `scripts/db/hardening_post_restore.sql` existe como fallback: `git -C <core> ls-tree develop scripts/db/hardening_post_restore.sql`.

## Validación manual / runtime

- **CI (Sub-PR A)**: abrir un PR de prueba a develop → confirmar que corren lint/build/tests/security-scan y que aparece el artefacto `coverage-report`. El gate imprime "Total coverage: X%" y un `::warning::` si <25% pero NO falla.
- **Deploy (Sub-PR B)**: tras deploy DEV, verificar en Cloud Run que el contenedor tiene env `COMPANION_BASE_URL`/`NEXUS_BASE_URL` y monta el secret; que el smoke no marque URLs taggeadas; que el servicio levanta con AUTH.
- **Reset DB (Sub-PR C)**: disparar `Reset DEV DB from PROD` manualmente con `reason` en ventana segura; observar que aplica migraciones hasta 224, restaura data-only sin colisión de seeds, backfillea tenant, sincroniza secuencias y corre hardening sin error.

## Tests sugeridos

- **BE**: `go build ./...` y `go test ./... -coverprofile=coverage.out -covermode=atomic` (lo que corre el CI). No hay paquete específico de esta feature.
- **Workflows**: `actionlint`, dry-run con `act` si está disponible (opcional).
- **FE** (otro repo, coordinar): `yarn test` / `yarn build` / e2e según feature-020-FE.

## Casos borde

- Cobertura 0% (sin tests): el gate debe imprimir warning y **pasar** (no fallar).
- `HARDENING_SQL_URI` vacío: el step de hardening se saltea (`if: env.HARDENING_SQL_URI != ''`).
- `HARDENING_SQL_URI` = path local vs `gs://` vs fallback `scripts/db/...`: el YAML cubre los 3; validar cada rama.
- PROD: `RUN_MIGRATIONS_ON_STARTUP=false` y `--min-instances=1` se conservan; NO debe cambiarse al portar.

## Qué revisar en UI / API / DB / env

- **UI/API**: ninguno directo. Indirecto: el servicio desplegado responde al smoke.
- **DB**: solo en reset-db (DEV). Verificar tablas `users`/`auth_*`/`fx_rates` post-reset.
- **env**: COMPANION/NEXUS en Cloud Run; `AUTH_ENABLED`/`APP_ENV` en DEV.

## Qué validar en el otro repo (FE)

- Que feature-020-FE haya rotado (o no) a Companion/Nexus de forma consistente.
- Que el FE DEV mande credenciales si BE DEV tiene `AUTH_ENABLED=true`.
- Estado del job de auditoría si está duplicado.

## Señales de incompletitud / incompatibilidad

- Quedan referencias a `AI_SERVICE_*`/`ai-service-keys` en algún workflow → extracción parcial.
- `gcloud run deploy` falla con "secret not found" → falta secret en GCP.
- El reset rompe en "no such file" sobre `migrations_v4/000224_*` → falta migración (features 003/019).
- DEV inaccesible tras deploy → `AUTH_ENABLED=true` sin clientes preparados.
