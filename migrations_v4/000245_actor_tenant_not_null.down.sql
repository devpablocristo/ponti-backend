BEGIN;

ALTER TABLE public.actors     ALTER COLUMN tenant_id DROP NOT NULL;
ALTER TABLE public.actor_keys ALTER COLUMN tenant_id DROP NOT NULL;

COMMIT;
