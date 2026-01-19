-- =============================================================================
-- MIGRACIÓN 000344: Consolidar costos de labores en v4_calc (UP)
-- =============================================================================
--
-- Propósito: Centralizar costos por categoría para field_crop_labores.
-- Enfoque: Crear v4_calc.field_crop_labor_costs_by_lot y reutilizarlo.
-- Nota: Comentarios en español, código en inglés.
--
BEGIN;

-- 1) v4_calc.field_crop_labor_costs_by_lot (fuente única)
CREATE OR REPLACE VIEW v4_calc.field_crop_labor_costs_by_lot AS
SELECT
  lb.project_id,
  lb.field_id,
  lb.crop_id,
  lb.lot_id,
  lb.sowed_area_ha,
  lb.surface_ha,
  -- Siembra
  COALESCE((
    SELECT SUM(lab.price * w.effective_area)
    FROM public.workorders w
    JOIN public.labors lab ON lab.id = w.labor_id
    JOIN public.categories cat ON cat.id = lab.category_id
    WHERE w.lot_id = lb.lot_id
      AND w.deleted_at IS NULL
      AND w.effective_area > 0
      AND lab.price IS NOT NULL
      AND cat.name = 'Siembra'
      AND cat.type_id = 4
  ), 0)::numeric AS siembra_usd,
  -- Pulverización
  COALESCE((
    SELECT SUM(lab.price * w.effective_area)
    FROM public.workorders w
    JOIN public.labors lab ON lab.id = w.labor_id
    JOIN public.categories cat ON cat.id = lab.category_id
    WHERE w.lot_id = lb.lot_id
      AND w.deleted_at IS NULL
      AND w.effective_area > 0
      AND lab.price IS NOT NULL
      AND cat.name = 'Pulverización'
      AND cat.type_id = 4
  ), 0)::numeric AS pulverizacion_usd,
  -- Riego
  COALESCE((
    SELECT SUM(lab.price * w.effective_area)
    FROM public.workorders w
    JOIN public.labors lab ON lab.id = w.labor_id
    JOIN public.categories cat ON cat.id = lab.category_id
    WHERE w.lot_id = lb.lot_id
      AND w.deleted_at IS NULL
      AND w.effective_area > 0
      AND lab.price IS NOT NULL
      AND cat.name = 'Riego'
      AND cat.type_id = 4
  ), 0)::numeric AS riego_usd,
  -- Cosecha
  COALESCE((
    SELECT SUM(lab.price * w.effective_area)
    FROM public.workorders w
    JOIN public.labors lab ON lab.id = w.labor_id
    JOIN public.categories cat ON cat.id = lab.category_id
    WHERE w.lot_id = lb.lot_id
      AND w.deleted_at IS NULL
      AND w.effective_area > 0
      AND lab.price IS NOT NULL
      AND cat.name = 'Cosecha'
      AND cat.type_id = 4
  ), 0)::numeric AS cosecha_usd,
  -- Otras labores
  COALESCE((
    SELECT SUM(lab.price * w.effective_area)
    FROM public.workorders w
    JOIN public.labors lab ON lab.id = w.labor_id
    JOIN public.categories cat ON cat.id = lab.category_id
    WHERE w.lot_id = lb.lot_id
      AND w.deleted_at IS NULL
      AND w.effective_area > 0
      AND lab.price IS NOT NULL
      AND cat.name NOT IN ('Siembra', 'Pulverización', 'Riego', 'Cosecha')
      AND cat.type_id = 4
  ), 0)::numeric AS otras_labores_usd
FROM v4_calc.field_crop_lot_base lb;

COMMENT ON VIEW v4_calc.field_crop_labor_costs_by_lot IS
'Costos de labores por lote (categorías).';

-- 2) field_crop_labores: usa v4_calc.field_crop_labor_costs_by_lot
CREATE OR REPLACE VIEW v4_report.field_crop_labores AS
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
FROM v4_calc.field_crop_labor_costs_by_lot
GROUP BY project_id, field_id, crop_id;

COMMIT;
