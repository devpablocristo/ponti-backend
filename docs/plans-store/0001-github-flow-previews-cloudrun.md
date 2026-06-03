# Informe técnico: Migración a GitHub Flow + Preview Environments en Cloud Run

> **Estado:** Deferred — **NADA implementado** · **Fecha:** 2026-06-03

> **Alcance de este documento:** auditoría + plan. **No se implementa nada**, no se modifican
> archivos ni configuración, no se toca `develop`. El retiro de `develop` queda como fase futura **opcional**.

---

## Context

El repositorio `ponti/core` (backend Go → Cloud Run) opera hoy con un modelo **GitFlow**:
`feature → develop → DEV` y `develop → main → STAGING → (promoción manual) → PROD`. Se quiere
evolucionar a **GitHub Flow** (`feature/* → PR → main`) añadiendo **Preview Environments
automáticos por PR** en Cloud Run, de modo que cada PR tenga un entorno desplegado, validable por
QA, y destruido al cerrarse — sin intervención manual y al menor costo posible en GCP.

El sistema CI/CD actual es **maduro y bien diseñado** (WIF sin claves estáticas, imágenes
inmutables por SHA, smoke tests, schema guardrails, promoción STG→PROD por re-tag de imagen,
rollback con guardas de migración, registro de deployments). El objetivo **no** es rehacerlo, sino
**añadir** la capa de previews y **reorientar el punto de integración** de `develop` a `main`,
preservando toda la maquinaria de staging/prod.

**Decisiones del usuario (confirmadas):**
1. **Topología preview:** un **servicio Cloud Run efímero por PR** (`ponti-backend-pr-<n>`).
2. **DB preview:** **instancia Cloud SQL propia y dedicada** (`ponti-preview-db`, tier shared-core, detenible cuando no hay PRs), con una **base por PR** (`ponti_preview_pr_<n>`) que migra al arranque y se `DROP` al cerrar. Aísla prod del riesgo (conexiones/contención/blast radius). **No** se usa la instancia compartida de dev/stg/prod.
3. **Auth preview:** `AUTH_ENABLED=false` (igual que DEV hoy).
4. **`develop`/DEV:** **no se tocan ahora.** El plan es aditivo; el retiro de `develop` es fase futura opcional.

---

## A. Estado actual

### A.1 Modelo de ramas
| Rama | Rol hoy | Despliega a |
|------|---------|-------------|
| `feature/*` | desarrollo | — (PR a `develop`) |
| `develop` | integración activa | DEV (`deploy-dev.yml`) |
| `main` | estable | STAGING (`deploy-staging.yml`) |
| tags `vX.Y.Z` | release | base para promoción PROD |
| `develop-problematico`, `backup/develop-*`, `develop-orig-updated` | trabajo pendiente / respaldos | — |

### A.2 Workflows (`.github/workflows/`)
| Workflow | Trigger | Función | Concurrency |
|----------|---------|---------|-------------|
| `ci-pr.yml` | `pull_request` → **develop** (paths Go) | lint + build + test + govulncheck | cancel por PR |
| `deploy-dev.yml` | push **develop** / dispatch | build+push imagen (SHA) → Cloud Run DEV, migraciones on-startup, smoke, guardrails | cancel |
| `deploy-staging.yml` | push **main** / dispatch | igual, → Cloud Run STAGING | cancel |
| `deploy-prod.yml` | `workflow_dispatch` (staging_sha) | valida `QA_APPROVED`, re-tag imagen STG→PROD, deploy, `min-instances=1`, `RUN_MIGRATIONS_ON_STARTUP=false` | serial |
| `release.yml` | dispatch | SemVer + tag anotado + GitHub Release | serial |
| `rollback-staging.yml` / `rollback-prod.yml` | dispatch | redeploy imagen previa, guarda de compatibilidad de migración | serial |
| `apply-migrations-prod.yml` | dispatch (dry-run/apply) | aplica migraciones PROD por separado + backup | serial |
| `reset-dev-db-from-prod.yml` / `reset-stg-db-from-prod.yml` | dispatch | clona datos PROD→DEV/STG con guardas hardcodeadas | serial |
| `approve-staging.yml` | dispatch | marca último `SMOKE_OK` como `QA_APPROVED` | serial |
| `audit-service-alignment.yml` | cron lunes 09:00 UTC | audita 100% tráfico en latestRevision + URLs no taggeadas (dev/stg/prod) | cancel |

