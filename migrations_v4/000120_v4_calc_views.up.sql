-- ========================================
-- MIGRATION 000120 V4 CALC VIEWS (UP)
-- ========================================
-- Nota: SQL en inglés, comentarios en español

BEGIN;

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
  JOIN public.labors lb ON lb.id = w.labor_id AND lb.deleted_at IS NULL
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
    ON s.id = wi.supply_id AND s.deleted_at IS NULL
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
  JOIN public.labors lb ON lb.id = w.labor_id AND lb.deleted_at IS NULL
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
    ON s.id = wi.supply_id AND s.deleted_at IS NULL
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

CREATE OR REPLACE VIEW v4_calc.lot_base_costs AS
WITH
raw AS (
  SELECT
    f.project_id,
    f.id AS field_id,
    l.current_crop_id,
    l.id AS lot_id,
    l.name AS lot_name,
    COALESCE(l.hectares, 0) AS hectares,
    COALESCE(l.tons, 0) AS tons,
    l.sowing_date
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  WHERE l.deleted_at IS NULL
),
areas AS (
  SELECT
    r.lot_id,
    v4_ssot.seeded_area_for_lot(r.lot_id) AS seeded_area_ha,
    v4_ssot.harvested_area_for_lot(r.lot_id) AS harvested_area_ha
  FROM raw r
),
costs AS (
  SELECT
    lot_id,
    MAX(COALESCE(labor_cost_usd, 0))::numeric AS labor_cost_usd,
    MAX(COALESCE(supplies_cost_usd, 0))::numeric AS supplies_cost_usd,
    MAX(COALESCE(direct_cost_usd, 0))::numeric AS direct_cost_usd
  FROM v4_calc.workorder_metrics_raw
  GROUP BY lot_id
),
ssot_values AS (
  SELECT
    r.lot_id,
    v4_ssot.yield_tn_per_ha_for_lot(r.lot_id) AS yield_tn_per_ha,
    v4_ssot.income_net_total_for_lot(r.lot_id) AS income_net_total_usd,
    v4_ssot.rent_per_ha_for_lot(r.lot_id) AS rent_per_ha_usd,
    v4_ssot.rent_fixed_only_for_lot(r.lot_id) AS rent_fixed_per_ha_usd,
    v4_ssot.admin_cost_per_ha_for_lot(r.lot_id) AS admin_cost_per_ha_usd
  FROM raw r
),
derived AS (
  SELECT
    r.project_id, r.field_id, r.current_crop_id, r.lot_id, r.lot_name,
    r.hectares, r.tons, r.sowing_date,
    COALESCE(a.seeded_area_ha, 0)::numeric AS seeded_area_ha,
    COALESCE(a.harvested_area_ha, 0)::numeric AS harvested_area_ha,
    COALESCE(c.labor_cost_usd, 0)::numeric AS labor_cost_usd,
    COALESCE(c.supplies_cost_usd, 0)::numeric AS supplies_cost_usd,
    COALESCE(c.direct_cost_usd, 0)::numeric AS direct_cost_usd,
    COALESCE(s.yield_tn_per_ha, 0) AS yield_tn_per_ha,
    COALESCE(s.income_net_total_usd, 0)::numeric AS income_net_total_usd,
    COALESCE(s.rent_per_ha_usd, 0)::numeric AS rent_per_ha_usd,
    COALESCE(s.rent_fixed_per_ha_usd, 0)::numeric AS rent_fixed_per_ha_usd,
    COALESCE(s.admin_cost_per_ha_usd, 0)::numeric AS admin_cost_per_ha_usd
  FROM raw r
  LEFT JOIN areas a ON a.lot_id = r.lot_id
  LEFT JOIN costs c ON c.lot_id = r.lot_id
  LEFT JOIN ssot_values s ON s.lot_id = r.lot_id
)
SELECT
  project_id,
  field_id,
  current_crop_id,
  lot_id,
  lot_name,
  hectares,
  tons,
  sowing_date,
  seeded_area_ha,
  harvested_area_ha,
  yield_tn_per_ha,
  labor_cost_usd,
  supplies_cost_usd,
  direct_cost_usd,
  income_net_total_usd,
  v4_core.per_ha(income_net_total_usd, hectares::numeric) AS income_net_per_ha_usd,
  v4_core.per_ha(direct_cost_usd, hectares::numeric) AS direct_cost_per_ha_usd,
  rent_per_ha_usd,
  rent_fixed_per_ha_usd,
  admin_cost_per_ha_usd
