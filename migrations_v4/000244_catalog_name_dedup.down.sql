BEGIN;

DROP TRIGGER IF EXISTS trg_prevent_dup_name ON public.crops;
DROP TRIGGER IF EXISTS trg_prevent_dup_name ON public.types;
DROP TRIGGER IF EXISTS trg_prevent_dup_name ON public.lease_types;
DROP TRIGGER IF EXISTS trg_prevent_dup_name ON public.categories;
DROP FUNCTION IF EXISTS public.prevent_duplicate_category_name();

COMMIT;
