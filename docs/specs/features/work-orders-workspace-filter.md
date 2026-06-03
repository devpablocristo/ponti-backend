# Spec — Listado de órdenes de trabajo acotado al workspace (cliente · proyecto · campaña · campo)

> Las órdenes de trabajo deben listarse **siempre** dentro de un workspace con **cliente + proyecto +
> campaña** elegidos (no archivados); el **campo es opcional**: sin campo = todos los campos del
> proyecto, con campo = solo ese.

- **slug:** `work-orders-workspace-filter`
- **nombre:** Listado de órdenes de trabajo acotado al workspace (cliente · proyecto · campaña · campo)
- **tipo:** refactor / hardening de contrato de filtros (sin migraciones)
- **repo:** `ponti-backend` (core)
- **merge:** BE + FE (BFF/UI), implementados juntos
- **estado:** ✅ **Implementado y verificado e2e** (core + web, rama `work-orders-workspace-filter`). Detalle en §10.

---

## 1. Propósito

Garantizar que las órdenes de trabajo solo se listen **dentro de un workspace válido**:

- **Cliente + proyecto + campaña → obligatorios** y **activos** (no archivados, `deleted_at IS NULL`).
- **Campo → opcional**: sin `field_id` = **todos los campos** del proyecto; con `field_id` = **solo ese**.

Hoy esa regla no se aplica en la capa de lectura de work-orders: los endpoints aceptan los filtros como
opcionales y, peor aún, tragan el error de parseo. El efecto es que `GET /work-orders` **sin filtros
devuelve todas las órdenes de todos los clientes**. Este spec define cerrar ese agujero **reutilizando**
el contrato de workspace que ya usan `report` y `dashboard`, sin features nuevas.

**Aclaración de "activo":** en el sistema, cliente/proyecto/campaña/campo se **archivan** vía soft-delete
(`deleted_at`); "activo" = `deleted_at IS NULL`. **No existe** ni se introduce un concepto de
"campaña vigente/actual": la campaña es solo una **etiqueta** del proyecto (ej. "Campaña 2024") y cada
proyecto pertenece a una sola campaña.

**Principio rector — una sola base de filtros, sin módulos repetidos.** El filtro de workspace
(cliente→proyecto→campaña→campo) es **una única base compartida** (`WorkspaceFilter` + `ParseWorkspaceFilter`
+ `ValidateRequiredWorkspaceFilter` + `ResolveProjectIDs`). Cada consumidor puede pasarle **config distinta**
(qué columnas filtra, filtros extra de dominio, en qué capa valida) — eso es esperable y correcto —, pero
**no debe duplicar el módulo**: nada de re-parsear los IDs a mano, recodificar la regla de requerido ni
reescribir la resolución de proyectos. Esta corrección hace que órdenes de trabajo **llame** a esa base en
vez de tener la suya. Una auditoría detectó que **otros módulos sí la duplicaron**; esa deuda se documenta en
**§9 y queda explícitamente FUERA del alcance de este spec** — acá **solo** se corrige órdenes de trabajo. La
deuda **no se arregla de golpe**: converge **gradualmente, módulo por módulo, a medida que se van aplicando
specs** (plan en §9).

---

## 2. Estado vs `develop`

### Ya existe y funciona (no se toca)

- Órdenes de trabajo **completamente implementadas en `core/`** (no migradas a `platform/`): tablas
  `workorders`, `workorder_items`, `workorder_investor_splits`, drafts digitales, CRUD, export, métricas,
  vista `v4_report.workorder_list`.
