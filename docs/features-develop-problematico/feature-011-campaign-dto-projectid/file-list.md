# file-list.md — feature-011 · campaign-dto-projectid (BE)

Flist autoritativa: `/tmp/flists/be-011.txt` (9 entradas).
Rango: `0972e565..777e5f6a`. SOURCE = `develop-problematico~1` (`777e5f6a`). Destino = `develop`.

> ADVERTENCIA TRANSVERSAL: la rama `777e5f6a` mezcla en estos mismos archivos un bundle
> (archive/CRUD = 009, tenancy = 001/003, limpieza json-tags = 027). Para feature-011
> SOLO importan dos hunks: `+ProjectID` en el DTO y el ruteo de `ListCampaigns` por DTO.

## Propios (núcleo real de feature-011)

| path | status | tipo | rol en la feature | extracción | motivo | riesgo | confianza |
|---|---|---|---|---|---|---|---|
| `internal/campaign/handler/dto/campaign.go` | M | DTO salida | Agrega `ProjectID int64 \`json:"project_id"\`` a struct + `ToDomain`/`FromDomain`. ES el corazón del bugfix. | **whole-file** | El archivo final en 777e5f6a es exactamente el DTO mínimo correcto (ID/Name/ProjectID, sin extras del bundle). Se puede traer entero. | bajo | alta |
| `internal/campaign/handler.go` | M | handler HTTP | Hunk de `ListCampaigns`: pasar de `c.JSON(...,campaigns)` a `[]dto.Campaign` con `dto.FromDomain`. | **partial-hunks** | El whole-file trae rutas CRUD/archive (009), `ginmw`, `types`, `runCampaignIDAction`, GetCampaign/Update/etc. SOLO traer el hunk de ListCampaigns (+ opcional GetCampaign vía DTO). | medio | alta |

## Compartidos (partial-hunks) — diff sirve a varias intenciones

| path | status | tipo | rol | extracción | motivo | riesgo | confianza |
|---|---|---|---|---|---|---|---|
| `internal/campaign/repository.go` | M | repositorio GORM | En 777e5f6a su diff es 95% tenancy (001/003) + archive lifecycle (009). Para 011 NO se necesita: `ListCampaigns` en develop ya devuelve `project_id` correcto vía `mapProject`. | **do-not-extract-yet** | Trae `tenancy.Scope`, `authz`, `lifecycle` (no existen en develop), columna soft-delete, CRUD. Romperia el build. | alto | alta |
| `internal/campaign/usecases.go` | M | usecases | El diff agrega métodos Archive/Restore/HardDelete/Update/ListArchived (009). Para 011 no se toca. | **do-not-extract-yet** | Métodos de 009; sin ellos el DTO/handler de 011 compila igual. | bajo | alta |
| `internal/campaign/repository/models/campaign.go` | M | modelo GORM | Agrega `TenantID uuid.UUID` y quita `unique` del name (001/003). No es 011. | **do-not-extract-yet** | Columna multitenant; pertenece a tenancy. | medio | alta |
| `internal/campaign/usecases/domain/campaign.go` | M | dominio | Quita los tags `json:"…"` del dominio (limpieza de pureza). | **do-not-extract-yet** | Es **feature-027** (be-cleanup-domain-purity). Con 011 el dominio ya no se serializa, así que quitar tags es cosmético y va en 027. | bajo | alta |

## Requeridos por dependencia
Ninguno DURO para la versión mínima. `dto.FromDomain`, `sharedhandlers.RespondOK`/`RespondError`,
`ParseOptionalInt64Query` ya existen en `develop`. (Si se trae GetCampaign vía DTO, también
existe `RespondOK`.) No hace falta `authz`/`lifecycle`/`tenancy` para 011.

## Dudosos

| path | status | tipo | rol | extracción | motivo | riesgo | confianza |
|---|---|---|---|---|---|---|---|
| `internal/campaign/handler.go` (hunk de `GetCampaign`) | M | handler | `GetCampaign` enruta por `dto.FromDomain`. Mejora el contrato del get individual. | **manual-port** (opcional) | Útil para consistencia de shape, pero `GetCampaign` no está ruteado en develop (no hay `GET /:id`). Traerlo obliga a también agregar la ruta → entra en zona de 009. Decisión humana. | medio | media |

## NO traer todavía (pertenecen a otras features — NO extraer aquí)

| path | status | feature dueña | motivo |
|---|---|---|---|
| `internal/campaign/handler/dto/requests.go` | A | 009 (archive-surface) | DTOs `CreateCampaignRequest`/`UpdateCampaignRequest` para POST/PUT. No es 011. |
| `internal/campaign/handler_actions_test.go` | A | 009 | Testea Archive/Restore/HardDelete handlers. |
| `internal/campaign/repository_tenant_test.go` | A | 001/003 (tenancy) | Testea aislamiento por tenant con sqlite + `contextkeys.OrgID`. |

## Resumen de decisión
- **Traer:** `dto/campaign.go` (whole) + hunk `ListCampaigns` de `handler.go`.
- **Opcional (humano):** hunk `GetCampaign` vía DTO (solo si se agrega también la ruta).
- **NO traer:** `repository.go`, `usecases.go`, `models/campaign.go`, `domain/campaign.go`,
  `requests.go`, ambos `_test.go`. Quedan para 009 / 001 / 003 / 027.
