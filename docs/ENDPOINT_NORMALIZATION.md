TLDR:
- Las rutas locales se normalizaron (kebab-case).
- El remoto usa rutas legacy.
- Esta tabla indica el mapeo local -> legacy para el FE.

# Normalizacion de endpoints (local -> remoto legacy)

## Mapeos de paths
| Local (normalizado) | Remoto (legacy) |
|---|---|
| `/api/v1/work-orders` | `/api/v1/workorders` |
| `/api/v1/invoices/{id}` | `/api/v1/invoice/{id}` |
| `/api/v1/projects/customers/{id}` | `/api/v1/projects/customer/{id}` |
| `/api/v1/projects/{id}/supply-movements/providers` | `/api/v1/providers` |
| `/api/v1/labors/group/{project_id}` | `/api/v1/labors/group/{project_id}` |

## Cambios de query params
- `field_id` (local) -> `fieldID` (remoto legacy)

## Endpoints sin equivalente remoto
- `/healthz`
- `/api/v1/business-parameters`
  - `/api/v1/business-parameters/{key}`
  - `/api/v1/business-parameters/category/{category}`
