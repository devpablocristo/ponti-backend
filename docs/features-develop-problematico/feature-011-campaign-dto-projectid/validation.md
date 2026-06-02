# validation.md — feature-011 · campaign-dto-projectid (BE)

## Checklist PRE-PR
- [ ] El diff toca SOLO: `internal/campaign/handler/dto/campaign.go` (entero) y el cuerpo
      de `ListCampaigns` en `internal/campaign/handler.go`.
- [ ] `dto.Campaign` tiene `ProjectID int64 \`json:"project_id"\`` y aparece en
      `ToDomain()` y `FromDomain()`.
- [ ] `ListCampaigns` construye `[]dto.Campaign` con `dto.FromDomain` y responde ese slice.
- [ ] NO aparecen en el diff: `tenancy.Scope`, `authz.`, `lifecycle.`, `runCampaignIDAction`,
      `ArchiveCampaign`/`RestoreCampaign`/`HardDeleteCampaign`/`UpdateCampaign`/
      `ListArchivedCampaigns`, imports `ginmw`/`types`, columna `TenantID`.
- [ ] No se agregaron archivos `requests.go`, `handler_actions_test.go`,
      `repository_tenant_test.go` (son de 009 / 001 / 003).

### Comandos de verificación (read-only)
```bash
cd /home/pablocristo/Proyectos/pablo/ponti/core
git grep -n "ProjectID" internal/campaign/handler/dto/campaign.go        # debe aparecer
git grep -n "dto.FromDomain" internal/campaign/handler.go                # debe aparecer
git grep -nE "tenancy\.Scope|authz\.|lifecycle\.|runCampaignIDAction|ArchiveCampaign" \
   internal/campaign/handler.go internal/campaign/repository.go          # debe estar VACÍO
git diff --check
```

## Build / tests (BE)
```bash
go build ./...
go vet ./internal/campaign/...
go test ./internal/campaign/...        # (sin tests propios de 011; debe seguir verde)
```

## Test sugerido (mejora-futura, opcional)
Agregar en el paquete `dto` (o un `_test.go` en `internal/campaign/handler/dto`):
- `dto.FromDomain(domain.Campaign{ID:1, Name:"x", ProjectID:7})` → marshaliza a JSON
  exactamente `{"id":1,"name":"x","project_id":7}` (sin claves del Base).

## Validación manual (API)
```bash
# Con el server local levantado:
curl -s "$API/campaigns" | jq '.[0]'
# Esperado: { "id": <int>, "name": <string>, "project_id": <int> }  (sin created_at/updated_at/etc.)

curl -s "$API/campaigns?customer_id=<id>&project_name=<name>" | jq '.'
# Esperado: lista filtrada con el mismo shape; project_id = id real del proyecto del workspace.
```

## Casos borde
- Lista vacía → `[]` (no `null`). En develop `ListCampaigns` retorna slice vacío bien;
  el handler usa `make([]dto.Campaign, 0, len(...))` → emite `[]`.
- Campaña sin proyecto asociado → `project_id` será `0` (cero-value). Confirmar con FE
  que `0` se trata como "sin proyecto" y no rompe el dropdown.
- Filtro sin coincidencias → lista vacía (en develop el repo ya devuelve `[]`).

## Qué revisar en UI / API / DB / env
- **UI (FE):** el dropdown de campañas se puebla; al seleccionar, el `project_id` viaja correcto.
- **API:** shape `{id,name,project_id}`, sin filtrado de campos del Base.
- **DB:** sin cambios de esquema en 011.
- **Env:** ninguno.

## Qué validar en el OTRO repo (FE feature-011)
- El servicio/hook de campañas mapea `project_id` (snake_case) y no espera `projectId`.
- El dropdown muestra campañas reales tras desplegar el BE.
- Build/test del FE: `yarn build`, `yarn test`; e2e del flujo de campañas/proyectos si existe.

## Señales de incompletitud / incompatibilidad
- Dropdown FE vacío tras desplegar BE → revisar que `project_id` realmente sale en la
  respuesta (`curl`) y que el FE lee snake_case.
- Build BE roto con "undefined: tenancy/authz/lifecycle" → se coló bundle de otra feature;
  revertir y re-extraer solo los hunks de 011.
- Respuesta con `created_at`/`updated_at` → el handler todavía serializa domain crudo
  (no se aplicó el ruteo por DTO).
