-- =============================================================================
-- MIGRACIÓN 000329: Ajustar costo por categoría a total_used (SSOT)
-- =============================================================================
--
-- Propósito: Alinear costos por categoría con exportación (Excel)
-- Regla: Costo de insumos = SUM(total_used * price) por categoría
-- Impacto: dashboard_management_balance, field_crop_insumos, etc.
-- Nota: Comentarios en español, código en inglés
--

BEGIN;

-- Reemplazar la función por categoría para usar total_used
CREATE OR REPLACE FUNCTION v3_lot_ssot.supply_cost_for_lot_by_category(
  p_lot_id bigint,
  p_category_name text
)
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    SUM(wi.total_used * s.price), 0
  )::numeric
  FROM public.workorders w
  JOIN public.workorder_items wi ON wi.workorder_id = w.id AND wi.deleted_at IS NULL
  JOIN public.supplies s ON s.id = wi.supply_id AND s.deleted_at IS NULL
  JOIN public.categories c ON c.id = s.category_id
  WHERE w.lot_id = p_lot_id
    AND c.name = p_category_name
    AND w.deleted_at IS NULL
    AND w.effective_area > 0
    AND wi.total_used > 0
    AND s.price IS NOT NULL
$$;

COMMIT;
