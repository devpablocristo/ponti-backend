# Estado actual de workflows

## TLDR
- `develop` → dev (DB fija del servicio)
- `main` → stg (DB fija del servicio)
- PROD: promoción manual desde STG (`promote-prod.yml`)
- Preview: `workflow_dispatch` (input `pr_number`) o label `preview` → DB `db_pr_<N>` y servicio `ponti-backend-preview-<N>`
- Limpieza automática al cerrar PR + cron semanal

## Comportamiento actual

### Resumen rápido
- No hay schema por rama: la aislación es por **DB**.
- `develop` y `main` usan la DB fija del servicio (dev y stg respectivamente).
- Deploy preview crea `db_pr_<N>` (DB) y `ponti-backend-preview-<N>` (servicio).
- Limpieza automática al cerrar PR + cron semanal.

### Por escenario
**PRs (Pull Requests)**
- Deploy automático: no (solo con label `preview` o manual)
- DB: no toca DB dev
- Cleanup: al cerrar PR se borran DB preview y servicio preview

**Deploy preview (`deploy-preview.yml`)**
- Trigger: `workflow_dispatch` (input `pr_number`) o label `preview` en PR
- Input: `pr_number` (número de PR)
- DB: `db_pr_<N>`
- Servicio: `ponti-backend-preview-<N>`
- Datos: opcional desde `PREVIEW_SEED_URI` (si está configurado)
- Resultado: servicio preview con DB aislada
- Cleanup: al cerrar PR + cron semanal

**Push a `develop`**
- Workflow: `deploy-dev.yml`
- DB: `new_ponti_db_dev` (instancia `new-ponti-db-dev`)

**Push a `main`**
- Workflow: `deploy-staging.yml`
- DB: `new_ponti_db_staging` (instancia `new-ponti-db-dev`, cross-project)

**Promoción a PROD**
- Workflow: `promote-prod.yml` (manual)
- Input: `commit_sha` (SHA probado en STG)
- Usa el mismo artefacto de imagen de STG, sin rebuild

### Limpieza automática
- PR cerrado (merge o close): borra DB `db_pr_<N>` y servicio `ponti-backend-preview-<N>`
- Cron semanal: borra DBs `db_pr_*` y servicios preview restantes
- DBs principales: nunca se limpian

---

## Workflows

### 1) Deploy DEV
Archivo: `.github/workflows/deploy-dev.yml`

**Triggers**
- `push` a `develop`

**Comportamiento**
- Deploy a Cloud Run en proyecto dev.
- Usa `DB_NAME_DEV` (`new_ponti_db_dev`), instancia `new-ponti-db-dev`.

### 2) Deploy STAGING
Archivo: `.github/workflows/deploy-staging.yml`

**Triggers**
- `push` a `main`

**Comportamiento**
- Deploy a Cloud Run en proyecto stg.
- Usa `DB_NAME_STG` (`new_ponti_db_staging`), instancia `new-ponti-db-dev` (cross-project).

### 3) Deploy Preview
Archivo: `.github/workflows/deploy-preview.yml`

**Triggers**
- `workflow_dispatch` (input `pr_number`)
- `pull_request` con label `preview`

**Comportamiento**
- Crea DB `db_pr_<N>` en instancia dev.
- Despliega servicio `ponti-backend-preview-<N>`.
- Opcional: importa seed desde `PREVIEW_SEED_URI` si está configurado.

### 4) Cleanup de previews
Archivo: `.github/workflows/cleanup-preview.yml`

**Triggers**
- `pull_request` cerrado (merge o close)
- `schedule` semanal

**Comportamiento**
- PR cerrado: borra DB `db_pr_<N>` y servicio `ponti-backend-preview-<N>`.
- Cron semanal: borra DBs `db_pr_*` y servicios preview restantes.

### 5) Otros
- `reset-dev.yml`: reset de DEV (borra y restaura desde Golden Snapshot).
- `refresh-golden-snapshot.yml`: exporta snapshot desde STG al bucket.
- `promote-prod.yml`: promoción manual a PROD (usa SHA de STG).
- `ci-pr.yml`: tests en PR a develop.
- `db-verify.yml`: verificación de migraciones (levanta PostgreSQL en CI).

## Seguridad
- Deploy preview solo opera en proyecto dev.
- Nunca toca DB prod ni DB dev principal en previews.

## Inputs manuales
- **Deploy preview**: input `pr_number`.
- **Promote prod**: input `commit_sha`.
