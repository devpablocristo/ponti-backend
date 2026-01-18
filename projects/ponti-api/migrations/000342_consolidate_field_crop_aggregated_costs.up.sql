-- =============================================================================
-- MIGRACIÓN 000342: Consolidar aggregated de field_crop_* en v4_calc (UP)
-- =============================================================================
--
-- Propósito: Eliminar duplicación de aggregated entre field_crop_economicos y field_crop_rentabilidad.
-- Enfoque: Crear v4_calc.field_crop_aggregated y reutilizarlo.
-- Nota: Comentarios en español, código en inglés.
--
BEGIN;

-- 1) v4_calc.field_crop_aggregated (fuente única)
CREATE OR REPLACE VIEW v4_calc.field_crop_aggregated AS
SELECT
  project_id,
  field_id,
  crop_id,
  MIN(lot_id) AS sample_lot_id,
  SUM(tons)::numeric AS production_tn,
  SUM(sowed_area_ha)::numeric AS sown_area_ha,
  SUM(surface_ha)::numeric AS surface_ha,
  SUM(v3_lot_ssot.labor_cost_for_lot(lot_id))::numeric AS labor_costs_usd,
  SUM(
    v3_lot_ssot.supply_cost_for_lot_by_category(lot_id, 'Semilla') +
    v3_lot_ssot.supply_cost_for_lot_by_category(lot_id, 'Curasemillas') +
    v3_lot_ssot.supply_cost_for_lot_by_category(lot_id, 'Herbicidas') +
    v3_lot_ssot.supply_cost_for_lot_by_category(lot_id, 'Insecticidas') +
    v3_lot_ssot.supply_cost_for_lot_by_category(lot_id, 'Fungicidas') +
    v3_lot_ssot.supply_cost_for_lot_by_category(lot_id, 'Coadyuvantes') +
    v3_lot_ssot.supply_cost_for_lot_by_category(lot_id, 'Fertilizantes') +
    v3_lot_ssot.supply_cost_for_lot_by_category(lot_id, 'Otros Insumos')
  )::numeric AS supply_costs_usd,
  -- Direct cost para rentabilidad
  SUM(
    v3_lot_ssot.labor_cost_for_lot(lot_id) +
    v3_lot_ssot.supply_cost_for_lot_by_category(lot_id, 'Semilla') +
    v3_lot_ssot.supply_cost_for_lot_by_category(lot_id, 'Curasemillas') +
    v3_lot_ssot.supply_cost_for_lot_by_category(lot_id, 'Herbicidas') +
    v3_lot_ssot.supply_cost_for_lot_by_category(lot_id, 'Insecticidas') +
    v3_lot_ssot.supply_cost_for_lot_by_category(lot_id, 'Fungicidas') +
    v3_lot_ssot.supply_cost_for_lot_by_category(lot_id, 'Coadyuvantes') +
    v3_lot_ssot.supply_cost_for_lot_by_category(lot_id, 'Fertilizantes') +
    v3_lot_ssot.supply_cost_for_lot_by_category(lot_id, 'Otros Insumos')
  )::numeric AS direct_cost_usd,
  -- Arriendo FIJO para mostrar en UI y total invertido
  SUM(v3_lot_ssot.rent_fixed_only_for_lot(lot_id) * surface_ha)::numeric AS rent_fixed_usd,
  -- Arriendo TOTAL (fijo + % ingresos) para resultado operativo
  SUM(v3_lot_ssot.rent_per_ha_for_lot(lot_id) * surface_ha)::numeric AS rent_total_usd,
  SUM(v3_lot_ssot.admin_cost_per_ha_for_lot(lot_id) * surface_ha)::numeric AS administration_usd
FROM v4_calc.field_crop_lot_base
GROUP BY project_id, field_id, crop_id;

COMMENT ON VIEW v4_calc.field_crop_aggregated IS
'Agregados comunes para field_crop_economicos y field_crop_rentabilidad.';

-- 2) field_crop_economicos: usa v4_calc.field_crop_aggregated
CREATE OR REPLACE VIEW v4_report.field_crop_economicos AS
SELECT
  project_id,
  field_id,
  crop_id AS current_crop_id,
  (labor_costs_usd + supply_costs_usd) AS gastos_directos_usd,
  v3_core_ssot.safe_div(
    (labor_costs_usd + supply_costs_usd),
    surface_ha
  ) AS gastos_directos_usd_ha,
  (
    (production_tn * v3_lot_ssot.net_price_usd_for_lot(sample_lot_id)) -
    (labor_costs_usd + supply_costs_usd)
  ) AS margen_bruto_usd,
  (
    (v3_core_ssot.safe_div(production_tn, surface_ha) * v3_lot_ssot.net_price_usd_for_lot(sample_lot_id)) -
    v3_core_ssot.safe_div((labor_costs_usd + supply_costs_usd), surface_ha)
  ) AS margen_bruto_usd_ha,
  -- Mostrar arriendo FIJO (capitalizable)
  rent_fixed_usd AS arriendo_usd,
  v3_core_ssot.safe_div(rent_fixed_usd, surface_ha) AS arriendo_usd_ha,
  administration_usd AS adm_estructura_usd,
  v3_core_ssot.safe_div(administration_usd, surface_ha) AS adm_estructura_usd_ha,
  -- Resultado operativo usa arriendo TOTAL
  (
    (production_tn * v3_lot_ssot.net_price_usd_for_lot(sample_lot_id)) -
    (labor_costs_usd + supply_costs_usd) - rent_total_usd - administration_usd
  ) AS resultado_operativo_usd,
  (
    (v3_core_ssot.safe_div(production_tn, surface_ha) * v3_lot_ssot.net_price_usd_for_lot(sample_lot_id)) -
    v3_core_ssot.safe_div((labor_costs_usd + supply_costs_usd), surface_ha) -
    v3_core_ssot.safe_div(rent_total_usd, surface_ha) -
    v3_core_ssot.safe_div(administration_usd, surface_ha)
  ) AS resultado_operativo_usd_ha
FROM v4_calc.field_crop_aggregated;

-- 3) field_crop_rentabilidad: usa v4_calc.field_crop_aggregated
CREATE OR REPLACE VIEW v4_report.field_crop_rentabilidad AS
SELECT
  project_id,
  field_id,
  crop_id AS current_crop_id,
  (direct_cost_usd + rent_fixed_usd + administration_usd) AS total_invertido_usd,
  v3_core_ssot.safe_div(
    (direct_cost_usd + rent_fixed_usd + administration_usd),
    surface_ha
  ) AS total_invertido_usd_ha,
  v3_lot_ssot.renta_pct(
    (
      (production_tn * v3_lot_ssot.net_price_usd_for_lot(sample_lot_id)) -
      direct_cost_usd - rent_fixed_usd - administration_usd
    ),
    (direct_cost_usd + rent_fixed_usd + administration_usd)
  ) AS renta_pct,
  v3_core_ssot.safe_div(
    v3_core_ssot.safe_div((direct_cost_usd + rent_fixed_usd + administration_usd), surface_ha),
    v3_lot_ssot.net_price_usd_for_lot(sample_lot_id)
  ) AS rinde_indiferencia_total_usd_tn
FROM v4_calc.field_crop_aggregated;

COMMIT;