### A.3 GCP / Cloud Run / Artifact Registry
- **Imagen:** `Dockerfile` multi-stage (`golang:1.26.3-alpine` → `alpine`), build `prod_binary` + `migrate_binary`, copia `migrations_v4/`, puerto **8080**, `entrypoint.sh` (migra si `RUN_MIGRATIONS_ON_STARTUP=true`, luego arranca API). Secreto de build `go_modules_token` (módulos privados `github.com/devpablocristo/*`).
- **Artifact Registry:** `${REGION}-docker.pkg.dev/${PROJECT}/${ARTIFACT_REGISTRY}/${IMAGE_NAME}:${SHA}` (tag = SHA, inmutable). PROD añade tag `:prod`.
- **Cloud Run deploy** (patrón en `deploy-dev.yml:121-133`): `--service-account`, `--add-cloudsql-instances`, `--set-env-vars` (~25), `--set-secrets` (Secret Manager `:latest`), `--allow-unauthenticated`. DEV/STG sin `min-instances` (escala a 0); PROD `--min-instances=1`.
- **Cloud SQL:** PostgreSQL + pgvector. Conexión por socket `/cloudsql/<INSTANCE>`. **Una instancia compartida** (proyecto DEV) sirve dev/stg vía bases distintas (`DB_NAME_*`); PROD tiene su instancia. Proxy `cloud-sql-proxy:2.11.0` se usa en CI para schema guardrails.

### A.4 WIF / IAM / Secrets / Environments
- **WIF por entorno:** `WIF_PROVIDER_{DEV,STG,PROD}` + `WIF_SERVICE_ACCOUNT_{DEV,STG,PROD}` (sin claves estáticas). Cloud Run corre bajo `CLOUD_RUN_SERVICE_ACCOUNT_{DEV,STG,PROD}`.
- **GitHub Environments:** `dev` (implícito), `staging`, `prod` (protegido en deploy/rollback/migrations).
- **GitHub Secrets:** `CORE_REPO_READ_TOKEN`, `DB_PASSWORD_{DEV,STG,PROD}`, `X_API_KEY_{DEV,STG,PROD}`, `SMOKE_AUTH_BEARER_TOKEN_{DEV,STG,PROD}`.
- **Secret Manager:** `db-password-{dev,stg,prod}`, `x-api-key-{dev,stg,prod}`, `ai-service-keys-dev`, `review-api-key-dev`.
- **Vars repo:** proyecto/región/registry/servicio/SA/WIF/cloudsql/DB por entorno + config app compartida (`API_VERSION`, `DB_*`, `MIGRATIONS_DIR`, `WORDS_SUGGESTER_*`, `REPORT_SCHEMA`, etc.).

### A.5 App / DB / migraciones (lo relevante para previews)
- Go monolito, `github.com/devpablocristo/ponti-backend`, Gin, GORM + `golang-migrate` v4.
- **Migraciones:** `migrations_v4/` (hasta `000231`), corren **on-startup** vía `entrypoint.sh`. Primer arranque sobre DB vacía ≈ 231 migraciones (~30–60 s); arranques posteriores son incrementales/rápidos.
- **Boot sin datos:** la app arranca sin seed. Auto-crea la DB solo en local (se salta si `K_SERVICE` está seteado, i.e. en Cloud Run).
- **Auth:** `AUTH_ENABLED=false` desadhiere Identity Platform (DEV ya lo hace). Servicios AI/Review son **opcionales** (degradan si la URL está vacía).
- **Health:** `GET /api/v1/healthz`, `/version`, `/ping`. Middleware exige header `X-API-KEY`.
- **pgvector:** la instancia compartida ya lo tiene habilitado; cada DB necesita su `CREATE EXTENSION` (asumido en migraciones — **verificar**, ver Riesgo R3).

