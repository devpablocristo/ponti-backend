-- ========================================
-- MIGRATION 000196 ALLOW ARCHIVED LABORS/SUPPLIES IN VIEWS (UP)
-- ========================================
-- Permite que labores e insumos archivados (soft deleted) sigan apareciendo
-- en reportes de OT existentes. Remueve filtros lb.deleted_at IS NULL y
-- s.deleted_at IS NULL de JOINs en contexto de workorders y funciones SSOT.
-- Los filtros w.deleted_at IS NULL (workorders) se MANTIENEN.

BEGIN;

-- ==========================================
-- 000110: SSOT FUNCTIONS
-- ==========================================

CREATE OR REPLACE FUNCTION v4_ssot.labor_cost_for_lot(p_lot_id bigint)
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(SUM(lb.price * w.effective_area), 0)::numeric
  FROM public.workorders w
  JOIN public.labors lb ON lb.id = w.labor_id
  WHERE w.deleted_at IS NULL
    AND w.effective_area > 0
    AND lb.price IS NOT NULL
    AND w.lot_id = p_lot_id
$$;

CREATE OR REPLACE FUNCTION v4_ssot.supply_cost_for_lot_base(p_lot_id bigint)
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    SUM(wi.total_used * s.price), 0
  )::numeric
  FROM public.workorders w
  JOIN public.workorder_items wi ON wi.workorder_id = w.id AND wi.deleted_at IS NULL
  JOIN public.supplies s ON s.id = wi.supply_id
  WHERE w.lot_id = p_lot_id
    AND w.deleted_at IS NULL
    AND w.effective_area IS NOT NULL
    AND w.effective_area > 0
$$;

CREATE OR REPLACE FUNCTION v4_ssot.supply_cost_for_lot_by_category(
  p_lot_id bigint,
  p_category_name text
)
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    SUM(wi.total_used * s.price), 0
  )::numeric
  FROM public.workorders w
  JOIN public.workorder_items wi ON wi.workorder_id = w.id AND wi.deleted_at IS NULL
  JOIN public.supplies s ON s.id = wi.supply_id
  JOIN public.categories c ON c.id = s.category_id AND c.deleted_at IS NULL
  WHERE w.lot_id = p_lot_id
    AND c.name = p_category_name
    AND w.deleted_at IS NULL
    AND w.effective_area > 0
    AND wi.total_used > 0
    AND s.price IS NOT NULL
$$;

CREATE OR REPLACE FUNCTION v4_ssot.liters_for_lot(p_lot_id bigint)
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    SUM(CASE WHEN s.unit_id = 1 THEN (wi.final_dose * w.effective_area) ELSE 0 END), 0
  )::numeric
  FROM public.workorders w
  LEFT JOIN public.workorder_items wi ON wi.workorder_id = w.id AND wi.deleted_at IS NULL
  LEFT JOIN public.supplies s ON s.id = wi.supply_id
  WHERE w.lot_id = p_lot_id
    AND w.deleted_at IS NULL
$$;

CREATE OR REPLACE FUNCTION v4_ssot.kilograms_for_lot(p_lot_id bigint)
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    SUM(CASE WHEN s.unit_id = 2 THEN (wi.final_dose * w.effective_area) ELSE 0 END), 0
  )::numeric
  FROM public.workorders w
  LEFT JOIN public.workorder_items wi ON wi.workorder_id = w.id AND wi.deleted_at IS NULL
  LEFT JOIN public.supplies s ON s.id = wi.supply_id
  WHERE w.lot_id = p_lot_id
    AND w.deleted_at IS NULL
$$;

CREATE OR REPLACE FUNCTION v4_ssot.supply_cost_for_lot(p_lot_id bigint)
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(SUM(wi.total_used * s.price), 0)::numeric
  FROM public.workorders w
  JOIN public.workorder_items wi ON wi.workorder_id = w.id AND wi.deleted_at IS NULL
  JOIN public.supplies s ON s.id = wi.supply_id
  WHERE w.deleted_at IS NULL
    AND w.effective_area > 0
    AND wi.total_used > 0
    AND s.price IS NOT NULL
    AND w.lot_id = p_lot_id
$$;

CREATE OR REPLACE FUNCTION v4_ssot.seeded_area_for_lot(p_lot_id bigint)
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(SUM(
    CASE WHEN lb.category_id = 9 THEN w.effective_area ELSE 0 END
  ), 0)::numeric
  FROM public.workorders w
  JOIN public.labors lb ON lb.id = w.labor_id
  WHERE w.lot_id = p_lot_id
    AND w.deleted_at IS NULL
    AND w.effective_area IS NOT NULL
    AND w.effective_area > 0
