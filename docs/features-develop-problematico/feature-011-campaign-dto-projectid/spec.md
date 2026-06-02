# spec.md — feature-011 · campaign-dto-projectid

## Identidad
- **ID:** feature-011
- **Slug:** campaign-dto-projectid
- **Nombre:** Campaign DTO project_id serialization
- **Tipo:** bugfix
- **Merge:** coordinado (shape change) — FULL-STACK BE+FE
- **Repo (este paquete):** Backend Go — `ponti-backend` (core) · `/home/pablocristo/Proyectos/pablo/ponti/core`
- **Existe en FE:** SÍ (mismo feature-011, paquete del repo FE)
- **Existe en BE:** SÍ (este paquete)

## Fuente de verdad
- **Rango diff:** `0972e565..777e5f6a`
- **SOURCE de extracción:** `develop-problematico~1` (SHA `777e5f6a`).
  NUNCA usar `develop-problematico` (su tip es un restore/vacío).
- **Rama destino:** `develop` (tip `003a9b8f`).

## Resumen
La respuesta del endpoint de campañas debe exponer un shape JSON estable y en minúscula
(`id`, `name`, `project_id`) construido desde un DTO de salida (`dto.Campaign` vía
`dto.FromDomain`), en lugar de serializar el objeto de dominio directamente. El campo
clave es `project_id`, que el FE usa para poblar el dropdown de campañas / vincular
campaña↔proyecto. Si el shape se desincroniza (mayúsculas, campo ausente, o se filtran
campos del `Base` embebido), el dropdown de campañas en el FE queda vacío.

## Objetivo
Garantizar que `GET /campaigns` (y `GET /campaigns/:id`) devuelva exactamente
`{ "id": <int>, "name": <string>, "project_id": <int> }` por campaña, agregando
`ProjectID` al DTO de salida y enrutando la respuesta del handler a través de
`dto.FromDomain` en lugar de devolver el dominio crudo.

## Problema (estado del bug)
En `develop` (tip actual `003a9b8f`):
- `internal/campaign/handler.go` → `ListCampaigns` hace `c.JSON(http.StatusOK, campaigns)`
  serializando **el slice de dominio** (`[]domain.Campaign`) directamente.
- El dominio `internal/campaign/usecases/domain/campaign.go` sí tiene `json:"id"`,
  `json:"name"`, `json:"project_id"` PERO embebe `shareddomain.Base`, por lo que la
  respuesta también filtra campos del Base (created_at, updated_at, created_by, ...).
- El DTO de salida `internal/campaign/handler/dto/campaign.go` **NO tiene** el campo
  `ProjectID` y de hecho **no se usa** en el camino de listado.

Resultado: el contrato de salida no es estable ni intencional. La feature lo arregla
definiendo `dto.Campaign{ ID, Name, ProjectID }` con tags `json` en minúscula y
mapeando con `dto.FromDomain` antes de responder.

## Alcance EN ESTE REPO (BE) — qué es realmente feature-011
Cambio MÍNIMO y autosuficiente (lo único que es el bugfix de serialización):
1. `internal/campaign/handler/dto/campaign.go` — agregar `ProjectID int64 \`json:"project_id"\``
   al struct `Campaign`, a `ToDomain()` y a `FromDomain()`.
2. `internal/campaign/handler.go` → `ListCampaigns` — reemplazar
   `c.JSON(http.StatusOK, campaigns)` por construir `[]dto.Campaign` con `dto.FromDomain`
   y responder ese slice. (Si además se incorpora `GetCampaign`, también vía `dto.FromDomain`.)

Esto NO requiere `authz`, `lifecycle`, `tenancy`, ni columna `tenant_id`. Compila contra
`develop` tal cual.

## Alcance EN EL OTRO REPO (FE)
- Consumir `project_id` (minúscula) en el componente/servicio de campañas (dropdown
  de campañas y/o vínculo campaña↔proyecto). Ver paquete `feature-011` del repo FE.
- El FE depende del shape `{ id, name, project_id }`. Si el BE no envía `project_id`,
  el dropdown queda vacío.