### A.6 Tagging y rollback actuales
- **Tags imagen:** SHA (inmutable) en todos; `:prod` mutable en prod; versiones de servicio `dev-<sha>` / `stg-<sha>` / `prod-<sha>`.
- **Releases:** SemVer (`release.yml`) sobre SHA aprobado en staging.
- **Rollback:** redeploy de imagen previa por SHA con guarda de compatibilidad de migración; PROD y STG tienen workflow dedicado.

---

## B. Estado objetivo

```
feature/*  ──PR──▶  [ci-pr: lint/build/test]  +  [preview.yml → Cloud Run preview por PR]
                                │                         │
                          review / QA  ◀── URL estable + comentario en PR
                                │
                          merge a main
                                │
                    [deploy-staging.yml] ──▶ STAGING (sin cambios)
                                │
                approve-staging → release → deploy-prod (promoción manual, sin cambios)

  Al cerrar/mergear el PR  ──▶  [cleanup-preview.yml] borra servicio + DROP DATABASE
```

### B.1 Topología de preview (decisión: servicio por PR)
- **Servicio efímero `ponti-backend-pr-<n>`** en el **proyecto DEV** (reutiliza WIF/SA/registry/secrets — cero IAM nuevo salvo el binding `cloudsql.client` sobre la instancia de preview).
- **URL estable por PR:** `https://ponti-backend-pr-<n>-<hash>-<region>.run.app` (estable mientras el servicio exista; el `<hash>` lo asigna Cloud Run y no cambia entre revisiones del mismo servicio).
- **DB por PR en instancia DEDICADA de preview** (`ponti-preview-db`, no la compartida con prod): base `ponti_preview_pr_<n>`, migrada on-startup, `DROP` al cerrar. Usuario propio + secret `db-password-preview`.
- **Sin `min-instances`** (escala a 0 → costo cómputo en reposo ≈ $0); `--max-instances` bajo (p.ej. 2) como tope de costo; `AUTH_ENABLED=false`.
- **No toca** los servicios dev/stg/prod → **no dispara** los guardrails de `audit-service-alignment.yml` ni los de deploy.

### B.2 Workflows objetivo (los 4 que pide el brief)
| Nuevo/cambio | Trigger | Función |
|--------------|---------|---------|
| **`preview.yml`** (nuevo) | `pull_request: [opened, synchronize, reopened]` → `main` | build+push imagen (SHA), `CREATE DATABASE IF NOT EXISTS` preview, `gcloud run deploy ponti-backend-pr-<n>`, resolver URL, comentario sticky en el PR |
| **`cleanup-preview.yml`** (nuevo) | `pull_request: [closed]` → `main` | `gcloud run services delete ponti-backend-pr-<n>`, `DROP DATABASE`, borrar tag de imagen del PR, actualizar comentario |
| **`deploy-staging.yml`** (se mantiene) | push `main` | STAGING (sin cambios) |
| **`deploy-prod.yml`** (se mantiene) | dispatch / promoción | PROD (sin cambios) |
| **`ci-pr.yml`** (cambio menor, fase 2) | añadir `pull_request → main` | gate de lint/build/test también en PRs a `main` |