$$;

CREATE OR REPLACE FUNCTION v4_ssot.labor_cost_pre_harvest_for_lot(p_lot_id bigint)
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(SUM(lb.price * w.effective_area), 0)::numeric
  FROM public.workorders w
  JOIN public.labors lb ON lb.id = w.labor_id
  JOIN public.categories cat ON cat.id = lb.category_id AND cat.deleted_at IS NULL
  WHERE w.deleted_at IS NULL
    AND w.effective_area > 0
    AND lb.price IS NOT NULL
    AND w.lot_id = p_lot_id
    AND cat.type_id = 4
    AND cat.name != 'Cosecha'
$$;

CREATE OR REPLACE FUNCTION v4_ssot.stock_value_for_project(p_project_id bigint)
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    (SELECT COALESCE(SUM(s.price * st.initial_units), 0)::numeric
     FROM public.supplies s
     JOIN public.stocks st ON st.supply_id = s.id AND st.deleted_at IS NULL
     WHERE s.project_id = p_project_id
       AND s.deleted_at IS NULL
       AND st.initial_units IS NOT NULL)
    -
    (SELECT COALESCE(SUM(wi.total_used * s.price), 0)::numeric
     FROM public.workorders w
     JOIN public.workorder_items wi ON wi.workorder_id = w.id AND wi.deleted_at IS NULL
     JOIN public.supplies s ON s.id = wi.supply_id
     WHERE w.project_id = p_project_id AND w.deleted_at IS NULL)
  , 0)::numeric
$$;

CREATE OR REPLACE FUNCTION v4_ssot.supply_cost_for_project(p_project_id bigint)
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    (SELECT COALESCE(SUM(wi.total_used * s.price), 0)::numeric
     FROM public.workorders w
     JOIN public.workorder_items wi ON wi.workorder_id = w.id AND wi.deleted_at IS NULL
     JOIN public.supplies s ON s.id = wi.supply_id
     WHERE w.deleted_at IS NULL
       AND w.effective_area > 0
       AND wi.total_used > 0
       AND s.price IS NOT NULL
       AND w.project_id = p_project_id)
    +
    (SELECT COALESCE(SUM(sm.quantity * s.price), 0)::numeric
     FROM public.supply_movements sm
     JOIN public.supplies s ON s.id = sm.supply_id
     WHERE sm.deleted_at IS NULL
       AND sm.movement_type = 'Movimiento interno'
       AND sm.is_entry = false
       AND sm.project_id = p_project_id
       AND s.price IS NOT NULL
       AND sm.quantity > 0)
  , 0)::numeric
$$;

-- ==========================================
-- 000120: CALC VIEWS
-- ==========================================

