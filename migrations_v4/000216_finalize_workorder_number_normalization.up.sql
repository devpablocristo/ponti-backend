-- ========================================
-- MIGRATION 000216 FINALIZE WORKORDER NUMBER NORMALIZATION (UP)
-- ========================================
-- Renumera únicamente workorders con número legacy/no entero, preserva el valor
-- visible anterior en legacy_number y endurece definitivamente el schema para que
-- number sea siempre entero oficial.

BEGIN;

DO $$
DECLARE
  project_row RECORD;
  workorder_row RECORD;
  candidate_number bigint;
  project_max bigint;
BEGIN
  CREATE TEMP TABLE tmp_workorder_number_assignments (
    workorder_id bigint PRIMARY KEY,
    project_id bigint NOT NULL,
    old_number text NOT NULL,
    new_number text NOT NULL
  ) ON COMMIT DROP;

  CREATE TEMP TABLE tmp_project_used_numbers (
    project_id bigint NOT NULL,
    assigned_number bigint NOT NULL,
    PRIMARY KEY (project_id, assigned_number)
  ) ON COMMIT DROP;

  INSERT INTO tmp_project_used_numbers (project_id, assigned_number)
  SELECT DISTINCT
    w.project_id,
    w.number::bigint
  FROM public.workorders w
  WHERE w.number ~ '^\d+$';

  FOR project_row IN
    SELECT DISTINCT w.project_id
    FROM public.workorders w
    ORDER BY w.project_id
  LOOP
    SELECT COALESCE(MAX(assigned_number), 0)
    INTO project_max
    FROM tmp_project_used_numbers
    WHERE project_id = project_row.project_id;

    FOR workorder_row IN
      SELECT
        w.id,
        w.number::text AS current_number,
        CASE
          WHEN w.number ~ '^\d+\.\d+$' THEN split_part(w.number::text, '.', 1)::bigint
          ELSE NULL
        END AS base_number,
        CASE
          WHEN w.number ~ '^\d+\.\d+$' THEN split_part(w.number::text, '.', 2)::bigint
          ELSE NULL
        END AS suffix_number,
        w.deleted_at,
        w.created_at
      FROM public.workorders w
      WHERE w.project_id = project_row.project_id
        AND w.number !~ '^\d+$'
      ORDER BY
        CASE WHEN w.deleted_at IS NULL THEN 0 ELSE 1 END,
        CASE WHEN w.number ~ '^\d+\.\d+$' THEN 0 ELSE 1 END,
        CASE WHEN w.number ~ '^\d+\.\d+$' THEN split_part(w.number::text, '.', 1)::bigint ELSE NULL END NULLS LAST,
        CASE WHEN w.number ~ '^\d+\.\d+$' THEN split_part(w.number::text, '.', 2)::bigint ELSE NULL END NULLS LAST,
        w.created_at,
        w.id
    LOOP
      IF workorder_row.base_number IS NOT NULL THEN
        candidate_number := workorder_row.base_number + COALESCE(workorder_row.suffix_number, 0);
      ELSE
        candidate_number := project_max + 1;
      END IF;

      IF candidate_number < 1 THEN
        candidate_number := 1;
      END IF;

      WHILE EXISTS (
        SELECT 1
        FROM tmp_project_used_numbers u
        WHERE u.project_id = project_row.project_id
          AND u.assigned_number = candidate_number
      ) LOOP
        candidate_number := candidate_number + 1;
      END LOOP;

      INSERT INTO tmp_project_used_numbers (project_id, assigned_number)
      VALUES (project_row.project_id, candidate_number);

      INSERT INTO tmp_workorder_number_assignments (
        workorder_id,
        project_id,
        old_number,
        new_number
      )
      VALUES (
        workorder_row.id,
        project_row.project_id,
        workorder_row.current_number,
        candidate_number::text
      );

      project_max := GREATEST(project_max, candidate_number);
    END LOOP;
  END LOOP;

  UPDATE public.workorders w
  SET
    legacy_number = COALESCE(NULLIF(btrim(w.legacy_number), ''), btrim(w.number)),
    number = assignments.new_number,
    updated_at = NOW()
  FROM tmp_workorder_number_assignments assignments
  WHERE w.id = assignments.workorder_id;
