BEGIN;

DROP TRIGGER IF EXISTS trg_prevent_dup_name ON public.customers;
DROP TRIGGER IF EXISTS trg_prevent_dup_name ON public.campaigns;
DROP TRIGGER IF EXISTS trg_prevent_dup_name ON public.managers;
DROP TRIGGER IF EXISTS trg_prevent_dup_name ON public.investors;
DROP TRIGGER IF EXISTS trg_prevent_dup_name ON public.providers;

DROP FUNCTION IF EXISTS public.prevent_duplicate_name();
DROP FUNCTION IF EXISTS public.normalize_name(text);

COMMIT;
