-- ========================================
-- MIGRATION 000140 COMPAT SEEDED AREA ALIASES (UP)
-- ========================================
-- Nota: SQL en inglés, comentarios en español

BEGIN;

-- Compatibilidad: agrega aliases sowed_area_ha y sown_area_ha
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
  admin_cost_per_ha_usd,
  seeded_area_ha AS sowed_area_ha,
  seeded_area_ha AS sown_area_ha
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
  c.income_net_per_ha_usd,
  c.seeded_area_ha AS sowed_area_ha,
  c.seeded_area_ha AS sown_area_ha
FROM v4_calc.lot_base_costs c;

CREATE OR REPLACE VIEW v4_calc.field_crop_lot_base AS
SELECT
  f.project_id,
  f.id AS field_id,
  l.current_crop_id AS crop_id,
  l.id AS lot_id,
  l.hectares AS surface_ha,
  v4_ssot.seeded_area_for_lot(l.id)::numeric AS seeded_area_ha,
  l.tons,
  v4_ssot.seeded_area_for_lot(l.id)::numeric AS sowed_area_ha,
  v4_ssot.seeded_area_for_lot(l.id)::numeric AS sown_area_ha
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
  
  (
    v4_ssot.supply_cost_for_lot_by_category(lot_id, 'Semilla') +
    v4_ssot.supply_cost_for_lot_by_category(lot_id, 'Curasemillas') +
    v4_ssot.supply_cost_for_lot_by_category(lot_id, 'Herbicidas') +
    v4_ssot.supply_cost_for_lot_by_category(lot_id, 'Insecticidas') +
    v4_ssot.supply_cost_for_lot_by_category(lot_id, 'Fungicidas') +
    v4_ssot.supply_cost_for_lot_by_category(lot_id, 'Coadyuvantes') +
    v4_ssot.supply_cost_for_lot_by_category(lot_id, 'Fertilizantes') +
    v4_ssot.supply_cost_for_lot_by_category(lot_id, 'Otros Insumos')
  )::numeric AS total_insumos_usd,
  seeded_area_ha AS sowed_area_ha,
  seeded_area_ha AS sown_area_ha
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
    JOIN public.labors lab ON lab.id = w.labor_id
    JOIN public.categories cat ON cat.id = lab.category_id
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
    JOIN public.labors lab ON lab.id = w.labor_id
    JOIN public.categories cat ON cat.id = lab.category_id
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
    JOIN public.labors lab ON lab.id = w.labor_id
    JOIN public.categories cat ON cat.id = lab.category_id
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
    JOIN public.labors lab ON lab.id = w.labor_id
    JOIN public.categories cat ON cat.id = lab.category_id
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
    JOIN public.labors lab ON lab.id = w.labor_id
    JOIN public.categories cat ON cat.id = lab.category_id
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
      JOIN public.labors lab ON lab.id = w.labor_id
      JOIN public.categories cat ON cat.id = lab.category_id
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
      JOIN public.labors lab ON lab.id = w.labor_id
      JOIN public.categories cat ON cat.id = lab.category_id
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
      JOIN public.labors lab ON lab.id = w.labor_id
      JOIN public.categories cat ON cat.id = lab.category_id
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
      JOIN public.labors lab ON lab.id = w.labor_id
      JOIN public.categories cat ON cat.id = lab.category_id
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
      JOIN public.labors lab ON lab.id = w.labor_id
      JOIN public.categories cat ON cat.id = lab.category_id
      WHERE w.lot_id = lb.lot_id
        AND w.deleted_at IS NULL
        AND w.effective_area > 0
        AND lab.price IS NOT NULL
        AND cat.name NOT IN ('Siembra', 'Pulverización', 'Riego', 'Cosecha')
        AND cat.type_id = 4
    ), 0)::numeric
  ) AS total_labores_usd,
  lb.seeded_area_ha AS sowed_area_ha,
  lb.seeded_area_ha AS sown_area_ha
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
  SUM(v4_ssot.admin_cost_prorated_per_ha_for_lot(lot_id) * surface_ha)::numeric AS administration_usd,
  SUM(seeded_area_ha)::numeric AS sowed_area_ha,
  SUM(seeded_area_ha)::numeric AS sown_area_ha
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
  COALESCE(v4_ssot.admin_cost_prorated_per_ha_for_lot(l.id), 0)::numeric AS admin_per_ha,
  COALESCE(v4_ssot.board_price_for_lot(l.id), 0)::numeric AS board_price,
  COALESCE(v4_ssot.freight_cost_for_lot(l.id), 0)::numeric AS freight_cost,
  COALESCE(v4_ssot.commercial_cost_for_lot(l.id), 0)::numeric AS commercial_cost,
  COALESCE(v4_ssot.seeded_area_for_lot(l.id), 0)::numeric AS sowed_area_ha,
  COALESCE(v4_ssot.seeded_area_for_lot(l.id), 0)::numeric AS sown_area_ha