### B.3 Diseño de `preview.yml` (boceto, no implementación)
```yaml
name: Preview (Cloud Run por PR)
on:
  pull_request:
    types: [opened, synchronize, reopened]
    branches: [main]
    paths: ["**/*.go","go.mod","go.sum","migrations_v4/**","scripts/**","Dockerfile",".github/workflows/preview.yml"]
permissions: { contents: read, id-token: write, pull-requests: write }
concurrency: { group: preview-pr-${{ github.event.pull_request.number }}, cancel-in-progress: true }
# Variables clave derivadas:
#   PR_NUM   = ${{ github.event.pull_request.number }}
#   SERVICE  = ponti-backend-pr-${PR_NUM}
#   DB_NAME  = ponti_preview_pr_${PR_NUM}
#   IMAGE    = ${REGION}-docker.pkg.dev/${PROJECT_DEV}/${REGISTRY}/${IMAGE_NAME}:<PR-head-SHA>
# Pasos: auth WIF_DEV → setup gcloud → docker build/push →
#   (proxy cloud-sql) CREATE DATABASE IF NOT EXISTS ${DB_NAME} →
#   gcloud run deploy ${SERVICE} --add-cloudsql-instances=${CLOUDSQL_INSTANCE_DEV}
#       --set-env-vars=...,DEPLOY_ENV=preview,DB_NAME=${DB_NAME},AUTH_ENABLED=false,RUN_MIGRATIONS_ON_STARTUP=true
#       --set-secrets=DB_PASSWORD=db-password-dev:latest,X_API_KEY=x-api-key-dev:latest
#       --labels=managed-by=preview,pr=${PR_NUM} --max-instances=2 --allow-unauthenticated →
#   resolver status.url → comentario sticky en el PR (actions/github-script, marcador <!--preview-->)
```
> Postgres no tiene `CREATE DATABASE IF NOT EXISTS`; se hace con guarda:
> `SELECT 1 FROM pg_database WHERE datname='...'` y `CREATE DATABASE` solo si no existe (idempotente).

### B.4 Diseño de `cleanup-preview.yml` (boceto)
```yaml
name: Cleanup Preview
on: { pull_request: { types: [closed], branches: [main] } }
permissions: { contents: read, id-token: write, pull-requests: write }
concurrency: { group: preview-pr-${{ github.event.pull_request.number }}, cancel-in-progress: false }
# Pasos: auth WIF_DEV →
#   gcloud run services delete ponti-backend-pr-${PR_NUM} --quiet || true →
#   (proxy) terminar conexiones + DROP DATABASE ponti_preview_pr_${PR_NUM} || true →
#   gcloud artifacts docker tags delete <imagen>:<SHA> (o dejar a política de retención) →
#   actualizar comentario: "🧹 preview destruido".
```

---

## C. Gap Analysis

| # | Brecha | Estado actual | Acción objetivo | Esfuerzo |
|---|--------|---------------|-----------------|----------|
| G1 | No existen previews por PR | — | crear `preview.yml` + `cleanup-preview.yml` | medio |
| G2 | CI solo gatea PRs a `develop` | `ci-pr.yml: branches:[develop]` | añadir `main` a triggers (fase 2) | bajo |
| G3 | Punto de integración es `develop` | GitFlow | dirigir features a `main` (GitHub Flow) | proceso |
| G4 | DB efímera por PR inexistente | bases fijas por entorno | `CREATE/DROP DATABASE` por PR en instancia compartida | bajo |
| G5 | SA/WIF preview sin permisos de borrar servicios/DB | SA DEV despliega | verificar `run.admin` (o `run.developer`) + `serviceAccountUser` + privilegio `CREATEDB` del usuario DB | verificación |
| G6 | Comentario automático en PR | — | `actions/github-script` con comentario sticky | bajo |
| G7 | Limpieza de imágenes por PR | sin política | borrar tag al cerrar o cleanup policy en Artifact Registry | bajo |
| G8 | Retiro de `develop`/`deploy-dev` | activos | **fase futura opcional, NO ahora** | aplazado |

---

## D. Riesgos

