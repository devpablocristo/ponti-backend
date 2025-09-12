-- ========================================
-- MIGRACIÓN 000082: CREAR VISTA v3_labor_list (UP)
-- ========================================
-- 
-- Objetivo: Listado de labores con info de WO, proyecto, field, lote, cultivo
-- Fecha: 2025-09-12
-- Autor: Sistema
-- 
-- Nota: Código en inglés, comentarios en español.

-- -------------------------------------------------------------------
-- v3_labor_list: listado de órdenes de trabajo con datos relevantes
-- -------------------------------------------------------------------
CREATE OR REPLACE VIEW public.v3_labor_list AS
SELECT
  w.id                  AS workorder_id,
  w.number              AS workorder_number,
  w.date,
  w.project_id,
  p.name                AS project_name,
  w.field_id,
  f.name                AS field_name,
  w.lot_id,
  l.name                AS lot_name,
  w.crop_id,
  c.name                AS crop_name,
  w.labor_id,
  lb.name               AS labor_name,
  cat_lb.id             AS labor_category_id,
  cat_lb.name           AS labor_category_name,
  w.contractor,
  lb.contractor_name,
  w.effective_area      AS surface_ha,
  lb.price              AS cost_per_ha,
  public.calculate_labor_cost(lb.price::numeric, w.effective_area::numeric) AS total_labor_cost,
  w.investor_id,
  i.name                AS investor_name
FROM public.workorders w
JOIN public.projects   p  ON p.id = w.project_id AND p.deleted_at IS NULL
JOIN public.fields     f  ON f.id = w.field_id   AND f.deleted_at IS NULL
LEFT JOIN public.lots  l  ON l.id = w.lot_id     AND l.deleted_at IS NULL
LEFT JOIN public.crops c  ON c.id = w.crop_id    AND c.deleted_at IS NULL
JOIN public.labors     lb ON lb.id = w.labor_id  AND lb.deleted_at IS NULL
LEFT JOIN public.categories cat_lb ON cat_lb.id = lb.category_id AND cat_lb.deleted_at IS NULL
LEFT JOIN public.investors  i      ON i.id = w.investor_id        AND i.deleted_at IS NULL
WHERE w.deleted_at IS NULL
  AND w.effective_area IS NOT NULL
  AND w.effective_area > 0
  AND lb.price IS NOT NULL;


