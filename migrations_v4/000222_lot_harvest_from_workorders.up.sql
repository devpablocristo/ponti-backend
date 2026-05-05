BEGIN;

CREATE OR REPLACE VIEW v4_calc.lot_base_costs AS
WITH
raw AS (
  SELECT
    f.project_id,
    f.id AS field_id,
    f.lease_type_id,
    f.lease_type_percent,
    f.lease_type_value,
    COALESCE(p.admin_cost, 0)::numeric AS admin_cost_per_ha_usd,
    l.current_crop_id,
    l.id AS lot_id,
    l.name AS lot_name,
    COALESCE(l.hectares, 0)::numeric AS hectares,
    COALESCE(l.tons, 0)::numeric AS tons,
    l.sowing_date,
    COALESCE(cc.net_price, 0)::numeric AS net_price_usd_tn
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  JOIN public.projects p ON p.id = f.project_id AND p.deleted_at IS NULL
  LEFT JOIN public.crop_commercializations cc
    ON cc.project_id = f.project_id
   AND cc.crop_id = l.current_crop_id
   AND cc.deleted_at IS NULL
  WHERE l.deleted_at IS NULL
),
workorder_base AS (
  SELECT
    w.id AS workorder_id,
    w.lot_id,
    w.effective_area,
    COALESCE(lb.price, 0)::numeric AS labor_price,
    cat.name AS labor_category_name,
    cat.type_id AS labor_category_type_id
  FROM public.workorders w
  JOIN raw r ON r.lot_id = w.lot_id
  JOIN public.labors lb ON lb.id = w.labor_id
  LEFT JOIN public.categories cat ON cat.id = lb.category_id
  WHERE w.deleted_at IS NULL
    AND w.effective_area IS NOT NULL
    AND w.effective_area > 0
),
workorder_totals AS (
  SELECT
    lot_id,
    SUM(
      CASE
        WHEN labor_category_name = 'Siembra' AND labor_category_type_id = 4
        THEN effective_area
        ELSE 0
      END
    )::numeric AS seeded_area_ha,
    SUM(
      CASE
        WHEN labor_category_name = 'Cosecha' AND labor_category_type_id = 4
        THEN effective_area
        ELSE 0
      END
    )::numeric AS harvested_area_ha,
    SUM(labor_price * effective_area)::numeric AS labor_cost_usd
  FROM workorder_base
  GROUP BY lot_id
),
supply_totals AS (
  SELECT
    wb.lot_id,
    SUM(COALESCE(wi.total_used, 0) * COALESCE(s.price, 0))::numeric AS supplies_cost_usd
  FROM workorder_base wb
  JOIN public.workorder_items wi ON wi.workorder_id = wb.workorder_id AND wi.deleted_at IS NULL
  LEFT JOIN public.supplies s ON s.id = wi.supply_id
  GROUP BY wb.lot_id
),
base AS (
  SELECT
    r.project_id,
    r.field_id,
    r.current_crop_id,
    r.lot_id,
    r.lot_name,
    r.hectares,
    r.tons,
    r.sowing_date,
    r.lease_type_id,
    r.lease_type_percent,
    r.lease_type_value,
    r.admin_cost_per_ha_usd,
    COALESCE(wt.seeded_area_ha, 0)::numeric AS seeded_area_ha,
    COALESCE(wt.harvested_area_ha, 0)::numeric AS harvested_area_ha,
    v4_core.per_ha(r.tons, NULLIF(COALESCE(wt.harvested_area_ha, 0), 0))::numeric AS yield_tn_per_ha,
    COALESCE(wt.labor_cost_usd, 0)::numeric AS labor_cost_usd,
    COALESCE(st.supplies_cost_usd, 0)::numeric AS supplies_cost_usd,
    (COALESCE(wt.labor_cost_usd, 0) + COALESCE(st.supplies_cost_usd, 0))::numeric AS direct_cost_usd,
    (r.tons * r.net_price_usd_tn)::numeric AS income_net_total_usd
  FROM raw r
  LEFT JOIN workorder_totals wt ON wt.lot_id = r.lot_id
  LEFT JOIN supply_totals st ON st.lot_id = r.lot_id
),
derived AS (
  SELECT
    b.*,
    v4_core.per_ha(b.income_net_total_usd, b.hectares)::numeric AS income_net_per_ha_usd,
    v4_core.per_ha(b.direct_cost_usd, b.hectares)::numeric AS direct_cost_per_ha_usd
  FROM base b
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
  income_net_per_ha_usd,
  direct_cost_per_ha_usd,
  v4_core.rent_per_ha(
    lease_type_id,
    lease_type_percent,
    lease_type_value,
    income_net_per_ha_usd,
    direct_cost_per_ha_usd,
    admin_cost_per_ha_usd
  )::numeric AS rent_per_ha_usd,
  CASE
    WHEN lease_type_id IN (3, 4) THEN COALESCE(lease_type_value, 0)
    ELSE 0
  END::numeric AS rent_fixed_per_ha_usd,
  admin_cost_per_ha_usd,
  seeded_area_ha AS sowed_area_ha,
  seeded_area_ha AS sown_area_ha
FROM derived;

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
  COALESCE(lm.yield_tn_per_ha, 0)::numeric AS yield_tn_per_ha,
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
  harvest_workorder.harvest_date AS lot_harvest_date,
  l.tons,
  (
    SELECT MIN(w.date)
    FROM public.workorders w
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
LEFT JOIN LATERAL (
  SELECT MAX(w.date) AS harvest_date
  FROM public.workorders w
  JOIN public.labors lb ON lb.id = w.labor_id
  JOIN public.categories cat ON cat.id = lb.category_id
  WHERE w.lot_id = l.id
    AND w.deleted_at IS NULL
    AND cat.type_id = 4
    AND cat.name = 'Cosecha'
) harvest_workorder ON true
WHERE l.deleted_at IS NULL;

COMMIT;
