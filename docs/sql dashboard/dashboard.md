#####################################################################
Metricas
#####################################################################

### 1) `dashboard_card_sowing.sql`

```sql
/* Avance de siembra (%)
   Hs sembradas (Hs_S) = suma de w.effective_area de órdenes con labor.category = 'Siembra'.
   Hs totales (Hs_T)    = projects.total_hectares.
   Porcentaje (P)       = (Hs_S / Hs_T) * 100   (si Hs_T = 0 => P = 0).
*/
CREATE OR REPLACE VIEW dashboard_card_sowing AS
WITH sowing_area_by_project AS (
  SELECT lb.project_id, SUM(w.effective_area)::numeric(14,2) AS sowed_area
  FROM labors lb
  JOIN workorders w
    ON w.labor_id = lb.id
   AND w.effective_area > 0
   AND w.deleted_at IS NULL
  WHERE lb.deleted_at IS NULL
    AND lb.category = 'Siembra'
  GROUP BY lb.project_id
),
levels AS (
  SELECT
    CASE WHEN GROUPING(p.customer_id)=1 THEN NULL ELSE p.customer_id END AS customer_id,
    CASE WHEN GROUPING(p.id)=1          THEN NULL ELSE p.id          END AS project_id,
    CASE WHEN GROUPING(p.campaign_id)=1 THEN NULL ELSE p.campaign_id END AS campaign_id,
    CASE WHEN GROUPING(f.id)=1          THEN NULL ELSE f.id          END AS field_id,
    COALESCE(p.total_hectares,0)::numeric(14,2) AS total_hectares
  FROM projects p
  LEFT JOIN fields f ON f.project_id = p.id AND f.deleted_at IS NULL
  WHERE p.deleted_at IS NULL
  GROUP BY GROUPING SETS (
    (p.customer_id, p.id, p.campaign_id, f.id, p.total_hectares),
    (p.customer_id, p.id, p.campaign_id,           p.total_hectares),
    (p.customer_id, p.id,                           p.total_hectares),
    (p.customer_id,                                  p.total_hectares),
    (                                                p.total_hectares)
  )
)
SELECT
  lvl.customer_id, lvl.project_id, lvl.campaign_id, lvl.field_id,
  COALESCE(SUM(sap.sowed_area),0)::numeric(14,2) AS sowing_hectares,
  lvl.total_hectares                              AS sowing_total_hectares,
  CASE WHEN lvl.total_hectares > 0
       THEN ROUND((COALESCE(SUM(sap.sowed_area),0) / lvl.total_hectares) * 100, 2)
       ELSE 0 END::numeric(14,2)                 AS sowing_progress_pct
FROM levels lvl
LEFT JOIN sowing_area_by_project sap ON sap.project_id = lvl.project_id
GROUP BY lvl.customer_id, lvl.project_id, lvl.campaign_id, lvl.field_id, lvl.total_hectares;
```

---

### 2) `dashboard_card_costs.sql`

```sql
/* Avance de costos (%)
   Ejecutado (E) = USD de órdenes (Labores ejecutadas + Insumos usados).
   Presupuestado (B) = (no existe en DB) → placeholder fijo (reemplazar cuando esté).
   Porcentaje (P) = (E / B) * 100  (si B = 0 => P = 0).
*/
CREATE OR REPLACE VIEW dashboard_card_costs AS
WITH params AS (
  -- ⚠️ Placeholder hasta que exista “presupuesto”: ajustá este valor cuando corresponda.
  SELECT 0.00::numeric(14,2) AS hardcoded_budget_usd
),
executed_labors_by_project AS (
  SELECT lb.project_id, SUM(lb.price)::numeric(14,2) AS labors_usd
  FROM labors lb
  WHERE lb.deleted_at IS NULL
    AND EXISTS (
      SELECT 1 FROM workorders w
      WHERE w.labor_id = lb.id AND w.effective_area > 0 AND w.deleted_at IS NULL
    )
  GROUP BY lb.project_id
),
used_supplies_by_project AS (
  SELECT sp.project_id, SUM(sp.price)::numeric(14,2) AS supplies_usd
  FROM supplies sp
  WHERE sp.deleted_at IS NULL
    AND EXISTS (
      SELECT 1 FROM workorder_items wi
      WHERE wi.supply_id = sp.id AND wi.final_dose > 0 AND wi.deleted_at IS NULL
    )
  GROUP BY sp.project_id
),
executed_costs AS (
  SELECT
    p.id AS project_id,
    COALESCE(el.labors_usd,0) + COALESCE(us.supplies_usd,0) AS executed_costs_usd
  FROM projects p
  LEFT JOIN executed_labors_by_project el ON el.project_id = p.id
  LEFT JOIN used_supplies_by_project  us ON us.project_id = p.id
  WHERE p.deleted_at IS NULL
)
SELECT
  p.customer_id, p.id AS project_id, p.campaign_id, NULL::bigint AS field_id,
  ec.executed_costs_usd,
  pr.hardcoded_budget_usd AS budget_cost_usd,
  CASE WHEN pr.hardcoded_budget_usd > 0
       THEN ROUND((ec.executed_costs_usd / pr.hardcoded_budget_usd) * 100, 2)
       ELSE 0 END::numeric(14,2) AS costs_progress_pct
FROM projects p
LEFT JOIN executed_costs ec ON ec.project_id = p.id
CROSS JOIN params pr
WHERE p.deleted_at IS NULL;
```

