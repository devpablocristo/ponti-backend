-- =============================================================================
-- MIGRACIÓN 000347: Consolidar totales de insumos y labores en v4_calc (UP)
-- =============================================================================
--
-- Propósito: Evitar recomputar totales a partir de categorías en vistas.
-- Enfoque: Agregar totales por lote en v4_calc y usar SUM(total_*).
-- Nota: Comentarios en español, código en inglés.
--
BEGIN;

-- 1) v4_calc.field_crop_labor_costs_by_lot: agregar total_labores_usd
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
  ), 0)::numeric AS otras_labores_usd,
  (
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
    ), 0)::numeric +
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
    ), 0)::numeric +
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
    ), 0)::numeric +
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
    ), 0)::numeric +
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
    ), 0)::numeric
  ) AS total_labores_usd
FROM v4_calc.field_crop_lot_base lb;

-- 2) field_crop_insumos: usar SUM(total_insumos_usd)
CREATE OR REPLACE VIEW v4_report.field_crop_insumos AS
WITH supply_costs AS (
  SELECT
    project_id,
    field_id,
    crop_id,
    lot_id,
    sowed_area_ha,
    surface_ha,
    semillas_usd,
    curasemillas_usd,
    herbicidas_usd,
    insecticidas_usd,
    fungicidas_usd,
    coadyuvantes_usd,
    fertilizantes_usd,
    otros_insumos_usd,
    total_insumos_usd
  FROM v4_calc.field_crop_supply_costs_by_lot
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
  COALESCE(SUM(total_insumos_usd), 0) AS total_insumos_usd,
  v3_core_ssot.safe_div(
    COALESCE(SUM(total_insumos_usd), 0),
    COALESCE(SUM(surface_ha), 1)::numeric
  ) AS total_insumos_usd_ha
FROM supply_costs
GROUP BY project_id, field_id, crop_id;

-- 3) field_crop_labores: usar SUM(total_labores_usd)
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
  COALESCE(SUM(total_labores_usd), 0) AS total_labores_usd,
  v3_core_ssot.safe_div(
    COALESCE(SUM(total_labores_usd), 0),
    COALESCE(SUM(surface_ha), 1)::numeric
  ) AS total_labores_usd_ha
FROM v4_calc.field_crop_labor_costs_by_lot
GROUP BY project_id, field_id, crop_id;

COMMIT;