END $$;

ALTER TABLE public.workorders
  ALTER COLUMN number SET NOT NULL;

ALTER TABLE public.workorders
  DROP CONSTRAINT IF EXISTS chk_workorders_number_integer;

ALTER TABLE public.workorders
  ADD CONSTRAINT chk_workorders_number_integer
  CHECK (number ~ '^\d+$');

CREATE OR REPLACE FUNCTION public.enforce_workorder_number_strict()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
  NEW.number := btrim(COALESCE(NEW.number, ''));

  IF NEW.number = '' THEN
    RAISE EXCEPTION 'work order number is required';
  END IF;

  IF NEW.number !~ '^[0-9]+$' THEN
    RAISE EXCEPTION 'work order number must contain digits only';
  END IF;

  RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS trg_enforce_workorders_number_transition ON public.workorders;
DROP TRIGGER IF EXISTS trg_enforce_workorders_number_strict ON public.workorders;

CREATE TRIGGER trg_enforce_workorders_number_strict
BEFORE INSERT OR UPDATE OF number
ON public.workorders
FOR EACH ROW
EXECUTE FUNCTION public.enforce_workorder_number_strict();

CREATE OR REPLACE FUNCTION v4_ssot.first_workorder_number_for_project(p_project_id bigint)
RETURNS text
LANGUAGE sql STABLE AS $$
  SELECT number::text
  FROM public.workorders
  WHERE project_id = p_project_id
    AND deleted_at IS NULL
  ORDER BY date ASC, id ASC
  LIMIT 1
$$;

CREATE OR REPLACE FUNCTION v4_ssot.last_workorder_number_for_project(p_project_id bigint)
RETURNS text
LANGUAGE sql STABLE AS $$
  SELECT number::text
  FROM public.workorders
  WHERE project_id = p_project_id
    AND deleted_at IS NULL
  ORDER BY date DESC, id DESC
  LIMIT 1
$$;

CREATE OR REPLACE VIEW v4_report.dashboard_operational_indicators AS
SELECT
  p.id AS project_id,
  v4_ssot.first_workorder_date_for_project(p.id) AS start_date,
  lw.date AS end_date,
  v4_core.calculate_campaign_closing_date(
    v4_ssot.last_workorder_date_for_project(p.id)
  ) AS campaign_closing_date,
  v4_ssot.first_workorder_number_for_project(p.id) AS first_workorder_id,
  lw.number::text AS last_workorder_id,
  v4_ssot.last_stock_count_date_for_project(p.id) AS last_stock_count_date
FROM public.projects p
LEFT JOIN LATERAL (
  SELECT
    w.date,
    w.number
  FROM public.workorders w
  WHERE w.project_id = p.id
    AND w.deleted_at IS NULL
  ORDER BY w.created_at DESC, w.id DESC
  LIMIT 1
) lw ON TRUE
WHERE p.deleted_at IS NULL;

CREATE OR REPLACE VIEW v4_report.dashboard_operational_indicators_field AS
SELECT
  p.id AS project_id,
  p.customer_id,
  p.campaign_id,
  f.id AS field_id,
  (SELECT w2.date
   FROM public.workorders w2
   WHERE w2.field_id = f.id
     AND w2.deleted_at IS NULL
   ORDER BY w2.date ASC, w2.id ASC
   LIMIT 1) AS start_date,
  lw.date AS end_date,
  v4_core.calculate_campaign_closing_date(
    (SELECT w2.date
     FROM public.workorders w2
     WHERE w2.field_id = f.id
       AND w2.deleted_at IS NULL
     ORDER BY w2.date DESC, w2.id DESC
     LIMIT 1)
  ) AS campaign_closing_date,
  (SELECT w2.number::text
   FROM public.workorders w2
   WHERE w2.field_id = f.id
     AND w2.deleted_at IS NULL
   ORDER BY w2.date ASC, w2.id ASC
   LIMIT 1) AS first_workorder_id,
  lw.number::text AS last_workorder_id,
  v4_ssot.last_stock_count_date_for_project(p.id) AS last_stock_count_date
