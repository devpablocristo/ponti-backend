# Estado actual de workflows

## TLDR
- `develop` → dev (DB fija del servicio)
- `main` → prod (DB fija del servicio)
- `workflow_dispatch` → preview con **DB por rama** y **snapshot automatico**
- Limpieza automatica al cerrar PR + cron semanal

## Workflows

### 1) Deploy principal
Archivo: `.github/workflows/deploy-cloud-run.yml`

**Triggers**
- `push` a `develop`
- `push` a `main`
- `workflow_dispatch` (manual)

**Comportamiento**
- **push a develop**: deploy a dev, usa `DB_NAME` del servicio dev.
- **push a main**: deploy a prod, usa `DB_NAME` del servicio prod.
- **manual**:
  - crea `DB_NAME=branch_<slug>`
  - borra y recrea la DB
  - exporta snapshot desde la DB dev real
  - importa el snapshot en `branch_<slug>`
  - despliega servicio preview con esa DB

**Notas**
- `DB_SCHEMA` siempre `public`.
- La aislacion de previews es por `DB_NAME`.

### 2) Cleanup de previews
Archivo: `.github/workflows/cleanup-preview.yml`

**Triggers**
- `pull_request` cerrado (merge o close)
- `schedule` semanal

**Comportamiento**
- PR cerrado: borra `branch_<slug>` y snapshots `preview_seed_branch_<slug>_*`.
- Cron semanal: borra DBs `branch_*` y snapshots restantes.

## Seguridad
- `workflow_dispatch` solo opera en proyecto dev.
- Nunca toca DB prod ni DB dev principal en previews.

## Inputs manuales
El deploy manual **solo pide la rama**.
