# spec.md — feature-009 · CRUDAR archive/restore/hard surface

- **id**: feature-009
- **slug**: crudar-archive-surface
- **nombre**: CRUDAR archive/restore/hard surface
- **tipo**: refactor (cambio de contrato HTTP, sin cambio de esquema DB propio)
- **repo**: Backend Go — `ponti-backend` (`/home/pablocristo/Proyectos/pablo/ponti/core`)
- **existe-en-BE**: Sí (esta extracción).
- **existe-en-FE**: No directamente. Es **Solo-BE**. La contraparte de UI vive en feature-014 (pages de master data que consumen los nuevos endpoints) y feature-006 (ArchivedListPage). En el repo FE NO hay carpeta para 009 → en su `cross-repo-map` figurar como "sin cambios FE".

## Resumen

Homogeneiza la superficie HTTP del ciclo de vida CRUDAR (Create-Read-Update-Archive-Restore-hard delete) a través de ~20 dominios del backend. Cambia el contrato:

```
ANTES:  DELETE /:id            (ambiguo: en unos dominios = hard delete, en lot = archive)
DESPUÉS:
  POST   /:id/archive          (soft delete: deleted_at = NOW)
  POST   /:id/restore          (limpia deleted_at)
  DELETE /:id/hard             (hard delete físico; suele exigir archivado previo y bloquear si hay hijos)
  GET    /archived             (lista archivados: deleted_at IS NOT NULL)
```

A nivel de código, cada handler:
1. Renombra el método de usecase/repo `DeleteX` → `HardDeleteX` (semántica explícita).
2. Quita el `DELETE /:id` legacy (en algunos dominios se conserva como alias hacia hard delete; ver `docs/crudar-lifecycle.md`).
3. Agrega `POST /:id/archive`, `POST /:id/restore`, `DELETE /:id/hard`, `GET /archived`.
4. Introduce un helper privado `run<Entity>IDAction(c, action func(ctx,int64) error)` que parsea el `:id`, ejecuta la acción y responde 204 — elimina la triplicación de parseo/respuesta.
5. Quita `GetProtected()` del `MiddlewaresEnginePort` (limpieza de puerto; ver nota de contaminación).

## Objetivo

Que el contrato de borrado/archivado sea uniforme y predecible para el FE (014/006): `DELETE` HTTP queda reservado a hard delete; archive/restore son verbos explícitos POST; y existe siempre `GET /archived` para alimentar la `ArchivedListPage`.

## Problema que resuelve

- Asimetría histórica: `lot.DELETE /:id` hacía archive (soft), mientras customer/project/work-order/supply hacían hard delete. `field` y `manager` ni siquiera exponían DELETE.
- El verbo `DeleteX` no decía si borraba físico o lógico → ambigüedad en usecase/repo.
- El FE no podía inventar reglas de ciclo de vida; necesitaba endpoints explícitos (regla de `docs/crudar-lifecycle.md`: "Frontend must not invent lifecycle rules; backend/usecase policy is the source of truth").

## Alcance en este repo (BE)

Dominios con superficie CRUDAR real tocada en mi flist (handler.go + usecases.go + tests):
`business-parameters`, `category`, `class-type`, `crop`, `customer`, `field`, `investor`, `labor`, `lot`, `manager`, `provider` (parcial), `supply` (+ supply-movements), `work-order`, `work-order-draft`.

Patrón por dominio (verificado en diffs):
- Interfaces `UseCasesPort`/`RepositoryPort`: `DeleteX`→`HardDeleteX`; se agregan `ArchiveX`, `RestoreX`, `ListArchivedX` cuando faltaban.
- `Routes()`: nuevas rutas `/archive`, `/restore`, `/hard`, `/archived`; baja del `DELETE /:id` legacy.
- Helper `run<Entity>IDAction`.
- Particularidades:
  - **lot**: `DELETE /:id` que antes era archive ahora es hard delete; el 409 de hard-delete bloqueado lleva prefijo machine-readable (commit `ca8a2208`). `GetMetrics` cambió firma a `LotListFilter` (eso es lot-metrics, feature DONE — NO es 009).
  - **supply**: agrega ciclo de vida también a **supply-movements** (scoped por `project_id`) y endpoints globales `GET /supply-movements/archived` y `GET /stock-movements/archived`. `runSupplyMovementAction(ctx, projectID, movementID)`.
  - **labor**: doble grupo (project-scoped y work-order-scoped); `ListArchivedLabors` + `ListArchivedLaborsGlobal`.
  - **work-order-draft**: ciclo CRUDAR completo nuevo.

## Alcance en el otro repo (FE)

Sin cambios FE en 009. El consumo de estos endpoints (`POST /:id/archive`, `DELETE /:id/hard`, `GET /archived`) ocurre en feature-014 (master-data pages) y feature-006 (ArchivedListPage / design system). 009 debe mergear **antes** que 014/006 (BE-first) para que el contrato exista.

## Fuera de alcance (NO traer en este PR)

