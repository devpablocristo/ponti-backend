# notes-for-future-agent.md — feature-011 · campaign-dto-projectid (BE)

## Resumen corto
Bugfix de contrato FULL-STACK. El BE debe devolver `GET /campaigns` con el shape estable
`{ id, name, project_id }` construido desde `dto.Campaign`/`dto.FromDomain`, en vez de
serializar el dominio crudo. El campo crítico es `project_id` (snake_case): el FE lo usa
para poblar el dropdown de campañas. Si falta o cambia de forma → dropdown vacío.

## Qué está en FE y qué en BE
- **BE (este paquete):** agregar `ProjectID` al DTO de salida + enrutar `ListCampaigns`
  por `dto.FromDomain`.
- **FE (feature-011 del otro repo):** consumir `project_id` en el servicio/dropdown.
- **Orden:** BE-first, luego FE.

## TRAMPA PRINCIPAL (leer esto antes de extraer)
La flist de feature-011 tiene 9 archivos, pero la rama de origen `777e5f6a` MEZCLA 4
features en los mismos archivos. El slice REAL de 011 son solo DOS hunks:
1. `internal/campaign/handler/dto/campaign.go` → `+ProjectID int64 \`json:"project_id"\``
   (este archivo SÍ se puede traer entero: su versión final es el DTO mínimo correcto).
2. `internal/campaign/handler.go` → cuerpo de `ListCampaigns`: pasar de
   `c.JSON(http.StatusOK, campaigns)` a `[]dto.Campaign` con `dto.FromDomain`.

Lo demás de la flist NO es 011:
- `repository.go`, `usecases.go`, `models/campaign.go` → tenancy (001/003) + archive (009).
- `usecases/domain/campaign.go` → quitar json-tags = **feature-027**.
- `handler/dto/requests.go`, `handler_actions_test.go` → archive surface = **009**.
- `repository_tenant_test.go` → tenancy = **001/003**.

## Archivos esenciales / peligrosos / mezclados
- **Esencial:** `internal/campaign/handler/dto/campaign.go` (entero, seguro).
- **Mezclado (peligroso, partial-hunks):** `internal/campaign/handler.go`.
  Si lo traés entero, arrastrás imports `ginmw`/`types`, rutas CRUD, `runCampaignIDAction`,
  y llamadas a métodos `ArchiveCampaign/...` que NO existen en develop → no compila.
- **Peligroso (NO traer):** `internal/campaign/repository.go` — usa `tenancy.Scope`,
  `authz.OptionalTenantOrStrict`, `lifecycle.*`; esos paquetes (`internal/shared/authz`,
  `internal/shared/lifecycle`) NO existen en develop.

## Decisiones ya tomadas
- Extraer SELECTIVO por hunks. NO cherry-pick del commit, NO whole-file de handler/repository.
- DTO entero OK. ListCampaigns por edición manual / `git restore -p`.
- Dejar archive (009), tenancy (001/003) y json-tag cleanup (027) a sus features.
- `GetCampaign` por DTO: NO en 011 (requiere ruta GET /:id, que es de 009).

## Dudas abiertas
- ¿`project_id == 0` (campaña sin proyecto) rompe el dropdown FE? Confirmar con FE.
- ¿Algún consumidor leía campos del Base (created_at...) del listado? El cambio los quita.

## Comandos a mirar primero (read-only)
```bash
cd /home/pablocristo/Proyectos/pablo/ponti/core
git show develop:internal/campaign/handler/dto/campaign.go     # confirma que NO tiene ProjectID
git show develop:internal/campaign/handler.go | sed -n '40,80p' # ListCampaigns hace c.JSON(...,campaigns)
git diff 0972e565..777e5f6a -- internal/campaign/handler/dto/campaign.go internal/campaign/handler.go
git ls-tree -r --name-only develop -- internal/shared/authz internal/shared/lifecycle  # vacío => no existen
```

## Errores a evitar
- Hacer whole-file de `handler.go` o `repository.go` → build roto.
- Traer `models/campaign.go` (columna `TenantID`) → AutoMigrate toca esquema, fuera de alcance.
- Mergear FE antes que BE → dropdown vacío.
- Confundir el archive surface (009) con este bugfix.

## Camino más seguro
1. Rama desde develop.
2. `git checkout develop-problematico~1 -- internal/campaign/handler/dto/campaign.go`.
3. Editar a mano `ListCampaigns` (import dto + loop FromDomain).
4. `go build ./... && go vet ./internal/campaign/...`.
5. Verificar greps de validation.md (sin tenancy/lifecycle/CRUD).
6. PR BE-first; coordinar FE feature-011.

## PR del otro repo: orden
- **Antes:** este PR de BE (feature-011 BE).
- **Después:** FE feature-011 (consume `project_id`).