| ID | Riesgo | Prob. | Impacto | Mitigación |
|----|--------|-------|---------|------------|
| R1 | El SA de DEV no puede crear/borrar servicios Cloud Run efímeros | media | bloqueante | verificar/otorgar `roles/run.admin` o `roles/run.developer` + `roles/iam.serviceAccountUser` sobre `CLOUD_RUN_SERVICE_ACCOUNT_DEV` antes de fase 1 |
| R2 | El usuario de preview no tiene `CREATEDB`/`DROP` | media | bloqueante | crear usuario propio con `CREATEDB` en la instancia dedicada (Fase 0) |
| R3 | `CREATE EXTENSION` falla en DB nueva | baja | medio | migraciones solo usan `unaccent`/`pg_trgm`/`pgcrypto` (estándar Cloud SQL). ✅ verificado |
| R3b | **Contención preview↔prod (instancia compartida)** | — | — | **eliminado**: previews en instancia DEDICADA, aislados de prod |
| R4 | 231 migraciones en cada cold start lentas | baja | medio (UX) | DB persiste entre pushes del mismo PR → solo el 1er deploy es lento; aceptable para preview |
| R5 | Migración divergente al re-pushear (down/cambio de migración en el PR) | media | medio | DB por PR es descartable → opción "recrear DB" (drop+create) en `workflow_dispatch`/reintento |
| R6 | Fuga de costos (muchos PRs, imágenes acumuladas) | media | bajo | escala a 0, `--max-instances=2`, borrado de imagen al cerrar + cleanup policy; `concurrency` cancela builds superados |
| R7 | Previews ensucian guardrails de dev/stg/prod | baja | medio | servicio **separado** `ponti-backend-pr-<n>`; auditoría solo mira frontend/backend/ai → no afecta |
| R8 | PR cerrado sin merge deja recursos huérfanos | baja | bajo | `cleanup` corre en `closed` (cubre merge **y** cierre); job de barrido semanal opcional por label `managed-by=preview` |
| R9 | Secretos expuestos en previews `--allow-unauthenticated` | media | medio | DB en instancia dedicada (sin datos de prod); `AUTH_ENABLED=false`; reusar `x-api-key-dev`; nunca secretos de prod |
| R10 | Romper el flujo actual durante la transición | baja | alto | plan **aditivo**: fases 1–3 no modifican `develop`/`deploy-dev`/staging/prod; todo reversible borrando 2 archivos |

---

## E. Roadmap por fases

> Principio rector: **aditivo y reversible**. `develop`, `deploy-dev` y staging/prod **no se tocan**
> en las fases 1–3. La fase 4 (retiro de `develop`) es **futura y opcional**, fuera de este ejercicio.

### Fase 0 — Verificaciones + provisión de instancia preview (sin cambios en repo)
- **Crear instancia dedicada `ponti-preview-db`** (Postgres, tier shared-core, proyecto `new-ponti-dev`), + usuario propio con `CREATEDB`, + secret `db-password-preview`. Opcional: programar stop/start cuando no hay PRs.
- **IAM:** binding `roles/cloudsql.client` sobre la nueva instancia para `cloudrun-sa` y `github-actions` (WIF) SA. Confirmar que el WIF SA ya tiene `run.developer/admin` + `serviceAccountUser` + `artifactregistry.writer` (lo usa deploy-dev).
- Confirmar que el usuario de preview puede `CREATE/DROP DATABASE` en la instancia nueva.
- Migraciones: solo `unaccent`/`pg_trgm`/`pgcrypto` (estándar en Cloud SQL) — sin riesgo. ✅ verificado.
- Definir variables nuevas: `CLOUDSQL_INSTANCE_PREVIEW`, `PREVIEW_DB_PREFIX` (= `ponti_preview_pr_`).
- Decidir política de retención de imágenes en Artifact Registry.