FROM public.lots l
JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
JOIN public.projects p ON p.id = f.project_id AND p.deleted_at IS NULL
LEFT JOIN public.crops c ON c.id = l.current_crop_id AND c.deleted_at IS NULL
WHERE l.deleted_at IS NULL
  AND l.current_crop_id IS NOT NULL
  AND l.hectares > 0;

CREATE OR REPLACE VIEW v4_report.lot_metrics AS
WITH 
project_totals AS (
  SELECT project_id, SUM(hectares)::numeric AS total_hectares
  FROM v4_calc.lot_base_costs GROUP BY project_id
),
field_totals AS (
  SELECT field_id, SUM(hectares)::numeric AS total_hectares
  FROM v4_calc.lot_base_costs GROUP BY field_id
)
SELECT
  c.project_id,
  c.field_id,
  c.lot_id,
  c.lot_name,
  c.hectares,
  c.seeded_area_ha,
  c.harvested_area_ha,
  c.yield_tn_per_ha,
  c.tons,
  c.sowing_date,
  c.labor_cost_usd,
  c.supplies_cost_usd AS supply_cost_usd,
  c.direct_cost_usd,
  c.income_net_total_usd,
  c.income_net_per_ha_usd,
  c.rent_per_ha_usd,
  c.admin_cost_per_ha_usd,
  c.direct_cost_per_ha_usd,
  
  (c.direct_cost_per_ha_usd + c.rent_per_ha_usd + c.admin_cost_per_ha_usd)::numeric AS active_total_per_ha_usd,
  (c.income_net_per_ha_usd - (c.direct_cost_per_ha_usd + c.rent_per_ha_usd + c.admin_cost_per_ha_usd))::numeric AS operating_result_per_ha_usd,
  (c.rent_per_ha_usd * c.hectares)::numeric AS rent_total_usd,
  (c.admin_cost_per_ha_usd * c.hectares)::numeric AS admin_total_usd,
  ((c.direct_cost_per_ha_usd + c.rent_per_ha_usd + c.admin_cost_per_ha_usd) * c.hectares)::numeric AS active_total_usd,
  ((c.income_net_per_ha_usd - (c.direct_cost_per_ha_usd + c.rent_per_ha_usd + c.admin_cost_per_ha_usd)) * c.hectares)::numeric AS operating_result_total_usd,
  c.direct_cost_usd AS direct_cost_total_usd,
  COALESCE(pt.total_hectares, 0) AS project_total_hectares,
  COALESCE(ft.total_hectares, 0) AS field_total_hectares,
  c.seeded_area_ha AS sowed_area_ha,
  c.seeded_area_ha AS sown_area_ha
FROM v4_calc.lot_base_costs c
LEFT JOIN project_totals pt ON pt.project_id = c.project_id
LEFT JOIN field_totals ft ON ft.field_id = c.field_id;

CREATE OR REPLACE VIEW v4_report.lot_list AS
SELECT
  f.project_id,
  p.name AS project_name,
  f.id AS field_id,
  f.name AS field_name,
  l.id AS id,
  l.name AS lot_name,
  l.variety,
  l.season,
  l.previous_crop_id,
  prev_crop.name AS previous_crop,
  l.current_crop_id,
  curr_crop.name AS current_crop,
  l.hectares,
  l.updated_at,
  COALESCE(lm.seeded_area_ha, 0)::numeric AS seeded_area_ha,
  COALESCE(lm.harvested_area_ha, 0)::numeric AS harvested_area_ha,
  COALESCE(lm.yield_tn_per_ha, 0) AS yield_tn_per_ha,
  COALESCE(lm.direct_cost_per_ha_usd, 0)::numeric AS cost_usd_per_ha,
  COALESCE(lm.income_net_per_ha_usd, 0)::numeric AS income_net_per_ha_usd,
  COALESCE(lm.rent_per_ha_usd, 0)::numeric AS rent_per_ha_usd,
  COALESCE(lm.admin_cost_per_ha_usd, 0)::numeric AS admin_cost_per_ha_usd,
  COALESCE(lm.active_total_per_ha_usd, 0)::numeric AS active_total_per_ha_usd,
  COALESCE(lm.operating_result_per_ha_usd, 0)::numeric AS operating_result_per_ha_usd,
  COALESCE(lm.income_net_total_usd, 0)::numeric AS income_net_total_usd,
  COALESCE(lm.direct_cost_total_usd, 0)::numeric AS direct_cost_total_usd,
  COALESCE(lm.rent_total_usd, 0)::numeric AS rent_total_usd,
  COALESCE(lm.admin_total_usd, 0)::numeric AS admin_total_usd,
  COALESCE(lm.active_total_usd, 0)::numeric AS active_total_usd,
  COALESCE(lm.operating_result_total_usd, 0)::numeric AS operating_result_total_usd,
  l.sowing_date AS lot_sowing_date,
  NULL::date AS lot_harvest_date,
  l.tons,
  (
    SELECT MIN(w.date)
    FROM public.workorders w
    JOIN public.labors lb ON lb.id = w.labor_id AND lb.deleted_at IS NULL
    WHERE w.lot_id = l.id AND w.deleted_at IS NULL
  ) AS raw_sowing_date,
  COALESCE(lm.seeded_area_ha, 0)::numeric AS sowed_area_ha,
  COALESCE(lm.seeded_area_ha, 0)::numeric AS sown_area_ha
