-- =============================================================================
-- MIGRACIÓN 000341: Consolidar lot_base de field_crop_metrics en v4_calc (DOWN)
-- =============================================================================
--
-- Propósito: Revertir field_crop_metrics con lot_base inline y eliminar v4_calc.field_crop_metrics_lot_base.
-- Nota: Comentarios en español, código en inglés.
--
BEGIN;

-- 1) field_crop_metrics: definición previa (lot_base inline)
CREATE OR REPLACE VIEW v4_report.field_crop_metrics AS
WITH
lot_base AS (
  SELECT
    f.project_id,
    f.id AS field_id,
    f.name AS field_name,
    l.current_crop_id,
    c.name AS crop_name,
    l.id AS lot_id,
    l.hectares,
    l.tons,
    COALESCE(v3_lot_ssot.seeded_area_for_lot(l.id), 0)::numeric AS sowed_area_ha,
    COALESCE(v3_lot_ssot.harvested_area_for_lot(l.id), 0)::numeric AS harvested_area_ha,
    COALESCE(v3_lot_ssot.yield_tn_per_ha_for_lot(l.id), 0) AS yield_tn_per_ha,
    COALESCE(v3_lot_ssot.labor_cost_for_lot(l.id), 0)::numeric AS labor_cost_usd,
    COALESCE(v3_lot_ssot.supply_cost_for_lot_base(l.id), 0)::numeric AS supply_cost_usd,
    COALESCE(v3_lot_ssot.net_price_usd_for_lot(l.id), 0)::numeric AS net_price_usd,
    COALESCE(v3_lot_ssot.rent_per_ha_for_lot(l.id), 0)::numeric AS rent_per_ha,
    COALESCE(v3_calc.admin_cost_per_ha_for_lot(l.id), 0)::numeric AS admin_per_ha,
    COALESCE(v3_report_ssot.board_price_for_lot(l.id), 0)::numeric AS board_price,
    COALESCE(v3_report_ssot.freight_cost_for_lot(l.id), 0)::numeric AS freight_cost,
    COALESCE(v3_report_ssot.commercial_cost_for_lot(l.id), 0)::numeric AS commercial_cost
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  JOIN public.projects p ON p.id = f.project_id AND p.deleted_at IS NULL
  LEFT JOIN public.crops c ON c.id = l.current_crop_id AND c.deleted_at IS NULL
  WHERE l.deleted_at IS NULL
    AND l.current_crop_id IS NOT NULL
    AND l.hectares > 0
),
aggregated AS (
  SELECT
    lb.project_id,
    lb.field_id,
    lb.field_name,
    lb.current_crop_id,
    lb.crop_name,
    SUM(lb.hectares)::numeric AS superficie_total,
    SUM(lb.sowed_area_ha)::numeric AS superficie_sembrada_ha,
    SUM(lb.harvested_area_ha)::numeric AS area_cosechada_ha,
    SUM(lb.tons)::numeric AS produccion_tn,
    CASE WHEN SUM(lb.sowed_area_ha) > 0
      THEN SUM(lb.yield_tn_per_ha * lb.sowed_area_ha) / SUM(lb.sowed_area_ha)
      ELSE 0
    END AS rendimiento_tn_ha,
    SUM(lb.labor_cost_usd)::numeric AS costos_labores_usd,
    SUM(lb.supply_cost_usd)::numeric AS costos_insumos_usd,
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
    -- Arriendo y admin por superficie total
    SUM(lb.rent_per_ha * lb.hectares)::numeric AS arriendo_total_usd,
    SUM(lb.admin_per_ha * lb.hectares)::numeric AS admin_total_usd,
    SUM(lb.tons * lb.net_price_usd)::numeric AS ingreso_neto_total
  FROM lot_base lb
  GROUP BY lb.project_id, lb.field_id, lb.field_name, lb.current_crop_id, lb.crop_name
)
SELECT
  a.project_id,
  a.field_id,
  a.field_name,
  a.current_crop_id,
  a.crop_name,
  a.superficie_total AS superficie_ha,
  a.produccion_tn,
  a.superficie_sembrada_ha AS area_sembrada_ha,
  a.area_cosechada_ha,
  a.rendimiento_tn_ha,
  a.precio_bruto_usd_tn,
  a.gasto_flete_usd_tn,
  a.gasto_comercial_usd_tn,
  a.precio_neto_usd_tn,
  a.ingreso_neto_total AS ingreso_neto_usd,
  v3_core_ssot.safe_div(a.ingreso_neto_total, a.superficie_total) AS ingreso_neto_usd_ha,
  a.costos_labores_usd,
  v3_core_ssot.safe_div(a.costos_labores_usd, a.superficie_total) AS costos_labores_usd_ha,
  a.costos_insumos_usd,
  v3_core_ssot.safe_div(a.costos_insumos_usd, a.superficie_total) AS costos_insumos_usd_ha,
  (a.costos_labores_usd + a.costos_insumos_usd)::numeric AS total_costos_directos_usd,
  v3_core_ssot.safe_div(a.costos_labores_usd + a.costos_insumos_usd, a.superficie_total) AS costos_directos_usd_ha,
  (a.ingreso_neto_total - a.costos_labores_usd - a.costos_insumos_usd)::numeric AS margen_bruto_usd,
  v3_core_ssot.safe_div(a.ingreso_neto_total - a.costos_labores_usd - a.costos_insumos_usd, a.superficie_total) AS margen_bruto_usd_ha,
  a.arriendo_total_usd AS arriendo_usd,
  v3_core_ssot.safe_div(a.arriendo_total_usd, a.superficie_total) AS arriendo_usd_ha,
  a.admin_total_usd AS administracion_usd,
  v3_core_ssot.safe_div(a.admin_total_usd, a.superficie_total) AS administracion_usd_ha,
  (a.ingreso_neto_total - a.costos_labores_usd - a.costos_insumos_usd - a.arriendo_total_usd - a.admin_total_usd)::numeric AS resultado_operativo_usd,
  v3_core_ssot.safe_div(
    a.ingreso_neto_total - a.costos_labores_usd - a.costos_insumos_usd - a.arriendo_total_usd - a.admin_total_usd,
    a.superficie_total
  ) AS resultado_operativo_usd_ha,
  (a.costos_labores_usd + a.costos_insumos_usd + a.arriendo_total_usd + a.admin_total_usd)::numeric AS total_invertido_usd,
  v3_core_ssot.safe_div(
    a.costos_labores_usd + a.costos_insumos_usd + a.arriendo_total_usd + a.admin_total_usd,
    a.superficie_total
  ) AS total_invertido_usd_ha,
  CASE WHEN (a.costos_labores_usd + a.costos_insumos_usd + a.arriendo_total_usd + a.admin_total_usd) > 0
    THEN ((a.ingreso_neto_total - a.costos_labores_usd - a.costos_insumos_usd - a.arriendo_total_usd - a.admin_total_usd) /
          (a.costos_labores_usd + a.costos_insumos_usd + a.arriendo_total_usd + a.admin_total_usd) * 100)::double precision
    ELSE 0
  END AS renta_pct,
  CASE WHEN a.precio_neto_usd_tn > 0 AND a.superficie_total > 0
    THEN ((a.costos_labores_usd + a.costos_insumos_usd + a.arriendo_total_usd + a.admin_total_usd) / a.superficie_total / a.precio_neto_usd_tn)::numeric
    ELSE 0
  END AS rinde_indiferencia_usd_tn
FROM aggregated a;

-- Eliminar base consolidada
DROP VIEW IF EXISTS v4_calc.field_crop_metrics_lot_base;

COMMIT;
