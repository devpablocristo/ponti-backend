# GitHub Secrets y Workload Identity

## Secrets requeridos (repo)

| Secret | Descripción |
|--------|-------------|
| `DB_PASSWORD_DEV` | Password DB dev |
| `X_API_KEY_DEV` | API key dev |
| `DB_PASSWORD_STG` | Password DB stg |
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
