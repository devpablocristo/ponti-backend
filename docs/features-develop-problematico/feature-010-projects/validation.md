# validation.md — feature-010 projects (BE)

## Checklist pre-PR (BE)
- [ ] Paso 0 de dependencias OK: `internal/actor`, `internal/shared/lifecycle`, `internal/shared/authz`, `internal/shared/text` presentes en develop y `platform/persistence/gorm/go` en `go.mod`.
- [ ] `go build ./internal/project/...` sin errores.
- [ ] `go build ./...` (repo entero) sin errores — detecta que el cambio de port no rompió wire/DI ni consumidores (`internal/data-integrity` por `GetRawAdminCostTotal`).
- [ ] `go vet ./internal/project/...`.
- [ ] `go test ./internal/project/...` verde.
- [ ] `git diff --check` (sin marcadores de conflicto ni whitespace roto).
- [ ] Ruta registrada: `DELETE /api/v1/projects/:project_id/hard`; ruta `DELETE /api/v1/projects/:project_id` (legacy) ausente.
- [ ] Port `UseCasesPort`/`RepositoryPort` sin `DeleteProject` ni `GetProjectByNameAndCampaignID` (renombrados).

## Tests sugeridos (BE)
```sh
go test ./internal/project/...
go test ./internal/project/... -run TestProjectActionRoutesCallExplicitUseCases
go test ./internal/project/... -run TestUpdateProjectPropagatesRenameToLegacyTables
go test ./internal/project/... -run 'Tenant|Archived'
# y, por el consumidor downstream:
go test ./internal/data-integrity/...
```

## Validación manual (API)
- [ ] `POST /api/v1/projects` con `customer.actor_id`/`managers[].actor_id`/`investors[].actor_id` → 201; verificar `tenant_id` escrito en `projects`, `customers`, `managers`, `investors`, tablas puente; y filas en `actors`/`legacy_actor_map`.
- [ ] `GET /api/v1/projects/:id` → la respuesta trae `actor_id` hidratado para managers/investors legacy.
- [ ] `PUT /api/v1/projects/:id` cambiando nombre de customer existente → renombra el customer (canonicalizado) y propaga al actor; managers/investors no se renombran.
- [ ] `POST /api/v1/projects/:id/archive` → 204; cascada archiva fields/lots/workorders/items/drafts/labors/supplies/movements/stocks/commercializations/dollar_values/puentes.
- [ ] `POST /api/v1/projects/:id/restore` → 204; restaura el grafo por `cause`.
- [ ] `DELETE /api/v1/projects/:id/hard` sobre proyecto NO archivado → 409 ("must be archived before hard delete").
- [ ] `DELETE /api/v1/projects/:id/hard` con dependiente activo → 409 con conteo y label.
- [ ] `DELETE /api/v1/projects/:id/hard` sobre proyecto archivado sin dependientes → 204; verificar que no quedan filas en el grafo.

## Casos borde
- Proyecto sin tenant en contexto y `TenantStrictModeEnabled()` → `domainerr.TenantMissing()`.
- Tenant A no ve/borra proyectos del tenant B (cubierto por `repository_tenant_test.go`).
- Referencia a customer/campaign/manager/investor/field/lot/crop/actor **archivado** en create/update → bloqueo (`repository_archived_refs_test.go`).
- Update con `customer.id=0` (free-solo) → link-by-name o create; con `id!=0` → strict + rename.
- Colisión de nombre canonicalizado → 409 ("customer already exists").
- Restore de proyecto cuyo customer padre está archivado → 409 ("project parent customer is archived").

## Qué revisar en UI / API / DB / env
- **UI (FE-010):** botón delete apunta a `/hard`; manejo de 409s; envío/lectura de `actor_id`.
- **API:** OpenAPI/clientes generados que aún apunten a `DELETE /:id` (sin `/hard`).
- **DB:** columnas `tenant_id` en todo el grafo, `actor_id` en customers, tablas `actors`/`legacy_actor_map`, columnas `archive_batch_id`/`archive_origin_entity`/`archive_origin_id`/`archive_reason` presentes.
- **Env:** flag/modo de tenant strict (lo gobierna 008/authz).

## Qué validar en el otro repo (FE-010)
- BFF `projects.ts`: endpoint de hard-delete migrado a `/hard`.
- Tipos del front incluyen `actor_id` opcional en customer/manager/investor.
- `pages/admin/projects`: flujo archivar-antes-de-borrar y mensajes de 409.
- `yarn test` / `yarn build` / e2e de projects en el repo FE.

## Señales de incompletitud / incompatibilidad
- Build BE: `undefined: actorsync.* / lifecycle.* / authz.* / tenancy.*` o `text.CanonicalizeName` → falta dependencia.
- Build BE: `(customerdom.Customer).ActorID undefined` → falta domain de 007/011.
- Runtime: `column "tenant_id"/"actor_id" does not exist` → faltan migraciones.
- FE: 404 en delete o `actor_id` ignorado → desincronía de contrato (FE no porteado).