FROM derived d;

CREATE OR REPLACE VIEW v4_calc.lot_base_income AS
SELECT
  c.project_id, c.field_id, c.current_crop_id, c.lot_id, c.lot_name,
  c.hectares,
  c.tons,
  c.seeded_area_ha,
  c.yield_tn_per_ha,
  COALESCE(v4_ssot.net_price_usd_for_lot(c.lot_id), 0)::numeric AS net_price_usd_tn,
  c.income_net_total_usd,
  c.income_net_per_ha_usd
FROM v4_calc.lot_base_costs c;

CREATE OR REPLACE VIEW v4_calc.field_crop_lot_base AS
SELECT
  f.project_id,
  f.id AS field_id,
  l.current_crop_id AS crop_id,
  l.id AS lot_id,
  l.hectares AS surface_ha,
  v4_ssot.seeded_area_for_lot(l.id)::numeric AS seeded_area_ha,
  l.tons
FROM public.lots l
JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
WHERE l.deleted_at IS NULL
  AND l.current_crop_id IS NOT NULL;

CREATE OR REPLACE VIEW v4_calc.field_crop_supply_costs_by_lot AS
SELECT
  project_id,
  field_id,
  crop_id,
  lot_id,
  surface_ha,
  seeded_area_ha,
  tons,
  
  v4_ssot.supply_cost_for_lot_by_category(lot_id, 'Semilla') AS semillas_usd,
  v4_ssot.supply_cost_for_lot_by_category(lot_id, 'Curasemillas') AS curasemillas_usd,
  v4_ssot.supply_cost_for_lot_by_category(lot_id, 'Herbicidas') AS herbicidas_usd,
  v4_ssot.supply_cost_for_lot_by_category(lot_id, 'Insecticidas') AS insecticidas_usd,
  v4_ssot.supply_cost_for_lot_by_category(lot_id, 'Fungicidas') AS fungicidas_usd,
  v4_ssot.supply_cost_for_lot_by_category(lot_id, 'Coadyuvantes') AS coadyuvantes_usd,
  v4_ssot.supply_cost_for_lot_by_category(lot_id, 'Fertilizantes') AS fertilizantes_usd,
  v4_ssot.supply_cost_for_lot_by_category(lot_id, 'Otros Insumos') AS otros_insumos_usd,
  
  v4_ssot.supply_cost_for_lot_base(lot_id)::numeric AS total_insumos_usd
FROM v4_calc.field_crop_lot_base;

