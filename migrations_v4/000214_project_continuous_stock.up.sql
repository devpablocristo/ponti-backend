BEGIN;

CREATE TABLE IF NOT EXISTS public.supply_stock_counts (
    id bigserial PRIMARY KEY,
    supply_id bigint NOT NULL,
    counted_units numeric(15,3) NOT NULL,
    counted_at timestamp with time zone NOT NULL,
    note text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    created_by character varying(255),
    updated_by character varying(255),
    deleted_by character varying(255),
    CONSTRAINT fk_supply_stock_counts_supply
      FOREIGN KEY (supply_id) REFERENCES public.supplies(id)
      ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_supply_stock_counts_supply_id
    ON public.supply_stock_counts USING btree (supply_id)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_supply_stock_counts_counted_at
    ON public.supply_stock_counts USING btree (counted_at DESC)
    WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS public.supply_stock_count_backfill_conflicts (
    project_id bigint NOT NULL,
    supply_id bigint NOT NULL,
    conflicting_stock_ids bigint[] NOT NULL,
    detected_at timestamp with time zone DEFAULT now() NOT NULL,
    PRIMARY KEY (project_id, supply_id)
);

ALTER TABLE public.supply_movements
    ALTER COLUMN stock_id DROP NOT NULL;

ALTER TABLE public.supply_movements
    DROP CONSTRAINT IF EXISTS chk_supply_movements_movement_type;

ALTER TABLE public.supply_movements
    ADD CONSTRAINT chk_supply_movements_movement_type CHECK (
      movement_type = ANY (
        ARRAY[
          'Movimiento interno'::text,
          'Remito oficial'::text,
          'Movimiento interno entrada'::text,
          'Devolución'::text
        ]
      )
    ) NOT VALID;

WITH conflicting_groups AS (
    SELECT
        project_id,
        supply_id,
        ARRAY_AGG(id ORDER BY id) AS conflicting_stock_ids
    FROM public.stocks
    WHERE deleted_at IS NULL
      AND close_date IS NULL
      AND has_real_stock_count = TRUE
    GROUP BY project_id, supply_id
    HAVING COUNT(*) > 1
)
INSERT INTO public.supply_stock_count_backfill_conflicts (project_id, supply_id, conflicting_stock_ids)
SELECT project_id, supply_id, conflicting_stock_ids
FROM conflicting_groups
ON CONFLICT (project_id, supply_id) DO UPDATE
SET conflicting_stock_ids = EXCLUDED.conflicting_stock_ids,
    detected_at = now();

WITH single_active_rows AS (
    SELECT
        st.id AS legacy_stock_id,
        st.project_id,
        st.supply_id,
        st.real_stock_units AS counted_units,
        COALESCE(st.updated_at, st.created_at, now()) AS counted_at,
        NULLIF(COALESCE(st.updated_by::text, st.created_by::text, ''), '') AS actor
    FROM public.stocks st
    WHERE st.deleted_at IS NULL
      AND st.close_date IS NULL
      AND st.has_real_stock_count = TRUE
      AND NOT EXISTS (
        SELECT 1
        FROM public.supply_stock_count_backfill_conflicts conflict
        WHERE conflict.project_id = st.project_id
          AND conflict.supply_id = st.supply_id
      )
)
INSERT INTO public.supply_stock_counts (
    supply_id,
    counted_units,
    counted_at,
    note,
    created_at,
    updated_at,
    created_by,
    updated_by
)
SELECT
    s.supply_id,
    s.counted_units,
    s.counted_at,
    'Backfill desde legacy stocks row ' || s.legacy_stock_id,
    s.counted_at,
    s.counted_at,
    s.actor,
    s.actor
FROM single_active_rows s
WHERE NOT EXISTS (
    SELECT 1
    FROM public.supply_stock_counts c
    WHERE c.deleted_at IS NULL
      AND c.supply_id = s.supply_id
      AND c.counted_at = s.counted_at
      AND c.note = 'Backfill desde legacy stocks row ' || s.legacy_stock_id
);

CREATE OR REPLACE FUNCTION v4_ssot.movement_in_units_for_supply_as_of(
    p_project_id bigint,
    p_supply_id bigint,
    p_cutoff_date date DEFAULT NULL
) RETURNS numeric
LANGUAGE sql STABLE AS $$
    SELECT COALESCE(SUM(sm.quantity), 0)::numeric
    FROM public.supply_movements sm
    WHERE sm.project_id = p_project_id
      AND sm.supply_id = p_supply_id
      AND sm.deleted_at IS NULL
      AND sm.movement_type <> 'Stock'
      AND sm.quantity > 0
      AND (p_cutoff_date IS NULL OR sm.movement_date::date <= p_cutoff_date)
$$;

CREATE OR REPLACE FUNCTION v4_ssot.movement_out_units_for_supply_as_of(
    p_project_id bigint,
    p_supply_id bigint,
    p_cutoff_date date DEFAULT NULL
) RETURNS numeric
LANGUAGE sql STABLE AS $$
    SELECT COALESCE(SUM(ABS(sm.quantity)), 0)::numeric
    FROM public.supply_movements sm
    WHERE sm.project_id = p_project_id
      AND sm.supply_id = p_supply_id
      AND sm.deleted_at IS NULL
      AND sm.movement_type <> 'Stock'
      AND sm.quantity < 0
      AND (p_cutoff_date IS NULL OR sm.movement_date::date <= p_cutoff_date)
$$;

CREATE OR REPLACE FUNCTION v4_ssot.consumed_units_for_supply_as_of(
    p_project_id bigint,
    p_supply_id bigint,
    p_cutoff_date date DEFAULT NULL
) RETURNS numeric
LANGUAGE sql STABLE AS $$
    SELECT COALESCE(SUM(woi.total_used), 0)::numeric
    FROM public.workorder_items woi
    JOIN public.workorders wo ON wo.id = woi.workorder_id
    WHERE wo.project_id = p_project_id
      AND woi.supply_id = p_supply_id
      AND wo.deleted_at IS NULL
      AND woi.deleted_at IS NULL
      AND (p_cutoff_date IS NULL OR wo.date::date <= p_cutoff_date)
$$;

CREATE OR REPLACE FUNCTION v4_ssot.latest_stock_count_for_supply_as_of(
    p_supply_id bigint,
    p_cutoff_date date DEFAULT NULL
) RETURNS TABLE (
    counted_units numeric,
    counted_at timestamp with time zone
)
LANGUAGE sql STABLE AS $$
    SELECT c.counted_units, c.counted_at
    FROM public.supply_stock_counts c
    WHERE c.supply_id = p_supply_id
      AND c.deleted_at IS NULL
      AND (p_cutoff_date IS NULL OR c.counted_at::date <= p_cutoff_date)
    ORDER BY c.counted_at DESC, c.id DESC
    LIMIT 1
$$;

CREATE OR REPLACE FUNCTION v4_ssot.system_stock_units_for_supply_as_of(
    p_project_id bigint,
    p_supply_id bigint,
    p_cutoff_date date DEFAULT NULL
) RETURNS numeric
LANGUAGE sql STABLE AS $$
    SELECT (
        COALESCE(v4_ssot.movement_in_units_for_supply_as_of(p_project_id, p_supply_id, p_cutoff_date), 0)
        - COALESCE(v4_ssot.movement_out_units_for_supply_as_of(p_project_id, p_supply_id, p_cutoff_date), 0)
        - COALESCE(v4_ssot.consumed_units_for_supply_as_of(p_project_id, p_supply_id, p_cutoff_date), 0)
    )::numeric
$$;

CREATE OR REPLACE FUNCTION v4_report.stock_summary_for_project_as_of(
    p_project_id bigint,
    p_cutoff_date date DEFAULT NULL
) RETURNS TABLE (
    supply_id bigint,
    project_id bigint,
    supply_name text,
    class_type text,
    supply_unit_id bigint,
    supply_unit_name text,
    supply_unit_price numeric,
    entry_stock numeric,
    out_stock numeric,
    consumed numeric,
    stock_units numeric,
    real_stock_units numeric,
    has_real_stock_count boolean,
    last_count_at timestamp with time zone
)
LANGUAGE sql STABLE AS $$
    WITH base_supplies AS (
        SELECT
            s.id,
            s.project_id,
            s.name,
            s.unit_id,
            CASE
                WHEN s.unit_id = 1 THEN 'kg'
                WHEN s.unit_id = 2 THEN 'lt'
                ELSE ''
            END AS unit_name,
            COALESCE(s.price, 0)::numeric AS price,
            COALESCE(cat.name, '') AS class_type
        FROM public.supplies s
        LEFT JOIN public.categories cat ON cat.id = s.category_id
        WHERE s.project_id = p_project_id
          AND (
              s.deleted_at IS NULL
              OR EXISTS (
                  SELECT 1
                  FROM public.supply_movements sm
                  WHERE sm.project_id = p_project_id
                    AND sm.supply_id = s.id
                    AND sm.deleted_at IS NULL
              )
              OR EXISTS (
                  SELECT 1
                  FROM public.workorder_items woi
                  JOIN public.workorders wo ON wo.id = woi.workorder_id
                  WHERE wo.project_id = p_project_id
                    AND woi.supply_id = s.id
                    AND wo.deleted_at IS NULL
                    AND woi.deleted_at IS NULL
              )
              OR EXISTS (
                  SELECT 1
                  FROM public.supply_stock_counts ssc
                  WHERE ssc.supply_id = s.id
                    AND ssc.deleted_at IS NULL
              )
          )
    )
    SELECT
        b.id AS supply_id,
        b.project_id,
        b.name AS supply_name,
        b.class_type,
        b.unit_id AS supply_unit_id,
        b.unit_name AS supply_unit_name,
        b.price AS supply_unit_price,
        COALESCE(v4_ssot.movement_in_units_for_supply_as_of(p_project_id, b.id, p_cutoff_date), 0)::numeric AS entry_stock,
        COALESCE(v4_ssot.movement_out_units_for_supply_as_of(p_project_id, b.id, p_cutoff_date), 0)::numeric AS out_stock,
        COALESCE(v4_ssot.consumed_units_for_supply_as_of(p_project_id, b.id, p_cutoff_date), 0)::numeric AS consumed,
        COALESCE(v4_ssot.system_stock_units_for_supply_as_of(p_project_id, b.id, p_cutoff_date), 0)::numeric AS stock_units,
        COALESCE(latest.counted_units, 0)::numeric AS real_stock_units,
        (latest.counted_units IS NOT NULL) AS has_real_stock_count,
        latest.counted_at AS last_count_at
    FROM base_supplies b
    LEFT JOIN LATERAL v4_ssot.latest_stock_count_for_supply_as_of(b.id, p_cutoff_date) latest ON TRUE
$$;

CREATE OR REPLACE FUNCTION v4_ssot.last_stock_count_date_for_project(p_project_id bigint)
RETURNS date
LANGUAGE sql STABLE AS $$
    SELECT MAX(ssc.counted_at::date)
    FROM public.supply_stock_counts ssc
    JOIN public.supplies s ON s.id = ssc.supply_id
    WHERE s.project_id = p_project_id
      AND ssc.deleted_at IS NULL
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
        v4_ssot.stock_value_for_project(p_project_id)
        +
        v4_ssot.supply_cost_received_for_project(p_project_id)
    , 0)::numeric
