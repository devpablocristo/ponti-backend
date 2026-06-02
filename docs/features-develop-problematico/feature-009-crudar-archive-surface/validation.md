# validation.md — feature-009 · CRUDAR archive/restore/hard surface

## Checklist pre-PR (BE)

- [ ] **Prerequisito 002 en develop**: `git -C <repo> grep -n "func (r \*Repository) HardDeleteCustomer" internal/customer/repository.go` devuelve resultado.
- [ ] `shared/models/base.go` tiene `DeletedAt gorm.DeletedAt`.
- [ ] No se coló ruido ajeno:
  ```
  git -C <repo> diff develop...HEAD | grep -E "csvexport|devpablocristo/platform|GetProtected" 
  # debe estar vacío (salvo decisión explícita sobre platform)
  ```
- [ ] No quedan `DeleteX` ambiguos en interfaces:
  ```
  git -C <repo> grep -nE "DeleteCustomer|DeleteLot|DeleteField|DeleteManager|DeleteWorkOrderByID|DeleteParameter|DeleteClassType|DeleteCrop|DeleteCategory" -- internal/**/handler.go internal/**/usecases.go
  ```
- [ ] Rutas nuevas presentes en cada dominio:
  ```
  git -C <repo> grep -nE '/:.*_id/hard|/:.*_id/archive|/:.*_id/restore|GET\("/archived"' -- internal/**/handler.go
  ```
- [ ] `go build ./...` verde.
- [ ] `go test ./internal/...` verde para paquetes tocados (ver abajo).
- [ ] `gofmt`/lint (staticcheck) verde — el rango original tuvo varios fixes de lint CRUDAR.
- [ ] `git diff --check` sin errores de whitespace.

## Tests sugeridos (BE)

```
go test ./internal/customer/...
go test ./internal/lot/...
go test ./internal/supply/...
go test ./internal/work-order/... ./internal/work-order-draft/...
go test ./internal/field/... ./internal/manager/... ./internal/investor/...
go test ./internal/class-type/... ./internal/category/... ./internal/crop/... ./internal/business-parameters/...
go test ./internal/labor/...
```

Tests clave a verificar verdes:
- `internal/lot/handler_actions_test.go::TestLotIDActionHandlersCallExplicitUseCases` (archive/restore/hard → 204).
- `internal/customer/handler_delete_test.go`, `internal/field/handler_actions_test.go`, `internal/investor/handler_actions_test.go`, `internal/manager/handler_actions_test.go`.
- `internal/lot/repository_crudar_test.go` (cobertura CRUDAR + 409 bloqueado) — requiere 002.
- `internal/class-type/usecases_test.go`.

## Validación manual (API)

Para un dominio (ej. customers, base `/v1`):
- [ ] `POST /v1/customers/:id/archive` → 204; el registro deja de aparecer en `GET /v1/customers`.
- [ ] `GET /v1/customers/archived` → incluye el registro archivado.
- [ ] `POST /v1/customers/:id/restore` → 204; vuelve a `GET /v1/customers`.
- [ ] `DELETE /v1/customers/:id/hard` sobre un registro con hijos activos → 409 (con prefijo machine-readable en lot).
- [ ] `DELETE /v1/customers/:id/hard` sobre archivado sin hijos → 204; desaparece de `/archived`.
- [ ] (supply) `GET /v1/supply-movements/archived` y `GET /v1/stock-movements/archived` globales responden.
- [ ] (supply) `POST /v1/projects/:project_id/supply-movements/:supply_movement_id/archive` → 204.

## Casos borde

- [ ] Restaurar un hijo cuyo padre sigue archivado → debe rechazar.
- [ ] Hard delete de un registro NO archivado → política (rechazar si exige archivado previo).
- [ ] `:id` inválido/no numérico → 400 (vía `ParseParamID`).
- [ ] lot: confirmar que `DELETE /v1/lots/:id` (alias legacy) hace HARD, no archive (cambio de semántica).
- [ ] Paginación de `/archived` (page/perPage; lot usa max 1000).

## Qué revisar en UI / API / DB / env

- **UI**: nada en 009 (FE-014/006).
- **API**: contrato de la tabla anterior; 204 en acciones por id.
- **DB**: que `deleted_at`/`deleted_by`/`archive_batch_id` existan (migraciones de 002 aplicadas).
- **Env**: ninguna var nueva.

## Qué validar en el otro repo (FE)

- FE-014 (master-data pages) y FE-006 (ArchivedListPage) apuntan a `/archive`, `/restore`, `/hard`, `/archived`.
- No quedan llamadas a `DELETE /:id` esperando archive.

## Señales de incompletitud / incompatibilidad

- Build falla por `HardDeleteX` no definido → falta 002.
- Import `github.com/devpablocristo/platform/...` no resuelve → develop está en `core/*`; rechazaste mal los hunks de import o falta 001.
- Mock de supply no implementa la interfaz → regenerar.
- Un dominio sin `GET /archived` → hunk no extraído.
- `grep csvexport` aparece en el diff del PR → arrastraste feature-013.
