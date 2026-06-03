# Informe técnico: KPIs de Órdenes de Trabajo — reflejar órdenes digitales y respetar `is_digital`/`status`

> **Estado:** Draft — **NADA implementado** · **Fecha:** 2026-06-03

> **Alcance de este documento:** auditoría + plan implementable (Opción B, **con migración**). **No se
> implementa nada**: no se crean migraciones, no se modifica código Go ni FE. La implementación queda
> pendiente de aprobación, con su propia ejecución por fases.

---

## Context

La pantalla de Órdenes de Trabajo muestra dos cosas alimentadas por endpoints distintos:

- **La lista / tabla** → `GET /work-orders` (y `GET /work-orders/filter-rows` para el dataset completo).
- **Las tarjetas de KPIs** (superficie, litros, kilos, costo directo, cantidad) → `GET /work-orders/metrics`.

La **lista incluye las órdenes digitales / borradores abiertos** (las que vienen del flujo de la app
móvil) y aplica los filtros `is_digital` y `status`. Las **métricas no**: se calculan desde una vista
que solo mira `public.workorders` (órdenes publicadas), **no incluye los borradores digitales** y **no
expone** las columnas `is_digital`/`status`, por lo que no puede filtrarlas. Resultado: **las tarjetas no
reflejan las órdenes de la app** y pueden divergir de la lista.

Hoy el problema es **latente** en el flujo de UI normal porque la UI filtra `status`/`is_digital` del
lado del cliente y recalcula las tarjetas en el navegador; pero el **BFF ya está cableado para reenviar
esos filtros a `/metrics`**, así que cualquier caller directo (o mover el filtro al servidor) lo destapa.
El spec del feature original ya lo dejó anotado como follow-up bajo la restricción "sin migraciones"
(ver `docs/specs/features/work-orders-workspace-filter.md` §10).

**Objetivo de este plan (Opción B):** que los KPIs **reflejen el mismo universo que la lista** — incluir
los borradores digitales abiertos y respetar `is_digital`/`status` — **vía migración**, sin romper los
otros consumidores de la vista de métricas.

---

## Hallazgo clave que condiciona el diseño (blast radius)

La vista de métricas `v4_report.workorder_metrics` (passthrough de `v4_calc.workorder_metrics`, definida
en `migrations_v4/000196_allow_archived_labors_supplies_in_views.up.sql`) **no la consume solo este
endpoint**. También la usa **`internal/data-integrity/usecases.go`** para los recálculos de integridad:

- `RecalcA`: `dashboard.DirectCostsExecutedUSD = ∑(workorder_metrics.direct_cost_usd) = ∑(wo RAW cost)`
  (`usecases.go:348,364,369,371`).
- `RecalcB`: `workorder_metrics.direct_cost_usd + aportes(...)` (`usecases.go:445,484-486,499-501`).

Estos chequeos asumen **solo costo ejecutado = órdenes publicadas**. Si se agregaran los borradores
digitales a `v4_calc.workorder_metrics`, `∑(direct_cost_usd)` crecería con costos **no ejecutados** y
**romperían** la reconciliación de integridad y el dashboard.

→ **Decisión de diseño: NO tocar `v4_calc.workorder_metrics` ni su passthrough.** Se crea una **vista
dedicada** para las KPIs de la pantalla de órdenes (`v4_calc.workorder_screen_metrics`), aislando el
cambio de los consumidores de costo ejecutado.

---

## A. Estado actual (verificado)