CREATE OR REPLACE VIEW v4_calc.field_crop_labor_costs_by_lot AS
SELECT
  lb.project_id,
  lb.field_id,
  lb.crop_id,
  lb.lot_id,
  lb.seeded_area_ha,
  lb.surface_ha,
  
  COALESCE((
    SELECT SUM(lab.price * w.effective_area)
    FROM public.workorders w
    JOIN public.labors lab ON lab.id = w.labor_id AND lab.deleted_at IS NULL
    JOIN public.categories cat ON cat.id = lab.category_id AND cat.deleted_at IS NULL
    WHERE w.lot_id = lb.lot_id
      AND w.deleted_at IS NULL
      AND w.effective_area > 0
      AND lab.price IS NOT NULL
      AND cat.name = 'Siembra'
      AND cat.type_id = 4
  ), 0)::numeric AS siembra_usd,
  
  COALESCE((
    SELECT SUM(lab.price * w.effective_area)
    FROM public.workorders w
    JOIN public.labors lab ON lab.id = w.labor_id AND lab.deleted_at IS NULL
    JOIN public.categories cat ON cat.id = lab.category_id AND cat.deleted_at IS NULL
    WHERE w.lot_id = lb.lot_id
      AND w.deleted_at IS NULL
      AND w.effective_area > 0
      AND lab.price IS NOT NULL
      AND cat.name = 'Pulverización'
      AND cat.type_id = 4
  ), 0)::numeric AS pulverizacion_usd,
  
  COALESCE((
    SELECT SUM(lab.price * w.effective_area)
    FROM public.workorders w
    JOIN public.labors lab ON lab.id = w.labor_id AND lab.deleted_at IS NULL
    JOIN public.categories cat ON cat.id = lab.category_id AND cat.deleted_at IS NULL
    WHERE w.lot_id = lb.lot_id
      AND w.deleted_at IS NULL
      AND w.effective_area > 0
      AND lab.price IS NOT NULL
      AND cat.name = 'Riego'
      AND cat.type_id = 4
  ), 0)::numeric AS riego_usd,
  
  COALESCE((
    SELECT SUM(lab.price * w.effective_area)
    FROM public.workorders w
    JOIN public.labors lab ON lab.id = w.labor_id AND lab.deleted_at IS NULL
    JOIN public.categories cat ON cat.id = lab.category_id AND cat.deleted_at IS NULL
    WHERE w.lot_id = lb.lot_id
      AND w.deleted_at IS NULL
      AND w.effective_area > 0
      AND lab.price IS NOT NULL
      AND cat.name = 'Cosecha'
      AND cat.type_id = 4
  ), 0)::numeric AS cosecha_usd,
  
  COALESCE((
    SELECT SUM(lab.price * w.effective_area)
    FROM public.workorders w
    JOIN public.labors lab ON lab.id = w.labor_id AND lab.deleted_at IS NULL
    JOIN public.categories cat ON cat.id = lab.category_id AND cat.deleted_at IS NULL
    WHERE w.lot_id = lb.lot_id
      AND w.deleted_at IS NULL
      AND w.effective_area > 0
      AND lab.price IS NOT NULL
      AND cat.name NOT IN ('Siembra', 'Pulverización', 'Riego', 'Cosecha')
      AND cat.type_id = 4
  ), 0)::numeric AS otras_labores_usd,
  (
    COALESCE((
      SELECT SUM(lab.price * w.effective_area)
      FROM public.workorders w
      JOIN public.labors lab ON lab.id = w.labor_id AND lab.deleted_at IS NULL
      JOIN public.categories cat ON cat.id = lab.category_id AND cat.deleted_at IS NULL
      WHERE w.lot_id = lb.lot_id
        AND w.deleted_at IS NULL
        AND w.effective_area > 0
        AND lab.price IS NOT NULL
        AND cat.name = 'Siembra'
        AND cat.type_id = 4
    ), 0)::numeric +
    COALESCE((
      SELECT SUM(lab.price * w.effective_area)
      FROM public.workorders w
      JOIN public.labors lab ON lab.id = w.labor_id AND lab.deleted_at IS NULL
      JOIN public.categories cat ON cat.id = lab.category_id AND cat.deleted_at IS NULL
      WHERE w.lot_id = lb.lot_id
        AND w.deleted_at IS NULL
        AND w.effective_area > 0
        AND lab.price IS NOT NULL
        AND cat.name = 'Pulverización'
        AND cat.type_id = 4
    ), 0)::numeric +
    COALESCE((
      SELECT SUM(lab.price * w.effective_area)
      FROM public.workorders w
      JOIN public.labors lab ON lab.id = w.labor_id AND lab.deleted_at IS NULL
      JOIN public.categories cat ON cat.id = lab.category_id AND cat.deleted_at IS NULL
      WHERE w.lot_id = lb.lot_id
        AND w.deleted_at IS NULL
        AND w.effective_area > 0
        AND lab.price IS NOT NULL
        AND cat.name = 'Riego'
        AND cat.type_id = 4
    ), 0)::numeric +
    COALESCE((
      SELECT SUM(lab.price * w.effective_area)
      FROM public.workorders w
      JOIN public.labors lab ON lab.id = w.labor_id AND lab.deleted_at IS NULL
      JOIN public.categories cat ON cat.id = lab.category_id AND cat.deleted_at IS NULL
      WHERE w.lot_id = lb.lot_id
        AND w.deleted_at IS NULL
        AND w.effective_area > 0
        AND lab.price IS NOT NULL
        AND cat.name = 'Cosecha'
        AND cat.type_id = 4
    ), 0)::numeric +
    COALESCE((
      SELECT SUM(lab.price * w.effective_area)
      FROM public.workorders w
      JOIN public.labors lab ON lab.id = w.labor_id AND lab.deleted_at IS NULL
      JOIN public.categories cat ON cat.id = lab.category_id AND cat.deleted_at IS NULL
      WHERE w.lot_id = lb.lot_id
        AND w.deleted_at IS NULL
        AND w.effective_area > 0
        AND lab.price IS NOT NULL
        AND cat.name NOT IN ('Siembra', 'Pulverización', 'Riego', 'Cosecha')
        AND cat.type_id = 4
    ), 0)::numeric
  ) AS total_labores_usd
