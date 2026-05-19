BEGIN;

DROP INDEX IF EXISTS public.idx_actors_tenant_normalized_name;
DROP INDEX IF EXISTS public.ux_customers_tenant_actor_id;
DROP INDEX IF EXISTS public.idx_customers_actor_id;

ALTER TABLE IF EXISTS public.customers
    DROP CONSTRAINT IF EXISTS fk_customers_actor;

ALTER TABLE IF EXISTS public.customers
    DROP COLUMN IF EXISTS actor_id;

COMMIT;