---

### 3) `dashboard_card_harvest.sql`

```sql
/* Avance de cosecha (%)
   Hs cosechadas (Hc) = suma de w.effective_area de órdenes con labor.category = 'Cosecha'.
   Hs totales (Ht)    = projects.total_hectares.
   Porcentaje (P)     = (Hc / Ht) * 100   (si Ht = 0 => P = 0).
*/
CREATE OR REPLACE VIEW dashboard_card_harvest AS
WITH harvest_area_by_project AS (
  SELECT lb.project_id, SUM(w.effective_area)::numeric(14,2) AS harvested_area
  FROM labors lb
  JOIN workorders w
    ON w.labor_id = lb.id
   AND w.effective_area > 0
   AND w.deleted_at IS NULL
  WHERE lb.deleted_at IS NULL
    AND lb.category = 'Cosecha'
  GROUP BY lb.project_id
),
levels AS (
  SELECT
    CASE WHEN GROUPING(p.customer_id)=1 THEN NULL ELSE p.customer_id END AS customer_id,
    CASE WHEN GROUPING(p.id)=1          THEN NULL ELSE p.id          END AS project_id,
    CASE WHEN GROUPING(p.campaign_id)=1 THEN NULL ELSE p.campaign_id END AS campaign_id,
    CASE WHEN GROUPING(f.id)=1          THEN NULL ELSE f.id          END AS field_id,
    COALESCE(p.total_hectares,0)::numeric(14,2) AS total_hectares
  FROM projects p
  LEFT JOIN fields f ON f.project_id = p.id AND f.deleted_at IS NULL
  WHERE p.deleted_at IS NULL
  GROUP BY GROUPING SETS (
    (p.customer_id, p.id, p.campaign_id, f.id, p.total_hectares),
    (p.customer_id, p.id, p.campaign_id,           p.total_hectares),
    (p.customer_id, p.id,                           p.total_hectares),
    (p.customer_id,                                  p.total_hectares),
    (                                                p.total_hectares)
  )
)
SELECT
  lvl.customer_id, lvl.project_id, lvl.campaign_id, lvl.field_id,
  COALESCE(SUM(hap.harvested_area),0)::numeric(14,2) AS harvest_hectares,
  lvl.total_hectares                                  AS harvest_total_hectares,
  CASE WHEN lvl.total_hectares > 0
       THEN ROUND((COALESCE(SUM(hap.harvested_area),0) / lvl.total_hectares) * 100, 2)
       ELSE 0 END::numeric(14,2)                     AS harvest_progress_pct
FROM levels lvl
LEFT JOIN harvest_area_by_project hap ON hap.project_id = lvl.project_id
GROUP BY lvl.customer_id, lvl.project_id, lvl.campaign_id, lvl.field_id, lvl.total_hectares;
```

---

### 4) `dashboard_card_contributions.sql`

```sql
/* Avance de aportes (%)
   Mostrar por inversor: Nombre + porcentaje teórico acordado (project_investors.percentage).
   (Card se nutre de este breakdown por inversor.)
*/
CREATE OR REPLACE VIEW dashboard_card_contributions AS
SELECT
  p.customer_id,
  p.id AS project_id,
  p.campaign_id,
  NULL::bigint AS field_id,
  pi.investor_id,
  COALESCE(i.name,'') AS investor_name,
  COALESCE(pi.percentage,0)::numeric(6,2) AS investor_percentage_pct
FROM projects p
JOIN project_investors pi ON pi.project_id = p.id AND pi.deleted_at IS NULL
LEFT JOIN investors i      ON i.id = pi.investor_id AND i.deleted_at IS NULL
WHERE p.deleted_at IS NULL;
```

---

### 5) `dashboard_card_operating_result.sql`