FROM public.lots l
JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
JOIN public.projects p ON p.id = f.project_id AND p.deleted_at IS NULL
LEFT JOIN public.crops prev_crop ON prev_crop.id = l.previous_crop_id AND prev_crop.deleted_at IS NULL
LEFT JOIN public.crops curr_crop ON curr_crop.id = l.current_crop_id AND curr_crop.deleted_at IS NULL
LEFT JOIN v4_report.lot_metrics lm ON lm.lot_id = l.id
WHERE l.deleted_at IS NULL;

CREATE OR REPLACE VIEW v4_report.dashboard_metrics AS
WITH lot_data AS (
  SELECT
    lm.project_id,
    lm.lot_id,
    lm.hectares,
    lm.seeded_area_ha,
    lm.harvested_area_ha,
    lm.direct_cost_per_ha_usd
  FROM v4_report.lot_metrics lm
),
project_hectares AS (
  SELECT
    project_id,
    SUM(hectares) AS total_hectares
  FROM v4_report.lot_metrics
  GROUP BY project_id
),
rent_fixed_ssot AS (
  SELECT
    f.project_id,
    SUM(v4_ssot.rent_fixed_only_for_lot(l.id) * l.hectares)::numeric AS rent_fixed_total_usd
  FROM public.fields f
  JOIN public.lots l ON l.field_id = f.id AND l.deleted_at IS NULL
  WHERE f.deleted_at IS NULL
  GROUP BY f.project_id
)
SELECT
  p.customer_id,
  p.id AS project_id,
  p.campaign_id,
  
  
  COALESCE(SUM(ld.seeded_area_ha), 0)::numeric AS sowing_hectares,
  COALESCE(SUM(ld.hectares), 0)::numeric AS sowing_total_hectares,
  v4_core.percentage(
    COALESCE(SUM(ld.seeded_area_ha), 0)::numeric,
    COALESCE(SUM(ld.hectares), 0)::numeric
  ) AS sowing_progress_pct,
  
  
  COALESCE(SUM(ld.harvested_area_ha), 0)::numeric AS harvest_hectares,
  COALESCE(SUM(ld.hectares), 0)::numeric AS harvest_total_hectares,
  v4_core.percentage(
    COALESCE(SUM(ld.harvested_area_ha), 0)::numeric,
    COALESCE(SUM(ld.hectares), 0)::numeric
  ) AS harvest_progress_pct,
  
  
  COALESCE(
    SUM(ld.direct_cost_per_ha_usd * ld.seeded_area_ha) / NULLIF(SUM(ld.seeded_area_ha), 0),
    0
  )::numeric AS executed_costs_usd,
  COALESCE(p.planned_cost, 0)::numeric AS budget_cost_usd,
  v4_core.percentage(
    COALESCE(
      SUM(ld.direct_cost_per_ha_usd * ld.seeded_area_ha) / NULLIF(SUM(ld.seeded_area_ha), 0),
      0
    )::numeric,
    COALESCE(p.planned_cost, 0)::numeric
  ) AS costs_progress_pct,
  
  
  COALESCE(SUM(v4_ssot.income_net_total_for_lot(ld.lot_id)), 0) AS operating_result_income_usd,
  v4_ssot.operating_result_total_for_project(p.id) AS operating_result_usd,
  (
    COALESCE(v4_ssot.direct_costs_total_for_project(p.id), 0) +
    COALESCE(p.admin_cost * ph.total_hectares, 0) +
    COALESCE(rfs.rent_fixed_total_usd, 0)
  )::numeric AS operating_result_total_costs_usd,
  v4_ssot.renta_pct(
    v4_ssot.operating_result_total_for_project(p.id),
    (
      COALESCE(v4_ssot.direct_costs_total_for_project(p.id), 0) +
      COALESCE(p.admin_cost * ph.total_hectares, 0) +
      COALESCE(rfs.rent_fixed_total_usd, 0)
    )::numeric
  ) AS operating_result_pct,
  
  COALESCE(ph.total_hectares, 0)::numeric AS project_total_hectares
  
FROM public.projects p
LEFT JOIN lot_data ld ON ld.project_id = p.id
LEFT JOIN project_hectares ph ON ph.project_id = p.id
LEFT JOIN rent_fixed_ssot rfs ON rfs.project_id = p.id
WHERE p.deleted_at IS NULL
GROUP BY p.customer_id, p.id, p.campaign_id, p.admin_cost, p.planned_cost, ph.total_hectares, rfs.rent_fixed_total_usd;

COMMIT;
