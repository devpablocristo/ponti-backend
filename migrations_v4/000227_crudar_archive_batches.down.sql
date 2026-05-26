BEGIN;

DROP INDEX IF EXISTS public.ux_customers_tenant_actor_id;

CREATE UNIQUE INDEX IF NOT EXISTS ux_customers_tenant_actor_id
    ON public.customers (tenant_id, actor_id)
    WHERE actor_id IS NOT NULL;

DO $$
DECLARE
    t text;
    mutable_tables text[] := ARRAY[
        'actors',
        'customers',
        'projects',
        'fields',
        'lots',
        'project_managers',
        'project_investors',
        'admin_cost_investors',
        'workorders',
        'labors',
        'supply_movements',
        'stocks',
        'crop_commercializations',
        'project_dollar_values'
    ];
BEGIN
    FOREACH t IN ARRAY mutable_tables LOOP
        IF to_regclass('public.' || t) IS NOT NULL THEN
            EXECUTE format('DROP INDEX IF EXISTS %I', 'idx_' || t || '_archive_cause');
            EXECUTE format('ALTER TABLE public.%I DROP CONSTRAINT IF EXISTS %I', t, t || '_archive_batch_id_fkey');
            EXECUTE format('ALTER TABLE public.%I DROP COLUMN IF EXISTS archive_reason', t);
            EXECUTE format('ALTER TABLE public.%I DROP COLUMN IF EXISTS archive_origin_id', t);
            EXECUTE format('ALTER TABLE public.%I DROP COLUMN IF EXISTS archive_origin_entity', t);
            EXECUTE format('ALTER TABLE public.%I DROP COLUMN IF EXISTS archive_batch_id', t);
        END IF;
    END LOOP;
END $$;

DROP TABLE IF EXISTS public.archive_batches;

COMMIT;
