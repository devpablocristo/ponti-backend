# GitHub Secrets y Workload Identity

## Secrets requeridos (repo)

| Secret | Descripción |
|--------|-------------|
| `DB_PASSWORD_DEV` | Password DB dev (usuario soalen-db-v3, new_ponti_db_dev) |
| `X_API_KEY_DEV` | API key dev |
| `DB_PASSWORD_STG` | Password usuario `app_stg` (DB new_ponti_db_staging en instancia dev) |
| `X_API_KEY_STG` | API key stg |
| `DB_PASSWORD_PROD` | Password DB prod |
| `X_API_KEY_PROD` | API key prod |

## Workload Identity Federation

Se usa OIDC desde GitHub Actions. No se usan keys JSON.

Valores actuales:

### DEV
- `WIF_PROVIDER_DEV`: `projects/1087442197188/locations/global/workloadIdentityPools/github-actions-pool/providers/github-actions-provider`
- `WIF_SERVICE_ACCOUNT_DEV`: `github-actions@new-ponti-dev.iam.gserviceaccount.com`

### STG
- `WIF_PROVIDER_STG`: `projects/65243764597/locations/global/workloadIdentityPools/github-actions-pool/providers/github-actions-provider`
- `WIF_SERVICE_ACCOUNT_STG`: `github-actions@new-ponti-stg.iam.gserviceaccount.com`

### PROD
- `WIF_PROVIDER_PROD`: `projects/875939220111/locations/global/workloadIdentityPools/github-actions-pool/providers/github-actions-provider`
- `WIF_SERVICE_ACCOUNT_PROD`: `github-actions@new-ponti-prod.iam.gserviceaccount.com`

> Nota: estos valores se cargan como **variables** (no secrets).

---

## Permiso IAM pendiente: refresh-golden-snapshot

El workflow `refresh-golden-snapshot.yml` exporta desde la instancia `new-ponti-db-dev` (proyecto **new-ponti-dev**) usando el SA `github-actions@new-ponti-stg`. Ese SA debe tener permiso de export en el proyecto donde vive la instancia:

```bash
# Ejecutar en new-ponti-dev (o con --project=new-ponti-dev)
gcloud projects add-iam-policy-binding new-ponti-dev \
  --member="serviceAccount:github-actions@new-ponti-stg.iam.gserviceaccount.com" \
  --role="roles/cloudsql.admin"
```

Alternativa más restrictiva: `roles/cloudsql.instances.export` si existe o el rol mínimo que permita `gcloud sql export sql`.

**Verificación:** Si el workflow falla con "Permission denied" al exportar, aplicar el binding anterior.

---

## Permiso IAM: promote-prod (smoke tests)

El workflow `promote-prod.yml` ejecuta smoke tests contra el servicio Cloud Run en PROD (que tiene `--no-allow-unauthenticated`). El SA `github-actions@new-ponti-prod` debe poder invocar el servicio:

```bash
gcloud run services add-iam-policy-binding ponti-backend \
  --project=new-ponti-prod \
  --region=us-central1 \
  --member="serviceAccount:github-actions@new-ponti-prod.iam.gserviceaccount.com" \
  --role="roles/run.invoker"
```

**Verificación:** Si los smoke tests fallan con 403 al llamar a la API, aplicar el binding anterior.