```sql
/* Resultado operativo (% y USD)
   Definiciones de la doc:
   A (Ingreso neto)                 = comercializaciones = SUM(lots.tons * 200) por proyecto.
   B (Total invertido - directos)   = Labores EJECUTADAS + Insumos UTILIZADOS (USD).
   C (Costo administrativo)         = projects.admin_cost.
   % Rentabilidad (P)               = (A / B) * 100         (si B = 0 => P = 0).
   Números grises: gris1 = A ; gris2 = (B + C).
*/
CREATE OR REPLACE VIEW dashboard_card_operating_result AS
WITH income_by_field AS (
  SELECT f.project_id, f.id AS field_id,
         COALESCE(SUM(l.tons * 200),0)::numeric(14,2) AS income_usd
  FROM fields f
  LEFT JOIN lots l ON l.field_id = f.id AND l.deleted_at IS NULL
  WHERE f.deleted_at IS NULL
  GROUP BY f.project_id, f.id
),
income_by_project AS (
  SELECT f.project_id, COALESCE(SUM(ibf.income_usd),0)::numeric(14,2) AS income_usd
  FROM fields f
  LEFT JOIN income_by_field ibf ON ibf.field_id = f.id
  WHERE f.deleted_at IS NULL
  GROUP BY f.project_id
),
executed_labors_by_project AS (
  SELECT lb.project_id, SUM(lb.price)::numeric(14,2) AS labors_usd
  FROM labors lb
  WHERE lb.deleted_at IS NULL
    AND EXISTS (
      SELECT 1 FROM workorders w
      WHERE w.labor_id = lb.id AND w.effective_area > 0 AND w.deleted_at IS NULL
    )
  GROUP BY lb.project_id
),
used_supplies_by_project AS (
  SELECT sp.project_id, SUM(sp.price)::numeric(14,2) AS supplies_usd
  FROM supplies sp
  WHERE sp.deleted_at IS NULL
    AND EXISTS (
      SELECT 1 FROM workorder_items wi
      WHERE wi.supply_id = sp.id AND wi.final_dose > 0 AND wi.deleted_at IS NULL
    )
  GROUP BY sp.project_id
),
direct_costs_by_project AS (
  SELECT
    p.id AS project_id,
    COALESCE(el.labors_usd,0)   AS direct_labors_usd,
    COALESCE(us.supplies_usd,0) AS direct_supplies_usd,
    (COALESCE(el.labors_usd,0) + COALESCE(us.supplies_usd,0))::numeric(14,2) AS total_direct_costs_usd
  FROM projects p
  LEFT JOIN executed_labors_by_project el ON el.project_id = p.id
  LEFT JOIN used_supplies_by_project  us ON us.project_id = p.id
  WHERE p.deleted_at IS NULL
)
SELECT
  p.customer_id, p.id AS project_id, p.campaign_id, NULL::bigint AS field_id,

  -- % rentabilidad = A / B * 100
  CASE WHEN dcp.total_direct_costs_usd > 0
       THEN ROUND((COALESCE(ibp.income_usd,0) / dcp.total_direct_costs_usd) * 100, 2)
       ELSE 0 END::numeric(14,2) AS operating_result_progress_pct,

  -- Números grises
  COALESCE(ibp.income_usd,0)::numeric(14,2)                                 AS operating_result_income_usd,       -- A
  (COALESCE(dcp.total_direct_costs_usd,0) + COALESCE(p.admin_cost,0))::numeric(14,2) AS operating_result_total_costs_usd, -- B + C

  -- Detalle
  COALESCE(dcp.direct_labors_usd,0)::numeric(14,2)   AS operating_result_direct_labors_usd,
  COALESCE(dcp.direct_supplies_usd,0)::numeric(14,2) AS operating_result_direct_supplies_usd,
  COALESCE(p.admin_cost,0)::numeric(14,2)            AS operating_result_admin_cost_usd

FROM projects p
LEFT JOIN income_by_project      ibp ON ibp.project_id = p.id
LEFT JOIN direct_costs_by_project dcp ON dcp.project_id = p.id
WHERE p.deleted_at IS NULL;
```

#####################################################################
Balance de Gestión
#####################################################################

