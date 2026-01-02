-- =============================================================================
-- Migration: 000325_optimize_summary_results_ssot
-- Descripción: Optimiza summary_results para usar field_crop_metrics sin timeout
-- Técnica: CTE MATERIALIZED fuerza a PostgreSQL a calcular primero
-- =============================================================================

DROP VIEW IF EXISTS v4_report.summary_results;

CREATE VIEW v4_report.summary_results AS
WITH 
-- MATERIALIZED fuerza a PostgreSQL a materializar este CTE primero
-- evitando la reevaluación de funciones SSOT
field_crop_data AS MATERIALIZED (
  SELECT
    project_id,
    current_crop_id,
    crop_name,
    area_sembrada_ha,
    ingreso_neto_usd,
    total_costos_directos_usd,
    arriendo_usd,
    administracion_usd,
    total_invertido_usd,
    resultado_operativo_usd
  FROM v4_report.field_crop_metrics
  WHERE current_crop_id IS NOT NULL
),
by_crop AS (
  SELECT
    project_id,
    current_crop_id,
    crop_name,
    SUM(area_sembrada_ha)::numeric AS surface_ha,
    SUM(ingreso_neto_usd)::numeric AS net_income_usd,
    SUM(total_costos_directos_usd)::numeric AS direct_costs_usd,
    SUM(arriendo_usd)::numeric AS rent_usd,
    SUM(administracion_usd)::numeric AS structure_usd,
    SUM(total_invertido_usd)::numeric AS total_invested_usd,
    SUM(resultado_operativo_usd)::numeric AS operating_result_usd
  FROM field_crop_data
  GROUP BY project_id, current_crop_id, crop_name
),
project_totals AS (
  SELECT
    project_id,
    SUM(surface_ha)::numeric AS total_surface_ha,
    SUM(net_income_usd)::numeric AS total_net_income_usd,
    SUM(direct_costs_usd)::numeric AS total_direct_costs_usd,
    SUM(rent_usd)::numeric AS total_rent_usd,
    SUM(structure_usd)::numeric AS total_structure_usd,
    SUM(total_invested_usd)::numeric AS total_invested_project_usd,
    SUM(operating_result_usd)::numeric AS total_operating_result_usd
  FROM by_crop
  GROUP BY project_id
)
SELECT
  bc.project_id,
  bc.current_crop_id,
  bc.crop_name,
  bc.surface_ha,
  bc.net_income_usd,
  bc.direct_costs_usd,
  bc.rent_usd,
  bc.structure_usd,
  bc.total_invested_usd,
  bc.operating_result_usd,
  CASE WHEN bc.total_invested_usd > 0 
    THEN (bc.operating_result_usd / bc.total_invested_usd * 100)::numeric
    ELSE 0::numeric 
  END AS crop_return_pct,
  pt.total_surface_ha,
  pt.total_net_income_usd,
  pt.total_direct_costs_usd,
  pt.total_rent_usd,
  pt.total_structure_usd,
  pt.total_invested_project_usd,
  pt.total_operating_result_usd,
  CASE WHEN pt.total_invested_project_usd > 0 
    THEN (pt.total_operating_result_usd / pt.total_invested_project_usd * 100)::numeric
    ELSE 0::numeric 
  END AS project_return_pct
FROM by_crop bc
JOIN project_totals pt ON pt.project_id = bc.project_id;

COMMENT ON VIEW v4_report.summary_results IS 
'SSOT: Agrega desde field_crop_metrics. CTE MATERIALIZED evita timeout.';