CREATE OR REPLACE VIEW v4_calc.workorder_metrics AS
WITH base AS (
  SELECT
    w.id AS workorder_id,
    w.project_id,
    w.field_id,
    w.lot_id,
    w.effective_area,
    lb.price AS labor_price
  FROM public.workorders w
  JOIN public.labors lb ON lb.id = w.labor_id
  WHERE w.deleted_at IS NULL
    AND w.effective_area IS NOT NULL
    AND w.effective_area > 0
),
surface AS (
  SELECT project_id, field_id, lot_id, SUM(effective_area)::numeric AS surface_ha
  FROM base
  GROUP BY project_id, field_id, lot_id
),
labor_costs AS (
  SELECT
    project_id, field_id, lot_id,
    SUM((labor_price * effective_area))::numeric AS labor_cost_usd
  FROM base
  GROUP BY project_id, field_id, lot_id
),
supply_metrics AS (
  SELECT
    b.project_id, b.field_id, b.lot_id,
    SUM(CASE WHEN s.unit_id = 1 THEN (wi.final_dose * b.effective_area) ELSE 0 END)::numeric AS liters,
    SUM(CASE WHEN s.unit_id = 2 THEN (wi.final_dose * b.effective_area) ELSE 0 END)::numeric AS kilograms,
    SUM(COALESCE(wi.total_used, 0) * COALESCE(s.price, 0))::numeric AS supplies_cost_usd
  FROM base b
  LEFT JOIN public.workorder_items wi
    ON wi.workorder_id = b.workorder_id AND wi.deleted_at IS NULL
  LEFT JOIN public.supplies s
    ON s.id = wi.supply_id
  GROUP BY b.project_id, b.field_id, b.lot_id
)
SELECT
  COALESCE(sur.project_id, lc.project_id, sm.project_id) AS project_id,
  COALESCE(sur.field_id, lc.field_id, sm.field_id) AS field_id,
  COALESCE(sur.lot_id, lc.lot_id, sm.lot_id) AS lot_id,
  COALESCE(sur.surface_ha, 0)::numeric AS surface_ha,
  COALESCE(sm.liters, 0)::numeric AS liters,
  COALESCE(sm.kilograms, 0)::numeric AS kilograms,
  COALESCE(lc.labor_cost_usd, 0)::numeric AS labor_cost_usd,
  COALESCE(sm.supplies_cost_usd, 0)::numeric AS supplies_cost_usd,
  (COALESCE(lc.labor_cost_usd, 0)::numeric +
   COALESCE(sm.supplies_cost_usd, 0)::numeric) AS direct_cost_usd,
  v4_core.cost_per_ha(
    COALESCE(lc.labor_cost_usd, 0)::numeric + COALESCE(sm.supplies_cost_usd, 0)::numeric,
    COALESCE(sur.surface_ha, 0)::numeric
  ) AS avg_cost_per_ha_usd,
  v4_core.per_ha(COALESCE(sm.liters, 0)::numeric, COALESCE(sur.surface_ha, 0)::numeric) AS liters_per_ha,
  v4_core.per_ha(COALESCE(sm.kilograms, 0)::numeric, COALESCE(sur.surface_ha, 0)::numeric) AS kilograms_per_ha
FROM surface sur
FULL JOIN labor_costs lc USING (project_id, field_id, lot_id)
FULL JOIN supply_metrics sm USING (project_id, field_id, lot_id);

CREATE OR REPLACE VIEW v4_calc.workorder_metrics_raw AS
WITH base AS (
  SELECT
    w.id AS workorder_id,
    w.project_id,
    w.field_id,
    w.lot_id,
    w.effective_area,
    lb.price AS labor_price
  FROM public.workorders w
  JOIN public.labors lb ON lb.id = w.labor_id
  WHERE w.deleted_at IS NULL
    AND w.effective_area IS NOT NULL
    AND w.effective_area > 0
),
surface AS (
  SELECT project_id, field_id, lot_id, SUM(effective_area)::numeric AS surface_ha
  FROM base
  GROUP BY project_id, field_id, lot_id
),
labor_costs AS (
  SELECT
    project_id, field_id, lot_id,
    SUM((labor_price * effective_area))::numeric AS labor_cost_usd
  FROM base
  GROUP BY project_id, field_id, lot_id
),
supply_metrics AS (
  SELECT
    b.project_id, b.field_id, b.lot_id,
    SUM(CASE WHEN s.unit_id = 1 THEN (wi.final_dose * b.effective_area) ELSE 0 END)::numeric AS liters,
    SUM(CASE WHEN s.unit_id = 2 THEN (wi.final_dose * b.effective_area) ELSE 0 END)::numeric AS kilograms,
    SUM(COALESCE(wi.total_used, 0) * COALESCE(s.price, 0))::numeric AS supplies_cost_usd
  FROM base b
  LEFT JOIN public.workorder_items wi
    ON wi.workorder_id = b.workorder_id AND wi.deleted_at IS NULL
  LEFT JOIN public.supplies s
    ON s.id = wi.supply_id
  GROUP BY b.project_id, b.field_id, b.lot_id
)
SELECT
  COALESCE(sur.project_id, lc.project_id, sm.project_id) AS project_id,
  COALESCE(sur.field_id, lc.field_id, sm.field_id) AS field_id,
  COALESCE(sur.lot_id, lc.lot_id, sm.lot_id) AS lot_id,
  COALESCE(sur.surface_ha, 0)::numeric AS surface_ha,
  COALESCE(sm.liters, 0)::numeric AS liters,
  COALESCE(sm.kilograms, 0)::numeric AS kilograms,
  COALESCE(lc.labor_cost_usd, 0)::numeric AS labor_cost_usd,
  COALESCE(sm.supplies_cost_usd, 0)::numeric AS supplies_cost_usd,
  (COALESCE(lc.labor_cost_usd, 0)::numeric +
   COALESCE(sm.supplies_cost_usd, 0)::numeric) AS direct_cost_usd,
  v4_core.cost_per_ha(
    COALESCE(lc.labor_cost_usd, 0)::numeric + COALESCE(sm.supplies_cost_usd, 0)::numeric,
    COALESCE(sur.surface_ha, 0)::numeric
  ) AS avg_cost_per_ha_usd,
  v4_core.per_ha(COALESCE(sm.liters, 0)::numeric, COALESCE(sur.surface_ha, 0)::numeric) AS liters_per_ha,
  v4_core.per_ha(COALESCE(sm.kilograms, 0)::numeric, COALESCE(sur.surface_ha, 0)::numeric) AS kilograms_per_ha
