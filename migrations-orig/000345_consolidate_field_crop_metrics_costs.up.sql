-- =============================================================================
-- MIGRACIÓN 000345: Consolidar costos en field_crop_metrics (UP)
-- =============================================================================
--
-- Propósito: Usar costos agregados SSOT en field_crop_metrics.
-- Enfoque: Reutilizar v4_calc.field_crop_aggregated para costos y rent/admin.
-- Nota: Comentarios en español, código en inglés.
--
BEGIN;

CREATE OR REPLACE VIEW v4_report.field_crop_metrics AS
WITH
lot_base AS (
  SELECT
    project_id,
    field_id,
    field_name,
    current_crop_id,
    crop_name,
    lot_id,
    hectares,
    tons,
    sowed_area_ha,
    harvested_area_ha,
    yield_tn_per_ha,
    net_price_usd,
    board_price,
    freight_cost,
    commercial_cost
  FROM v4_calc.field_crop_metrics_lot_base
),
aggregated_base AS (
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
  FROM lot_base lb
  GROUP BY lb.project_id, lb.field_id, lb.field_name, lb.current_crop_id, lb.crop_name
),
costs AS (
  SELECT
    project_id,
    field_id,
    crop_id,
    labor_costs_usd,
    supply_costs_usd,
    rent_total_usd AS arriendo_total_usd,
    administration_usd AS admin_total_usd
  FROM v4_calc.field_crop_aggregated
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
  c.labor_costs_usd AS costos_labores_usd,
  v3_core_ssot.safe_div(c.labor_costs_usd, a.superficie_total) AS costos_labores_usd_ha,
  c.supply_costs_usd AS costos_insumos_usd,
  v3_core_ssot.safe_div(c.supply_costs_usd, a.superficie_total) AS costos_insumos_usd_ha,
  (c.labor_costs_usd + c.supply_costs_usd)::numeric AS total_costos_directos_usd,
  v3_core_ssot.safe_div(c.labor_costs_usd + c.supply_costs_usd, a.superficie_total) AS costos_directos_usd_ha,
  (a.ingreso_neto_total - c.labor_costs_usd - c.supply_costs_usd)::numeric AS margen_bruto_usd,
  v3_core_ssot.safe_div(a.ingreso_neto_total - c.labor_costs_usd - c.supply_costs_usd, a.superficie_total) AS margen_bruto_usd_ha,
  c.arriendo_total_usd AS arriendo_usd,
  v3_core_ssot.safe_div(c.arriendo_total_usd, a.superficie_total) AS arriendo_usd_ha,
  c.admin_total_usd AS administracion_usd,
  v3_core_ssot.safe_div(c.admin_total_usd, a.superficie_total) AS administracion_usd_ha,
  (a.ingreso_neto_total - c.labor_costs_usd - c.supply_costs_usd - c.arriendo_total_usd - c.admin_total_usd)::numeric AS resultado_operativo_usd,
  v3_core_ssot.safe_div(
    a.ingreso_neto_total - c.labor_costs_usd - c.supply_costs_usd - c.arriendo_total_usd - c.admin_total_usd,
    a.superficie_total
  ) AS resultado_operativo_usd_ha,
  (c.labor_costs_usd + c.supply_costs_usd + c.arriendo_total_usd + c.admin_total_usd)::numeric AS total_invertido_usd,
  v3_core_ssot.safe_div(
    c.labor_costs_usd + c.supply_costs_usd + c.arriendo_total_usd + c.admin_total_usd,
    a.superficie_total
  ) AS total_invertido_usd_ha,
  CASE WHEN (c.labor_costs_usd + c.supply_costs_usd + c.arriendo_total_usd + c.admin_total_usd) > 0
    THEN ((a.ingreso_neto_total - c.labor_costs_usd - c.supply_costs_usd - c.arriendo_total_usd - c.admin_total_usd) /
          (c.labor_costs_usd + c.supply_costs_usd + c.arriendo_total_usd + c.admin_total_usd) * 100)::double precision
    ELSE 0
  END AS renta_pct,
  CASE WHEN a.precio_neto_usd_tn > 0 AND a.superficie_total > 0
    THEN ((c.labor_costs_usd + c.supply_costs_usd + c.arriendo_total_usd + c.admin_total_usd) / a.superficie_total / a.precio_neto_usd_tn)::numeric
    ELSE 0
  END AS rinde_indiferencia_usd_tn
FROM aggregated_base a
LEFT JOIN costs c
  ON c.project_id = a.project_id
  AND c.field_id = a.field_id
  AND c.crop_id = a.current_crop_id;

COMMIT;
