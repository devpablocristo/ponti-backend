-- =============================================================================
-- MIGRACIÓN 000328: Rollback costo de insumos base (final_dose * area)
-- =============================================================================
--
-- Propósito: Restaurar lógica anterior de supply_cost_for_lot_base
-- Nota: Comentarios en español, código en inglés
--

BEGIN;

-- Restaurar fórmula anterior (final_dose × effective_area × price)
CREATE OR REPLACE FUNCTION v3_lot_ssot.supply_cost_for_lot_base(p_lot_id bigint)
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    SUM(v3_core_ssot.supply_cost(
      wi.final_dose::double precision,
      s.price::numeric,
      w.effective_area::numeric
    )), 0
  )::numeric
  FROM public.workorders w
  LEFT JOIN public.workorder_items wi ON wi.workorder_id = w.id AND wi.deleted_at IS NULL
  LEFT JOIN public.supplies s ON s.id = wi.supply_id AND s.deleted_at IS NULL
  WHERE w.lot_id = p_lot_id
    AND w.deleted_at IS NULL
$$;

COMMIT;