FROM v4_calc.field_crop_lot_base lb;

CREATE OR REPLACE VIEW v4_calc.field_crop_aggregated AS
SELECT
  project_id,
  field_id,
  crop_id,
  MIN(lot_id) AS sample_lot_id,
  SUM(tons)::numeric AS production_tn,
  SUM(seeded_area_ha)::numeric AS seeded_area_ha,
  SUM(surface_ha)::numeric AS surface_ha,
  SUM(v4_ssot.labor_cost_for_lot(lot_id))::numeric AS labor_costs_usd,
  SUM(total_insumos_usd)::numeric AS supply_costs_usd,
  SUM(v4_ssot.labor_cost_for_lot(lot_id) + total_insumos_usd)::numeric AS direct_cost_usd,
  
  SUM(v4_ssot.rent_fixed_only_for_lot(lot_id) * surface_ha)::numeric AS rent_fixed_usd,
  
  SUM(v4_ssot.rent_per_ha_for_lot(lot_id) * surface_ha)::numeric AS rent_total_usd,
  SUM(v4_ssot.admin_cost_per_ha_for_lot(lot_id) * surface_ha)::numeric AS administration_usd
FROM v4_calc.field_crop_supply_costs_by_lot
GROUP BY project_id, field_id, crop_id;

CREATE OR REPLACE VIEW v4_calc.field_crop_metrics_lot_base AS
SELECT
  f.project_id,
  f.id AS field_id,
  f.name AS field_name,
  l.current_crop_id,
  c.name AS crop_name,
  l.id AS lot_id,
  l.hectares,
  l.tons,
  COALESCE(v4_ssot.seeded_area_for_lot(l.id), 0)::numeric AS seeded_area_ha,
  COALESCE(v4_ssot.harvested_area_for_lot(l.id), 0)::numeric AS harvested_area_ha,
  COALESCE(v4_ssot.yield_tn_per_ha_for_lot(l.id), 0) AS yield_tn_per_ha,
  COALESCE(v4_ssot.labor_cost_for_lot(l.id), 0)::numeric AS labor_cost_usd,
  COALESCE(v4_ssot.supply_cost_for_lot_base(l.id), 0)::numeric AS supply_cost_usd,
  COALESCE(v4_ssot.net_price_usd_for_lot(l.id), 0)::numeric AS net_price_usd,
  COALESCE(v4_ssot.rent_per_ha_for_lot(l.id), 0)::numeric AS rent_per_ha,
  COALESCE(v4_ssot.admin_cost_per_ha_for_lot(l.id), 0)::numeric AS admin_per_ha,
  COALESCE(v4_ssot.board_price_for_lot(l.id), 0)::numeric AS board_price,
  COALESCE(v4_ssot.freight_cost_for_lot(l.id), 0)::numeric AS freight_cost,
  COALESCE(v4_ssot.commercial_cost_for_lot(l.id), 0)::numeric AS commercial_cost
FROM public.lots l
JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
JOIN public.projects p ON p.id = f.project_id AND p.deleted_at IS NULL
LEFT JOIN public.crops c ON c.id = l.current_crop_id AND c.deleted_at IS NULL
WHERE l.deleted_at IS NULL
  AND l.current_crop_id IS NOT NULL
  AND l.hectares > 0;

CREATE OR REPLACE VIEW v4_calc.field_crop_metrics_aggregated AS
SELECT
  lb.project_id,
  lb.field_id,
  lb.field_name,
  lb.current_crop_id,
  lb.crop_name,
  SUM(lb.hectares)::numeric AS superficie_total,
  SUM(lb.seeded_area_ha)::numeric AS superficie_sembrada_ha,
  SUM(lb.harvested_area_ha)::numeric AS area_cosechada_ha,
  SUM(lb.tons)::numeric AS produccion_tn,
  CASE WHEN SUM(lb.seeded_area_ha) > 0
    THEN SUM(lb.yield_tn_per_ha * lb.seeded_area_ha) / SUM(lb.seeded_area_ha)
    ELSE 0
  END AS rendimiento_tn_ha,
  CASE WHEN SUM(lb.tons) > 0
    THEN SUM(lb.board_price * lb.tons) / SUM(lb.tons)
    ELSE 0
  END AS precio_bruto_usd_tn,
  CASE WHEN SUM(lb.tons) > 0
    THEN SUM(lb.freight_cost * lb.tons) / SUM(lb.tons)
    ELSE 0
  END AS gasto_flete_usd_tn,
  CASE WHEN SUM(lb.tons) > 0
    THEN SUM(lb.commercial_cost * lb.tons) / SUM(lb.tons)
    ELSE 0
  END AS gasto_comercial_usd_tn,
  CASE WHEN SUM(lb.tons) > 0
    THEN SUM(lb.net_price_usd * lb.tons) / SUM(lb.tons)
    ELSE 0
  END AS precio_neto_usd_tn,
  SUM(lb.tons * lb.net_price_usd)::numeric AS ingreso_neto_total