FROM surface sur
FULL JOIN labor_costs lc USING (project_id, field_id, lot_id)
FULL JOIN supply_metrics sm USING (project_id, field_id, lot_id);

-- ==========================================
-- 000130: REPORT VIEWS
-- ==========================================

CREATE OR REPLACE VIEW v4_report.labor_metrics AS
WITH wo AS (
  SELECT
    w.id AS workorder_id,
    w.project_id,
    w.field_id,
    w.date,
    w.effective_area::numeric AS effective_area,
    lb.price::numeric AS labor_price_per_ha
  FROM public.workorders w
  JOIN public.labors lb ON lb.id = w.labor_id
  WHERE w.deleted_at IS NULL
    AND w.effective_area IS NOT NULL
    AND w.effective_area > 0
    AND lb.price IS NOT NULL
),
agg AS (
  SELECT
    project_id,
    field_id,
    COUNT(DISTINCT workorder_id) AS total_workorders,
    SUM(effective_area) AS surface_ha,
    SUM(v4_core.labor_cost(labor_price_per_ha, effective_area)) AS total_labor_cost,
    MIN(date) AS first_workorder_date,
    MAX(date) AS last_workorder_date
  FROM wo
  GROUP BY project_id, field_id
)
SELECT
  a.project_id,
  a.field_id,
  a.surface_ha,
  a.total_labor_cost,
  v4_core.cost_per_ha(a.total_labor_cost, a.surface_ha) AS avg_labor_cost_per_ha,
  a.total_workorders,
  a.first_workorder_date,
  a.last_workorder_date
FROM agg a;

-- ==========================================
-- 000192: LABOR LIST VIEW (supersedes 000130 version)
-- ==========================================

CREATE OR REPLACE VIEW v4_report.labor_list AS
WITH workorder_alloc AS (
  SELECT
    w.id AS workorder_id,
    w.investor_id,
    1::numeric AS factor
  FROM public.workorders w
  WHERE w.deleted_at IS NULL
    AND NOT EXISTS (
      SELECT 1
      FROM public.workorder_investor_splits wis
      WHERE wis.workorder_id = w.id
        AND wis.deleted_at IS NULL
    )
  UNION ALL
  SELECT
    w.id AS workorder_id,
    wis.investor_id,
    (wis.percentage::numeric / 100)::numeric AS factor
  FROM public.workorders w
  JOIN public.workorder_investor_splits wis
    ON wis.workorder_id = w.id
   AND wis.deleted_at IS NULL
  WHERE w.deleted_at IS NULL
)
SELECT
  w.id AS workorder_id,
  w.number AS workorder_number,
  w.date,
  w.project_id,
  p.name AS project_name,
  w.field_id,
  f.name AS field_name,
  w.lot_id,
  l.name AS lot_name,
  w.crop_id,
  c.name AS crop_name,
  w.labor_id,
  lb.name AS labor_name,
  lb.category_id AS labor_category_id,
  cat.name AS labor_category_name,
  w.contractor,
  lb.contractor_name,
  (w.effective_area * a.factor)::numeric(18,6) AS surface_ha,
  lb.price AS cost_per_ha,
  (lb.price * w.effective_area * a.factor)::numeric AS total_labor_cost,
  v4_core.dollar_average_for_month(w.project_id, w.date) AS dollar_average_month,
  lb.price::numeric AS usd_cost_ha,
  (lb.price * w.effective_area * a.factor)::numeric AS usd_net_total,
  a.investor_id,
  i.name AS investor_name
