# spec.md — feature-025 · BE test coverage sweep

- **id:** feature-025
- **slug:** be-test-coverage
- **nombre:** BE test coverage sweep
- **tipo:** tests
- **repo:** Backend Go (ponti-backend) — `/home/pablocristo/Proyectos/pablo/ponti/core`
- **existe-en-FE:** NO. Solo-BE. En el mapa cross-repo del FE se menciona como "sin cambios FE".
- **existe-en-BE:** SÍ (45 archivos de test).
- **merge:** "sigue a su módulo" — no es un PR independiente con vida propia; los tests deben acompañar (o ir como follow-up inmediato de) los PRs de las features que validan (001/002/009).

## Resumen

Barrido de cobertura de tests unitarios sobre los paquetes `internal/*` del backend. Son 45 archivos:
44 creados (`A`) + 1 modificado (`M`: `internal/work-order/handler_test.go`). Tres familias de tests:

1. **`handler_test.go`** (13 archivos creados + 1 modificado) — tests de los handlers HTTP con stubs de
   usecases y `gin.CreateTestContext` / `httptest`. Validan rutas, status codes, parseo de body,
   propagación del actor (`ctxkeys.Actor`) hacia `CreatedBy`/`UpdatedBy`, y las nuevas rutas de
   ciclo de vida (`/archive`, `/restore`, `/hard`).
2. **`repository_tenant_test.go`** (23 archivos) — tests de aislamiento multi-tenant sobre los repositorios,
   con sqlite `:memory:`. Inyectan contexto con `contextkeys.OrgID` (UUID de tenant), `Actor`, `Role`,
   `Scopes`; verifican que las queries filtran por tenant y que en `TENANT_STRICT_MODE=true` fallan
   sin tenant en contexto. Validan la feature **001 (platform-tenancy-refactor)** y **003 (multitenant-db-hardening)**.
3. **`repository_archived_refs_test.go`** (10 archivos, incl. `repository_movement_archived_refs_test.go` de supply)
   — tests de las funciones `assert<Entity>ReferencesActive(db, entity)` que rechazan crear/actualizar
   una entidad si referencia a otra entidad archivada (soft-deleted). Verifican kind `Conflict`
   (`domainerr.IsConflict`) y mensajes. Validan la feature **009 (crudar-archive-surface)**.

## Objetivo

Subir la cobertura de tests del backend tras los refactors grandes (tenancy 001, lifecycle/crudar 002,
archive surface 009) para fijar (lock-in) el comportamiento esperado: aislamiento por tenant, strict mode,
propagación del actor en auditoría, rutas explícitas de archive/restore/hard-delete y la integridad
referencial contra entidades archivadas.

## Problema

Los refactors 001/002/009 cambiaron firmas (`DeleteWorkOrderByID` → `HardDeleteWorkOrder`), agregaron
métodos (`ArchiveWorkOrder`, `RestoreWorkOrder`, `ListArchivedWorkOrders`, `GetArchivedParameters`),
nuevas rutas y nuevas reglas de integridad. Sin tests, estos cambios quedan sin red de seguridad.

## Alcance en este repo (BE)

- 44 archivos de test nuevos + 1 modificado, todos bajo `internal/<modulo>/`.
- **Son tests white-box**: declaran `package <modulo>` (NO `<modulo>_test`), por lo que acceden a
  símbolos no exportados de producción: `assertCustomerReferencesActive`, `assertSupplyMovementReferencesActive`,
  el campo `Handler{ucs: ...}`, etc. Esto los acopla fuertemente al código de producción de su paquete.
- Dependencias de test (ya presentes en `go.mod`/`go.sum` de develop): `gorm.io/driver/sqlite`,
  `gorm.io/gorm`, `github.com/google/uuid`, `github.com/gin-gonic/gin`,
  `github.com/devpablocristo/platform/security/go/contextkeys`,
  `github.com/devpablocristo/platform/errors/go/domainerr`.

## Alcance en el otro repo (FE)

Ninguno. Sin cambios FE. (Anotar en `cross-repo-map` del FE como "feature-025: sin carpeta FE".)

## Fuera de alcance

- NO tocar código de producción (handlers, repositories, usecases, domain). Si un test no compila contra
  develop es porque falta su feature productora (001/002/009), NO porque haya que editar producción aquí.
- NO tocar `go.mod`/`go.sum` (las deps de test ya están en develop por features previas).
- Lot-metrics / total_tons y tentative-prices YA están porteados (#117/#121/#124). Si algún test de
  `work-order`/`lot`/`commercialization` toca esas funciones, ya existen en develop.

## Comportamiento esperado

- `go test ./internal/...` pasa en verde una vez mergeadas 001/002/009.
- Cada test es hermético: sqlite `:memory:` con `CREATE TABLE` inline; no requiere Postgres ni migraciones reales.

## Estado en dp~1 (777e5f6a)

Completo y coherente: los 45 archivos existen y referencian símbolos de producción que TAMBIÉN existen en
777e5f6a (verificado: `assertCustomerReferencesActive` en `internal/customer/repository.go`,
`HardDeleteWorkOrder` en `internal/work-order/{handler,repository,usecases}.go`,
`GetArchivedParameters` en `internal/business-parameters/{handler,usecases}.go`).

## Criterios de aceptación

1. Los 45 archivos se copian sin modificación a la rama de extracción.
2. `go build ./...` y `go test ./internal/...` compilan y pasan.
3. NO se modificó ningún `.go` de producción ni `go.mod`/`go.sum` dentro de este PR.
4. Las features 001/002/009 ya están en la rama base antes de correr los tests.

## Endpoints / modelos / UI / DB / tests afectados

- **Endpoints ejercitados (no creados aquí):** `GET/POST/PUT /api/v1/business-parameters[...]`,
  `POST /api/v1/work-orders/:id/archive`, `POST /api/v1/work-orders/:id/restore`,
  `DELETE /api/v1/work-orders/:id/hard`, y rutas equivalentes por módulo.
- **Modelos/tipos referenciados:** `domain.BusinessParameter`, `domain.Customer`, `domain.SupplyMovement`,
  `domain.WorkOrderListElement`, `domain.ArchivedWorkOrderFilter`, etc.
- **DB:** ninguna migración. Solo `CREATE TABLE` sqlite en memoria dentro de cada test.
- **Tests:** este ES el contenido de la feature.

## Dependencias

- **Intra-repo (fuertes):** 001-be-platform-tenancy-refactor (contextkeys/tenant + `repository_tenant_test`),
  002-be-crudar-lifecycle-framework (`HardDeleteWorkOrder`, archive/restore en handler), 009-crudar-archive-surface
  (`assert*ReferencesActive` + `repository_archived_refs_test`). También se apoya en 003 (strict mode) y
  004 (shared-text-propername, vía dominio).
- **Cross-repo:** ninguna.

## Riesgos

- **Funcional:** bajo (son tests; no cambian runtime).
- **Técnico (ALTO de orden):** los tests NO compilan contra develop tip (003a9b8f) — los símbolos de
  producción NO existen ahí. Mergear este PR antes que 001/002/009 ROMPE el build de CI.

## DECISIÓN recomendada

**Postergar / follow-up.** Extraer tal cual los 45 archivos, pero NO mergear hasta que 001, 002 y 009 estén
en la rama base. Idealmente dividir por familia y enganchar cada grupo al PR de su feature productora
(tenant→001/003, archived-refs→009, lifecycle handler→002). Si se prefiere un PR único de tests, debe ir
DESPUÉS de los tres y validar `go test ./internal/...` en verde antes de abrir.