FROM v4_calc.field_crop_metrics_lot_base lb
GROUP BY lb.project_id, lb.field_id, lb.field_name, lb.current_crop_id, lb.crop_name;

CREATE OR REPLACE VIEW v4_calc.dashboard_fertilizers_invested_by_project AS
SELECT
  sm.project_id,
  COALESCE(SUM(sm.quantity * s.price), 0) AS fertilizantes_invertidos_usd
FROM public.supply_movements sm
JOIN public.supplies s ON s.id = sm.supply_id
JOIN public.categories c ON s.category_id = c.id AND c.deleted_at IS NULL
WHERE sm.deleted_at IS NULL
  AND s.deleted_at IS NULL
  AND sm.is_entry = TRUE
  AND sm.movement_type IN ('Stock', 'Remito oficial', 'Movimiento interno', 'Movimiento interno entrada')
  AND c.type_id = 3
GROUP BY sm.project_id;

CREATE OR REPLACE VIEW v4_calc.dashboard_supply_costs_by_project AS
SELECT
  p.id AS project_id,
  COALESCE(SUM(v4_ssot.supply_cost_for_lot_by_category(l.id, 'Semilla')), 0) AS semillas_ejecutados_usd,
  COALESCE(SUM(
    v4_ssot.supply_cost_for_lot_base(l.id)
    - v4_ssot.supply_cost_for_lot_by_category(l.id, 'Semilla')
    - v4_ssot.supply_cost_for_lot_by_category(l.id, 'Fertilizantes')
  ), 0) AS agroquimicos_ejecutados_usd,
  COALESCE(SUM(v4_ssot.supply_cost_for_lot_by_category(l.id, 'Fertilizantes')), 0) AS fertilizantes_ejecutados_usd
FROM public.projects p
LEFT JOIN public.fields f ON f.project_id = p.id AND f.deleted_at IS NULL
LEFT JOIN public.lots l ON l.field_id = f.id AND l.deleted_at IS NULL
WHERE p.deleted_at IS NULL
GROUP BY p.id;