- Implementaciones `repository.go` de Archive/Restore/HardDelete (transacciones, `archive_batch_id`, `tenancy.Scope`, `Unscoped`): **pertenecen a feature-002** (crudar-lifecycle-framework) y NO están en mi flist. 009 asume que ya existen.
- Columnas `deleted_at/deleted_by/archive_batch_id` en `shared/models/base.go`: feature-002.
- `tenant_id`/`TenantID` en `repository/models/*.go`: features 001/003.
- `actor_id`/`ActorID`: feature-007.
- Limpieza de json-tags de dominio (`usecases/domain/*`): feature-027.
- Swap de import `core/*` → `platform/*`: feature-001 (aparece como ruido en casi todos los archivos).
- Reemplazo XLSX→CSV (`csvexport`): feature-013.
- Quita de `GetProtected()` del puerto de middlewares: feature-008 (identity-tenant-context). Acompaña a 009 pero conceptualmente es de 008.

## Comportamiento esperado (contrato final)

| Verbo HTTP | Efecto | deleted_at |
|---|---|---|
| `POST /:id/archive` | soft delete | NULL → NOW |
| `POST /:id/restore` | restaurar | NOT NULL → NULL (rechaza si el padre sigue archivado) |
| `DELETE /:id/hard` | borrado físico | fila eliminada (suele exigir archivado previo; 409 si tiene hijos activos) |
| `GET /archived` | listar archivados | filtra `deleted_at IS NOT NULL` |
| Respuesta exitosa de acciones por id | 204 No Content | — |

## Estado en dp~1 (SHA 777e5f6a)

Completo y coherente a nivel de la superficie HTTP: todos los handlers de los dominios CRUDAR exponen el contrato nuevo, con tests de handler (`*_actions_test.go`) que verifican que cada ruta llama al usecase correcto y responde 204. Compila como parte del árbol `develop-problematico~1`.

## Criterios de aceptación

1. Cada dominio CRUDAR expone `POST /:id/archive`, `POST /:id/restore`, `DELETE /:id/hard`, `GET /archived`.
2. No queda ningún `DeleteX` ambiguo en las interfaces de los dominios tocados (renombrado a `HardDeleteX`), salvo `invoice` (clave compuesta, no es CRUDAR) que conserva `DeleteInvoice`.
3. Acciones por id responden 204.
4. `go build ./...` y `go test ./internal/<dominio>/...` verdes en los paquetes tocados.
5. `docs/crudar-lifecycle.md` refleja qué dominios conservan el alias legacy.

## Endpoints / modelos / UI / DB / tests afectados

- **Endpoints**: ver tabla de comportamiento; por dominio el `:id` es `customer_id`, `lot_id`, `field_id`, `manager_id`, `investor_id`, `class_type_id`, `parameter_id`, `work_order_id`, `work_order_draft_id`, `supply_id`, `supply_movement_id`/`stock_movement_id`, `labor_id`.
- **Modelos/DTOs**: 009 NO cambia DTOs propios. Los cambios en `handler/dto/*` de mi flist son ruido de otras features.
- **UI**: ninguna (FE en 014/006).
- **DB**: ninguna migración propia de 009 (el esquema soft-delete lo aporta 002).
- **Tests creados (009)**: `internal/customer/handler_delete_test.go`, `internal/field/handler_actions_test.go`, `internal/investor/handler_actions_test.go`, `internal/manager/handler_actions_test.go`, `internal/lot/handler_actions_test.go`, `internal/lot/repository_crudar_test.go`, `internal/class-type/usecases_test.go`. Tests modificados: `internal/customer/repository_harddelete_test.go`, `internal/supply/usecases_delete_test.go`, `internal/supply/repository_delete_test.go`, `internal/supply/repository_movement_delete_test.go`, `internal/work-order-draft/usecases_test.go`.

## Dependencias

- **Intra-repo (fuerte)**: feature-002 (crudar-lifecycle-framework) DEBE ir antes — aporta los `repository.go` con Archive/Restore/HardDelete y `shared/models/base.go` con soft-delete. Sin 002, 009 no compila.
- **Intra-repo (débil)**: feature-008 (GetProtected removido del puerto), feature-001 (platform import), feature-013 (csvexport en lot/supply handlers) coexisten en los mismos archivos.
- **Cross-repo (fuerte hacia adelante)**: bloquea a FE-014 y FE-006 (consumen el contrato). BE-first.

## Riesgos

- **Funcional**: clientes viejos que llamaban `DELETE /:id` esperando archive (caso lot) ahora reciben hard delete. Mitigación: alias legacy documentado + FE 014/006 migrado.
- **Técnico**: el diff de cada handler viene mezclado con import platform/csvexport/GetProtected → riesgo alto de arrastrar features ajenas si se hace `checkout` de archivo entero.
- **Extracción**: `repository.go` (la implementación real) no está en mi flist; si feature-002 no se mergea primero, esto rompe build.

## DECISIÓN recomendada

**Partir en subfeatures por entidad y extraer como hunks parciales, después de 002.**
- 009 NO se puede extraer "tal cual" archivo-entero: arrastraría 001/008/013/027.
- Recomendado: PRs por-entidad (customer, lot, supply, work-order, field/manager/investor, master-data simples) usando `git restore -p` para tomar solo los hunks de `archive/restore/hard/archived/runXIDAction` en handler.go + usecases.go + sus tests, **una vez** que 002/003/007/001 estén en `develop`.
- Si 002 aún no está mergeado: **postergar** 009 hasta que 002 entre, porque el build depende de los métodos de repositorio.
