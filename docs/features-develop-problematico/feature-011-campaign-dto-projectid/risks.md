# risks.md — feature-011 · campaign-dto-projectid (BE)

## Funcionales
- **R-F1 · Dropdown de campañas vacío (cross-repo).** Si BE no envía `project_id` o el FE
  espera otra capitalización, el dropdown queda vacío.
  *Mitigación:* tags json en minúscula (`project_id`); BE-first; verificar respuesta real
  con `curl GET /campaigns`.
- **R-F2 · Shape change rompe consumidores ocultos.** Pasar de domain crudo a DTO quita
  campos del `Base` embebido (created_at, updated_at, ...).
  *Mitigación:* confirmar que ningún cliente (FE u otro servicio) leía esos campos del
  listado de campañas. El FE solo necesita `{id,name,project_id}`.

## Técnicos
- **R-T1 · Arrastrar el bundle al extraer.** `handler.go` y `repository.go` mezclan
  011 + 009 + 001/003. Un whole-file incluiría `tenancy.Scope`, `authz.*`, `lifecycle.*`
  (inexistentes en develop) y métodos Archive/Restore/HardDelete → **no compila**.
  *Mitigación:* extracción por hunks (`git restore -p`) o edición manual; nunca whole-file
  de `handler.go`/`repository.go`. Verificar con
  `git grep -nE "tenancy\.Scope|authz\.|lifecycle\.|runCampaignIDAction"` (debe dar vacío).
- **R-T2 · Import muerto.** Si se agrega el ruteo por DTO pero se olvida el import
  `dto "…/internal/campaign/handler/dto"`, o si se trae el import `ginmw`/`types` sin
  usarlos. *Mitigación:* `go vet` + `go build`; `git diff --check`.

## Integración / cross-repo
- **R-I1 · Orden de merge.** Si FE feature-011 mergea primero, el dropdown queda vacío
  hasta que despliegue el BE. *Mitigación:* **BE-first**, luego FE.
- **R-I2 · Mergear solo este repo (BE).** Bajo riesgo: agrega un campo a la respuesta;
  los consumidores actuales lo ignoran. El FE seguirá con dropdown vacío hasta su PR,
  igual que hoy. **Aceptable** desplegar BE solo.
- **R-I3 · Mergear solo el otro repo (FE).** Si FE lee `project_id` y el BE aún no lo
  envía, el dropdown queda vacío (estado actual, sin regresión nueva). **No mergear FE
  antes que BE.**

## Datos / migración
- **R-D1 · Ninguna migración en 011.** La columna `tenant_id` y el soft-delete son de
  001/003 y 009. No tocar esquema acá. Si por error se trae `models/campaign.go`
  (`TenantID`), GORM AutoMigrate podría intentar alterar la tabla → fuera de alcance.
  *Mitigación:* no extraer `models/campaign.go`.

## Archivos compartidos
- **`internal/campaign/handler.go`** (011 + 009): extraer solo `ListCampaigns`.
- **`internal/campaign/repository.go`** (001/003 + 009): no extraer en 011.
- **`internal/campaign/usecases/domain/campaign.go`** (027): no extraer en 011.

## Extracción parcial
- **R-E1 · Quedarse corto:** olvidar el ruteo por DTO y solo agregar `ProjectID` al DTO →
  como `ListCampaigns` sigue serializando domain, no cambia nada visible.
  *Detección:* `git grep "dto.FromDomain" internal/campaign/handler.go` debe aparecer.
- **R-E2 · Pasarse:** colar CRUD/archive/tenancy. *Detección:* el grep de R-T1 vacío;
  el diff del PR debe tocar solo 2 archivos y ~10 líneas netas.

## Resumen de riesgo de mergear aislado
- **Solo BE:** seguro (aditivo). Recomendado primero.
- **Solo FE:** evitar; deja dropdown vacío hasta que llegue el BE.
