# extraction-plan.md — feature-009 · CRUDAR archive/restore/hard surface

- **repo**: `ponti-backend` (`/home/pablocristo/Proyectos/pablo/ponti/core`)
- **rama base**: `develop` (tip `003a9b8f`)
- **SOURCE**: `develop-problematico~1` (SHA `777e5f6a`). NUNCA usar `develop-problematico` (su tip es un restore/vacío).
- **rama sugerida**: `pr/feature-009-crudar-archive-surface-be` (o, si se parte: `pr/feature-009-crudar-customer-be`, `...-lot-be`, `...-supply-be`, `...-workorder-be`, `...-masterdata-be`).

## Pre-requisito DURO

`develop` debe contener **feature-002 (crudar-lifecycle-framework)** ya mergeada: aporta `internal/<dom>/repository.go` con `ArchiveX/RestoreX/HardDeleteX`, `shared/models/base.go` con `DeletedAt`, y los helpers `shared/handlers`. Verificar antes de empezar:

```
/usr/bin/git -C <repo> grep -l "func (r \*Repository) HardDeleteCustomer" -- internal/customer/repository.go
/usr/bin/git -C <repo> grep -n "DeletedAt gorm.DeletedAt" internal/shared/models/base.go
```

Si esos símbolos NO existen en `develop` → **postergar 009** hasta mergear 002.

## PR title + description (PR único, si se hace monolítico)

**Title**: `refactor(be): CRUDAR archive/restore/hard surface across domains (#009)`

**Description**:
```
Homogeneiza la superficie HTTP de ciclo de vida en ~14 dominios:
- POST /:id/archive, POST /:id/restore, DELETE /:id/hard, GET /archived
- DeleteX -> HardDeleteX en UseCasesPort/RepositoryPort
- helper run<Entity>IDAction (parse id -> action -> 204)
- baja DELETE /:id legacy (alias documentado en docs/crudar-lifecycle.md)

Depende de #002 (crudar-lifecycle-framework). BE-first: habilita FE-014 / FE-006.
NO incluye: platform import swap (001), csvexport (013), tenant/actor en models (001/003/007),
json-tags de dominio (027), GetProtected removal (008).
```

## Estrategia recomendada: PRs por entidad + hunks parciales

No hacer `checkout` de archivo entero (arrastra 001/008/013/027). Usar `git restore -p` para seleccionar solo los hunks CRUDAR. Agrupación sugerida:

1. `customer` (caso de referencia, hunks más limpios).
2. `lot` (cuidado: separar archive/restore de GetMetrics y csvexport).
3. `supply` + supply-movements + stock-movements.
4. `work-order` + `work-order-draft`.
5. `field` + `manager` + `investor`.
6. master-data simples: `business-parameters`, `category`, `class-type`, `crop`, `provider`.

## Pasos ordenados (por entidad, ejemplo customer)

1. Partir de develop actualizado y crear rama.
2. Traer tests creados enteros (son 100% 009).
3. Seleccionar hunks CRUDAR de handler.go y usecases.go con `restore -p` (aceptar solo los hunks que mencionan `Archive/Restore/HardDelete/archived/runCustomerIDAction`; **rechazar** los hunks `core→platform` y `GetProtected`).
4. Compilar y testear el paquete.
5. Repetir por entidad.
6. Actualizar/incluir `docs/crudar-lifecycle.md` (alias legacy por recurso).

## Archivos enteros vs parciales