- **Contrato de workspace compartido**, patrón establecido del repo:
  - `ParseWorkspaceFilter(c)` — parsea `customer_id`/`project_id`/`campaign_id`/`field_id`
    ([`internal/shared/handlers/workspace_filters.go:14`](../../../internal/shared/handlers/workspace_filters.go)).
  - `ValidateRequiredWorkspaceFilter(f)` — exige `customer_id` + `project_id` + `campaign_id`; `field_id`
    opcional = "todos los campos"
    ([`internal/shared/handlers/workspace_filters.go:46`](../../../internal/shared/handlers/workspace_filters.go)).
  - `ResolveProjectIDs(ctx, db, f)` — resuelve los proyectos filtrando por `p.deleted_at IS NULL`
    (= activo) y valida campo con `EXISTS (... fields f ... f.deleted_at IS NULL)`
    ([`internal/shared/filters/workspace.go:46`](../../../internal/shared/filters/workspace.go)).
  - `ValidateFieldBelongsToProject(ctx, db, projectID, fieldID)`
    ([`internal/shared/filters/workspace.go:106`](../../../internal/shared/filters/workspace.go)).
  - **Consumidores actuales del patrón:** `report`
    ([`internal/report/handler.go:137`](../../../internal/report/handler.go)) y `dashboard`
    ([`internal/dashboard/handler.go:72`](../../../internal/dashboard/handler.go)) llaman a
    `ValidateRequiredWorkspaceFilter` a nivel handler (en sus rutas principales; ambos tienen **además
    duplicaciones puntuales** en otras rutas — ver §9, fuera de alcance).
- **El repositorio de work-order ya usa `ResolveProjectIDs`**
  ([`internal/work-order/repository.go:440`](../../../internal/work-order/repository.go)). Es decir, la
  semántica "activo" y "campo = todos / uno" **ya está resuelta a nivel query**:
  - si los filtros no resuelven ningún proyecto → devuelve **lista vacía** (`repository.go:451-452`);
  - si `field_id` viene → además aplica `field_id = ?` sobre la vista (`repository.go:455-456`).

### Falta (el gap a cubrir)

- Los **4 endpoints de lectura** de work-order **no exigen el mínimo** cliente+proyecto+campaña: nunca
  llaman a `ValidateRequiredWorkspaceFilter`.
- `parseFilters` **traga el error** de `ParseWorkspaceFilter`
  ([`internal/work-order/handler.go:227-230`](../../../internal/work-order/handler.go)):

  ```go
  workspaceFilter, err := sharedhandlers.ParseWorkspaceFilter(c)
  if err != nil {
      return f // <- se descarta el error y se devuelve filtro vacío
  }
  ```

  Un `customer_id` mal formado (ej. `?customer_id=abc`) termina en un filtro vacío → lista global.

---

## 3. Alcance / Archivos

**Archivos de la app modificados (partial-hunks) — solo `work-order` (implementado, ver §10):**

- [`internal/work-order/handler.go`](../../../internal/work-order/handler.go)
  - `parseFilters` → pasar a devolver `(domain.WorkOrderFilter, error)` propagando el error de
    `ParseWorkspaceFilter` (dejar de tragarlo). Mantiene el parseo de `supply_id` / `is_digital` / `status`.
  - `ListWorkOrders`, `ListWorkOrderFilterRows`, `GetMetrics`, `ExportWorkOrders` → tras parsear, invocar
    `sharedhandlers.ValidateRequiredWorkspaceFilter(workspaceFilter)` y `sharedhandlers.RespondError(c, err)`
    si falla, antes de delegar al use case. Espejo exacto de `report`/`dashboard`.
- [`internal/work-order/usecases.go`](../../../internal/work-order/usecases.go)
  - `ListWorkOrderFilterRows` tiene hoy una **validación propia más débil y divergente** —exige
    `project_id` **OR** `field_id` (`usecases.go:246`)— en vez de la regla canónica. **Eliminarla**: la regla
    queda única (`ValidateRequiredWorkspaceFilter` en el handler). Es el "no repetir módulos" del Principio
    rector (§1) aplicado **dentro** de work-order.

**Endpoints alcanzados (los 4 de lectura):**

| Método | Ruta | Handler |
|--------|------|---------|
| `GET` | `/api/v1/work-orders` | `ListWorkOrders` |
| `GET` | `/api/v1/work-orders/filter-rows` | `ListWorkOrderFilterRows` |
| `GET` | `/api/v1/work-orders/metrics` | `GetMetrics` |
| `GET` | `/api/v1/work-orders/export` | `ExportWorkOrders` |