### A.1 Backend — la lista SÍ filtra; las métricas NO
| Componente | Archivo:línea | Qué hace |
|---|---|---|
| `workOrderListBaseQuery` | `internal/work-order/repository.go:432-498` | Aplica `is_digital = ?` (458-459), `status = ?` (461-462); rama `SupplyID` (464-494) que cruza `work_order_draft_items` e IDs negativos. Lee de `v4_report.workorder_list`. |
| `GetMetrics` (repo) | `internal/work-order/repository.go:531-612` | Lee `shareddb.ReportView("workorder_metrics")` = `v4_report.workorder_metrics`. **No** usa `filt.IsDigital`/`filt.Status`. Filtra solo `project_id`/`field_id`. |
| `orders_count` | `internal/work-order/repository.go:587-603` | `COUNT(DISTINCT split_part(number,'.',1))` **solo** de `public.workorders`. No cuenta borradores. |
| `getSupplyFilteredMetrics` | `internal/work-order/repository.go:614-662` | Con `supply_id`: lee crudo `workorders`+`workorder_items` (**sin** drafts). No usa is_digital/status. |
| `GetMetrics` (handler) | `internal/work-order/handler.go:274-300` | Parsea **solo** `supply_id` (289-292) + workspace. **No** parsea `is_digital`/`status`. |
| `parseFilters` (lista) | `internal/work-order/handler.go:233-271` | Sí parsea `is_digital`/`status` (usado por el handler de la lista). |
| `WorkOrderFilter` | `internal/work-order/usecases/domain/work_order.go:55-64` | Ya tiene `IsDigital *bool`, `Status *string`, `SupplyID *int64`. |

### A.2 Base de datos (fuente autoritativa = `migrations_v4/`, NO `scripts/db/schema.expected.sql`, que está desactualizado para la list-view)
- **Vista de métricas vigente** — `migrations_v4/000196_*.up.sql:181-240`:
  - CTE `base`: `FROM public.workorders w JOIN public.labors lb` con `deleted_at IS NULL AND effective_area > 0`.
  - `surface` = `SUM(effective_area)`, `labor_costs` = `SUM(labor_price * effective_area)`,
    `supply_metrics` = `liters`/`kilograms` vía `unit_id=1/2` sobre `final_dose*effective_area` y
    `supplies_cost_usd` = `SUM(total_used*price)`. Todas agregadas por `project_id, field_id, lot_id`.
  - Final: `FULL JOIN ... USING (project_id, field_id, lot_id)`. **No** expone `is_digital`/`status`; **no** incluye drafts.
  - `v4_report.workorder_metrics` = passthrough `SELECT project_id, field_id, lot_id, surface_ha, liters, kilograms, labor_cost_usd, supplies_cost_usd, direct_cost_usd, ... FROM v4_calc.workorder_metrics`.
- **Vista de lista vigente** (referencia de cómo incluir drafts) — `migrations_v4/000230_workorders_is_digital_origin.up.sql`:
  - `UNION ALL` de 4 ramas: (1) insumos de publicadas, (2) labor de publicadas, (3) insumos de borradores
    digitales abiertos, (4) labor de borradores. Las ramas de publicadas exponen
    `is_digital = COALESCE(w.is_digital,false)`, `status = 'published'`; las de drafts `is_digital = true`,
    `status = wod.status`, filtradas a `wod.is_digital = true AND wod.status = 'draft'`, con `id = -wod.id`.
- **Tablas de borradores** (`migrations_v4/000205_*`, `is_digital` agregada en `000207_*`):
  - `public.work_order_drafts`: `id, number, customer_id, project_id, campaign_id, field_id, lot_id,
    crop_id, labor_id, effective_area, investor_id, status (default 'draft'), is_digital,
    published_work_order_id, deleted_at`. Índices `idx_work_order_drafts_project_id`, `idx_work_order_drafts_status`.
  - `public.work_order_draft_items`: `id, draft_id, supply_id, total_used, final_dose, deleted_at`.
  - `supplies.unit_id` (1 = litros, 2 = kilos) y `supplies.price` aplican igual que para publicadas.
- **Migración 000230** ya agregó `public.workorders.is_digital boolean NOT NULL DEFAULT false` (true para
  publicadas que provienen de un borrador digital). `status` **no** existe como columna física: es un
  literal de vista (`'published'` para publicadas, `wod.status` para drafts).

### A.3 Consumidores de la vista de métricas (blast radius)
| Consumidor | Uso | ¿Tolera drafts? |
|---|---|---|
| `GET /work-orders/metrics` (este endpoint) | KPIs de la pantalla | **Debe** incluir drafts (objetivo) |
| `internal/data-integrity/usecases.go` (RecalcA/RecalcB) | reconciliación de costo ejecutado vs dashboard vs RAW | **NO** — solo publicadas |