CREATE OR REPLACE VIEW v4_calc.investor_contribution_categories AS
WITH lot_base AS (
  SELECT
    f.project_id,
    l.id AS lot_id,
    l.hectares,
    COALESCE((
      SELECT SUM(w.effective_area)
      FROM public.workorders w
      JOIN public.labors lab ON w.labor_id = lab.id AND lab.deleted_at IS NULL
      JOIN public.categories cat ON lab.category_id = cat.id AND cat.deleted_at IS NULL
      WHERE w.lot_id = l.id
        AND w.deleted_at IS NULL
        AND cat.name = 'Siembra'
        AND cat.type_id = 4
    ), 0)::numeric AS seeded_area_ha
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  JOIN public.projects p ON p.id = f.project_id AND p.deleted_at IS NULL
  WHERE l.deleted_at IS NULL
),
seed_area AS (
  SELECT
    lb.project_id,
    COALESCE(SUM(lb.seeded_area_ha), 0)::numeric AS total_seeded_area_ha,
    COALESCE(SUM(v4_ssot.rent_fixed_only_for_lot(lb.lot_id) * lb.hectares), 0)::numeric AS rent_capitalizable_total_usd,
    COALESCE(SUM(v4_ssot.admin_cost_per_ha_for_lot(lb.lot_id) * lb.hectares), 0)::numeric AS administration_total_usd
  FROM lot_base lb
  GROUP BY lb.project_id
),
labor_totals AS (
  SELECT
    lb.project_id,
    COALESCE(SUM(CASE WHEN cat.name IN ('Pulverización', 'Otras Labores') THEN lab.price * w.effective_area ELSE 0 END), 0)::numeric AS general_labors_total_usd,
    COALESCE(SUM(CASE WHEN cat.name = 'Siembra' THEN lab.price * w.effective_area ELSE 0 END), 0)::numeric AS sowing_total_usd,
    COALESCE(SUM(CASE WHEN cat.name = 'Riego' THEN lab.price * w.effective_area ELSE 0 END), 0)::numeric AS irrigation_total_usd
  FROM lot_base lb
  JOIN public.workorders w ON w.lot_id = lb.lot_id AND w.deleted_at IS NULL
  JOIN public.labors lab ON lab.id = w.labor_id AND lab.deleted_at IS NULL AND lab.deleted_at IS NULL
  JOIN public.categories cat ON cat.id = lab.category_id AND cat.deleted_at IS NULL
  WHERE cat.type_id = 4
  GROUP BY lb.project_id
),
invested_totals AS (
  SELECT
    p.project_id,
    v4_ssot.seeds_invested_for_project_mb(p.project_id)::numeric AS seeds_total_usd,
    v4_ssot.agrochemicals_invested_for_project_mb(p.project_id)::numeric AS agrochemicals_total_usd,
    (
      SELECT COALESCE(SUM(sm.quantity * s.price), 0)::numeric
      FROM public.supply_movements sm
      JOIN public.supplies s ON s.id = sm.supply_id AND s.deleted_at IS NULL
      JOIN public.categories cat ON cat.id = s.category_id AND cat.deleted_at IS NULL
      WHERE sm.project_id = p.project_id
        AND sm.deleted_at IS NULL
        AND sm.is_entry = TRUE
        AND sm.movement_type IN ('Stock', 'Remito oficial', 'Movimiento interno', 'Movimiento interno entrada')
        AND cat.type_id = 3
    ) AS fertilizers_total_usd
  FROM (SELECT DISTINCT project_id FROM lot_base) p
)
SELECT
  it.project_id,
  it.agrochemicals_total_usd,
  COALESCE(it.fertilizers_total_usd, 0)::numeric AS fertilizers_total_usd,
  it.seeds_total_usd,
  COALESCE(lt.general_labors_total_usd, 0)::numeric AS general_labors_total_usd,
  COALESCE(lt.sowing_total_usd, 0)::numeric AS sowing_total_usd,
  COALESCE(lt.irrigation_total_usd, 0)::numeric AS irrigation_total_usd,
  COALESCE(sa.rent_capitalizable_total_usd, 0)::numeric AS rent_capitalizable_total_usd,
  COALESCE(sa.administration_total_usd, 0)::numeric AS administration_total_usd,
  COALESCE(sa.total_seeded_area_ha, 0)::numeric AS total_seeded_area_ha
FROM invested_totals it
LEFT JOIN labor_totals lt ON lt.project_id = it.project_id
LEFT JOIN seed_area sa ON sa.project_id = it.project_id;

CREATE OR REPLACE VIEW v4_calc.investor_real_contributions AS
WITH investor_base AS (
  
  SELECT
    pi.project_id,
    pi.investor_id,
    i.name AS investor_name,
    pi.percentage AS share_pct_agreed
  FROM public.project_investors pi
  JOIN public.investors i ON i.id = pi.investor_id AND i.deleted_at IS NULL
  WHERE pi.deleted_at IS NULL
),
admin_config AS (
  
  SELECT
    project_id,
    COUNT(*) FILTER (WHERE deleted_at IS NULL) > 0 AS has_custom_admin
  FROM public.admin_cost_investors
  GROUP BY project_id
),
category_totals AS (
  
  SELECT
    project_id,
    agrochemicals_total_usd,
    fertilizers_total_usd,
    seeds_total_usd,
    general_labors_total_usd,
    sowing_total_usd,
    irrigation_total_usd,
    rent_capitalizable_total_usd,
    administration_total_usd,
    total_seeded_area_ha,
    (agrochemicals_total_usd + fertilizers_total_usd + seeds_total_usd +
     general_labors_total_usd + sowing_total_usd + irrigation_total_usd +
     rent_capitalizable_total_usd + administration_total_usd) AS total_contributions_usd
  FROM v4_calc.investor_contribution_categories
),



