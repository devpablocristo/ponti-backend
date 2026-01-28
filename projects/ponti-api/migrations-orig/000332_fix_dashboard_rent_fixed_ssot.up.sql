-- =============================================================================
-- MIGRACIÓN 000332: Dashboard arriendo fijo desde SSOT de lotes
-- =============================================================================
--
-- Propósito: Corregir arriendo fijo en proyectos mixtos
-- Regla: usar SUM(rent_fixed_only_for_lot * hectares) como SSOT
-- Impacto: dashboard_metrics (operating_result_total_costs_usd, operating_result_pct)
-- Nota: Comentarios en español, código en inglés
--

BEGIN;

CREATE OR REPLACE VIEW v4_report.dashboard_metrics AS
WITH lot_data AS (
  SELECT
    lm.project_id,
    lm.lot_id,
    lm.hectares,
    lm.sowed_area_ha,
    lm.harvested_area_ha,
    lm.direct_cost_per_ha_usd
  FROM public.v3_lot_metrics lm
),
project_hectares AS (
  SELECT
    project_id,
    SUM(hectares) AS total_hectares
  FROM public.v3_lot_metrics
  GROUP BY project_id
),
rent_fixed_ssot AS (
  SELECT
    f.project_id,
    SUM(v3_lot_ssot.rent_fixed_only_for_lot(l.id) * l.hectares)::double precision AS rent_fixed_total_usd
  FROM public.fields f
  JOIN public.lots l ON l.field_id = f.id AND l.deleted_at IS NULL
  WHERE f.deleted_at IS NULL
  GROUP BY f.project_id
)
SELECT
  p.customer_id,
  p.id AS project_id,
  p.campaign_id,
  
  -- CARD 1: AVANCE DE SIEMBRA
  COALESCE(SUM(ld.sowed_area_ha), 0)::double precision AS sowing_hectares,
  COALESCE(SUM(ld.hectares), 0)::double precision AS sowing_total_hectares,
  v3_core_ssot.percentage(
    COALESCE(SUM(ld.sowed_area_ha), 0)::numeric,
    COALESCE(SUM(ld.hectares), 0)::numeric
  ) AS sowing_progress_pct,
  
  -- CARD 2: AVANCE DE COSECHA
  COALESCE(SUM(ld.harvested_area_ha), 0)::double precision AS harvest_hectares,
  COALESCE(SUM(ld.hectares), 0)::double precision AS harvest_total_hectares,
  v3_core_ssot.percentage(
    COALESCE(SUM(ld.harvested_area_ha), 0)::numeric,
    COALESCE(SUM(ld.hectares), 0)::numeric
  ) AS harvest_progress_pct,
  
  -- CARD 3: AVANCE DE COSTOS
  COALESCE(
    SUM(ld.direct_cost_per_ha_usd * ld.sowed_area_ha) / NULLIF(SUM(ld.sowed_area_ha), 0),
    0
  )::double precision AS executed_costs_usd,
  COALESCE(p.planned_cost, 0)::double precision AS budget_cost_usd,
  v3_core_ssot.percentage(
    COALESCE(
      SUM(ld.direct_cost_per_ha_usd * ld.sowed_area_ha) / NULLIF(SUM(ld.sowed_area_ha), 0),
      0
    )::numeric,
    COALESCE(p.planned_cost, 0)::numeric
  ) AS costs_progress_pct,
  
  -- CARD 4: RESULTADO OPERATIVO
  COALESCE(SUM(v3_lot_ssot.income_net_total_for_lot(ld.lot_id)), 0) AS operating_result_income_usd,
  v3_dashboard_ssot.operating_result_total_for_project(p.id) AS operating_result_usd,
  (
    COALESCE(v3_dashboard_ssot.direct_costs_total_for_project(p.id), 0) +
    COALESCE(p.admin_cost * ph.total_hectares, 0) +
    COALESCE(rfs.rent_fixed_total_usd, 0)
  )::double precision AS operating_result_total_costs_usd,
  v3_lot_ssot.renta_pct(
    v3_dashboard_ssot.operating_result_total_for_project(p.id),
    (
      COALESCE(v3_dashboard_ssot.direct_costs_total_for_project(p.id), 0) +
      COALESCE(p.admin_cost * ph.total_hectares, 0) +
      COALESCE(rfs.rent_fixed_total_usd, 0)
    )::double precision
  ) AS operating_result_pct,
  
  COALESCE(ph.total_hectares, 0)::numeric AS project_total_hectares
  
FROM public.projects p
LEFT JOIN lot_data ld ON ld.project_id = p.id
LEFT JOIN project_hectares ph ON ph.project_id = p.id
LEFT JOIN rent_fixed_ssot rfs ON rfs.project_id = p.id
WHERE p.deleted_at IS NULL
GROUP BY p.customer_id, p.id, p.campaign_id, p.admin_cost, p.planned_cost, ph.total_hectares, rfs.rent_fixed_total_usd;

COMMENT ON VIEW v4_report.dashboard_metrics IS 'Dashboard metrics migrada a v4_report (000332: arriendo fijo desde SSOT).';

COMMIT;