## FUERA DE ALCANCE (presente en la flist pero pertenece a OTRAS features)
La rama `777e5f6a` mezcla en los mismos archivos un bundle mucho mayor. NO es feature-011:
- **Archive/CRUD surface de campaigns** (Create/Update/Archive/Restore/HardDelete,
  `ListArchivedCampaigns`, rutas nuevas, `requests.go`, `handler_actions_test.go`) →
  pertenece a **feature-009 (crudar-archive-surface)**.
- **Tenancy / multitenant** (`tenancy.Scope`, columna `TenantID uuid.UUID`,
  `authz.OptionalTenantOrStrict`, `repository_tenant_test.go`) → **features 001/003**.
- **Limpieza de json-tags del dominio** (quitar `json:"…"` de
  `usecases/domain/campaign.go`) → **feature-027 (be-cleanup-domain-purity)**.
- **Lifecycle helpers** (`lifecycle.RootCause/ArchiveUpdates/RestoreUpdates/RequireArchived`)
  → **feature-009 + shared lifecycle (no existe en develop)**.

## Comportamiento esperado
- `GET /campaigns` → `200` con `[{ "id":1, "name":"X", "project_id":7 }, ...]`.
- `GET /campaigns?customer_id=&project_name=` → mismo shape, filtrado.
- `GET /campaigns/:campaign_id` (si se incluye GetCampaign) → `200` con
  `{ "id":1, "name":"X", "project_id":7 }`.
- Ningún campo del `Base` embebido (created_at, ...) debe filtrarse en la respuesta.

## Estado en dp~1 (777e5f6a)
- DTO con `ProjectID`: **presente**.
- Handler enruta por `dto.FromDomain`: **presente** (pero acoplado al bundle archive+tenancy).
- En `develop`: **ninguno de los dos** — el bug sigue vivo.

## Criterios de aceptación
- [ ] `dto.Campaign` expone `id`, `name`, `project_id` (tags json en minúscula).
- [ ] `ListCampaigns` responde `[]dto.Campaign` construido con `dto.FromDomain`.
- [ ] La respuesta NO incluye campos del `Base` embebido del dominio.
- [ ] `project_id` viene del valor real del workspace (lo resuelve `ListCampaigns` del repo).
- [ ] FE: el dropdown de campañas se puebla con datos reales (validar en el repo FE).
- [ ] `go build ./...` y `go vet ./internal/campaign/...` OK.

## Afectados
- **Endpoints:** `GET /campaigns`, `GET /campaigns/:campaign_id` (read paths).
  (Las rutas POST/PUT/POST archive/restore/DELETE hard son de feature-009, NO de 011.)
- **Modelos/DTOs:** `dto.Campaign` (salida); `domain.Campaign` (lectura, sin tocar tags acá).
- **UI (otro repo):** componente/servicio dropdown de campañas (FE feature-011).
- **DB:** ninguna migración para feature-011. (La columna `tenant_id` es de 001/003.)
- **Tests:** ninguno propio en la versión mínima. Los tests de la flist
  (`handler_actions_test.go`, `repository_tenant_test.go`) son de 009 y 001/003.

## Dependencias
- **Intra-repo:** ninguna dura para la versión mínima. (Los helpers `RespondOK`,
  `dto.FromDomain` ya existen en develop.) Si se trae GetCampaign vía DTO, también OK.
- **Cross-repo:** coordinar con FE feature-011 (shape change). **BE-first** recomendado.

## Riesgos
- **Funcional:** shape change → cualquier cliente que dependiera de campos extra del
  Base embebido dejará de recibirlos (deseado, pero verificar FE).
- **Técnico:** riesgo de arrastrar el bundle completo (archive+tenancy+027) al extraer
  el diff de `handler.go`/`repository.go` con whole-file. Hay que extraer SOLO hunks.

## DECISIÓN recomendada
**PARTIR EN SUBFEATURES + extraer solo el slice mínimo tal cual.**
Portar únicamente: el `ProjectID` en `dto.Campaign` + el ruteo de `ListCampaigns`
(y opcionalmente `GetCampaign`) vía `dto.FromDomain`. Dejar archive/CRUD para 009,
tenancy para 001/003, y json-tags del dominio para 027. NO traer `repository.go`,
`models/campaign.go`, `requests.go`, ni los tests con whole-file.