FROM public.projects p
JOIN public.fields f ON f.project_id = p.id AND f.deleted_at IS NULL
LEFT JOIN LATERAL (
  SELECT
    w2.date,
    w2.number
  FROM public.workorders w2
  WHERE w2.field_id = f.id
    AND w2.deleted_at IS NULL
  ORDER BY w2.created_at DESC, w2.id DESC
  LIMIT 1
) lw ON TRUE
WHERE p.deleted_at IS NULL;

CREATE OR REPLACE VIEW v4_report.labor_list AS
WITH workorder_alloc AS (
  SELECT
    w.id AS workorder_id,
    w.investor_id,
    1::numeric AS factor,
    NULL::text AS investor_payment_status,
    false AS investor_payment_enabled
  FROM public.workorders w
  WHERE w.deleted_at IS NULL
    AND NOT EXISTS (
      SELECT 1
      FROM public.workorder_investor_splits wis
      WHERE wis.workorder_id = w.id
        AND wis.deleted_at IS NULL
    )
  UNION ALL
  SELECT
    w.id AS workorder_id,
    wis.investor_id,
    (wis.percentage::numeric / 100)::numeric AS factor,
    wis.payment_status AS investor_payment_status,
    true AS investor_payment_enabled
  FROM public.workorders w
  JOIN public.workorder_investor_splits wis
    ON wis.workorder_id = w.id
   AND wis.deleted_at IS NULL
  WHERE w.deleted_at IS NULL
)
SELECT
  w.id AS workorder_id,
  w.number::character varying(100) AS workorder_number,
  w.date,
  w.project_id,
  p.name AS project_name,
  w.field_id,
  f.name AS field_name,
  w.lot_id,
  l.name AS lot_name,
  w.crop_id,
  c.name AS crop_name,
  w.labor_id,
  lb.name AS labor_name,
  lb.category_id AS labor_category_id,
  cat.name AS labor_category_name,
  w.contractor,
  lb.contractor_name,
  (w.effective_area * a.factor)::numeric(18,6) AS surface_ha,
  lb.price AS cost_per_ha,
  (lb.price * w.effective_area * a.factor)::numeric AS total_labor_cost,
  v4_core.dollar_average_for_month(w.project_id, w.date) AS dollar_average_month,
  lb.price::numeric AS usd_cost_ha,
  (lb.price * w.effective_area * a.factor)::numeric AS usd_net_total,
  a.investor_id,
  i.name AS investor_name,
  a.investor_payment_status,
  a.investor_payment_enabled
FROM public.workorders w
JOIN workorder_alloc a ON a.workorder_id = w.id
JOIN public.projects p ON p.id = w.project_id AND p.deleted_at IS NULL
JOIN public.fields f ON f.id = w.field_id AND f.deleted_at IS NULL
LEFT JOIN public.lots l ON l.id = w.lot_id AND l.deleted_at IS NULL
LEFT JOIN public.crops c ON c.id = w.crop_id AND c.deleted_at IS NULL
JOIN public.labors lb ON lb.id = w.labor_id AND lb.deleted_at IS NULL
LEFT JOIN public.categories cat ON cat.id = lb.category_id AND cat.deleted_at IS NULL
LEFT JOIN public.investors i ON i.id = a.investor_id AND i.deleted_at IS NULL
WHERE w.deleted_at IS NULL
  AND w.effective_area IS NOT NULL
  AND w.effective_area > 0
  AND lb.price IS NOT NULL;