**Fuera de alcance (no se toca):**

- `CreateWorkOrder`, `UpdateWorkOrderByID`, `DeleteWorkOrderByID`, `ArchiveWorkOrder`, `RestoreWorkOrder`,
  `UpdateInvestorPaymentStatus`, `GetWorkOrderByID` — operan **por ID**, no por workspace filter.
- `DuplicateWorkOrder` — es un stub (`handler.go:117`), no se implementa acá.
- La vista SQL `v4_report.workorder_list`, el repositorio (`repository.go`) y las migraciones — **sin
  cambios** (la semántica activo/campo ya está en `ResolveProjectIDs`). De los use cases **solo** se toca la
  validación divergente de `ListWorkOrderFilterRows` (arriba); el resto queda igual.
- Cualquier concepto de "campaña vigente/actual" — **explícitamente descartado**.

---

## 4. Migraciones

**Ninguna.** No hay cambios de schema. "Activo" = `deleted_at IS NULL` ya está implementado vía soft-delete
de GORM (`internal/shared/models/base.go`) y aplicado en `ResolveProjectIDs`.

---

## 5. Dependencias

- **Intra-repo:** los helpers `internal/shared/filters` e `internal/shared/handlers` ya están presentes y en
  uso. No hay features bloqueantes.
- **Cross-repo (FE) — RESUELTO (ver §10):** este cambio hace obligatorio enviar `customer_id` + `project_id` +
  `campaign_id` en los 4 endpoints. Se actualizó el **BFF y la UI del repo `web`** (rama
  `work-orders-workspace-filter`) para enviar siempre los tres + `field_id` cuando aplica. La UI ya tenía los
  tres en su workspace global; el bug real estaba en el **BFF**, que con un campo seleccionado **dropeaba**
  `project_id` (contrato viejo "field OR project"). Verificado e2e (§10).
- **Plataforma:** usa `domainerr` de `platform/errors/go/domainerr` (ya importado por los helpers). Sin
  bumps de `go.mod`.

---

## 6. Plan de implementación (ejecutado — ver §10)

1. Refactorizar `parseFilters` → `parseWorkOrderFilter(c) (domain.WorkOrderFilter, error)`:
   - propagar el error de `ParseWorkspaceFilter` (no tragarlo);
   - conservar el mapeo a `domain.WorkOrderFilter` y el parseo de `supply_id` / `is_digital` / `status`.
2. En `ListWorkOrders`, `ListWorkOrderFilterRows`, `GetMetrics`, `ExportWorkOrders`:
   - parsear el workspace filter → si error, `RespondError`;
   - `ValidateRequiredWorkspaceFilter(workspaceFilter)` → si error, `RespondError`;
   - recién entonces delegar al use case.
3. Eliminar la validación divergente de `ListWorkOrderFilterRows` (`usecases.go:246`, `project_id` OR
   `field_id`): la regla queda unificada en el handler, sin una segunda copia más débil.
4. No duplicar la validación de campo: `field_id` opcional y la pertenencia campo↔proyecto ya quedan
   cubiertos por `ResolveProjectIDs` (vía `EXISTS ... fields ... deleted_at IS NULL`).
5. Mantener intactos el orden (`date desc, sequence_day desc, id desc`) y la paginación.

> Nota de diseño: la validación se hace **en el handler** (consistente con `dashboard`). `report` además la
> repite en el use case mediante un validator dedicado; acá no es necesario duplicarla.

---

## 7. Validación

**Build & test:**

- `go build ./...`
- `go vet ./...`
- `go test ./internal/work-order/...`

**Smoke (aplicar a cada endpoint: `list`, `filter-rows`, `metrics`, `export`):**

