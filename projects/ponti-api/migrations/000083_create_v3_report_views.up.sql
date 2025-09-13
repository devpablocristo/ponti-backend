-- ========================================
-- MIGRATION 000083: CREATE v3_report_views (UP)
-- ========================================
-- 
-- Purpose: Create report field crop metrics view
-- Date: 2025-09-12
-- Author: System
-- 
-- Note: Code in English, comments in Spanish.

-- -------------------------------------------------------------------
-- v3_report_field_crop_metrics_view: métricas por field y cultivo
-- -------------------------------------------------------------------
CREATE OR REPLACE VIEW public.v3_report_field_crop_metrics_view AS
WITH lot_base AS (
  SELECT
    l.id AS lot_id,
    f.project_id,
    f.id AS field_id,
    f.name AS field_name,
    l.current_crop_id,
    c.name AS crop_name,
    l.hectares,
    COALESCE(l.tons, 0)::numeric AS tons,
    v3_calc.seeded_area(l.sowing_date, l.hectares::numeric)::double precision AS area_sembrada_ha,
    v3_calc.harvested_area(l.tons::numeric, l.hectares::numeric)::double precision AS area_cosechada_ha
  FROM public.lots l
  JOIN public.fields   f ON f.id = l.field_id AND f.deleted_at IS NULL
  JOIN public.projects p ON p.id = f.project_id AND p.deleted_at IS NULL
  LEFT JOIN public.crops c ON c.id = l.current_crop_id AND c.deleted_at IS NULL
  WHERE l.deleted_at IS NULL AND l.hectares > 0
)
SELECT
  lb.project_id,
  lb.field_id,
  lb.field_name::text AS field_name,
  lb.current_crop_id,
  lb.crop_name::text  AS crop_name,

  COALESCE(SUM(v3_calc.income_net_total_for_lot(lb.lot_id)), 0)             AS income_usd,

  -- Costos directos ejecutados (labores + insumos)
  COALESCE(SUM(v3_calc.direct_cost_for_lot(lb.lot_id)), 0)                  AS costos_directos_ejecutados_usd,

  -- Costos directos invertidos (labores+insumos+arriendo+estructura)
  (
    COALESCE(SUM(v3_calc.direct_cost_for_lot(lb.lot_id)), 0)::double precision
    + COALESCE(SUM(v3_calc.rent_per_ha_for_lot(lb.lot_id) * lb.hectares), 0)::double precision
    + COALESCE(SUM(v3_calc.admin_cost_per_ha_for_lot(lb.lot_id) * lb.hectares), 0)::double precision
  )                                                                        AS costos_directos_invertidos_usd,

  -- Componentes de invertidos
  COALESCE(SUM(v3_calc.rent_per_ha_for_lot(lb.lot_id) * lb.hectares), 0)     AS arriendo_invertidos_usd,
  COALESCE(SUM(v3_calc.admin_cost_per_ha_for_lot(lb.lot_id) * lb.hectares), 0) AS estructura_invertidos_usd,

  -- Resultado operativo y ratio
  COALESCE(SUM(v3_calc.operating_result_per_ha_for_lot(lb.lot_id) * lb.hectares), 0) AS operating_result_usd,
  v3_calc.renta_pct(
    COALESCE(SUM(v3_calc.operating_result_per_ha_for_lot(lb.lot_id) * lb.hectares), 0)::double precision,
    COALESCE(SUM(v3_calc.direct_cost_for_lot(lb.lot_id)), 0)::double precision
  )                                                                          AS operating_result_pct
FROM lot_base lb
LEFT JOIN public.crop_commercializations cc
  ON cc.project_id = lb.project_id
 AND cc.crop_id   = lb.current_crop_id
 AND cc.deleted_at IS NULL
WHERE lb.current_crop_id IS NOT NULL;
