BEGIN;

ALTER TABLE IF EXISTS public.customers
    ADD COLUMN IF NOT EXISTS actor_id bigint;

WITH unambiguous_customer_actor AS (
    SELECT m.tenant_id, m.source_id, m.actor_id
    FROM public.legacy_actor_map m
    WHERE m.source_table = 'customers'
      AND m.actor_id IS NOT NULL
      AND NOT EXISTS (
          SELECT 1
          FROM public.legacy_actor_map other
          WHERE other.tenant_id = m.tenant_id
            AND other.source_table = 'customers'
            AND other.actor_id = m.actor_id
            AND other.source_id IS DISTINCT FROM m.source_id
      )
)
UPDATE public.customers c
SET actor_id = m.actor_id
FROM unambiguous_customer_actor m
WHERE m.source_id = c.id
  AND m.tenant_id = c.tenant_id
  AND c.actor_id IS NULL;

UPDATE public.projects p
SET customer_actor_id = c.actor_id
FROM public.customers c
WHERE c.id = p.customer_id
  AND c.tenant_id = p.tenant_id
  AND c.actor_id IS NOT NULL
  AND p.customer_actor_id IS DISTINCT FROM c.actor_id;

ALTER TABLE IF EXISTS public.customers
    ADD CONSTRAINT fk_customers_actor
    FOREIGN KEY (actor_id) REFERENCES public.actors(id) ON DELETE RESTRICT NOT VALID;

CREATE INDEX IF NOT EXISTS idx_customers_actor_id
    ON public.customers (actor_id);

CREATE UNIQUE INDEX IF NOT EXISTS ux_customers_tenant_actor_id
    ON public.customers (tenant_id, actor_id)
    WHERE actor_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_actors_tenant_normalized_name
    ON public.actors (tenant_id, normalized_name)
    WHERE deleted_at IS NULL AND merged_into_actor_id IS NULL;

COMMIT;
