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
  NEW.legacy_number := NULLIF(btrim(COALESCE(NEW.legacy_number, '')), '');

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
BEFORE INSERT OR UPDATE OF number, legacy_number
ON public.workorders
FOR EACH ROW
EXECUTE FUNCTION public.enforce_workorder_number_strict();

COMMIT;
