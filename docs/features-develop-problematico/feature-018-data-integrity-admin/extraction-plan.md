# feature-018 · data-integrity-admin · extraction-plan (BE)

- **repo**: ponti-backend (`/home/pablocristo/Proyectos/pablo/ponti/core`)
- **rama base**: `develop` (tip `003a9b8f`)
- **SOURCE**: `develop-problematico~1` (SHA `777e5f6a`) — NUNCA `develop-problematico` (tip = restore vacío)
- **rama sugerida**: `pr/feature-018-data-integrity-admin-be`
- **rango de referencia**: `0972e565..777e5f6a`

> Todos los comandos `git` de abajo son **SUGERENCIAS para un humano**. Este documento no
> ejecuta mutaciones. El agente que lo generó solo hizo lectura + escritura de docs.

## PR title

`feat(be): data-integrity refactor — 5 controles SSOT-vs-RAW (feature-018)`

## PR description (borrador)

> Reescribe `internal/data-integrity` (1068→343 líneas en usecases.go): reemplaza los 17
> controles encadenados módulo-vs-módulo por **5 controles independientes** que comparan el
> valor SSOT del dashboard (`v4_*`) contra un recálculo RAW directo sobre `public.*`.
> Cada control corre en paralelo y hace 2 queries cortas (timeout handler 8m→30s).
> Agrega `check_type/severity/recommendation` al DTO. Tenantiza las queries RAW vía
> `internal/shared/authz`. Incluye tests de handler, usecases, tenant-propagation y mocks.
>
> Depende de: base de tenancy (`internal/shared/authz`, features 001/003) y de 4 métodos
> RAW nuevos en `report/supply/project/lot/repository.go` (traídos por partial-hunks).
> Coordinar merge con FE `feature-018` (`pages/admin/data-integrity` + `useDatabase`).

## Pre-requisitos (deben estar en develop ANTES de este PR)

1. **`internal/shared/authz`** disponible en develop (lo aportan features 001/003 be-platform-tenancy / be-multitenant-db-hardening). Sin esto, los métodos RAW no compilan.
   - Verificar: `git -C <repo> ls-tree -r --name-only develop -- internal/shared/authz` debe listar `authz.go`.

## Archivos enteros (whole-file) — desde SOURCE

```
internal/data-integrity/usecases.go
internal/data-integrity/handler.go
internal/data-integrity/handler/dto/integrity_check.go
internal/data-integrity/usecases/domain/types.go
internal/data-integrity/handler_test.go
internal/data-integrity/usecases_tenant_test.go
internal/data-integrity/usecases_test.go
internal/data-integrity/usecases_mock_test.go
```

## Archivos parciales (partial-hunks) — solo los hunks `GetRaw*` / bloque DataIntegrity

```
internal/report/repository.go      → solo func GetRawNetIncome (≈ línea 743 en SOURCE)
internal/supply/repository.go      → solo func GetRawSupplyInvestment (≈ línea 837)
internal/project/repository.go     → solo func GetRawAdminCostTotal (≈ línea 2653)
internal/lot/repository.go         → solo func GetRawLeaseExecuted (≈ línea 829)
internal/work-order/repository.go  → (opcional) solo el filtro tenant de GetRawDirectCost
wire/data_integrity_providers.go   → solo el bloque ProvideDataIntegrity* + DataIntegritySet
```

## Wire regenerado (NO copiar a mano)

```
wire/wire_gen.go → regenerar con: go generate ./wire/...   (o: cd wire && wire)
```
NO traer `wire/wire.go` (su único cambio en rango — `ActorSet`/`ActorHandler` — es de feature-007).

## Migraciones / tests a incluir

- Migraciones: **ninguna** (el módulo no agrega `.sql`).
- Tests: los 4 archivos `*_test.go` del módulo (incluidos como whole-file arriba).

## Pasos ordenados (sugeridos)

