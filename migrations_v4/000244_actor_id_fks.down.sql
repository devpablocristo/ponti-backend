BEGIN;

ALTER TABLE public.customers  DROP CONSTRAINT IF EXISTS fk_customers_actor;
ALTER TABLE public.investors  DROP CONSTRAINT IF EXISTS fk_investors_actor;
ALTER TABLE public.managers   DROP CONSTRAINT IF EXISTS fk_managers_actor;
ALTER TABLE public.providers  DROP CONSTRAINT IF EXISTS fk_providers_actor;
ALTER TABLE public.workorders DROP CONSTRAINT IF EXISTS fk_workorders_actor;
ALTER TABLE public.labors     DROP CONSTRAINT IF EXISTS fk_labors_actor;
ALTER TABLE public.invoices   DROP CONSTRAINT IF EXISTS fk_invoices_actor;

ALTER TABLE public.customers  DROP COLUMN IF EXISTS actor_id;
ALTER TABLE public.investors  DROP COLUMN IF EXISTS actor_id;
ALTER TABLE public.managers   DROP COLUMN IF EXISTS actor_id;
ALTER TABLE public.providers  DROP COLUMN IF EXISTS actor_id;
ALTER TABLE public.workorders DROP COLUMN IF EXISTS contractor_actor_id;
ALTER TABLE public.labors     DROP COLUMN IF EXISTS contractor_actor_id;
ALTER TABLE public.invoices   DROP COLUMN IF EXISTS biller_actor_id;

COMMIT;
