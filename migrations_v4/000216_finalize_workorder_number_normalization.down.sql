-- ========================================
-- MIGRATION 000216 FINALIZE WORKORDER NUMBER NORMALIZATION (DOWN)
-- ========================================
-- La renumeración de datos es irreversible. Este down revierte el endurecimiento
-- final del schema y vuelve al modo transicional de 000215.

BEGIN;

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

COMMIT;