investor_agrochemicals_real AS (
  SELECT
    sm.project_id,
    sm.investor_id,
    COALESCE(SUM(sm.quantity * s.price), 0)::numeric AS agrochemicals_real_usd
  FROM public.supply_movements sm
  JOIN public.supplies s ON s.id = sm.supply_id AND s.deleted_at IS NULL
  JOIN public.categories cat ON cat.id = s.category_id AND cat.deleted_at IS NULL
  WHERE sm.deleted_at IS NULL
    AND sm.is_entry = TRUE
    AND sm.movement_type IN ('Stock', 'Remito oficial', 'Movimiento interno', 'Movimiento interno entrada')
    AND cat.type_id = 2
  GROUP BY sm.project_id, sm.investor_id
),
investor_fertilizers_real AS (
  SELECT
    sm.project_id,
    sm.investor_id,
    COALESCE(SUM(sm.quantity * s.price), 0)::numeric AS fertilizers_real_usd
  FROM public.supply_movements sm
  JOIN public.supplies s ON s.id = sm.supply_id AND s.deleted_at IS NULL
  JOIN public.categories cat ON cat.id = s.category_id AND cat.deleted_at IS NULL
  WHERE sm.deleted_at IS NULL
    AND sm.is_entry = TRUE
    AND sm.movement_type IN ('Stock', 'Remito oficial', 'Movimiento interno', 'Movimiento interno entrada')
    AND cat.type_id = 3
  GROUP BY sm.project_id, sm.investor_id
),
investor_seeds_real AS (
  SELECT
    sm.project_id,
    sm.investor_id,
    COALESCE(SUM(sm.quantity * s.price), 0)::numeric AS seeds_real_usd
  FROM public.supply_movements sm
  JOIN public.supplies s ON s.id = sm.supply_id AND s.deleted_at IS NULL
  JOIN public.categories cat ON cat.id = s.category_id AND cat.deleted_at IS NULL
  WHERE sm.deleted_at IS NULL
    AND sm.is_entry = TRUE
    AND sm.movement_type IN ('Stock', 'Remito oficial', 'Movimiento interno', 'Movimiento interno entrada')
    AND cat.type_id = 1
  GROUP BY sm.project_id, sm.investor_id
),



investor_general_labors_real AS (
  SELECT
    w.project_id,
    w.investor_id,
    COALESCE(SUM(lab.price * w.effective_area), 0)::numeric AS general_labors_real_usd
  FROM public.workorders w
  JOIN public.labors lab ON w.labor_id = lab.id AND lab.deleted_at IS NULL AND lab.deleted_at IS NULL
  JOIN public.categories cat ON cat.id = lab.category_id AND cat.deleted_at IS NULL
  WHERE w.deleted_at IS NULL
    AND cat.type_id = 4
    AND cat.name IN ('Pulverización', 'Otras Labores')
  GROUP BY w.project_id, w.investor_id
),
investor_sowing_real AS (
  SELECT
    w.project_id,
    w.investor_id,
    COALESCE(SUM(lab.price * w.effective_area), 0)::numeric AS sowing_real_usd
  FROM public.workorders w
  JOIN public.labors lab ON w.labor_id = lab.id AND lab.deleted_at IS NULL AND lab.deleted_at IS NULL
  JOIN public.categories cat ON cat.id = lab.category_id AND cat.deleted_at IS NULL
  WHERE w.deleted_at IS NULL
    AND cat.type_id = 4
    AND cat.name = 'Siembra'
  GROUP BY w.project_id, w.investor_id
),
investor_irrigation_real AS (
  SELECT
    w.project_id,
    w.investor_id,
    COALESCE(SUM(lab.price * w.effective_area), 0)::numeric AS irrigation_real_usd
  FROM public.workorders w
  JOIN public.labors lab ON w.labor_id = lab.id AND lab.deleted_at IS NULL AND lab.deleted_at IS NULL
  JOIN public.categories cat ON cat.id = lab.category_id AND cat.deleted_at IS NULL
  WHERE w.deleted_at IS NULL
    AND cat.type_id = 4
    AND cat.name = 'Riego'
  GROUP BY w.project_id, w.investor_id
),



investor_rent_real AS (
  SELECT
    f.project_id,
    fi.investor_id,
    SUM(
      v4_ssot.rent_fixed_only_for_lot(l.id) * l.hectares *
      COALESCE(fi.percentage, 0)::numeric / 100
    )::numeric AS rent_real_usd
  FROM public.fields f
  JOIN public.lots l ON l.field_id = f.id AND l.deleted_at IS NULL
  LEFT JOIN public.field_investors fi ON fi.field_id = f.id AND fi.deleted_at IS NULL
  WHERE f.deleted_at IS NULL
  GROUP BY f.project_id, fi.investor_id
),