$$;

CREATE OR REPLACE FUNCTION v4_ssot.stock_value_for_project(p_project_id bigint)
RETURNS numeric
LANGUAGE sql STABLE AS $$
    SELECT COALESCE(SUM(
        v4_ssot.system_stock_units_for_supply_as_of(p_project_id, s.id, NULL) * COALESCE(s.price, 0)
    ), 0)::numeric
    FROM public.supplies s
    WHERE s.project_id = p_project_id
      AND (
          s.deleted_at IS NULL
          OR EXISTS (
              SELECT 1
              FROM public.supply_movements sm
              WHERE sm.project_id = p_project_id
                AND sm.supply_id = s.id
                AND sm.deleted_at IS NULL
          )
          OR EXISTS (
              SELECT 1
              FROM public.workorder_items woi
              JOIN public.workorders wo ON wo.id = woi.workorder_id
              WHERE wo.project_id = p_project_id
                AND woi.supply_id = s.id
                AND wo.deleted_at IS NULL
                AND woi.deleted_at IS NULL
          )
          OR EXISTS (
              SELECT 1
              FROM public.supply_stock_counts ssc
              WHERE ssc.supply_id = s.id
                AND ssc.deleted_at IS NULL
          )
      )
$$;

CREATE OR REPLACE FUNCTION v4_ssot.stock_value_for_project_mb(p_project_id bigint)
RETURNS numeric
LANGUAGE sql STABLE AS $$
    SELECT v4_ssot.stock_value_for_project(p_project_id) - v4_ssot.direct_costs_total_for_project(p_project_id)
$$;

COMMIT;