```sql
-- =========================================================
-- management_balance_view  (ÚNICA VISTA)
-- =========================================================
/*
Balance de gestión — fórmulas simples por línea:

Seed
  E (executed) = Σ precio supplies categoría "Semilla" UTILIZADOS en OT
  I (invested) = Σ precio supplies categoría "Semilla" (comprados/cargados)
  S (stock)    = I - E

Supplies (no incluye Semilla)
  E = Σ precio supplies NO "Semilla" UTILIZADOS en OT
  I = Σ precio supplies NO "Semilla" (comprados/cargados)
  S = I - E

Labors
  E = Σ precio labors EJECUTADAS (OT con effective_area > 0)
  I = Σ precio labors (cargadas; usadas o por usar)
  S = 0   -- no aplica stock a labores

Rent
  E = % de arriendo * Ingresos   (por ahora 0 si no hay columna de %/monto)
  I = E
  S = 0

Structure
  E = admin_cost (projects)
  I = E
  S = 0
*/

CREATE OR REPLACE VIEW management_balance_view AS
WITH
-- -----------------------------------------------
-- Reglas de “ejecutado / utilizado”
--   • Labor EJECUTADA:        EXISTS workorders (w.labor_id = lb.id AND w.effective_area > 0)
--   • Insumo UTILIZADO:       EXISTS workorder_items (wi.supply_id = sp.id AND wi.final_dose > 0)
-- -----------------------------------------------
exec_lab AS (
  SELECT lb.project_id, SUM(lb.price) AS executed_labors_usd
  FROM labors lb
  WHERE lb.deleted_at IS NULL
    AND EXISTS (
      SELECT 1 FROM workorders w
      WHERE w.labor_id = lb.id
        AND w.effective_area > 0
        AND w.deleted_at IS NULL
    )
  GROUP BY lb.project_id
),
inv_lab AS (
  SELECT lb.project_id, SUM(lb.price) AS invested_labors_usd
  FROM labors lb
  WHERE lb.deleted_at IS NULL
  GROUP BY lb.project_id
),

-- Seed (categoría semilla)
used_seed AS (
  SELECT sp.project_id, SUM(sp.price) AS executed_seed_usd
  FROM supplies sp
  WHERE sp.deleted_at IS NULL
    AND (sp.category ILIKE 'semilla%' OR sp.category ILIKE 'seed%')
    AND EXISTS (
      SELECT 1 FROM workorder_items wi
      WHERE wi.supply_id = sp.id
        AND wi.final_dose > 0
        AND wi.deleted_at IS NULL
    )
  GROUP BY sp.project_id
),
inv_seed AS (
  SELECT sp.project_id, SUM(sp.price) AS invested_seed_usd
  FROM supplies sp
  WHERE sp.deleted_at IS NULL
    AND (sp.category ILIKE 'semilla%' OR sp.category ILIKE 'seed%')
  GROUP BY sp.project_id
),

-- Supplies (no semilla)
used_sup AS (
  SELECT sp.project_id, SUM(sp.price) AS executed_supplies_usd
  FROM supplies sp
  WHERE sp.deleted_at IS NULL
    AND NOT (sp.category ILIKE 'semilla%' OR sp.category ILIKE 'seed%')
    AND EXISTS (
      SELECT 1 FROM workorder_items wi
      WHERE wi.supply_id = sp.id
        AND wi.final_dose > 0
        AND wi.deleted_at IS NULL
    )
  GROUP BY sp.project_id
),
inv_sup AS (
  SELECT sp.project_id, SUM(sp.price) AS invested_supplies_usd
  FROM supplies sp
  WHERE sp.deleted_at IS NULL
    AND NOT (sp.category ILIKE 'semilla%' OR sp.category ILIKE 'seed%')
  GROUP BY sp.project_id
),

-- Ingresos por proyecto (por si querés cablear Rent como % de ingresos)
income_by_project AS (
  SELECT f.project_id, COALESCE(SUM(l.tons * 200),0)::numeric(14,2) AS income_usd
  FROM fields f
  LEFT JOIN lots l ON l.field_id = f.id AND l.deleted_at IS NULL
  WHERE f.deleted_at IS NULL
  GROUP BY f.project_id
)

-- -----------------------------------------------
-- SALIDA: 5 filas por proyecto (Seed, Supplies, Labors, Rent, Structure)
-- -----------------------------------------------
SELECT
  p.customer_id,
  p.id             AS project_id,
  p.campaign_id,
  NULL::bigint     AS field_id,
  'Seed'::text     AS label,
  COALESCE(us.executed_seed_usd,0)::numeric(14,2)  AS executed_usd,
  COALESCE(isd.invested_seed_usd,0)::numeric(14,2) AS invested_usd,
  (COALESCE(isd.invested_seed_usd,0) - COALESCE(us.executed_seed_usd,0))::numeric(14,2) AS stock_usd
FROM projects p
LEFT JOIN used_seed us ON us.project_id = p.id
LEFT JOIN inv_seed  isd ON isd.project_id = p.id
WHERE p.deleted_at IS NULL

UNION ALL
SELECT
  p.customer_id,
  p.id,
  p.campaign_id,
  NULL::bigint,
  'Supplies'::text,
  COALESCE(uu.executed_supplies_usd,0)::numeric(14,2),
  COALESCE(ii.invested_supplies_usd,0)::numeric(14,2),
  (COALESCE(ii.invested_supplies_usd,0) - COALESCE(uu.executed_supplies_usd,0))::numeric(14,2)
FROM projects p
LEFT JOIN used_sup uu ON uu.project_id = p.id
LEFT JOIN inv_sup  ii ON ii.project_id = p.id
WHERE p.deleted_at IS NULL

UNION ALL
SELECT
  p.customer_id,
  p.id,
  p.campaign_id,
  NULL::bigint,
  'Labors'::text,
  COALESCE(el.executed_labors_usd,0)::numeric(14,2),
  COALESCE(il.invested_labors_usd,0)::numeric(14,2),
  0::numeric(14,2)  -- sin stock para labores
FROM projects p
LEFT JOIN exec_lab el ON el.project_id = p.id
LEFT JOIN inv_lab  il ON il.project_id = p.id
WHERE p.deleted_at IS NULL

UNION ALL
SELECT
  p.customer_id,
  p.id,
  p.campaign_id,
  NULL::bigint,
  'Rent'::text,
  /* Si tenés projects.rent_pct (0–1): COALESCE(ip.income_usd,0) * COALESCE(p.rent_pct,0)
     o si tenés projects.rent_usd: COALESCE(p.rent_usd,0)
     Por ahora, sin columna definida → 0 */
  0::numeric(14,2) AS executed_usd,
  0::numeric(14,2) AS invested_usd,
  0::numeric(14,2) AS stock_usd
FROM projects p
LEFT JOIN income_by_project ip ON ip.project_id = p.id
WHERE p.deleted_at IS NULL

UNION ALL
SELECT
  p.customer_id,
  p.id,
  p.campaign_id,
  NULL::bigint,
  'Structure'::text,
  COALESCE(p.admin_cost,0)::numeric(14,2) AS executed_usd,
  COALESCE(p.admin_cost,0)::numeric(14,2) AS invested_usd,
  0::numeric(14,2)                         AS stock_usd
FROM projects p
WHERE p.deleted_at IS NULL

ORDER BY customer_id NULLS LAST, project_id NULLS LAST, label;
```

