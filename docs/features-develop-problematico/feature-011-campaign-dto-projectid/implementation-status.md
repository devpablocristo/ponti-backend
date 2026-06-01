# implementation-status.md — feature-011 · campaign-dto-projectid (BE)

## Estado general
- **Estado:** COMPLETA en el SOURCE (`777e5f6a`), pero **ACOPLADA** a un bundle de otras
  features. El slice puro de 011 es pequeño y limpio. En `develop` (destino): **NO portado**.
- **% completitud del slice 011 en el source:** ~100% (DTO con `project_id` + ruteo por DTO).
- **% completitud en develop:** 0% (el bug sigue presente).

## Estado en ESTE repo (BE)
- `dto.Campaign` con `ProjectID` (json `project_id`): **implementado en source**, ausente en develop.
- `ListCampaigns` responde `[]dto.Campaign` vía `dto.FromDomain`: **implementado en source**,
  ausente en develop (allí hace `c.JSON(http.StatusOK, campaigns)`).
- Confianza: **alta** (verificado con `git show develop:` y `git show 777e5f6a:`).

## Estado en el OTRO repo (FE)
- **Desconocido desde aquí.** Existe paquete `feature-011` en el repo FE (FULL-STACK).
  El FE debe consumir `project_id` para el dropdown de campañas. Validar en ese paquete.

## Tests
- **Propios de 011:** ninguno en la flist. La versión mínima no trae tests.
- **En la flist pero NO son de 011:**
  - `internal/campaign/handler_actions_test.go` (009 — archive handlers).
  - `internal/campaign/repository_tenant_test.go` (001/003 — aislamiento por tenant).
- **Sugerido (mejora-futura):** test de shape JSON de `dto.FromDomain` (ver validation.md).

## Pendientes / clasificación

### BLOQUEANTE para mergear (este repo)
- [ ] Portar `ProjectID` al DTO y el ruteo de `ListCampaigns` por `dto.FromDomain`.
- [ ] `go build ./...` OK sin arrastrar símbolos de tenancy/lifecycle.
- [ ] Confirmar que el diff del PR NO incluye CRUD/archive/tenancy/json-tag removal.

### BLOQUEANTE para mergear (cross-repo)
- [ ] Coordinar con FE feature-011 (shape change). BE-first.

### Mejora futura
- [ ] Test unitario del shape JSON (`{id,name,project_id}` exacto).
- [ ] Enrutar `GetCampaign` por DTO (depende de que exista la ruta GET /:id → 009).

### Deuda aceptable
- Los json-tags del dominio (`usecases/domain/campaign.go`) siguen presentes en develop;
  tras 011 ya no se usan para serializar. Se limpian en 027. No bloquea.

### Duda humana
- ¿Se quiere también el read individual `GET /campaigns/:id` con shape DTO? Eso pertenece
  mejor a 009; decidir si se adelanta acá o se espera.

## Bugs / observaciones
- El bug original: serializar `[]domain.Campaign` filtra campos del `Base` embebido y
  acopla el contrato a los tags del dominio. El fix lo aísla en el DTO. Sin riesgo de
  regresión funcional siempre que se extraiga SOLO el slice mínimo.