- **Sin filtros** → `400` con `"customer_id, project_id and campaign_id are required"`.
- **Filtro mal formado** (ej. `?customer_id=abc`) → `400` (ya no devuelve lista global).
- **Con `customer_id` + `project_id` + `campaign_id`** → `200`.
- **Con `field_id`** → solo las órdenes de ese campo.
- **Sin `field_id`** → todas las órdenes de todos los campos del proyecto.
- **Con proyecto archivado / `field_id` que no pertenece al proyecto / filtros incoherentes** → **`400`**
  con `"project_id does not match provided filters"`. Como `project_id` ahora es **obligatorio**,
  `ResolveProjectIDs` entra siempre por la rama "project_id explícito": cuenta los proyectos que matchean
  (`p.deleted_at IS NULL`, `customer_id`/`campaign_id` coincidentes, `EXISTS` del campo activo) y si da `0`
  devuelve un error de validación (`workspace.go:68`), **no una lista vacía**. La lista vacía solo ocurriría
  por la rama sin `project_id` (no alcanzable en este flujo). Ver §8.
- **`filter-rows`** ya **no** acepta su antigua regla `project_id` OR `field_id`: aplica el mismo `400` que
  el resto de los endpoints.

**Resultado (verificado el 2026-06-02 contra el stack local docker):**

- `go build ./...`, `go vet`, `go test ./internal/work-order/...` → **verde** (incluye 2 tests nuevos:
  `parseFilters` exige el workspace; `filter-rows` delega sin validación propia).
- e2e contra **core** (`:8080`) con IDs reales (project 30 / customer 17 / campaign 2 / field SJDD 39):
  sin filtros → `400`; solo project → `400`; los 3 + field → **`200`** (filtra el campo); los 3 sin field →
  **`200`** (todos los campos).
- e2e a través del **BFF** (`:3000`) como la UI, con campo seleccionado: list / metrics / filter-rows /
  export → **`200`**; solo project (sin cliente/campaña) → `400`.

---

## 8. Riesgos y decisiones pendientes

- **Breaking change de contrato (RESUELTO):** quien liste sin los tres IDs recibe `400`. El FE (BFF + UI de
  work-orders) se actualizó para enviarlos siempre; verificado e2e (§7, §10).
- **`filter-rows` (RESUELTO):** la UI de órdenes manda los tres (su workspace global ya los tiene). El acceso
  desde **Stock** ("ver órdenes que consumen este insumo") **no se rompe**: Stock solo muestra datos con
  cliente+proyecto+campaña ya elegidos (`web · ui/src/pages/admin/stock/Stock.tsx:666`) y esa selección
  persiste al navegar. Único caso teórico de `400`: abrir `/admin/work-orders?project_id=N` por URL pelada sin
  workspace — no es un flujo de la app (ver follow-ups en §10).
- **Capa de validación:** se elige el **handler** (espejo de `dashboard`). `report` además valida en el use
  case con un validator; acá se considera innecesario duplicarlo. Decisión menor, documentada.
- **"Activo" transitivo:** `ResolveProjectIDs` chequea `deleted_at IS NULL` en `projects` y `fields`. **No**
  hay chequeo explícito de `customers.deleted_at` ni `campaigns.deleted_at`. Esto es aceptado porque:
  - la campaña no tiene lifecycle (es solo una etiqueta del proyecto);
  - un cliente archivado no puede tener proyectos activos (regla de archivado de customer), por lo que un
    `project_id` de un cliente archivado no matchea → `400` (ver abajo).

  **Forma del rechazo:** como `project_id` es obligatorio y explícito, cuando los filtros no resuelven a un
  proyecto activo coherente (proyecto archivado, `field_id` ajeno, customer/campaign que no matchean),
  `ResolveProjectIDs` devuelve `domainerr.Validation("project_id does not match provided filters")` → **`400`**,
  no una lista vacía (la rama de lista vacía requiere `project_id` ausente, no alcanzable en este flujo). Si en
  el futuro se prefiriera devolver lista vacía en estos casos, sería un cambio adicional fuera de este spec.

---

## 9. Plan de convergencia gradual (deuda transversal — fuera de alcance de esta entrega)