### A.4 BFF / FE
- BFF: `web/api/src/utils/workOrdersRoute.ts:50-84` `buildWorkOrderScopeParams` reenvía `is_digital`/`status`
  si vienen; usado por las 4 rutas, incluida `/work-orders/metrics` (`web/api/src/routes/workorders.ts`).
- UI: `web/ui/src/pages/admin/workorders/WorkOrders.tsx` — `workOrdersBaseQuery` (624-648) NO envía
  `is_digital`/`status`; filtra client-side; `derivedMetrics` (1065-1087) recalcula KPIs en cliente con un
  **over-count de `surface_ha`** (suma cada fila de la list-view, que repite `surface_ha` por orden);
  `displayedMetrics` (1094) alterna entre métricas BE y `derivedMetrics` según haya filtros de columna.

---

## B. Estado objetivo

```
GET /work-orders/metrics?scope&[is_digital]&[status]
        │
        ▼
  v4_report.workorder_screen_metrics            (NUEVA — passthrough)
        │
        ▼
  v4_calc.workorder_screen_metrics              (NUEVA — publicadas ∪ borradores digitales abiertos,
        │                                          agregada por project/field/lot/is_digital/status)
        ▼
  GetMetrics aplica WHERE is_digital=? / status=?  → KPIs == universo de la lista

  v4_calc.workorder_metrics (SIN CAMBIOS) ──▶ data-integrity / dashboard (costo ejecutado, solo publicadas)
```

Semántica objetivo (**a documentar y aceptar explícitamente**): por defecto (sin `is_digital`/`status`),
las KPIs pasan a **incluir los borradores digitales abiertos**, igual que la lista. Esto **cambia los
números actuales** de las tarjetas en la vista por defecto (hoy excluyen drafts).

---

## C. Cambios propuestos (por fases)

### Fase 1 — Migración: vista dedicada con drafts + dimensiones de filtro
Archivos: `migrations_v4/000232_workorder_screen_metrics_digital_and_status.up.sql` y `.down.sql`
(siguiente número libre = **000232**; el último es 000231).