### Fase 1 — Previews (aditivo, riesgo casi nulo)
- Añadir `preview.yml` (PR→main, opened/synchronize/reopened) y `cleanup-preview.yml` (PR→main, closed).
- Servicio por PR en proyecto DEV, DB por PR, `AUTH_ENABLED=false`, comentario sticky.
- **No** modifica ningún workflow existente.

### Fase 2 — CI en `main` (cambio mínimo)
- Extender `ci-pr.yml` para gatear también `pull_request → main` (mantener `develop`).

### Fase 3 — Validación funcional
- Abrir 1–2 PRs reales a `main`, validar: preview se crea/actualiza/comenta, smoke OK, merge → staging despliega, cierre → cleanup borra servicio + DB.
- Medir tiempos y costo real (build minutes, storage).

### Fase 4 — Retiro de `develop` (FUTURO / OPCIONAL — no ahora)
- Solo cuando el trabajo de `develop` (incl. descomposición `develop-problematico`) esté mergeado a `main`.
- Quitar `develop` de triggers de `ci-pr.yml`; **congelar** `deploy-dev.yml` (mantener archivo, deshabilitar trigger) y observar; luego eliminar; finalmente borrar la rama `develop`.
- Reevaluar destino del entorno DEV (retirar vs. conservar como demo).

---

## F. Definition of Done por fase

- **Fase 0 DoD:** checklist IAM/DB/extension confirmado por escrito; sin cambios pendientes que bloqueen previews.
- **Fase 1 DoD:** al abrir un PR a `main` se crea `ponti-backend-pr-<n>`, responde `200` en `/api/v1/healthz`, publica URL en el PR; un nuevo push actualiza el mismo servicio/URL; `develop`/staging/prod sin cambios de comportamiento.
- **Fase 2 DoD:** un PR a `main` ejecuta lint+build+test (igual que hoy en `develop`); PRs a `develop` siguen funcionando.
- **Fase 3 DoD:** ciclo completo demostrado (crear→actualizar→merge→staging→cleanup) en PR real; recursos cero tras cerrar; costo incremental medido y aceptable.
- **Fase 4 DoD (futuro):** `develop` drenado y eliminado, `deploy-dev` retirado, sin despliegues rotos, documentación actualizada.

---

## G. Plan de rollback completo

| Escenario | Rollback |
|-----------|----------|
| `preview.yml`/`cleanup-preview.yml` fallan o molestan | **Borrar los 2 archivos** (o vaciar su `on:`). Nada más cambió → flujo intacto. |
| Recursos preview huérfanos | `gcloud run services list --filter="labels.managed-by=preview"` + `services delete`; `DROP DATABASE` de bases `ponti_preview_pr_*` en la instancia de preview; borrar tags de imagen. (Si todo falla: borrar la instancia de preview entera, no afecta prod.) |
| Cambio en `ci-pr.yml` (fase 2) molesta | revertir el trigger `main` (1 línea); `develop` sigue gateado. |
| Fase 4 sale mal (futuro) | restaurar trigger de `deploy-dev.yml`/`ci-pr.yml` (reversibles); las ramas `backup/develop-*` permiten recrear `develop`. |
| Staging/Prod | **sin cambios** en este plan → siguen los rollbacks existentes (`rollback-staging.yml` / `rollback-prod.yml`). |

---

## Estrategia de costo (Opción A vs B vs C)

| Criterio | A: Revisions+Tags (1 servicio) | **B: Servicio por PR (elegida)** | C: Alternativas |
|----------|-------------------------------|----------------------------------|-----------------|
| Costo cómputo en reposo | ~$0 (escala a 0) | ~$0 (escala a 0) | Postgres sidecar: pierde datos y cold-start lento; GKE namespaces: caro/pesado |
| Aislamiento env/DB | a nivel revision (servicio comparte cloudsql/labels) | **total por PR** | variable |
| Cleanup | quitar tag + delete revision + revision "parking" al 100% | **`services delete` (1 comando)** | variable |
| Complejidad | gestión de tags/tráfico | **baja** | alta |
| Riesgo guardrails | mayor (URLs `---` taggeadas) | **nulo (servicio aparte)** | — |
| Mantenimiento | 1 servicio en consola | N servicios efímeros | — |