> **Esta sección NO se implementa en este spec.** Es la hoja de ruta para que la app **converja sola, con el
> tiempo**, a una única base de filtros: registra dónde otros consumidores **duplicaron** el módulo en vez de
> reutilizarlo (ver Principio rector, §1) y define **cómo se corrige gradualmente** —no en un refactor único—.
> Hallazgos verificados por auditoría (con cita `file:line`).

**Lo que está bien — config distinta sobre la MISMA base (NO es deuda):** `dashboard` varía las columnas por
query vía `WorkspaceFilterColumns`; `work-order` agrega filtros extra (`supply_id` / `is_digital` / `status`).
Eso es exactamente lo que el patrón permite y debe seguir así.

**Lo que es deuda — módulo duplicado (debería llamar al compartido):**

| Módulo | Qué duplicó | Por qué importa |
|---|---|---|
| `dashboard` | Reimplementa su propio resolutor de proyectos ([`repository.go:134-142`](../../../internal/dashboard/repository.go)) | **Divergencia real de comportamiento**: con `project_id` no chequea `deleted_at` ni coherencia → acepta proyectos **archivados/incoherentes** que el canónico rechazaría |
| `lot` | Con `field_id` filtra solo por campo ([`repository.go:455-457`](../../../internal/lot/repository.go)) e ignora cliente/campaña; `GetMetrics` parsea a mano ([`handler.go:234-261`](../../../internal/lot/handler.go)) | **Divergencia real** (campo "solo uno" se desvía) + parser propio |
| `report` | `summary-results` parsea por gin form-binding en vez de `ParseWorkspaceFilter` ([`handler.go:150`](../../../internal/report/handler.go)); regla de requerido **triplicada**, incl. un validator con `fmt.Errorf` en vez de `domainerr` ([`usecases/validators.go:20-25`](../../../internal/report/usecases/validators.go)) | Inconsistencia intra-dominio + tipo de error distinto |
| `labor` | list/export parsean `field_id`/`project_id` a mano e ignoran cliente/campaña ([`handler.go:308,345`](../../../internal/labor/handler.go)) | Parser propio + regla ad-hoc |
| `data-integrity` | `CheckCostsCoherence` parsea solo `project_id` a mano ([`handler.go:81`](../../../internal/data-integrity/handler.go)) | Parser propio |

Además, casi todos los dominios **redeclaran** los 4 campos en un struct propio (`ReportFilter`, `SupplyFilter`,
`LaborFilter`, `DashboardFilter`, `LotListFilter`, …) en lugar de **embeber** `WorkspaceFilter`. No es urgente
(la forma es idéntica), pero es el mismo principio: convendría reusar el tipo, no copiarlo.

### Cómo se corrige: gradualmente, a medida que se aplican specs

**Regla de convergencia (boy-scout):** toda spec o feature futura que **toque** uno de los módulos de la
tabla **debe, en la misma entrega, migrarlo a la base compartida** (`ParseWorkspaceFilter` +
`ValidateRequiredWorkspaceFilter` + `ResolveProjectIDs`) y borrar su copia. **Prohibido agregar nuevas
copias** del módulo de filtros. Así la deuda baja sola con el tiempo, sin un refactor big-bang ni una rama
larga.

**Backlog priorizado (se va tildando con cada entrega):**

- [x] `work-order` — **este spec** (primer paso aplicado).
- [ ] `dashboard` — *prioridad alta*: divergencia **real** de comportamiento (acepta proyectos archivados).
- [ ] `lot` — *prioridad alta*: divergencia **real** (campo "solo uno" ignora cliente/campaña).
- [ ] `report` (`summary-results`) — *media*: parser propio + regla de requerido triplicada.
- [ ] `labor` (list/export) — *media*: parser propio + regla ad-hoc.
- [ ] `data-integrity` (`CheckCostsCoherence`) — *baja*.
- [ ] Structs de filtro: **embeber** `WorkspaceFilter` en vez de redeclarar los 4 campos — *oportunista*.

