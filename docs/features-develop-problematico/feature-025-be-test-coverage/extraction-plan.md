# extraction-plan.md — feature-025 · BE test coverage sweep

- **repo:** ponti-backend (`/home/pablocristo/Proyectos/pablo/ponti/core`)
- **rama base:** `develop` (tip 003a9b8f)
- **SOURCE:** `develop-problematico~1` (SHA **777e5f6a**). NUNCA usar `develop-problematico` (su tip es un restore/vacío).
- **rama sugerida:** `pr/feature-025-be-test-coverage-be`
- **orden:** SOLO-BE. No hay coordinación con FE.

## PR title

`test(be): cobertura unitaria de tenancy, archive-refs y lifecycle (feature-025)`

## PR description (borrador)

> Barrido de cobertura unitaria del backend (45 archivos: 44 nuevos + 1 modificado).
> Tres familias: `handler_test.go` (rutas/actor/lifecycle), `repository_tenant_test.go` (aislamiento
> multi-tenant + strict mode, sqlite in-memory) y `repository_archived_refs_test.go` (integridad
> referencial contra entidades archivadas). Tests white-box (`package <modulo>`).
>
> **PRE-REQUISITO DURO:** este PR depende del código de producción de las features 001
> (platform-tenancy-refactor), 002 (crudar-lifecycle-framework) y 009 (crudar-archive-surface).
> NO mergear hasta que esos tres estén en `develop`, o CI rompe el build.
> Sin migraciones, sin cambios de producción, sin tocar `go.mod`/`go.sum`.

## Decisión de granularidad (RECOMENDADA)

Dado que "merge = sigue a su módulo", lo más seguro es **partir en 3 sub-PRs** y enganchar cada uno a su
feature productora, en este orden:

1. Con/después de **001+003**: los 23 `repository_tenant_test.go`.
2. Con/después de **009**: los 10 `repository_archived_refs_test.go` + los `handler_test.go` que usan `GetArchivedParameters`.
3. Con/después de **002**: `work-order/handler_test.go` (M) + handler_test con rutas archive/restore/hard.

Alternativa (un solo PR de tests): válida solo si 001/002/003/009 ya están todas en `develop`.

## Pasos ordenados (para el PR único; adaptar si se parte en 3)

1. Confirmar que 001, 002, 003 y 009 ya están en `develop`. Si NO, FRENAR.
2. Crear la rama:
   ```
   git checkout develop
   git checkout -b pr/feature-025-be-test-coverage-be
   ```
3. Traer los 45 archivos enteros desde el SOURCE (todos son whole-file, ninguno es partial):
   ```
   git checkout 777e5f6a -- \
     internal/business-parameters/handler_test.go \
     internal/business-parameters/repository_tenant_test.go \
     internal/businessinsights/handler_test.go \
     internal/businessinsights/repository_tenant_test.go \
     internal/category/handler_test.go \
     internal/category/repository_tenant_test.go \
     internal/class-type/handler_test.go \
     internal/class-type/repository_tenant_test.go \
     internal/commercialization/handler_test.go \
     internal/commercialization/repository_tenant_test.go \
     internal/crop/handler_test.go \
     internal/crop/repository_tenant_test.go \
     internal/customer/repository_archived_refs_test.go \
     internal/customer/repository_tenant_test.go \
     internal/dashboard/handler_test.go \
     internal/dashboard/repository_tenant_test.go \
     internal/dollar/handler_test.go \
     internal/dollar/repository_tenant_test.go \
     internal/field/repository_archived_refs_test.go \
     internal/field/repository_tenant_test.go \
     internal/investor/repository_archived_refs_test.go \
     internal/investor/repository_tenant_test.go \
     internal/invoice/handler_test.go \
     internal/invoice/repository_tenant_test.go \
     internal/labor/repository_archived_refs_test.go \
     internal/labor/repository_tenant_test.go \
     internal/lease-type/handler_test.go \
     internal/lease-type/repository_tenant_test.go \
     internal/lot/repository_archived_refs_test.go \
     internal/lot/repository_tenant_test.go \
     internal/manager/repository_archived_refs_test.go \
     internal/manager/repository_tenant_test.go \
     internal/provider/handler_test.go \
     internal/provider/repository_tenant_test.go \
     internal/report/handler_test.go \
     internal/report/repository_tenant_test.go \
     internal/stock/repository_tenant_test.go \
     internal/supply/repository_archived_refs_test.go \
     internal/supply/repository_movement_archived_refs_test.go \
     internal/supply/repository_tenant_test.go \
     internal/work-order-draft/repository_archived_refs_test.go \
     internal/work-order-draft/repository_tenant_test.go \
     internal/work-order/handler_test.go \
     internal/work-order/repository_archived_refs_test.go \
     internal/work-order/repository_tenant_test.go
   ```
   Nota: `internal/work-order/handler_test.go` ya existe en develop; `git checkout 777e5f6a -- <path>`
   lo sobreescribe con la versión del SOURCE (que es lo que queremos: incluye los stubs renombrados).
4. Verificar que no quedó nada a medio aplicar: `git status` y `git diff --check`.
5. Compilar y testear (ver `validation.md`):
   ```
   go build ./...
   go test ./internal/... 2>&1 | tail -40
   ```
6. Commit (NO ejecutar sin pedido del humano) y abrir PR contra `develop`.

## Archivos enteros vs parciales

- **Enteros (whole-file):** los 45. Incluso el `M` (`work-order/handler_test.go`) se trae entero desde
  el SOURCE; no hace falta `git restore -p` porque la versión del SOURCE es la final deseada y el archivo
  es 100% de test (no mezcla producción).
- **Parciales (partial-hunks):** ninguno.

## Migraciones / tests a incluir

- Migraciones: **ninguna**. Cada test usa sqlite `:memory:` con `CREATE TABLE` inline.
- Tests: los 45 archivos SON el contenido.

## Dependencias previas (deben estar en develop ANTES)

001 (tenancy) · 003 (strict mode / multitenant-db-hardening) · 009 (archive surface) · 002 (lifecycle, para work-order handler).

## Coordinación con el otro repo

Ninguna. Sin cambios FE.

## Qué NO traer

- Ningún `.go` de producción (handlers, repositories, usecases, domain): vienen de 001/002/009, no de aquí.
- `go.mod` / `go.sum`: las deps de test ya están en develop.

## Qué podría romperse

- **Build de CI** si se mergea antes que 001/002/009: errores `undefined: assertCustomerReferencesActive`,
  `undefined: HardDeleteWorkOrder`, `unknown field ucs`, `contextkeys.OrgID` no usado, etc.
- Si la firma de producción de algún `assert*ReferencesActive` o de `Routes()`/`NewHandler` difiere entre
  el SOURCE y lo que finalmente se mergeó en 001/002/009, el test no compilará: hay que ajustar el test
  (no la producción) o re-extraer la versión coherente.

## Cómo detectar extracción incompleta

- `go test ./internal/...` falla a compilar en algún paquete → falta su feature productora o se trajo el
  test sin su par (ej. trajiste `repository_tenant_test.go` pero el repo de producción aún no tiene tenancy).
- `grep -rL "^package " internal/.../*_test.go` no debería devolver nada.

## Qué validar antes del PR

- `go build ./...` OK.
- `go test ./internal/...` verde.
- `git diff develop --name-only` lista SOLO archivos `*_test.go` (cero archivos de producción/config).

## Qué hacer después de mergear

- Asegurar que CI corre `go test ./internal/...` (no solo `go build`).
- Si se partió en 3 sub-PRs, cerrar feature-025 cuando los tres estén mergeados.