`up.sql` (boceto, no implementación):
```sql
BEGIN;

CREATE OR REPLACE VIEW v4_calc.workorder_screen_metrics AS
WITH base AS (
  -- (a) Publicadas / manuales
  SELECT w.id AS source_id, w.number,
         w.project_id, w.field_id, w.lot_id,
         w.effective_area, lb.price AS labor_price,
         COALESCE(w.is_digital, false) AS is_digital,
         'published'::varchar(30) AS status,
         false AS is_draft
  FROM public.workorders w
  JOIN public.labors lb ON lb.id = w.labor_id
  WHERE w.deleted_at IS NULL AND w.effective_area IS NOT NULL AND w.effective_area > 0

  UNION ALL

  -- (b) Borradores digitales abiertos
  SELECT wod.id AS source_id, wod.number,
         wod.project_id, wod.field_id, wod.lot_id,
         wod.effective_area, lb.price AS labor_price,
         true AS is_digital,
         wod.status::varchar(30) AS status,
         true AS is_draft
  FROM public.work_order_drafts wod
  JOIN public.labors lb ON lb.id = wod.labor_id
  WHERE wod.deleted_at IS NULL AND wod.is_digital = true AND wod.status = 'draft'
    AND wod.effective_area IS NOT NULL AND wod.effective_area > 0
),
surface AS (
  SELECT project_id, field_id, lot_id, is_digital, status,
         SUM(effective_area)::numeric AS surface_ha
  FROM base GROUP BY project_id, field_id, lot_id, is_digital, status
),
labor_costs AS (
  SELECT project_id, field_id, lot_id, is_digital, status,
         SUM(labor_price * effective_area)::numeric AS labor_cost_usd
  FROM base GROUP BY project_id, field_id, lot_id, is_digital, status
),
supply_metrics AS (
  -- items de publicadas y de borradores, unificados, cruzando supplies para unit_id/price
  SELECT b.project_id, b.field_id, b.lot_id, b.is_digital, b.status,
         SUM(CASE WHEN s.unit_id = 1 THEN it.final_dose * b.effective_area ELSE 0 END)::numeric AS liters,
         SUM(CASE WHEN s.unit_id = 2 THEN it.final_dose * b.effective_area ELSE 0 END)::numeric AS kilograms,
         SUM(COALESCE(it.total_used,0) * COALESCE(s.price,0))::numeric AS supplies_cost_usd
  FROM base b
  LEFT JOIN LATERAL (
      SELECT wi.supply_id, wi.final_dose, wi.total_used
      FROM public.workorder_items wi
      WHERE b.is_draft = false AND wi.workorder_id = b.source_id AND wi.deleted_at IS NULL
      UNION ALL
      SELECT wodi.supply_id, wodi.final_dose, wodi.total_used
      FROM public.work_order_draft_items wodi
      WHERE b.is_draft = true AND wodi.draft_id = b.source_id AND wodi.deleted_at IS NULL
  ) it ON true
  LEFT JOIN public.supplies s ON s.id = it.supply_id
  GROUP BY b.project_id, b.field_id, b.lot_id, b.is_digital, b.status
)
SELECT
  COALESCE(sur.project_id, lc.project_id, sm.project_id) AS project_id,
  COALESCE(sur.field_id,  lc.field_id,  sm.field_id)  AS field_id,
  COALESCE(sur.lot_id,    lc.lot_id,    sm.lot_id)    AS lot_id,
  COALESCE(sur.is_digital, lc.is_digital, sm.is_digital) AS is_digital,
  COALESCE(sur.status,     lc.status,     sm.status)     AS status,
  COALESCE(sur.surface_ha, 0)::numeric AS surface_ha,
  COALESCE(sm.liters, 0)::numeric AS liters,
  COALESCE(sm.kilograms, 0)::numeric AS kilograms,
  COALESCE(lc.labor_cost_usd, 0)::numeric AS labor_cost_usd,
  COALESCE(sm.supplies_cost_usd, 0)::numeric AS supplies_cost_usd,
  (COALESCE(lc.labor_cost_usd,0) + COALESCE(sm.supplies_cost_usd,0))::numeric AS direct_cost_usd
FROM surface sur
FULL JOIN labor_costs   lc USING (project_id, field_id, lot_id, is_digital, status)
FULL JOIN supply_metrics sm USING (project_id, field_id, lot_id, is_digital, status);

CREATE OR REPLACE VIEW v4_report.workorder_screen_metrics AS
  SELECT project_id, field_id, lot_id, is_digital, status,
         surface_ha, liters, kilograms, labor_cost_usd, supplies_cost_usd, direct_cost_usd
  FROM v4_calc.workorder_screen_metrics;

COMMIT;
```
`down.sql`: `DROP VIEW IF EXISTS v4_report.workorder_screen_metrics; DROP VIEW IF EXISTS v4_calc.workorder_screen_metrics;`
**No** toca `v4_calc.workorder_metrics` ni `v4_report.workorder_metrics`.

> Notas de diseño:
> - El `FULL JOIN ... USING (..., is_digital, status)` debe alinear las mismas combinaciones de claves en
>   las 3 CTEs; el `COALESCE` en el SELECT final cubre lotes presentes en una CTE y no en otra (igual que hoy).
> - El `LATERAL` unifica items de publicadas y de borradores sin duplicar la rama base; alternativa: dos
>   `supply_metrics` separadas unidas por `UNION ALL` antes de agregar. Validar plan de ejecución.
> - Confirmar nombre final de la vista (`workorder_screen_metrics`) con la convención del equipo.

### Fase 2 — Backend Go
- `repository.go GetMetrics`: cambiar la fuente a `shareddb.ReportView("workorder_screen_metrics")` y
  añadir al WHERE dinámico `AND is_digital = ?` y `AND status = ?` cuando `filt.IsDigital`/`filt.Status`
  estén presentes (sin filtro → suma todo: publicadas + drafts).
