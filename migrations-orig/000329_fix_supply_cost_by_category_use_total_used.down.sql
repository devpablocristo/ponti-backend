-- =============================================================================
-- MIGRACIÓN 000329: Rollback costo por categoría (final_dose * area)
-- =============================================================================
--
-- Propósito: Restaurar lógica anterior de supply_cost_for_lot_by_category
-- Nota: Comentarios en español, código en inglés
--

BEGIN;

-- Restaurar fórmula anterior (final_dose × effective_area × price)
CREATE OR REPLACE FUNCTION v3_lot_ssot.supply_cost_for_lot_by_category(
  p_lot_id bigint,
  p_category_name text
)
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    SUM(wi.final_dose * s.price * w.effective_area), 0
  )::numeric
  FROM public.workorders w
  JOIN public.workorder_items wi ON wi.workorder_id = w.id
  JOIN public.supplies s ON s.id = wi.supply_id
  JOIN public.categories c ON c.id = s.category_id
  WHERE w.lot_id = p_lot_id
    AND c.name = p_category_name
    AND w.deleted_at IS NULL
    AND w.effective_area > 0
    AND wi.final_dose > 0
    AND s.price IS NOT NULL
$$;

COMMIT;