**Definición de "convergido" (por módulo):** parsea con `ParseWorkspaceFilter`, exige el mínimo con
`ValidateRequiredWorkspaceFilter` y resuelve con `ResolveProjectIDs`; **cero** parser propio, **cero** copia
de la regla de requerido, **cero** resolutor propio; `go build` / `go test` en verde. Se prioriza primero
`dashboard` y `lot` porque cambian el comportamiento observable, no solo el estilo.

---

## 10. Implementación realizada y verificación (estado real)

> Este spec se escribió como **plan previo (BE)**. Esta sección registra lo que **efectivamente se implementó
> y verificó**, incluida la parte FE que el plan había dejado como "a confirmar".

### Core (`ponti-backend`, rama `work-orders-workspace-filter`)

- `internal/work-order/handler.go`: `parseFilters` ahora devuelve `(filter, error)` y propaga el error de
  parseo; los 4 read handlers llaman `ValidateRequiredWorkspaceFilter` antes de delegar.
- `internal/work-order/usecases.go`: eliminada la validación divergente (`project_id` OR `field_id`) de
  `ListWorkOrderFilterRows`.
- `internal/work-order/usecases_test.go`: tests nuevos.
- Commits: `docs(work-orders): spec…` + `fix(work-orders): exigir cliente+proyecto+campaña…`.

### FE (`web`, rama `work-orders-workspace-filter`) — acá estaba el bug que se reportó

- **Causa raíz:** el BFF (`api/src/utils/workOrdersRoute.ts` → `buildWorkOrderScopeParams`) usaba `field_id`
  **O** `project_id` (excluyente, contrato viejo). Con un campo seleccionado **dropeaba `project_id`** → core
  respondía `400`. Lo mismo en la copia inline de `/metrics` y en `/export` (solo `project_id`).
- **Fix:** enviar `project_id` **y** `field_id` juntos; `/metrics` y `/export` convergen a los helpers de
  scope compartidos; la UI (`ui/.../WorkOrders.tsx` → `handleExport`) exporta con la query completa. Test del
  BFF actualizado (`api/test/workOrdersRoute.test.js`).
- Commit: `fix(work-orders): el BFF/UI envían project_id y field_id juntos…`.

### Verificación e2e (2026-06-02, stack local docker)

Elegir un campo **filtra** (200) y el error solo aparece sin cliente/proyecto/campaña — exactamente el
comportamiento pedido. Detalle de casos en §7.

### Follow-ups detectados en el code review (NO incluidos — fuera de alcance de esta entrega)

- **Guards del FE laxos:** `hasWorkOrderScope` (UI y BFF) sigue con la regla vieja `project OR field`. En el
  uso normal no molesta (la UI ya tiene los tres en su workspace), pero un acceso con solo `project` (URL
  pelada, o un futuro caller) muestra el `400` crudo en inglés en vez de un gate claro. Pendiente: alinear a
  cliente+proyecto+campaña y localizar el mensaje.
- **Acceso desde Stock** ("ver órdenes que consumen este insumo"): **no se rompe** hoy (Stock exige el
  workspace completo para mostrar datos), pero su deep-link solo lleva `project_id`+`supply_id`; conviene que
  incluya cliente+campaña por robustez. **Pertenece al módulo Stock, fuera de este spec.**
- **`errorMetrics`** oculta el bloque de KPIs sin aviso si metrics falla mientras la lista está cacheada.
- **`handleExport`** chequea `projectId` crudo vs `effectiveProjectId` (el botón export puede quedar
  deshabilitado con la lista ya cargada).
- **`GetMetrics` (core)** ignora `is_digital`/`status` que el BFF ahora reenvía (KPIs vs lista podrían diferir
  si se usan esos filtros).
- **Defensa en profundidad (core):** la validación vive solo en el handler; el use case ya no valida scope
  (no alcanzable hoy, pero un futuro caller no-HTTP podría listar sin scope).
- **Seguridad (no relacionado):** `verifyToken` del BFF (`api/src/routes/authMiddleware.ts`) no valida la
  firma del JWT, solo lo decodifica.