investor_admin_real AS (
  SELECT
    ib.project_id,
    ib.investor_id,
    (
      CASE
        WHEN COALESCE(ac.has_custom_admin, FALSE)
          THEN (ct.administration_total_usd * COALESCE(aci.percentage, 0)::numeric / 100)
        ELSE (ct.administration_total_usd * ib.share_pct_agreed / 100)
      END
    )::numeric AS admin_real_usd
  FROM investor_base ib
  JOIN category_totals ct ON ct.project_id = ib.project_id
  LEFT JOIN admin_config ac ON ac.project_id = ib.project_id
  LEFT JOIN public.admin_cost_investors aci
    ON aci.project_id = ib.project_id
   AND aci.investor_id = ib.investor_id
   AND aci.deleted_at IS NULL
)



SELECT
  ib.project_id,
  ib.investor_id,
  ib.investor_name,
  ib.share_pct_agreed,
  
  COALESCE(agro.agrochemicals_real_usd, 0)::numeric AS agrochemicals_real_usd,
  COALESCE(fert.fertilizers_real_usd, 0)::numeric AS fertilizers_real_usd,
  COALESCE(seed.seeds_real_usd, 0)::numeric AS seeds_real_usd,
  COALESCE(glabor.general_labors_real_usd, 0)::numeric AS general_labors_real_usd,
  COALESCE(sow.sowing_real_usd, 0)::numeric AS sowing_real_usd,
  COALESCE(irrig.irrigation_real_usd, 0)::numeric AS irrigation_real_usd,
  COALESCE(rri.rent_real_usd, 0)::numeric AS rent_real_usd,
  COALESCE(ia.admin_real_usd, 0)::numeric AS administration_real_usd,
  
  (
    COALESCE(agro.agrochemicals_real_usd, 0) +
    COALESCE(fert.fertilizers_real_usd, 0) +
    COALESCE(seed.seeds_real_usd, 0) +
    COALESCE(glabor.general_labors_real_usd, 0) +
    COALESCE(sow.sowing_real_usd, 0) +
    COALESCE(irrig.irrigation_real_usd, 0) +
    COALESCE(rri.rent_real_usd, 0) +
    COALESCE(ia.admin_real_usd, 0)
  )::numeric AS total_real_contribution_usd,
  
  ct.total_contributions_usd AS project_total_contributions_usd,
  
  CASE
    WHEN ct.total_contributions_usd > 0
    THEN (
      (
        COALESCE(agro.agrochemicals_real_usd, 0) +
        COALESCE(fert.fertilizers_real_usd, 0) +
        COALESCE(seed.seeds_real_usd, 0) +
        COALESCE(glabor.general_labors_real_usd, 0) +
        COALESCE(sow.sowing_real_usd, 0) +
        COALESCE(irrig.irrigation_real_usd, 0) +
        COALESCE(rri.rent_real_usd, 0) +
        COALESCE(ia.admin_real_usd, 0)
      ) / ct.total_contributions_usd * 100
    )::numeric
    ELSE 0::numeric
  END AS contributions_progress_pct
FROM investor_base ib
JOIN category_totals ct ON ct.project_id = ib.project_id
LEFT JOIN investor_agrochemicals_real agro
  ON agro.project_id = ib.project_id AND agro.investor_id = ib.investor_id
LEFT JOIN investor_fertilizers_real fert
  ON fert.project_id = ib.project_id AND fert.investor_id = ib.investor_id
LEFT JOIN investor_seeds_real seed
  ON seed.project_id = ib.project_id AND seed.investor_id = ib.investor_id
LEFT JOIN investor_general_labors_real glabor
  ON glabor.project_id = ib.project_id AND glabor.investor_id = ib.investor_id
LEFT JOIN investor_sowing_real sow
  ON sow.project_id = ib.project_id AND sow.investor_id = ib.investor_id
LEFT JOIN investor_irrigation_real irrig
  ON irrig.project_id = ib.project_id AND irrig.investor_id = ib.investor_id
LEFT JOIN investor_rent_real rri
  ON rri.project_id = ib.project_id AND rri.investor_id = ib.investor_id
LEFT JOIN investor_admin_real ia
  ON ia.project_id = ib.project_id AND ia.investor_id = ib.investor_id;

COMMIT;
