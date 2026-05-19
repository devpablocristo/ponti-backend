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
), project_hectares AS (
    SELECT
      project_id,
      SUM(hectares) AS total_hectares
    FROM v4_report.lot_metrics
    GROUP BY project_id
), rent_fixed_ssot AS (
    SELECT
      f.project_id,
      SUM(v4_ssot.rent_fixed_only_for_lot(l.id) * l.hectares) AS rent_fixed_total_usd
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
  v4_core.percentage(COALESCE(SUM(ld.seeded_area_ha), 0)::numeric, COALESCE(SUM(ld.hectares), 0)::numeric) AS sowing_progress_pct,
  COALESCE(SUM(ld.harvested_area_ha), 0)::numeric AS harvest_hectares,
  COALESCE(SUM(ld.hectares), 0)::numeric AS harvest_total_hectares,
  v4_core.percentage(COALESCE(SUM(ld.harvested_area_ha), 0)::numeric, COALESCE(SUM(ld.hectares), 0)::numeric) AS harvest_progress_pct,
  COALESCE(SUM(ld.direct_cost_per_ha_usd * ld.seeded_area_ha) / NULLIF(SUM(ld.seeded_area_ha), 0), 0)::numeric AS executed_costs_usd,
  COALESCE(p.planned_cost, 0)::numeric AS budget_cost_usd,
  v4_core.percentage(
    COALESCE(SUM(ld.direct_cost_per_ha_usd * ld.seeded_area_ha) / NULLIF(SUM(ld.seeded_area_ha), 0), 0)::numeric,
    COALESCE(p.planned_cost, 0)::numeric
  ) AS costs_progress_pct,
  COALESCE(SUM(v4_ssot.income_net_total_for_lot(ld.lot_id)), 0)::numeric AS operating_result_income_usd,
  v4_ssot.operating_result_total_for_project(p.id) AS operating_result_usd,
  COALESCE(v4_ssot.direct_costs_total_for_project(p.id), 0)::numeric
    + COALESCE(p.admin_cost * ph.total_hectares, 0)::numeric
    + COALESCE(rfs.rent_fixed_total_usd, 0)::numeric AS operating_result_total_costs_usd,
  v4_ssot.renta_pct(
    v4_ssot.operating_result_total_for_project(p.id),
    COALESCE(v4_ssot.direct_costs_total_for_project(p.id), 0)::numeric
      + COALESCE(p.admin_cost * ph.total_hectares, 0)::numeric
      + COALESCE(rfs.rent_fixed_total_usd, 0)::numeric
  ) AS operating_result_pct,
  COALESCE(ph.total_hectares, 0)::numeric AS project_total_hectares
FROM public.projects p
LEFT JOIN lot_data ld ON ld.project_id = p.id
LEFT JOIN project_hectares ph ON ph.project_id = p.id
LEFT JOIN rent_fixed_ssot rfs ON rfs.project_id = p.id
WHERE p.deleted_at IS NULL
GROUP BY p.customer_id, p.id, p.campaign_id, p.admin_cost, p.planned_cost, ph.total_hectares, rfs.rent_fixed_total_usd;

CREATE OR REPLACE VIEW v4_report.dashboard_metrics_field AS
WITH lot_data AS (
    SELECT
      lm.project_id,
      lm.field_id,
      lm.lot_id,
      lm.hectares,
      lm.seeded_area_ha,
      lm.harvested_area_ha,
      lm.direct_cost_per_ha_usd
    FROM v4_report.lot_metrics lm
), field_hectares AS (
    SELECT
      project_id,
      field_id,
      SUM(hectares) AS total_hectares
    FROM v4_report.lot_metrics
    GROUP BY project_id, field_id
), project_hectares AS (
    SELECT
      project_id,
      SUM(hectares) AS total_hectares
    FROM v4_report.lot_metrics
    GROUP BY project_id
), rent_fixed_ssot AS (
    SELECT
      f.project_id,
      f.id AS field_id,
      SUM(v4_ssot.rent_fixed_only_for_lot(l.id) * l.hectares) AS rent_fixed_total_usd
    FROM public.fields f
    JOIN public.lots l ON l.field_id = f.id AND l.deleted_at IS NULL
    WHERE f.deleted_at IS NULL
    GROUP BY f.project_id, f.id
), direct_costs AS (
    SELECT
      project_id,
      field_id,
      SUM(v4_ssot.direct_cost_for_lot(lot_id)) AS direct_costs_total_usd
    FROM v4_report.lot_metrics
    GROUP BY project_id, field_id
), admin_costs AS (
    SELECT
      project_id,
      field_id,
      SUM(v4_ssot.admin_cost_per_ha_for_lot(lot_id) * hectares) AS admin_total_usd
    FROM v4_report.lot_metrics
    GROUP BY project_id, field_id
), rent_totals AS (
    SELECT
      project_id,
      field_id,
      SUM(v4_ssot.rent_per_ha_for_lot(lot_id) * hectares) AS rent_total_usd
    FROM v4_report.lot_metrics
    GROUP BY project_id, field_id
), income_totals AS (
    SELECT
      project_id,
      field_id,
      SUM(v4_ssot.income_net_total_for_lot(lot_id)) AS income_net_total_usd
    FROM v4_report.lot_metrics
    GROUP BY project_id, field_id
)
SELECT
  p.customer_id,
  p.id AS project_id,
  p.campaign_id,
  ld.field_id,
  COALESCE(SUM(ld.seeded_area_ha), 0)::numeric AS sowing_hectares,
  COALESCE(SUM(ld.hectares), 0)::numeric AS sowing_total_hectares,
  v4_core.percentage(COALESCE(SUM(ld.seeded_area_ha), 0)::numeric, COALESCE(SUM(ld.hectares), 0)::numeric) AS sowing_progress_pct,
  COALESCE(SUM(ld.harvested_area_ha), 0)::numeric AS harvest_hectares,
  COALESCE(SUM(ld.hectares), 0)::numeric AS harvest_total_hectares,
  v4_core.percentage(COALESCE(SUM(ld.harvested_area_ha), 0)::numeric, COALESCE(SUM(ld.hectares), 0)::numeric) AS harvest_progress_pct,
  COALESCE(SUM(ld.direct_cost_per_ha_usd * ld.seeded_area_ha) / NULLIF(SUM(ld.seeded_area_ha), 0), 0)::numeric AS executed_costs_usd,
  COALESCE(p.planned_cost, 0)::numeric * v4_core.safe_div(COALESCE(fh.total_hectares, 0)::numeric, COALESCE(ph.total_hectares, 0)::numeric) AS budget_cost_usd,
  v4_core.percentage(
    COALESCE(SUM(ld.direct_cost_per_ha_usd * ld.seeded_area_ha) / NULLIF(SUM(ld.seeded_area_ha), 0), 0)::numeric,
    COALESCE(p.planned_cost, 0)::numeric * v4_core.safe_div(COALESCE(fh.total_hectares, 0)::numeric, COALESCE(ph.total_hectares, 0)::numeric)
  ) AS costs_progress_pct,
  COALESCE(it.income_net_total_usd, 0)::numeric AS operating_result_income_usd,
  COALESCE(it.income_net_total_usd, 0)::numeric
    - COALESCE(dc.direct_costs_total_usd, 0)::numeric
    - COALESCE(rt.rent_total_usd, 0)::numeric
    - COALESCE(ac.admin_total_usd, 0)::numeric AS operating_result_usd,
  COALESCE(dc.direct_costs_total_usd, 0)::numeric
    + COALESCE(ac.admin_total_usd, 0)::numeric
    + COALESCE(rfs.rent_fixed_total_usd, 0)::numeric AS operating_result_total_costs_usd,
  v4_ssot.renta_pct(
    COALESCE(it.income_net_total_usd, 0)::numeric
      - COALESCE(dc.direct_costs_total_usd, 0)::numeric
      - COALESCE(rt.rent_total_usd, 0)::numeric
      - COALESCE(ac.admin_total_usd, 0)::numeric,
    COALESCE(dc.direct_costs_total_usd, 0)::numeric
      + COALESCE(ac.admin_total_usd, 0)::numeric
      + COALESCE(rfs.rent_fixed_total_usd, 0)::numeric
  ) AS operating_result_pct,
  COALESCE(ph.total_hectares, 0)::numeric AS project_total_hectares
FROM public.projects p
LEFT JOIN lot_data ld ON ld.project_id = p.id
LEFT JOIN field_hectares fh ON fh.project_id = p.id AND fh.field_id = ld.field_id
LEFT JOIN project_hectares ph ON ph.project_id = p.id
LEFT JOIN rent_fixed_ssot rfs ON rfs.project_id = p.id AND rfs.field_id = ld.field_id
LEFT JOIN direct_costs dc ON dc.project_id = p.id AND dc.field_id = ld.field_id
LEFT JOIN admin_costs ac ON ac.project_id = p.id AND ac.field_id = ld.field_id
LEFT JOIN rent_totals rt ON rt.project_id = p.id AND rt.field_id = ld.field_id
LEFT JOIN income_totals it ON it.project_id = p.id AND it.field_id = ld.field_id
WHERE p.deleted_at IS NULL
GROUP BY
  p.customer_id,
  p.id,
  p.campaign_id,
  p.planned_cost,
  ph.total_hectares,
  fh.total_hectares,
  rfs.rent_fixed_total_usd,
  dc.direct_costs_total_usd,
  ac.admin_total_usd,
  rt.rent_total_usd,
  it.income_net_total_usd,
  ld.field_id;

COMMIT;
