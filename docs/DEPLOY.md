# Despliegue de ponti-backend en Google Cloud Run

## TLDR
- Deploy automatico: `push` a `develop`/`main`.
- Deploy manual: `workflow_dispatch` (solo pide `branch`).
- Previews: **DB por rama** `branch_<slug>` con **snapshot automatico** desde la DB dev real.
- Limpieza: al cerrar PR y cron semanal (ver `cleanup-preview.yml`).

## Requisitos previos
- `gcloud` configurado
- Docker instalado
- Acceso a GCP:
  - `new-ponti-dev`
  - `new-ponti-prod`

## Variables de aplicacion (Cloud Run)
Estas variables se configuran en el servicio, no en Actions.

### Minimas
| Variable | Descripcion | Ejemplo |
|----------|-------------|---------|
| `DB_TYPE` | `postgres` | `postgres` |
| `DB_HOST` | Host/IP de Postgres | `136.112.24.122` |
| `DB_PORT` | Puerto | `5432` |
| `DB_USER` | Usuario | `soalen-db` |
| `DB_PASSWORD` | Password | `****` |
| `DB_NAME` | DB base del servicio | `ponti_api_db` |
| `DB_SSL_MODE` | Modo SSL | `disable` |
| `HTTP_SERVER_PORT` | Puerto HTTP | `8080` |

> Nota: `DB_SCHEMA` siempre se usa como `public`. La aislacion es por `DB_NAME`.

## Deploy con GitHub Actions
Workflow principal: `.github/workflows/deploy-cloud-run.yml`.

### Triggers
- `push` a `develop` → dev (DB fija del servicio)
- `push` a `main` → prod (DB fija del servicio)
- `workflow_dispatch` → preview en dev (**DB por rama**)

### Deploy manual por rama (preview)
1. GitHub → Actions → **Deploy to Cloud Run**
2. Run workflow
3. Completar **solo**:
   - `branch`: rama a desplegar

**Que pasa en cada run manual:**
- Se calcula `DB_NAME=branch_<slug>`.
- Se borra y recrea la DB `branch_<slug>`.
- Se **exporta** un snapshot de la DB dev real.
- Se **importa** ese snapshot en la DB `branch_<slug>`.
- Se despliega un servicio preview con esa DB.

## Limpieza automatica
Workflow: `.github/workflows/cleanup-preview.yml`
- Al **cerrar PR** (merge o close): borra DB `branch_*` y snapshots asociados.
- **Cron semanal**: limpia DBs `branch_*` y snapshots restantes.

## Troubleshooting rapido
### Error: "container failed to start" + errores de DB
**Causa**: variables de DB faltantes o DB preview sin datos.  
**Solucion**: correr `workflow_dispatch` (se resetea y se importan datos siempre).

## Documentacion relacionada
- `SETUP_PROD.md`
- `GITHUB_SECRETS.md`
