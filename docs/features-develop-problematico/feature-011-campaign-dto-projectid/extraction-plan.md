# extraction-plan.md — feature-011 · campaign-dto-projectid (BE)

## Contexto
- **Repo:** `ponti-backend` (core) — `/home/pablocristo/Proyectos/pablo/ponti/core`
- **Rama base:** `develop` (tip `003a9b8f`)
- **SOURCE:** `develop-problematico~1` (SHA `777e5f6a`) — NUNCA `develop-problematico` (tip vacío/restore)
- **Rama sugerida:** `pr/feature-011-campaign-dto-projectid-be`

## PR title
`fix(be): serializar project_id/id/name de campañas vía DTO (dropdown FE)`

## PR description (borrador)
> Bugfix de contrato. `GET /campaigns` serializaba el objeto de dominio crudo
> (`[]domain.Campaign`), filtrando campos del `Base` embebido y dependiendo de tags
> json del dominio. Se enruta la respuesta a través de `dto.Campaign`/`dto.FromDomain`
> y se agrega el campo `project_id` al DTO de salida, garantizando el shape estable
> `{ id, name, project_id }` que consume el dropdown de campañas del FE.
>
> Cambio de shape coordinado con FE feature-011 (mergear BE primero).
> NO incluye: archive/CRUD de campaigns (009), tenancy (001/003), ni limpieza de
> json-tags del dominio (027) — aunque aparecen mezclados en la rama de origen.

## Estrategia: extracción SELECTIVA por hunks
El diff de origen mezcla 4 features. NO se hace cherry-pick del commit ni whole-file de
`handler.go`/`repository.go`. Se trae:
1. **Archivo entero:** `internal/campaign/handler/dto/campaign.go` (su versión final
   es exactamente el DTO mínimo correcto).
2. **Hunk parcial:** en `internal/campaign/handler.go`, solo el cuerpo de `ListCampaigns`
   (de `c.JSON(http.StatusOK, campaigns)` → construir `[]dto.Campaign` con `dto.FromDomain`).

## Pasos ordenados
1. Verificar que el bug existe en develop (read-only):
   `git -C <core> show develop:internal/campaign/handler/dto/campaign.go` (no tiene ProjectID)
   `git -C <core> show develop:internal/campaign/handler.go` (ListCampaigns hace c.JSON(...,campaigns))
2. Crear rama desde develop.
3. Traer el DTO entero desde el SOURCE.
4. Editar a mano el cuerpo de `ListCampaigns` en `handler.go` (NO usar checkout whole-file:
   arrastraría rutas/CRUD/imports `ginmw`+`types`).
5. `go build ./internal/campaign/...` y `go vet`.
6. (Opcional, decisión humana) Si se quiere `GetCampaign` con shape DTO, requiere también
   agregar la ruta `GET /campaigns/:campaign_id` — eso roza 009; preferible dejarlo para 009.
7. Abrir PR. Coordinar merge con FE feature-011.

## Comandos git SUGERIDOS (para un humano — NO ejecutar desde el agente)
```bash
cd /home/pablocristo/Proyectos/pablo/ponti/core
git checkout develop
git pull
git checkout -b pr/feature-011-campaign-dto-projectid-be

# (1) DTO entero (su versión final es el DTO mínimo correcto):
git checkout develop-problematico~1 -- internal/campaign/handler/dto/campaign.go

# (2) handler.go: NO traer whole-file. Editar a mano SOLO ListCampaigns.
#     Edición manual equivalente (reemplazar el cuerpo de ListCampaigns):
#       campaigns -> out := make([]dto.Campaign, 0, len(campaigns))
#       for _, d := range campaigns { out = append(out, *dto.FromDomain(d)) }
#       c.JSON(http.StatusOK, out)
#     y agregar import: dto "github.com/devpablocristo/ponti-backend/internal/campaign/handler/dto"
#  Alternativa interactiva (revisar hunk por hunk y RECHAZAR todo lo de archive/tenancy):
git restore -p --source=develop-problematico~1 -- internal/campaign/handler.go
#   -> aceptar SOLO el hunk de ListCampaigns + el import de dto; rechazar rutas, CRUD,
#      runCampaignIDAction, GetCampaign/Update/Archive/Restore/HardDelete, imports ginmw/types.

# Verificación:
git diff --check
go build ./internal/campaign/...
go vet ./internal/campaign/...
```

## Archivos enteros vs parciales
- **Entero:** `internal/campaign/handler/dto/campaign.go`.
- **Parcial (manual / restore -p):** `internal/campaign/handler.go` (solo ListCampaigns + import dto).

## Migraciones / tests a incluir
- **Migraciones:** NINGUNA para 011. (La columna `tenant_id` es de 001/003.)
- **Tests:** NINGUNO de la flist es de 011. `handler_actions_test.go` (009) y
  `repository_tenant_test.go` (001/003) NO se traen. Sugerido: agregar un test propio
  pequeño que verifique el shape JSON de `FromDomain` (ver validation.md).

## Dependencias previas
- Ninguna dura. `dto.FromDomain` ya existe; `sharedhandlers.RespondError` y
  `ParseOptionalInt64Query` ya existen en develop. No requiere 001/003/009/027.

## Coordinación con el otro repo (FE)
- **Orden: BE-first.** Mergear este PR de BE antes que FE feature-011, así el FE ya recibe
  `project_id` al desplegarse. El cambio es retrocompatible-aditivo en sentido FE→BE
  (BE solo agrega un campo), pero FE depende del campo para poblar el dropdown.
- Mientras BE no envíe `project_id`, el dropdown del FE quedará vacío. Mientras FE no lea
  `project_id`, no rompe nada (campo extra ignorado).

## Qué NO traer
- `repository.go`, `usecases.go`, `models/campaign.go` (tenancy/archive → 001/003/009).
- `usecases/domain/campaign.go` (quitar json-tags → 027).
- `handler/dto/requests.go`, `handler_actions_test.go` (archive surface → 009).
- `repository_tenant_test.go` (tenancy → 001/003).

## Qué podría romperse
- Si se hace whole-file de `handler.go`: imports `ginmw`/`types` y llamadas a métodos
  `ArchiveCampaign/...` inexistentes en `develop` → no compila.
- Si se trae `repository.go`: `tenancy.Scope`, `authz.*`, `lifecycle.*` no existen en
  develop → no compila.

## Cómo detectar extracción incompleta
- `git grep -n "ProjectID" internal/campaign/handler/dto/campaign.go` → debe aparecer.
- `git grep -n "dto.FromDomain" internal/campaign/handler.go` → debe aparecer en ListCampaigns.
- `git grep -nE "tenancy\.Scope|authz\.|lifecycle\.|runCampaignIDAction|ArchiveCampaign" internal/campaign/handler.go internal/campaign/repository.go`
  → debe estar VACÍO (si aparece, se coló bundle de otra feature).

## Qué validar antes del PR
- `go build ./...` (todo el repo) OK.
- Respuesta de `GET /campaigns` = `[{ "id":…, "name":…, "project_id":… }]`, sin campos del Base.
- Diff del PR toca SOLO `dto/campaign.go` y el cuerpo de `ListCampaigns`.

## Qué hacer después de mergear
- Avisar al equipo FE que `project_id` ya viaja en la respuesta; mergear FE feature-011.
- Verificar en staging que el dropdown de campañas se puebla.
- Cuando se porte 027, recordar quitar los json-tags del dominio (ya no usados por el handler).
