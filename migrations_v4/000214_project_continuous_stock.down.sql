BEGIN;

DROP FUNCTION IF EXISTS v4_report.stock_summary_for_project_as_of(bigint, date);
DROP FUNCTION IF EXISTS v4_ssot.system_stock_units_for_supply_as_of(bigint, bigint, date);
DROP FUNCTION IF EXISTS v4_ssot.latest_stock_count_for_supply_as_of(bigint, date);
DROP FUNCTION IF EXISTS v4_ssot.consumed_units_for_supply_as_of(bigint, bigint, date);
DROP FUNCTION IF EXISTS v4_ssot.movement_out_units_for_supply_as_of(bigint, bigint, date);
DROP FUNCTION IF EXISTS v4_ssot.movement_in_units_for_supply_as_of(bigint, bigint, date);

DROP TABLE IF EXISTS public.supply_stock_count_backfill_conflicts;
DROP TABLE IF EXISTS public.supply_stock_counts;

ALTER TABLE public.supply_movements
    DROP CONSTRAINT IF EXISTS chk_supply_movements_movement_type;

ALTER TABLE public.supply_movements
    ADD CONSTRAINT chk_supply_movements_movement_type CHECK (
      movement_type = ANY (
        ARRAY[
          'Stock'::text,
          'Movimiento interno'::text,
          'Remito oficial'::text,
          'Movimiento interno entrada'::text,
          'Devolución'::text
        ]
      )
    );

CREATE OR REPLACE FUNCTION v4_ssot.last_stock_count_date_for_project(p_project_id bigint)
RETURNS date
LANGUAGE sql STABLE AS $$
  SELECT MAX(close_date)
  FROM public.stocks
  WHERE project_id = p_project_id
    AND deleted_at IS NULL
$$;

CREATE OR REPLACE FUNCTION v4_ssot.direct_costs_invested_for_project(p_project_id bigint)
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    (SELECT COALESCE(SUM(lb.price * l.hectares), 0)::numeric
     FROM public.lots l
     JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
     JOIN public.labors lb ON lb.project_id = f.project_id AND lb.deleted_at IS NULL
     WHERE f.project_id = p_project_id AND l.deleted_at IS NULL)
    +
    (SELECT COALESCE(SUM(s.price * st.initial_units), 0)::numeric
     FROM public.supplies s
     JOIN public.stocks st ON st.supply_id = s.id AND st.deleted_at IS NULL
     WHERE s.project_id = p_project_id AND s.deleted_at IS NULL
       AND st.initial_units IS NOT NULL)
    +
    v4_ssot.supply_cost_received_for_project(p_project_id)
  , 0)::numeric
$$;

CREATE OR REPLACE FUNCTION v4_ssot.stock_value_for_project(p_project_id bigint)
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    (SELECT COALESCE(SUM(s.price * st.initial_units), 0)::numeric
     FROM public.supplies s
     JOIN public.stocks st ON st.supply_id = s.id AND st.deleted_at IS NULL
     WHERE s.project_id = p_project_id
       AND s.deleted_at IS NULL
       AND st.initial_units IS NOT NULL)
    -
    (SELECT COALESCE(SUM(wi.total_used * s.price), 0)::numeric
     FROM public.workorders w
     JOIN public.workorder_items wi ON wi.workorder_id = w.id AND wi.deleted_at IS NULL
     JOIN public.supplies s ON s.id = wi.supply_id
     WHERE w.project_id = p_project_id AND w.deleted_at IS NULL)
  , 0)::numeric
$$;

CREATE OR REPLACE FUNCTION v4_ssot.stock_value_for_project_mb(p_project_id bigint)
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    (SELECT SUM(sm.quantity * s.price)
     FROM public.supply_movements sm
     JOIN public.supplies s ON s.id = sm.supply_id
     WHERE sm.project_id = p_project_id
       AND sm.deleted_at IS NULL
       AND s.deleted_at IS NULL
       AND sm.movement_type IN ('Stock', 'Remito oficial', 'Devolución'))
  , 0)::numeric - v4_ssot.direct_costs_total_for_project(p_project_id)
$$;

COMMIT;