- **Enteros** (creados por 009): los `*_actions_test.go`, `handler_delete_test.go` (customer), `repository_crudar_test.go` (lot — depende de 002), `class-type/usecases_test.go`.
- **Parciales** (`restore -p`): todos los `handler.go` y `usecases.go`, los tests modificados, `supply/usecases_movement.go`, `supply/mocks/mock_repository.go` (mejor regenerar el mock).
- **NO traer**: todo lo de la sección E de file-list.md (repository/models/*, dollar/*, commercialization/*, invoice/*, stock/*, report/*, dashboard/*, businessinsights/*, domain/* json-tags, DTOs de actor/csv).

## Migraciones / tests a incluir

- Migraciones: ninguna propia de 009 (esquema soft-delete = feature-002).
- Tests a incluir: los listados en file-list.md sección A.1.

## Comandos git SUGERIDOS (NO ejecutar aquí; para un humano)

```
R=/home/pablocristo/Proyectos/pablo/ponti/core
# 0) confirmar prerequisito 002 (ver arriba)
git -C "$R" checkout develop
git -C "$R" pull
git -C "$R" checkout -b pr/feature-009-crudar-customer-be

# 1) tests creados enteros (customer)
git -C "$R" checkout develop-problematico~1 -- internal/customer/handler_delete_test.go

# 2) hunks parciales del handler/usecases (aceptar SOLO hunks Archive/Restore/Hard/archived/runCustomerIDAction)
git -C "$R" restore -p --source=develop-problematico~1 -- \
  internal/customer/handler.go internal/customer/usecases.go internal/customer/repository_harddelete_test.go

# 3) chequeo de whitespace y build
git -C "$R" diff --check
go -C "$R" build ./...
go test "$R/internal/customer/..."
```

Para lot/supply (mezclados con csvexport y metrics) usar SIEMPRE `restore -p` y rechazar a mano los hunks de `csvexport`, `GetMetrics(LotListFilter)`, `core→platform` y `GetProtected`.

Solo-lectura permitido para auditar: `git -C "$R" diff 0972e565..777e5f6a -- <path>`, `git -C "$R" show 777e5f6a:<path>`, `git -C "$R" log --oneline 0972e565..777e5f6a`.

## Qué NO traer

- `core/* → platform/*` (001), `csvexport` (013), `GetProtected()` removal (008), `tenant_id/actor_id` en models (001/003/007), json-tags (027), `DeleteInvoice` (no es CRUDAR).

## Qué podría romperse

- Si 002 no está en develop: build roto (faltan `HardDeleteX` en repository.go).
- Si se trae `handler.go` entero: `ginmw` apuntará a `platform/*` que quizá no exista en develop (rompe import) → arrastra 001.
- `supply/mocks/mock_repository.go`: si los métodos del mock no coinciden con la interfaz tras la extracción parcial, no compila → preferible regenerar el mock con la herramienta del repo.

## Cómo detectar extracción incompleta

```
# no debe quedar ningún DeleteX ambiguo en interfaces de dominios CRUDAR
git -C "$R" grep -n "DeleteCustomer\|DeleteLot\|DeleteField\|DeleteManager\|DeleteWorkOrder\b" -- internal/**/handler.go internal/**/usecases.go
# deben existir las rutas nuevas
git -C "$R" grep -n "/:.*_id/hard\|/:.*_id/archive\|/:.*_id/restore\|GET(\"/archived\"" -- internal/**/handler.go
```

## Qué validar antes del PR

- `go build ./...` y `go test ./internal/<dominio>/...` verdes.
- No se coló `csvexport`/`platform`/`GetProtected` ajeno: `git diff develop... | grep -E "csvexport|devpablocristo/platform|GetProtected"` debe estar vacío salvo lo intencional.
- Lint del repo (staticcheck) verde — varios commits del rango fueron fixes de lint CRUDAR (`656074b7`, `d629d547`).

## Coordinación cross-repo

- **BE-first**: mergear 009 (y su prerequisito 002) antes que FE-014 y FE-006.
- Avisar al equipo FE el cambio de contrato (DELETE ya no archiva; usar `/hard`, `/archive`, `/restore`, `/archived`).

## Qué hacer después de mergear

- Verificar que FE-014/006 apunten a los endpoints nuevos.
- Eventual deprecación de los alias legacy `DELETE /:id` (documentado como temporal en `docs/crudar-lifecycle.md`).