FROM public.workorders w
JOIN workorder_alloc a ON a.workorder_id = w.id
JOIN public.projects p ON p.id = w.project_id AND p.deleted_at IS NULL
JOIN public.fields f ON f.id = w.field_id AND f.deleted_at IS NULL
LEFT JOIN public.lots l ON l.id = w.lot_id AND l.deleted_at IS NULL
LEFT JOIN public.crops c ON c.id = w.crop_id AND c.deleted_at IS NULL
JOIN public.labors lb ON lb.id = w.labor_id
LEFT JOIN public.categories cat ON cat.id = lb.category_id AND cat.deleted_at IS NULL
LEFT JOIN public.investors i ON i.id = a.investor_id AND i.deleted_at IS NULL
WHERE w.deleted_at IS NULL
  AND w.effective_area IS NOT NULL
  AND w.effective_area > 0
  AND lb.price IS NOT NULL;

-- ==========================================
-- 000194: WORKORDER LIST VIEW (supersedes 000130 version)
-- ==========================================

CREATE OR REPLACE VIEW v4_report.workorder_list AS
-- Filas de insumos
SELECT
  w.id,
  w.number,
  w.project_id,
  w.field_id,
  p.name  AS project_name,
  f.name  AS field_name,
  l.name  AS lot_name,
  w.date,
  c.name  AS crop_name,
  lb.name AS labor_name,
  cat_lb.name AS labor_category_name,
  t.name  AS type_name,
  w.contractor,
  w.effective_area AS surface_ha,
  s.name AS supply_name,
  wi.total_used AS consumption,
  cat.name AS category_name,
  wi.final_dose AS dose_per_ha,
  s.price AS unit_price,
  CASE
    WHEN wi.final_dose IS NOT NULL AND s.price IS NOT NULL
    THEN v4_core.cost_per_ha(
      (wi.final_dose::numeric * s.price)::numeric,
      1::numeric
    )
    ELSE 0
  END AS supply_cost_per_ha,
  v4_core.supply_cost(
    wi.final_dose::numeric,
    s.price::numeric,
    w.effective_area::numeric
  ) AS supply_total_cost
FROM public.workorders w
JOIN public.projects p ON p.id = w.project_id AND p.deleted_at IS NULL
JOIN public.fields f ON f.id = w.field_id AND f.deleted_at IS NULL
JOIN public.lots l ON l.id = w.lot_id AND l.deleted_at IS NULL
JOIN public.crops c ON c.id = w.crop_id AND c.deleted_at IS NULL
JOIN public.labors lb ON lb.id = w.labor_id
LEFT JOIN public.categories cat_lb ON cat_lb.id = lb.category_id AND cat_lb.deleted_at IS NULL
LEFT JOIN public.workorder_items wi ON wi.workorder_id = w.id AND wi.deleted_at IS NULL
LEFT JOIN public.supplies s ON s.id = wi.supply_id
LEFT JOIN public.types t ON t.id = s.type_id AND t.deleted_at IS NULL
LEFT JOIN public.categories cat ON cat.id = s.category_id AND cat.deleted_at IS NULL
WHERE w.deleted_at IS NULL

UNION ALL

-- Fila de costo de labor (una por workorder)
SELECT
  w.id,
  w.number,
  w.project_id,
  w.field_id,
  p.name  AS project_name,
  f.name  AS field_name,
  l.name  AS lot_name,
  w.date,
  c.name  AS crop_name,
  lb.name AS labor_name,
  cat_lb.name AS labor_category_name,
  'Labor'::varchar(250) AS type_name,
  w.contractor,
  w.effective_area AS surface_ha,
  lb.name::varchar(100) AS supply_name,
  0::numeric(18,6) AS consumption,
  cat_lb.name::varchar(250) AS category_name,
  0::numeric(18,6) AS dose_per_ha,
  lb.price::numeric(18,6) AS unit_price,
  lb.price::numeric AS supply_cost_per_ha,
  (lb.price * w.effective_area)::numeric AS supply_total_cost
FROM public.workorders w
JOIN public.projects p ON p.id = w.project_id AND p.deleted_at IS NULL
JOIN public.fields f ON f.id = w.field_id AND f.deleted_at IS NULL
JOIN public.lots l ON l.id = w.lot_id AND l.deleted_at IS NULL
JOIN public.crops c ON c.id = w.crop_id AND c.deleted_at IS NULL
JOIN public.labors lb ON lb.id = w.labor_id
LEFT JOIN public.categories cat_lb ON cat_lb.id = lb.category_id AND cat_lb.deleted_at IS NULL
WHERE w.deleted_at IS NULL
  AND w.effective_area IS NOT NULL
  AND w.effective_area > 0
  AND lb.price IS NOT NULL;

COMMIT;
