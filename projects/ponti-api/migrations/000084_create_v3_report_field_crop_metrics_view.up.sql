-- ========================================
-- MIGRACIÓN 000083: CREAR VISTA v3_report_field_crop_metrics_view (UP)
-- ========================================
-- 
-- Objetivo: Vista de métricas por field/crop apoyada en vistas base v3
-- Fecha: 2025-09-12
-- Autor: Sistema
-- 
-- Nota: Código en inglés, comentarios en español.

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
    calc.seeded_area(l.sowing_date, l.hectares::numeric)::double precision AS area_sembrada_ha,
    calc.harvested_area(l.tons::numeric, l.hectares::numeric)::double precision AS area_cosechada_ha
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

  lb.hectares::numeric(14,2) AS superficie_ha,
  lb.tons::numeric(14,2)      AS produccion_tn,
  lb.area_sembrada_ha::numeric(14,2)  AS area_sembrada_ha,
  lb.area_cosechada_ha::numeric(14,2) AS area_cosechada_ha,

  calc.per_ha_dp(lb.tons::double precision, lb.hectares)::numeric(14,2) AS rendimiento_tn_ha,

  COALESCE(cc.board_price,      0)::numeric(14,2) AS precio_bruto_usd_tn,
  COALESCE(cc.freight_cost,     0)::numeric(14,2) AS gasto_flete_usd_tn,
  COALESCE(cc.commercial_cost,  0)::numeric(14,2) AS gasto_comercial_usd_tn,
  COALESCE(cc.net_price,        0)::numeric(14,2) AS precio_neto_usd_tn,

  COALESCE(calc.income_net_total_for_lot(lb.lot_id), 0)::numeric(14,2) AS ingreso_neto_usd,
  calc.per_ha_dp(COALESCE(calc.income_net_total_for_lot(lb.lot_id),0)::double precision, lb.hectares)::numeric(14,2) AS ingreso_neto_usd_ha,

  COALESCE(calc.labor_cost_for_lot(lb.lot_id),  0)::numeric(14,2) AS costos_labores_usd,
  COALESCE(calc.supply_cost_for_lot(lb.lot_id), 0)::numeric(14,2) AS costos_insumos_usd,
  COALESCE(calc.direct_cost_for_lot(lb.lot_id), 0)::numeric(14,2) AS total_costos_directos_usd,

  calc.per_ha_dp(COALESCE(calc.direct_cost_for_lot(lb.lot_id),0)::double precision, lb.hectares)::numeric(14,2) AS costos_directos_usd_ha,

  (COALESCE(calc.income_net_total_for_lot(lb.lot_id),0)::double precision
   - COALESCE(calc.direct_cost_for_lot(lb.lot_id),0)::double precision)::numeric(14,2) AS margen_bruto_usd,

  calc.per_ha_dp((COALESCE(calc.income_net_total_for_lot(lb.lot_id),0)::double precision - COALESCE(calc.direct_cost_for_lot(lb.lot_id),0)::double precision), lb.hectares)::numeric(14,2) AS margen_bruto_usd_ha,

  (COALESCE(calc.rent_per_ha_for_lot(lb.lot_id),0)::double precision * lb.hectares)::numeric(14,2) AS arriendo_usd,
  COALESCE(calc.rent_per_ha_for_lot(lb.lot_id),0)::numeric(14,2) AS arriendo_usd_ha,

  (COALESCE(calc.admin_cost_per_ha_for_lot(lb.lot_id),0)::double precision * lb.hectares)::numeric(14,2) AS administracion_usd,
  COALESCE(calc.admin_cost_per_ha_for_lot(lb.lot_id),0)::numeric(14,2) AS administracion_usd_ha,

  (COALESCE(calc.operating_result_per_ha_for_lot(lb.lot_id),0)::double precision * lb.hectares)::numeric(14,2) AS resultado_operativo_usd,
  COALESCE(calc.operating_result_per_ha_for_lot(lb.lot_id),0)::numeric(14,2) AS resultado_operativo_usd_ha,

  -- total invertido
  ( COALESCE(calc.direct_cost_for_lot(lb.lot_id),0)::double precision
    + (COALESCE(calc.rent_per_ha_for_lot(lb.lot_id),0)::double precision * lb.hectares)
    + (COALESCE(calc.admin_cost_per_ha_for_lot(lb.lot_id),0)::double precision * lb.hectares)
  )::numeric(14,2) AS total_invertido_usd,

  ( calc.per_ha_dp(COALESCE(calc.direct_cost_for_lot(lb.lot_id),0)::double precision, lb.hectares)
    + COALESCE(calc.rent_per_ha_for_lot(lb.lot_id),0)::double precision
    + COALESCE(calc.admin_cost_per_ha_for_lot(lb.lot_id),0)::double precision
  )::numeric(14,2) AS total_invertido_usd_ha,

  calc.renta_pct(
    (COALESCE(calc.operating_result_per_ha_for_lot(lb.lot_id),0)::double precision * lb.hectares),
    ( COALESCE(calc.direct_cost_for_lot(lb.lot_id),0)::double precision
      + (COALESCE(calc.rent_per_ha_for_lot(lb.lot_id),0)::double precision * lb.hectares)
      + (COALESCE(calc.admin_cost_per_ha_for_lot(lb.lot_id),0)::double precision * lb.hectares)
    )
  )::numeric(6,2) AS renta_pct,

  calc.indifference_price_usd_tn(
    calc.per_ha_dp(COALESCE(calc.direct_cost_for_lot(lb.lot_id),0)::double precision, lb.hectares)
    + COALESCE(calc.rent_per_ha_for_lot(lb.lot_id),0)::double precision
    + COALESCE(calc.admin_cost_per_ha_for_lot(lb.lot_id),0)::double precision,
    calc.per_ha_dp(lb.tons::double precision, lb.hectares)
  )::numeric(14,2) AS rinde_indiferencia_usd_tn
FROM lot_base lb
LEFT JOIN public.crop_commercializations cc
  ON cc.project_id = lb.project_id
 AND cc.crop_id   = lb.current_crop_id
 AND cc.deleted_at IS NULL
WHERE lb.current_crop_id IS NOT NULL;


