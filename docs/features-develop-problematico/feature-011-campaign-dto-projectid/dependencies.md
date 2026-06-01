# dependencies.md — feature-011 · campaign-dto-projectid (BE)

## Depende de
- **Ninguna feature (versión mínima).** El slice real (DTO `project_id` + ruteo de
  `ListCampaigns` por DTO) compila contra `develop` sin features previas.
  - `dto.FromDomain` ya existe en `develop`.
  - `sharedhandlers.RespondError`, `ParseOptionalInt64Query` ya existen en `develop`.
  - El repo `ListCampaigns` de `develop` ya resuelve y devuelve `project_id` real (vía
    `mapProject`), así que el DTO solo lo expone.

## Bloquea a
- **FE feature-011 (cross-repo, débil-fuerte):** el dropdown de campañas del FE necesita
  `project_id` en la respuesta. Hasta que este BE mergee, el FE no recibe el campo.

## Clasificación de dependencias

### Fuertes (cross-repo)
- **FE feature-011** ⇄ **BE feature-011**: contrato de salida `{ id, name, project_id }`.
  Es un *shape change* coordinado. Romper el shape = dropdown vacío en FE.
  Orden recomendado: **BE-first**.

### Débiles (intra-repo)
- Ninguna para la versión mínima.

### Inciertas
- **GetCampaign vía DTO (opcional):** si se decide enrutar también `GET /campaigns/:id`
  por `dto.FromDomain`, eso obliga a registrar la ruta, lo cual solapa con
  **feature-009 (archive/CRUD surface)**. Confianza media: preferible NO hacerlo en 011.

## NO confundir (estas SÍ aparecen en la flist pero son de otras features)
La rama de origen `777e5f6a` acopla en los mismos archivos:
- **feature-001 / 003 (tenancy / multitenant hardening):** `tenancy.Scope(...)`,
  `authz.OptionalTenantOrStrict`, columna `TenantID uuid.UUID` en `models/campaign.go`,
  `repository_tenant_test.go`. Paquetes `internal/shared/authz` y
  `internal/shared/lifecycle` **NO existen en develop**.
- **feature-009 (crudar-archive-surface):** Create/Update/Archive/Restore/HardDelete,
  `ListArchivedCampaigns`, rutas nuevas, `handler/dto/requests.go`,
  `handler_actions_test.go`, helpers `lifecycle.*`.
- **feature-027 (be-cleanup-domain-purity):** remover tags `json:"…"` de
  `usecases/domain/campaign.go`.

→ Para feature-011 NINGUNO de estos es dependencia. Si se extrae con whole-file se
arrastran y rompen el build (símbolos inexistentes en develop).

## Archivos / tipos / config / migraciones / APIs compartidos
- **Compartido (partial-hunks):** `internal/campaign/handler.go` — su diff sirve a 011
  (ListCampaigns) Y a 009 (CRUD/archive). Extraer SOLO el hunk de ListCampaigns.
- **Compartido (partial-hunks):** `internal/campaign/repository.go` — su diff sirve a
  001/003 (tenancy) Y 009 (archive). Para 011 NO se necesita → do-not-extract-yet.
- **Migraciones:** ninguna para 011.
- **API contract compartido cross-repo:** `GET /campaigns` response shape — owner de la
  forma `project_id` es esta feature; consumidor es FE feature-011.

## Recomendación de orden
1. **BE feature-011** (este PR) — agrega `project_id` al contrato.
2. **FE feature-011** — consume `project_id`.
3. (Independiente) 001/003, 009, 027 portan el resto del bundle en sus propios PRs;
   cuando 027 mergee, quitar los json-tags del dominio (ya inertes tras 011).
