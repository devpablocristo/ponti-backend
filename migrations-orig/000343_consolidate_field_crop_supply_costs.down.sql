-- =============================================================================
-- MIGRACIÓN 000343: Consolidar costos de insumos en v4_calc (DOWN)
-- =============================================================================
--
-- Propósito: Revertir field_crop_insumos y field_crop_aggregated a definiciones previas
-- y eliminar v4_calc.field_crop_supply_costs_by_lot.
-- Nota: Comentarios en español, código en inglés.
--
BEGIN;

-- 1) field_crop_insumos: definición previa (usa v4_calc.field_crop_lot_base)
CREATE OR REPLACE VIEW v4_report.field_crop_insumos AS
WITH lot_base AS (
  SELECT
    project_id,
    field_id,
    crop_id,
    lot_id,
    surface_ha,
    sowed_area_ha
  FROM v4_calc.field_crop_lot_base
),
supply_costs AS (
  SELECT
    lb.project_id,
    lb.field_id,
    lb.crop_id,
    lb.lot_id,
    lb.sowed_area_ha,
    lb.surface_ha,
    v3_lot_ssot.supply_cost_for_lot_by_category(lb.lot_id, 'Semilla') AS semillas_usd,
    v3_lot_ssot.supply_cost_for_lot_by_category(lb.lot_id, 'Curasemillas') AS curasemillas_usd,
    v3_lot_ssot.supply_cost_for_lot_by_category(lb.lot_id, 'Herbicidas') AS herbicidas_usd,
    v3_lot_ssot.supply_cost_for_lot_by_category(lb.lot_id, 'Insecticidas') AS insecticidas_usd,
    v3_lot_ssot.supply_cost_for_lot_by_category(lb.lot_id, 'Fungicidas') AS fungicidas_usd,
    v3_lot_ssot.supply_cost_for_lot_by_category(lb.lot_id, 'Coadyuvantes') AS coadyuvantes_usd,
    v3_lot_ssot.supply_cost_for_lot_by_category(lb.lot_id, 'Fertilizantes') AS fertilizantes_usd,
    v3_lot_ssot.supply_cost_for_lot_by_category(lb.lot_id, 'Otros Insumos') AS otros_insumos_usd
  FROM lot_base lb
)
SELECT
  project_id,
  field_id,
  crop_id AS current_crop_id,
  COALESCE(SUM(semillas_usd), 0) AS semillas_total_usd,
  COALESCE(SUM(curasemillas_usd), 0) AS curasemillas_total_usd,
  COALESCE(SUM(herbicidas_usd), 0) AS herbicidas_total_usd,
  COALESCE(SUM(insecticidas_usd), 0) AS insecticidas_total_usd,
  COALESCE(SUM(fungicidas_usd), 0) AS fungicidas_total_usd,
  COALESCE(SUM(coadyuvantes_usd), 0) AS coadyuvantes_total_usd,
  COALESCE(SUM(fertilizantes_usd), 0) AS fertilizantes_total_usd,
  COALESCE(SUM(otros_insumos_usd), 0) AS otros_insumos_total_usd,
  v3_core_ssot.safe_div(COALESCE(SUM(semillas_usd), 0), COALESCE(SUM(surface_ha), 1)::numeric) AS semillas_usd_ha,
  v3_core_ssot.safe_div(COALESCE(SUM(curasemillas_usd), 0), COALESCE(SUM(surface_ha), 1)::numeric) AS curasemillas_usd_ha,
  v3_core_ssot.safe_div(COALESCE(SUM(herbicidas_usd), 0), COALESCE(SUM(surface_ha), 1)::numeric) AS herbicidas_usd_ha,
  v3_core_ssot.safe_div(COALESCE(SUM(insecticidas_usd), 0), COALESCE(SUM(surface_ha), 1)::numeric) AS insecticidas_usd_ha,
  v3_core_ssot.safe_div(COALESCE(SUM(fungicidas_usd), 0), COALESCE(SUM(surface_ha), 1)::numeric) AS fungicidas_usd_ha,
  v3_core_ssot.safe_div(COALESCE(SUM(coadyuvantes_usd), 0), COALESCE(SUM(surface_ha), 1)::numeric) AS coadyuvantes_usd_ha,
  v3_core_ssot.safe_div(COALESCE(SUM(fertilizantes_usd), 0), COALESCE(SUM(surface_ha), 1)::numeric) AS fertilizantes_usd_ha,
  v3_core_ssot.safe_div(COALESCE(SUM(otros_insumos_usd), 0), COALESCE(SUM(surface_ha), 1)::numeric) AS otros_insumos_usd_ha,
  COALESCE(SUM(semillas_usd) + SUM(curasemillas_usd) + SUM(herbicidas_usd) +
   SUM(insecticidas_usd) + SUM(fungicidas_usd) + SUM(coadyuvantes_usd) +
   SUM(fertilizantes_usd) + SUM(otros_insumos_usd), 0) AS total_insumos_usd,
  v3_core_ssot.safe_div(
    COALESCE(SUM(semillas_usd) + SUM(curasemillas_usd) + SUM(herbicidas_usd) +
     SUM(insecticidas_usd) + SUM(fungicidas_usd) + SUM(coadyuvantes_usd) +
     SUM(fertilizantes_usd) + SUM(otros_insumos_usd), 0),
    COALESCE(SUM(surface_ha), 1)::numeric
  ) AS total_insumos_usd_ha
FROM supply_costs
GROUP BY project_id, field_id, crop_id;

-- 2) field_crop_aggregated: definición previa (usa v4_calc.field_crop_lot_base inline)
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

-- Eliminar base consolidada
DROP VIEW IF EXISTS v4_calc.field_crop_supply_costs_by_lot;

COMMIT;