**Conclusión:** B (servicio por PR) + **instancia Cloud SQL dedicada para previews**. Drivers de costo,
todos acotados: **minutos de build** (GitHub Actions), **storage de imágenes** (prune/retención), e
**instancia de preview** (shared-core ~USD 8–12/mes 24/7, ~USD 1–3/mes si se detiene cuando no hay PRs).
Cloud Run en reposo ≈ $0. El costo extra de la instancia compra **aislamiento total de prod** (sin
contención de conexiones/CPU ni blast radius) — la razón por la que NO se reusa la instancia compartida.

---

## Impacto verificado (env / DB / secretos / migraciones / IAM / auth / red / dominios)

| Área | Impacto en previews | Acción |
|------|---------------------|--------|
| Env vars | reusar set de DEV; override `DB_NAME`, `DEPLOY_ENV=preview`, `SERVICE_VERSION=pr-<n>-<sha>`, `AUTH_ENABLED=false` | en `preview.yml` |
| Base de datos | DB por PR en **instancia dedicada de preview**; migra on-startup; `DROP` al cerrar | guard `^ponti_preview_pr_[0-9]+$` en el DROP |
| Secretos | nuevo `db-password-preview`; reusar `x-api-key-dev`; AI/Review opcionales | SA preview con `secretAccessor` |
| Migraciones | on-startup; 1er deploy ~30–60 s; persisten entre pushes del PR | opción recrear DB ante divergencia |
| IAM | reusar `WIF_PROVIDER_DEV` + SA de DEV; **+1 binding** `cloudsql.client` sobre la instancia de preview | **verificar** `run.developer`/`serviceAccountUser` (R1) y `CREATEDB` del usuario preview (R2) |
| Auth | `AUTH_ENABLED=false`; QA usa `X-API-KEY` | sin Identity Platform |
| Networking | URL `run.app`, socket `/cloudsql/<INSTANCE>` (`--add-cloudsql-instances`), `--allow-unauthenticated` | sin VPC connector |
| Dominios/certs | no requeridos (URL `run.app` por servicio) | — |

---

## Bloqueos a resolver antes de implementar (Fase 0)
1. **Instancia preview:** crear `ponti-preview-db` (shared-core) + usuario con `CREATEDB` + secret `db-password-preview`.
2. **IAM:** binding `cloudsql.client` sobre la instancia nueva; confirmar que el WIF SA puede crear/borrar servicios y `actAs` el runtime SA.
3. **Retención de imágenes:** definir política/cleanup en Artifact Registry para tags de PR.
   *(pgvector/extensiones ya verificado: solo `unaccent`/`pg_trgm`/`pgcrypto`, estándar.)*

---

## Verificación (cómo se probará end-to-end, en la futura implementación)
1. Abrir PR a `main` → ver job `preview` verde, comentario con URL, `curl https://<url>/api/v1/healthz` → `200`.
2. Push adicional al PR → mismo servicio/URL actualizado (nueva revision), smoke OK.
3. Merge a `main` → `deploy-staging.yml` despliega STAGING (comportamiento intacto); `cleanup-preview.yml` borra servicio + `DROP DATABASE`.
4. Cerrar un PR sin merge → cleanup también borra recursos.
5. `gcloud run services list --filter="labels.managed-by=preview"` → vacío tras cerrar; comprobar que dev/stg/prod y `audit-service-alignment.yml` siguen OK.

---

> **Recordatorio:** este documento es solo el informe/plan. No se ha modificado ningún archivo del
> repositorio ni configuración, y `develop` no se toca. La implementación (crear `preview.yml` y
> `cleanup-preview.yml`, tocar `ci-pr.yml`) queda pendiente de tu aprobación.