CREATE OR REPLACE VIEW v4_report.workorder_list AS
SELECT
  w.id,
  w.number::character varying(100) AS number,
  w.project_id,
  w.field_id,
  p.name AS project_name,
  f.name AS field_name,
  l.name AS lot_name,
  w.date,
  c.name AS crop_name,
  lb.name AS labor_name,
  cat_lb.name AS labor_category_name,
  t.name AS type_name,
  w.contractor,
  w.effective_area AS surface_ha,
  s.name AS supply_name,
  wi.total_used AS consumption,
  cat.name AS category_name,
  wi.final_dose AS dose_per_ha,
  s.price AS unit_price,
  CASE
    WHEN wi.final_dose IS NOT NULL AND s.price IS NOT NULL
    THEN v4_core.cost_per_ha(
      (wi.final_dose::numeric * s.price)::numeric,
      1::numeric
    )
    ELSE 0
  END AS supply_cost_per_ha,
  v4_core.supply_cost(
    wi.final_dose::numeric,
    s.price::numeric,
    w.effective_area::numeric
  ) AS supply_total_cost
FROM public.workorders w
JOIN public.projects p ON p.id = w.project_id AND p.deleted_at IS NULL
JOIN public.fields f ON f.id = w.field_id AND f.deleted_at IS NULL
JOIN public.lots l ON l.id = w.lot_id AND l.deleted_at IS NULL
JOIN public.crops c ON c.id = w.crop_id AND c.deleted_at IS NULL
JOIN public.labors lb ON lb.id = w.labor_id AND lb.deleted_at IS NULL
JOIN public.categories cat_lb ON cat_lb.id = lb.category_id AND cat_lb.deleted_at IS NULL
LEFT JOIN public.workorder_items wi ON wi.workorder_id = w.id AND wi.deleted_at IS NULL
LEFT JOIN public.supplies s ON s.id = wi.supply_id AND s.deleted_at IS NULL
LEFT JOIN public.types t ON t.id = s.type_id AND t.deleted_at IS NULL
LEFT JOIN public.categories cat ON cat.id = s.category_id AND cat.deleted_at IS NULL
WHERE w.deleted_at IS NULL

UNION ALL

SELECT
  w.id,
  w.number::character varying(100) AS number,
  w.project_id,
  w.field_id,
  p.name AS project_name,
  f.name AS field_name,
  l.name AS lot_name,
  w.date,
  c.name AS crop_name,
  lb.name AS labor_name,
  cat_lb.name AS labor_category_name,
  'Labor'::character varying(250) AS type_name,
  w.contractor,
  w.effective_area AS surface_ha,
  lb.name::character varying(100) AS supply_name,
  0::numeric(18,6) AS consumption,
  cat_lb.name AS category_name,
  0::numeric(18,6) AS dose_per_ha,
  lb.price::numeric(18,6) AS unit_price,
  lb.price AS supply_cost_per_ha,
  lb.price * w.effective_area AS supply_total_cost
FROM public.workorders w
JOIN public.projects p ON p.id = w.project_id AND p.deleted_at IS NULL
JOIN public.fields f ON f.id = w.field_id AND f.deleted_at IS NULL
JOIN public.lots l ON l.id = w.lot_id AND l.deleted_at IS NULL
JOIN public.crops c ON c.id = w.crop_id AND c.deleted_at IS NULL
JOIN public.labors lb ON lb.id = w.labor_id AND lb.deleted_at IS NULL
JOIN public.categories cat_lb ON cat_lb.id = lb.category_id AND cat_lb.deleted_at IS NULL
WHERE w.deleted_at IS NULL
  AND w.effective_area IS NOT NULL
  AND w.effective_area > 0::numeric
  AND lb.price IS NOT NULL;

ALTER TABLE public.workorders
  DROP COLUMN IF EXISTS legacy_number;

COMMIT;
