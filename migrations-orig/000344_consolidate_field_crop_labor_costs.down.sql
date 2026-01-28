-- =============================================================================
-- MIGRACIÓN 000344: Consolidar costos de labores en v4_calc (DOWN)
-- =============================================================================
--
-- Propósito: Revertir field_crop_labores a definición previa y eliminar v4_calc.field_crop_labor_costs_by_lot.
-- Nota: Comentarios en español, código en inglés.
--
BEGIN;

-- 1) field_crop_labores: definición previa (usa v4_calc.field_crop_lot_base inline)
CREATE OR REPLACE VIEW v4_report.field_crop_labores AS
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
labor_costs AS (
  SELECT
    lb.project_id,
    lb.field_id,
    lb.crop_id,
    lb.lot_id,
    lb.sowed_area_ha,
    lb.surface_ha,
    COALESCE(SUM(CASE WHEN cat.name = 'Siembra' THEN lab.price * w.effective_area ELSE 0 END), 0)::numeric AS siembra_usd,
    COALESCE(SUM(CASE WHEN cat.name = 'Pulverización' THEN lab.price * w.effective_area ELSE 0 END), 0)::numeric AS pulverizacion_usd,
    COALESCE(SUM(CASE WHEN cat.name = 'Riego' THEN lab.price * w.effective_area ELSE 0 END), 0)::numeric AS riego_usd,
    COALESCE(SUM(CASE WHEN cat.name = 'Cosecha' THEN lab.price * w.effective_area ELSE 0 END), 0)::numeric AS cosecha_usd,
    COALESCE(SUM(CASE WHEN cat.name = 'Otras Labores' THEN lab.price * w.effective_area ELSE 0 END), 0)::numeric AS otras_labores_usd
  FROM lot_base lb
  LEFT JOIN public.workorders w ON w.lot_id = lb.lot_id AND w.deleted_at IS NULL
  LEFT JOIN public.labors lab ON lab.id = w.labor_id AND lab.deleted_at IS NULL
  LEFT JOIN public.categories cat ON cat.id = lab.category_id
  GROUP BY lb.project_id, lb.field_id, lb.crop_id, lb.lot_id, lb.sowed_area_ha, lb.surface_ha
)
SELECT
  project_id,
  field_id,
  crop_id AS current_crop_id,
  COALESCE(SUM(siembra_usd), 0) AS siembra_total_usd,
  COALESCE(SUM(pulverizacion_usd), 0) AS pulverizacion_total_usd,
  COALESCE(SUM(riego_usd), 0) AS riego_total_usd,
  COALESCE(SUM(cosecha_usd), 0) AS cosecha_total_usd,
  COALESCE(SUM(otras_labores_usd), 0) AS otras_labores_total_usd,
  v3_core_ssot.safe_div(COALESCE(SUM(siembra_usd), 0), COALESCE(SUM(surface_ha), 1)::numeric) AS siembra_usd_ha,
  v3_core_ssot.safe_div(COALESCE(SUM(pulverizacion_usd), 0), COALESCE(SUM(surface_ha), 1)::numeric) AS pulverizacion_usd_ha,
  v3_core_ssot.safe_div(COALESCE(SUM(riego_usd), 0), COALESCE(SUM(surface_ha), 1)::numeric) AS riego_usd_ha,
  v3_core_ssot.safe_div(COALESCE(SUM(cosecha_usd), 0), COALESCE(SUM(surface_ha), 1)::numeric) AS cosecha_usd_ha,
  v3_core_ssot.safe_div(COALESCE(SUM(otras_labores_usd), 0), COALESCE(SUM(surface_ha), 1)::numeric) AS otras_labores_usd_ha,
  COALESCE(SUM(siembra_usd) + SUM(pulverizacion_usd) + SUM(riego_usd) +
   SUM(cosecha_usd) + SUM(otras_labores_usd), 0) AS total_labores_usd,
  v3_core_ssot.safe_div(
    COALESCE(SUM(siembra_usd) + SUM(pulverizacion_usd) + SUM(riego_usd) +
     SUM(cosecha_usd) + SUM(otras_labores_usd), 0),
    COALESCE(SUM(surface_ha), 1)::numeric
  ) AS total_labores_usd_ha
FROM labor_costs
GROUP BY project_id, field_id, crop_id;

-- Eliminar base consolidada
DROP VIEW IF EXISTS v4_calc.field_crop_labor_costs_by_lot;

COMMIT;