```bash
# 0) partir de develop limpio
git -C <repo> checkout develop
git -C <repo> pull --ff-only
git -C <repo> checkout -b pr/feature-018-data-integrity-admin-be

# 1) traer el módulo entero desde SOURCE
git -C <repo> checkout develop-problematico~1 -- \
  internal/data-integrity/usecases.go \
  internal/data-integrity/handler.go \
  internal/data-integrity/handler/dto/integrity_check.go \
  internal/data-integrity/usecases/domain/types.go \
  internal/data-integrity/handler_test.go \
  internal/data-integrity/usecases_tenant_test.go \
  internal/data-integrity/usecases_test.go \
  internal/data-integrity/usecases_mock_test.go

# 2) traer SOLO los hunks GetRaw* de los repos compartidos (interactivo)
git -C <repo> restore -p --source=develop-problematico~1 -- internal/report/repository.go
git -C <repo> restore -p --source=develop-problematico~1 -- internal/supply/repository.go
git -C <repo> restore -p --source=develop-problematico~1 -- internal/project/repository.go
git -C <repo> restore -p --source=develop-problematico~1 -- internal/lot/repository.go
# opcional (tenantizar el método existente):
git -C <repo> restore -p --source=develop-problematico~1 -- internal/work-order/repository.go
#   → en cada hunk: aceptar SOLO el bloque del método GetRaw* (y sus imports authz/domainerr/fmt si faltan).

# 3) re-cablear wire: traer solo el bloque DataIntegrity de los providers
git -C <repo> restore -p --source=develop-problematico~1 -- wire/data_integrity_providers.go

# 4) regenerar wire_gen.go (NO copiarlo a mano)
cd <repo>/wire && go generate ./... ; cd -

# 5) validar
git -C <repo> diff --check
cd <repo> && go build ./... && go vet ./internal/data-integrity/... && go test ./internal/data-integrity/...
```

> Nota sobre `restore -p`: si tu git no soporta `restore -p --source`, usar el equivalente
> `git checkout -p develop-problematico~1 -- <path>`.

## Qué NO traer

- `wire/wire.go` (ActorSet — feature-007).
- Hunks NO-`GetRaw*` de `report/supply/project/lot/work-order/repository.go` (son 001/003/009/dominio-de-otras-features).
- Cualquier archivo de tentative-prices / lot-metrics (ya DONE).

## Qué podría romperse

- **Compilación**: si `internal/shared/authz` no está en develop, los 4 RAW methods (que llaman `authz.TenantFromContext/TenantStrictModeEnabled/TenantMissing`) no compilan.
- **Wire**: si `wire_gen.go` queda desincronizado del providers (firma `ProvideDataIntegrityUseCases` cambió: 6 args reordenados, sin `stockRepo`, con `projectRepo` y repos concretos), `Initialize()` no compila. Por eso hay que regenerar, no copiar.
- **FE**: shape de DTO cambió (3 campos nuevos + `status` set distinto). Si el FE valida estrictamente el JSON, romperá.

## Cómo detectar extracción incompleta

- `go build ./...` falla con `undefined: authz.TenantFromContext` → falta tenancy (001/003).
- `go build ./...` falla con `r.GetRawNetIncome undefined` / `GetRawSupplyInvestment` / `GetRawAdminCostTotal` / `GetRawLeaseExecuted` → faltan hunks RAW en repos compartidos.
- `wire` se queja de provider faltante (`ProvideDataIntegrityProjectRepositoryPort`) o sobrante (`...StockRepositoryPort`) → providers/wire_gen desincronizados.
- `git diff --check` muestra conflict markers o whitespace roto en los repos compartidos (típico al cherry-pickear hunks parciales).

## Qué validar antes del PR

- `go build ./...`, `go vet ./internal/data-integrity/...`, `go test ./internal/data-integrity/...` verdes.
- `wire_gen.go` idempotente (re-correr `go generate` no produce diff).
- `grep -rn "StockRepositoryPort" internal/data-integrity wire/data_integrity_providers.go` → 0 resultados (se eliminó).
- DTO incluye `check_type/severity/recommendation`.

## Coordinación con el otro repo (FE)

- **Orden recomendado: BE-first.** Mergear BE 018 (con el nuevo DTO) y luego el FE 018, que ya
  espera `check_type/severity/recommendation`. Si se mergea FE-first y el BE viejo no envía esos
  campos, el FE debe tolerar ausencia (validar con el paquete FE).
- El PR del FE `feature-018` debe ir **después** del BE para que `useDatabase` y
  `pages/admin/data-integrity` reciban el shape nuevo.

## Qué hacer después de mergear

- Smoke test del endpoint en un entorno con datos: `GET /data-integrity/costs-check?project_id=<real>`.
- Confirmar que el FE admin renderiza los 5 checks y los nuevos campos.
- Si la tenantización de `GetRawDirectCost` quedó fuera, abrir follow-up para igualar el filtro
  tenant con los otros 4 RAW.