- `orders_count`: reescribir para contar `DISTINCT` número base sobre **publicadas ∪ borradores abiertos**
  respetando los mismos filtros — p. ej. una subconsulta `UNION ALL` de `public.workorders` (status
  'published') y `public.work_order_drafts` (is_digital, status) y luego
  `COUNT(DISTINCT split_part(number,'.',1))`. (Alternativa: incorporar el conteo a la vista.)
- `getSupplyFilteredMetrics`: alinear con la nueva semántica: con `supply_id`, considerar también
  `work_order_draft_items` (como ya hace la rama supply de la list-view) y respetar `is_digital`/`status`.
  Evaluar unificar este camino con la nueva vista para no mantener dos cálculos.
- `handler.go GetMetrics`: parsear `is_digital`/`status`. **DRY recomendado:** extraer el parseo de filtros
  que ya hace `parseFilters` (233-271) a un helper reutilizable por el handler de lista y el de métricas,
  para que ambos interpreten los params igual.
- `WorkOrderFilter` no cambia (ya tiene los campos).

### Fase 3 — BFF / FE (consistencia)
- BFF: ya reenvía `is_digital`/`status` a `/metrics`; queda correcto en cuanto el BE los respeta. Agregar/ajustar
  test en `web/api/test/workOrdersRoute.test.js` para fijar que `/metrics` los reenvía.
- FE (dos sub-opciones, a decidir al implementar):
  - **C1 (server-driven):** la UI **envía** `is_digital`/`status` al backend y muestra directamente la
    respuesta de `/metrics`; se retira/reduce `derivedMetrics`/`displayedMetrics`. KPIs y lista siempre
    consistentes desde una sola fuente (BE).
  - **C2 (mantener client-side):** conservar `derivedMetrics` pero **corregir el over-count de surface**
    (sumar `surface_ha` una vez por número base de orden). Mantiene la matemática en dos lugares.
  - Recomendación: **C1**, ya que el BE pasa a ser la fuente correcta única.

---

## D. Gap Analysis
| # | Brecha | Estado actual | Acción objetivo | Esfuerzo |
|---|--------|---------------|-----------------|----------|
| G1 | Métricas no incluyen borradores digitales | vista solo `public.workorders` | nueva vista con UNION de drafts | medio |
| G2 | Métricas no exponen/filtran `is_digital`/`status` | columnas inexistentes en la vista | dimensiones nuevas + WHERE en GetMetrics | medio |
| G3 | `orders_count` solo cuenta publicadas | query a `public.workorders` | UNION publicadas ∪ drafts con filtros | bajo-medio |
| G4 | Handler no parsea `is_digital`/`status` | solo `supply_id` | reutilizar `parseFilters` | bajo |
| G5 | `getSupplyFilteredMetrics` sin drafts | crudo workorders+items | considerar `work_order_draft_items` | medio |
| G6 | FE recalcula y sobre-cuenta surface | `derivedMetrics` 1065-1094 | C1 (server-driven) o C2 (fix surface) | bajo |
| G7 | Riesgo de romper data-integrity/dashboard | vista compartida | **vista dedicada** (no tocar la SSOT) | diseño |

---

## E. Riesgos
| ID | Riesgo | Prob. | Impacto | Mitigación |
|----|--------|-------|---------|------------|
| R1 | Romper reconciliación de `data-integrity` / dashboard | — | alto | **No tocar** `v4_calc.workorder_metrics`; usar vista dedicada. Test de regresión que confirme `∑(direct_cost_usd)` sin cambios. |
| R2 | Doble conteo / fan-out del `FULL JOIN` con claves extra | media | medio | Alinear claves en las 3 CTEs; tests de paridad contra la lista; revisar combinaciones NULL. |
| R3 | Cambio de números por defecto de las tarjetas (ahora incluyen drafts) | alta | medio (UX/expectativa) | **Decisión de producto explícita**: KPIs reflejan lo mismo que la lista. Documentar y comunicar. |
| R4 | `orders_count` mal contado con drafts/filtros | media | medio | Camino UNION dedicado + test `?status=draft`/`?is_digital=true`. |
| R5 | Performance de la vista (UNION + LATERAL + JOINs) | media | medio | Reusar índices `idx_work_order_drafts_*`; revisar `EXPLAIN`; agregar índices si hace falta. |
| R6 | Divergencia de unidades (litros/kilos) entre publicadas y drafts | baja | bajo | Misma fórmula `unit_id`/`final_dose*effective_area` para ambas ramas. |
| R7 | `schema.expected.sql` quede más desfasado | media | bajo | Regenerar el dump tras aplicar 000232 (ya está desfasado para la list-view). |