<!-- ### Notas rápidas

* **Seed/Supplies** usan `workorder_items.final_dose > 0` para contar “utilizado”.
* **Labors** usa `workorders.effective_area > 0` para contar “ejecutada”.
* **Structure** toma `projects.admin_cost`.
* **Rent** queda en `0` hasta que me confirmes dónde está el dato (porcentaje o monto). Cuando lo tengas, reemplazo el cálculo en la sección marcada. -->

#####################################################################
Incidencia de Costos por Cultivo
#####################################################################

```sql
-- =========================================================
-- MIGRATION 0000XX: Dashboard - Crop Cost Incidence View
-- =========================================================
-- Purpose: "Incidencia de costos por cultivo" como vista atómica
-- Lógica 100% alineada con la vista general:
--   • Labor EJECUTADA:        EXISTS workorders (w.labor_id = lb.id AND w.effective_area > 0)
--   • Insumo UTILIZADO:       EXISTS workorder_items (wi.supply_id = sp.id AND wi.final_dose > 0)
--   • Costos directos = labors ejecutadas + supplies utilizados (a nivel proyecto)
--   • Se prorratea a cada cultivo por participación de hectáreas
--
-- 📌 Cálculos simples (lo que muestra la card):
--   Hc (hectáreas del cultivo)         = SUM(lots.hectares) del cultivo
--   Ht (hectáreas totales del proyecto)= SUM(lots.hectares) de todos los cultivos
--   Cdir_total (USD)                   = labors_ejecutadas + supplies_utilizados (proyecto)
--   C_cultivo (USD)                    = Cdir_total * (Hc / Ht)
--   Costo/ha (USD/ha)                  = C_cultivo / Hc                  (si Hc=0 → 0)
--   Rotación (%)                       = (Hc / Ht) * 100                 (si Ht=0 → 0)
--   Incidencia (%)                     = (C_cultivo / Cdir_total) * 100  (si Cdir_total=0 → 0)
--
-- Salida:
--   • Filas por cultivo  (row_kind='crop')
--   • Fila TOTAL proyecto (row_kind='total')
--   • Nivel: Customer / Project / Campaign (field_id NULL)
-- =========================================================

DROP VIEW IF EXISTS dashboard_crop_incidence_view;

CREATE OR REPLACE VIEW dashboard_crop_incidence_view AS
WITH
-- ---------------------------------------------------------
-- Labors ejecutadas por proyecto
-- ---------------------------------------------------------
executed_labors_by_project AS (
  SELECT lb.project_id, SUM(lb.price) AS labors_usd
  FROM labors lb
  WHERE lb.deleted_at IS NULL
    AND EXISTS (
      SELECT 1
      FROM workorders w
      WHERE w.labor_id = lb.id
        AND w.effective_area > 0
        AND w.deleted_at IS NULL
    )
  GROUP BY lb.project_id
),

-- ---------------------------------------------------------
-- Supplies utilizados por proyecto
-- ---------------------------------------------------------
used_supplies_by_project AS (
  SELECT sp.project_id, SUM(sp.price) AS supplies_usd
  FROM supplies sp
  WHERE sp.deleted_at IS NULL
    AND EXISTS (
      SELECT 1
      FROM workorder_items wi
      WHERE wi.supply_id = sp.id
        AND wi.final_dose > 0
        AND wi.deleted_at IS NULL
    )
  GROUP BY sp.project_id
),

-- ---------------------------------------------------------
-- Costos directos por proyecto (solo ejecutado/utilizado)
-- ---------------------------------------------------------
direct_costs_by_project AS (
  SELECT
    p.id AS project_id,
    COALESCE(el.labors_usd,   0)::numeric(14,2) AS executed_labors_usd,
    COALESCE(us.supplies_usd, 0)::numeric(14,2) AS used_supplies_usd,
    (COALESCE(el.labors_usd,0) + COALESCE(us.supplies_usd,0))::numeric(14,2) AS total_direct_costs_usd
  FROM projects p
  LEFT JOIN executed_labors_by_project el ON el.project_id = p.id
  LEFT JOIN used_supplies_by_project  us ON us.project_id = p.id
  WHERE p.deleted_at IS NULL
),

-- ---------------------------------------------------------
-- Hectáreas por cultivo (se toma el cultivo desde fields)
-- ⚠️ Si tu esquema guarda el cultivo en otra columna/tabla,
--    reemplazá f.crop_name por la correcta (p.ej. f.crop, l.crop_name, etc.)
-- ---------------------------------------------------------
crop_area AS (
  SELECT
    p.customer_id,
    p.id          AS project_id,
    p.campaign_id,
    NULL::bigint  AS field_id,
    COALESCE(f.crop_name, 'Sin cultivo')::text AS crop_name,
    COALESCE(SUM(l.hectares),0)::numeric(14,2) AS crop_hectares
  FROM projects p
  JOIN fields  f ON f.project_id = p.id AND f.deleted_at IS NULL
  LEFT JOIN lots l ON l.field_id   = f.id AND l.deleted_at IS NULL
  WHERE p.deleted_at IS NULL
  GROUP BY p.customer_id, p.id, p.campaign_id, f.crop_name
),

-- ---------------------------------------------------------
-- Hectáreas totales por proyecto
-- ---------------------------------------------------------
project_hectares AS (
  SELECT
    ca.customer_id,
    ca.project_id,
    ca.campaign_id,
    SUM(ca.crop_hectares)::numeric(14,2) AS total_hectares
  FROM crop_area ca
  GROUP BY ca.customer_id, ca.project_id, ca.campaign_id
),

-- ---------------------------------------------------------
-- Base por cultivo con prorrateo de costos
-- ---------------------------------------------------------
crop_costs AS (
  SELECT
    ca.customer_id,
    ca.project_id,
    ca.campaign_id,
    ca.field_id,                   -- siempre NULL en esta vista
    ca.crop_name,
    ca.crop_hectares                         AS hectares,
    ph.total_hectares,
    dc.total_direct_costs_usd,
    -- prorrateo por hectáreas
    CASE
      WHEN ph.total_hectares > 0
      THEN ROUND( (dc.total_direct_costs_usd * (ca.crop_hectares / ph.total_hectares))::numeric, 2)
      ELSE 0::numeric(14,2)
    END AS crop_cost_usd
  FROM crop_area ca
  JOIN project_hectares ph ON ph.customer_id = ca.customer_id
                           AND ph.project_id  = ca.project_id
                           AND ph.campaign_id = ca.campaign_id
  LEFT JOIN direct_costs_by_project dc ON dc.project_id = ca.project_id
),

-- ---------------------------------------------------------
-- Totales por proyecto (para la fila 'total')
-- ---------------------------------------------------------
project_totals AS (
  SELECT
    ca.customer_id,
    ca.project_id,
    ca.campaign_id,
    NULL::bigint AS field_id,
    SUM(ca.hectares)::numeric(14,2)          AS total_hectares,
    COALESCE(MAX(ca.total_hectares),0)::numeric(14,2) AS check_total_hectares, -- debe igualar total_hectares
    COALESCE(MAX(ca.total_direct_costs_usd),0)::numeric(14,2) AS total_direct_costs_usd
  FROM crop_costs ca
  GROUP BY ca.customer_id, ca.project_id, ca.campaign_id
)

-- ---------------------------------------------------------
-- OUTPUT
--   • Filas por cultivo (row_kind='crop')
--   • Fila total del proyecto (row_kind='total')
-- ---------------------------------------------------------
SELECT
  cc.customer_id,
  cc.project_id,
  cc.campaign_id,
  cc.field_id,             -- NULL
  cc.crop_name,
  -- métricas por cultivo
  cc.hectares                                       AS crop_hectares,
  cc.total_hectares                                  AS project_total_hectares,
  CASE WHEN cc.total_hectares > 0
       THEN ROUND((cc.hectares / cc.total_hectares) * 100, 2)
       ELSE 0 END                                    AS rotation_pct,
  cc.crop_cost_usd                                   AS crop_cost_usd,
  CASE WHEN cc.hectares > 0
       THEN ROUND((cc.crop_cost_usd / cc.hectares), 2)
       ELSE 0 END                                    AS cost_usd_per_ha,
  CASE WHEN cc.total_direct_costs_usd > 0
       THEN ROUND((cc.crop_cost_usd / cc.total_direct_costs_usd) * 100, 2)
       ELSE 0 END                                    AS incidence_pct,
  'crop'::text                                       AS row_kind
FROM crop_costs cc

UNION ALL

-- Fila TOTAL del proyecto (resumen)
SELECT
  pt.customer_id,
  pt.project_id,
  pt.campaign_id,
  pt.field_id,            -- NULL
  'TOTAL'::text           AS crop_name,
  pt.total_hectares       AS crop_hectares,
  pt.total_hectares       AS project_total_hectares,
  100.00::numeric(14,2)   AS rotation_pct,
  pt.total_direct_costs_usd AS crop_cost_usd,
  CASE WHEN pt.total_hectares > 0
       THEN ROUND((pt.total_direct_costs_usd / pt.total_hectares), 2)
       ELSE 0 END         AS cost_usd_per_ha,
  100.00::numeric(14,2)   AS incidence_pct,
  'total'::text           AS row_kind
FROM project_totals pt

ORDER BY
  customer_id NULLS LAST,
  project_id NULLS LAST,
  campaign_id NULLS LAST,
  CASE WHEN row_kind='total' THEN 1 ELSE 0 END,  -- total al final
  crop_name;
```

