# validation.md — feature-025 · BE test coverage sweep

## Pre-condición (verificar PRIMERO)

- [ ] 001 (tenancy), 003 (strict mode), 009 (archive surface) y 002 (lifecycle) están en `develop`.
  - `git grep -ln "func assertCustomerReferencesActive" develop -- internal/customer/` → debe devolver `repository.go`.
  - `git grep -ln "func.*HardDeleteWorkOrder" develop -- internal/work-order/` → debe devolver `usecases.go`/`repository.go`/`handler.go`.
  - `git grep -n "GetArchivedParameters" develop -- internal/business-parameters/` → debe existir en `usecases.go`/`handler.go`.
  - Si alguno NO aparece → FRENAR, falta su feature productora.

## Checklist pre-PR

- [ ] La rama solo contiene archivos `*_test.go`: `git diff develop --name-only` → cero archivos de producción/config.
- [ ] No se tocó `go.mod`/`go.sum`: `git diff develop -- go.mod go.sum` vacío.
- [ ] `go build ./...` sin errores.
- [ ] `go vet ./internal/...` limpio (opcional pero recomendado).
- [ ] `git diff --check` sin whitespace errors.

## Tests sugeridos (BE)

```
# suite completa de los paquetes tocados
go test ./internal/... 2>&1 | tail -60

# por familia, si se parte en sub-PRs:
go test ./internal/customer/... ./internal/field/... ./internal/investor/... \
        ./internal/labor/... ./internal/lot/... ./internal/manager/... \
        ./internal/supply/... ./internal/work-order/... ./internal/work-order-draft/...   # archived-refs
go test ./internal/business-parameters/... ./internal/work-order/...                       # handlers/lifecycle

# tests puntuales clave
go test ./internal/business-parameters/ -run TestBusinessParameterRepositoryTenantIsolation -v
go test ./internal/business-parameters/ -run TestBusinessParameterRepositoryRequiresTenantInStrictMode -v
go test ./internal/work-order/ -run TestWorkOrderActionRoutesCallExplicitUseCases -v
go test ./internal/customer/ -run TestAssertCustomerReferencesActive -v
go test ./internal/supply/ -run TestAssertSupplyMovementReferencesActive -v
```

Esperado: todo verde. Si un paquete no compila → falta su feature productora o hay desalineación de firmas.

## FE

N/A — sin cambios FE. No correr `yarn test` / build / e2e para esta feature.

## Casos borde que cubren los tests (verificar que pasen)

- **Tenant:** cross-tenant `GetByKey`/`Update`/`HardDelete` deben FALLAR; `ListAll` solo devuelve filas del
  tenant en contexto; `TENANT_STRICT_MODE=true` sin tenant → error en `ListAll` y `Create`.
- **Archived refs:** referencia a entidad con `deleted_at != NULL` (id=99) → error kind `Conflict`
  (`domainerr.IsConflict`) con mensaje "<entity> is archived"; actor nil / id=0 / entidad nil → sin error.
- **Handlers:** create → 201 con actor en `CreatedBy`; update id=42 → 204 con `UpdatedBy`; rutas
  `/work-orders/:id/archive|restore|hard` → 204 y llaman al usecase correcto (`actionCall`).

## Qué revisar en UI / API / DB / env

- **UI:** nada.
- **API:** los tests no cambian contratos; solo verifican rutas existentes (definidas en 002/009).
- **DB:** sin migraciones; sqlite in-memory por test. No tocar la DB local/prod.
- **env:** `TENANT_STRICT_MODE` se setea con `t.Setenv` (local al test); no requiere config de entorno real.

## Qué validar en el otro repo

Nada. Anotar en cross-repo-map del FE: "feature-025: sin cambios FE".

## Señales de incompletitud / incompatibilidad

- `undefined: assert...ReferencesActive` / `undefined: HardDeleteWorkOrder` / `GetArchivedParameters` →
  falta 009/002 en la base.
- `unknown field ucs in struct literal of type Handler` o `contextkeys.OrgID` no resuelto → falta 001.
- `cannot use ... as ... in argument` en `assertSupplyMovementReferencesActive`/`Routes()`/`NewHandler` →
  firma de producción divergente entre SOURCE y develop: ajustar el TEST, no producción.