---

## F. Roadmap por fases
- **Fase 1 — Migración `000232`** (vista dedicada `*_screen_metrics`). Reversible con `down` (solo DROP de vistas nuevas).
- **Fase 2 — Backend Go** (`GetMetrics`, `orders_count`, `getSupplyFilteredMetrics`, parseo de filtros DRY).
- **Fase 3 — BFF/FE** (sub-opción C1 recomendada; ajustar tests del BFF).
- **Fase 4 — Validación** (paridad lista↔KPIs; regresión data-integrity; e2e; smoke).

---

## G. Definition of Done por fase
- **Fase 1 DoD:** `000232` aplica/revierte limpio; `v4_report.workorder_screen_metrics` devuelve filas por
  `project/field/lot/is_digital/status`; `v4_calc.workorder_metrics` **idéntica** (diff vacío).
- **Fase 2 DoD:** `GET /work-orders/metrics` respeta `is_digital`/`status`; `orders_count` cuenta publicadas+drafts;
  sin filtros, los totales coinciden con el agregado de `/work-orders/filter-rows`.
- **Fase 3 DoD:** la UI muestra KPIs consistentes con la tabla bajo cualquier filtro; sin over-count de surface.
- **Fase 4 DoD:** tests de paridad verdes; chequeos de `data-integrity` sin cambios; e2e con publicada + draft OK.

---

## H. Plan de rollback
| Escenario | Rollback |
|-----------|----------|
| La vista nueva da números mal | `down` de `000232` (DROP de vistas nuevas) + revertir `GetMetrics` a `ReportView("workorder_metrics")`. La SSOT de costo nunca se tocó. |
| Backend Go con bug | revertir el commit de Fase 2; la vista nueva puede quedar (no la consume nadie más). |
| FE inconsistente | revertir Fase 3; el BFF seguir reenviando params es inocuo. |
| data-integrity/dashboard afectados | no debería ocurrir (vista dedicada); si ocurre, confirmar que ningún consumidor migró por error a la vista nueva. |

---

## I. Verificación (cómo se probará end-to-end, en la futura implementación)
1. **Unit BE (paridad):** espejar el harness embebido de `internal/work-order/repository_list_test.go`
   (crea `v4_report.workorder_list` como tabla con `is_digital`/`status` e inserta el fixture); crear
   también `v4_report.workorder_screen_metrics` y assertear que, con el **mismo `WorkOrderFilter`**,
   `GetMetrics` == agregado deduplicado de `ListWorkOrderFilterRows` para `surface_ha`/`direct_cost`/
   `orders_count`, con y sin `?status=draft` y `?is_digital=true`.
2. **Regresión data-integrity:** ejecutar los recálculos de `internal/data-integrity` y confirmar que
   `∑(workorder_metrics.direct_cost_usd)` (vista vieja) **no cambió**.
3. **e2e (docker local):** sembrar un proyecto con 1 WO publicada multi-insumo + 1 borrador digital abierto;
   comparar `/work-orders` vs `/work-orders/metrics` con el scope completo y con `?status=draft` /
   `?is_digital=true` → tabla y tarjetas **coinciden**.
4. **Smoke:** `curl` a ambos endpoints con esos params; pre-fix la lista cambia y las métricas no
   (divergencia), post-fix coinciden.

---

> **Recordatorio:** este documento es solo el informe/plan. No se ha creado ninguna migración ni modificado
> código (Go/FE). La implementación (migración `000232`, backend, BFF/FE) queda pendiente de aprobación.
>
> **Relacionado:** `docs/specs/features/work-orders-workspace-filter.md` §10 (follow-up que originó este plan).