<!-- **Notas rápidas**

* Si tu esquema guarda el nombre del cultivo en otra columna (p. ej. `fields.crop`, `lots.crop_name`, o vía `fields.crop_id → crops.name`), cambiá **solo** la expresión `COALESCE(f.crop_name, 'Sin cultivo')` en el CTE `crop_area`.
* La **incidencia** usa el mismo universo de costos que la vista general: **solo** labors *ejecutadas* e insumos *utilizados*.
* La asignación de costos por cultivo se hace **proporcional a hectáreas** (criterio estándar cuando los costos no vienen asignados a cultivo en origen). Si más adelante tenés trazabilidad directa por cultivo en `workorders`, se puede reemplazar el prorrateo por una suma directa por cultivo. -->

#####################################################################
Indicadores Operativos
#####################################################################

```sql
-- =========================================================
-- VIEW 1: dashboard_op_first_workorder_view
-- =========================================================
/*
Indicador: Primera orden de trabajo (first_workorder)
- Tomar solo órdenes EJECUTADAS (effective_area > 0, no borradas).
- Fecha = mínima performed_at (o created_at si fuera nulo).
- Se expone: fecha, id, code, status de esa primera OT.
*/
CREATE OR REPLACE VIEW dashboard_op_first_workorder_view AS
WITH wo_exec AS (
  SELECT
    p.customer_id,
    p.id  AS project_id,
    p.campaign_id,
    NULL::bigint AS field_id,
    w.id,
    w.code,
    COALESCE(w.performed_at, w.created_at) AS performed_at,
    COALESCE(w.status, '') AS status
  FROM workorders w
  JOIN labors lb   ON lb.id = w.labor_id AND lb.deleted_at IS NULL
  JOIN projects p  ON p.id = lb.project_id AND p.deleted_at IS NULL
  WHERE w.deleted_at IS NULL
    AND COALESCE(w.effective_area,0) > 0
)
SELECT
  CASE WHEN GROUPING(customer_id)=1 THEN NULL ELSE customer_id END AS customer_id,
  CASE WHEN GROUPING(project_id)=1  THEN NULL ELSE project_id  END AS project_id,
  CASE WHEN GROUPING(campaign_id)=1 THEN NULL ELSE campaign_id END AS campaign_id,
  CASE WHEN GROUPING(field_id)=1    THEN NULL ELSE field_id    END AS field_id,
  MIN(performed_at) AS first_workorder_date,
  (ARRAY_AGG(id    ORDER BY performed_at ASC, id ASC))[1]   AS first_workorder_id,
  (ARRAY_AGG(code  ORDER BY performed_at ASC, id ASC))[1]   AS first_workorder_code,
  (ARRAY_AGG(status ORDER BY performed_at ASC, id ASC))[1]  AS first_workorder_status
FROM wo_exec
GROUP BY GROUPING SETS (
  (customer_id, project_id, campaign_id, field_id),
  (customer_id, project_id, campaign_id),
  (customer_id, project_id),
  (customer_id),
  ()
);
```

