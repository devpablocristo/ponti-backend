-- ========================================
-- MIGRATION 000216 FINALIZE WORKORDER NUMBER NORMALIZATION (DOWN)
-- ========================================
-- La renumeración de datos es irreversible. Este down revierte el endurecimiento
-- final del schema y vuelve al modo transicional de 000215.

BEGIN;

ALTER TABLE public.workorders
  ADD COLUMN IF NOT EXISTS legacy_number character varying(100);

DROP TRIGGER IF EXISTS trg_enforce_workorders_number_strict ON public.workorders;
DROP FUNCTION IF EXISTS public.enforce_workorder_number_strict();

CREATE OR REPLACE FUNCTION public.enforce_workorder_number_transition()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
  NEW.number := btrim(COALESCE(NEW.number, ''));
  NEW.legacy_number := NULLIF(btrim(COALESCE(NEW.legacy_number, '')), '');

  IF NEW.number = '' THEN
    RAISE EXCEPTION 'work order number is required';
  END IF;

  IF TG_OP = 'UPDATE' AND NEW.number = OLD.number THEN
    RETURN NEW;
  END IF;

  IF NEW.number !~ '^[0-9]+$' THEN
    RAISE EXCEPTION 'work order number must contain digits only';
  END IF;

  RETURN NEW;
END;
$$;

CREATE TRIGGER trg_enforce_workorders_number_transition
BEFORE INSERT OR UPDATE OF number, legacy_number
ON public.workorders
FOR EACH ROW
EXECUTE FUNCTION public.enforce_workorder_number_transition();

ALTER TABLE public.workorders
  DROP CONSTRAINT IF EXISTS chk_workorders_number_integer;

ALTER TABLE public.workorders
  ALTER COLUMN number DROP NOT NULL;

CREATE OR REPLACE FUNCTION v4_ssot.first_workorder_number_for_project(p_project_id bigint)
RETURNS text
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(NULLIF(btrim(legacy_number), ''), number)::text
  FROM public.workorders
  WHERE project_id = p_project_id
    AND deleted_at IS NULL
  ORDER BY date ASC, id ASC
  LIMIT 1
$$;

CREATE OR REPLACE FUNCTION v4_ssot.last_workorder_number_for_project(p_project_id bigint)
RETURNS text
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(NULLIF(btrim(legacy_number), ''), number)::text
  FROM public.workorders
  WHERE project_id = p_project_id
    AND deleted_at IS NULL
  ORDER BY date DESC, id DESC
  LIMIT 1
$$;

COMMIT;
