-- ========================================
-- MIGRATION 000170 FIX DASHBOARD COSTS HECTARES (DOWN)
-- ========================================
-- Revierte al cálculo anterior usando seeded_area_ha

BEGIN;

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