---

```sql
-- =========================================================
-- VIEW 2: dashboard_op_last_workorder_view
-- =========================================================
/*
Indicador: Última orden de trabajo (last_workorder)
- Tomar solo órdenes EJECUTADAS (effective_area > 0, no borradas).
- Fecha = máxima performed_at (o created_at si fuera nulo).
- Se expone: fecha, id, code, status de esa última OT.
*/
CREATE OR REPLACE VIEW dashboard_op_last_workorder_view AS
WITH wo_exec AS (
  SELECT
    p.customer_id,
    p.id  AS project_id,
    p.campaign_id,
    NULL::bigint AS field_id,
    w.id,
    w.code,
    COALESCE(w.performed_at, w.created_at) AS performed_at,
    COALESCE(w.status, '') AS status
  FROM workorders w
  JOIN labors lb   ON lb.id = w.labor_id AND lb.deleted_at IS NULL
  JOIN projects p  ON p.id = lb.project_id AND p.deleted_at IS NULL
  WHERE w.deleted_at IS NULL
    AND COALESCE(w.effective_area,0) > 0
)
SELECT
  CASE WHEN GROUPING(customer_id)=1 THEN NULL ELSE customer_id END AS customer_id,
  CASE WHEN GROUPING(project_id)=1  THEN NULL ELSE project_id  END AS project_id,
  CASE WHEN GROUPING(campaign_id)=1 THEN NULL ELSE campaign_id END AS campaign_id,
  CASE WHEN GROUPING(field_id)=1    THEN NULL ELSE field_id    END AS field_id,
  MAX(performed_at) AS last_workorder_date,
  (ARRAY_AGG(id    ORDER BY performed_at DESC, id DESC))[1]  AS last_workorder_id,
  (ARRAY_AGG(code  ORDER BY performed_at DESC, id DESC))[1]  AS last_workorder_code,
  (ARRAY_AGG(status ORDER BY performed_at DESC, id DESC))[1] AS last_workorder_status
FROM wo_exec
GROUP BY GROUPING SETS (
  (customer_id, project_id, campaign_id, field_id),
  (customer_id, project_id, campaign_id),
  (customer_id, project_id),
  (customer_id),
  ()
);
```

---

```sql
-- =========================================================
-- VIEW 3: dashboard_op_last_stock_audit_view
-- =========================================================
/*
Indicador: Último arqueo de stock (last_stock_audit)
- Tomar auditorías de stock no borradas.
- Fecha = máxima audited_at (o created_at si fuera nulo).
- Se expone: fecha, id, code, status de esa auditoría.
NOTA: Ajustá nombres de columna si en tu esquema difieren
      (ej.: audited_at / performed_at, code, status).
*/
CREATE OR REPLACE VIEW dashboard_op_last_stock_audit_view AS
WITH audits AS (
  SELECT
    p.customer_id,
    p.id  AS project_id,
    p.campaign_id,
    NULL::bigint AS field_id,
    sa.id,
    sa.code,
    COALESCE(sa.audited_at, sa.created_at) AS audited_at,
    COALESCE(sa.status,'') AS status
  FROM stock_audits sa
  JOIN projects p ON p.id = sa.project_id AND p.deleted_at IS NULL
  WHERE sa.deleted_at IS NULL
)
SELECT
  CASE WHEN GROUPING(customer_id)=1 THEN NULL ELSE customer_id END AS customer_id,
  CASE WHEN GROUPING(project_id)=1  THEN NULL ELSE project_id  END AS project_id,
  CASE WHEN GROUPING(campaign_id)=1 THEN NULL ELSE campaign_id END AS campaign_id,
  CASE WHEN GROUPING(field_id)=1    THEN NULL ELSE field_id    END AS field_id,
  MAX(audited_at) AS last_stock_audit_date,
  (ARRAY_AGG(id    ORDER BY audited_at DESC, id DESC))[1]  AS last_stock_audit_id,
  (ARRAY_AGG(code  ORDER BY audited_at DESC, id DESC))[1]  AS last_stock_audit_code,
  (ARRAY_AGG(status ORDER BY audited_at DESC, id DESC))[1] AS last_stock_audit_status
FROM audits
GROUP BY GROUPING SETS (
  (customer_id, project_id, campaign_id, field_id),
  (customer_id, project_id, campaign_id),
  (customer_id, project_id),
  (customer_id),
  ()
);
```

---

```sql
-- =========================================================
-- VIEW 4: dashboard_op_campaign_close_view
-- =========================================================
/*
Indicador: Cierre de campaña (campaign_close)
- Tomar campaigns de los projects (no borrados).
- Fecha = closed_at de la campaña (puede haber varias → tomar la última).
- Estado: 'closed' si TODAS tienen closed_at, si alguna no, 'pending'.
- Se expone: fecha cierre (última), campaign_id (si aplica), status.
NOTA: Ajustá si tu esquema usa end_date/closed_at o status distinto.
*/
CREATE OR REPLACE VIEW dashboard_op_campaign_close_view AS
WITH cx AS (
  SELECT
    p.customer_id,
    p.id  AS project_id,
    p.campaign_id,
    NULL::bigint AS field_id,
    c.id AS current_campaign_id,
    c.closed_at,
    COALESCE(c.status, CASE WHEN c.closed_at IS NULL THEN 'pending' ELSE 'closed' END) AS campaign_status
  FROM projects p
  LEFT JOIN campaigns c ON c.id = p.campaign_id AND c.deleted_at IS NULL
  WHERE p.deleted_at IS NULL
)
SELECT
  CASE WHEN GROUPING(customer_id)=1 THEN NULL ELSE customer_id END AS customer_id,
  CASE WHEN GROUPING(project_id)=1  THEN NULL ELSE project_id  END AS project_id,
  CASE WHEN GROUPING(campaign_id)=1 THEN NULL ELSE campaign_id END AS campaign_id,
  CASE WHEN GROUPING(field_id)=1    THEN NULL ELSE field_id    END AS field_id,
  MAX(closed_at) AS campaign_close_date,
  -- Si alguna campaña del grupo no tiene closed_at → 'pending', si no → 'closed'
  CASE WHEN COUNT(*) FILTER (WHERE closed_at IS NULL) > 0 THEN 'pending' ELSE 'closed' END AS campaign_close_status,
  -- Retornamos el campaign_id actual si el nivel es por proyecto/campaign; en niveles agregados puede quedar NULL
  CASE WHEN GROUPING(campaign_id)=0 THEN MAX(current_campaign_id) ELSE NULL END AS campaign_close_campaign_id
FROM cx
GROUP BY GROUPING SETS (
  (customer_id, project_id, campaign_id, field_id),
  (customer_id, project_id, campaign_id),
  (customer_id, project_id),
  (customer_id),
  ()
);
```

