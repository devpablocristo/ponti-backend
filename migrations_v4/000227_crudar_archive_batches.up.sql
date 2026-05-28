BEGIN;

CREATE TABLE IF NOT EXISTS public.archive_batches (
    id bigserial PRIMARY KEY,
    tenant_id uuid NULL,
    root_entity text NOT NULL,
    root_id bigint NOT NULL,
    action text NOT NULL DEFAULT 'archive',
    reason text NULL,
    created_by text NULL,
    created_at timestamp with time zone NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_archive_batches_tenant_root
    ON public.archive_batches (tenant_id, root_entity, root_id);

CREATE INDEX IF NOT EXISTS idx_archive_batches_created_at
    ON public.archive_batches (created_at);

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
            EXECUTE format('ALTER TABLE public.%I ADD COLUMN IF NOT EXISTS archive_batch_id bigint NULL', t);
            EXECUTE format('ALTER TABLE public.%I ADD COLUMN IF NOT EXISTS archive_origin_entity text NULL', t);
            EXECUTE format('ALTER TABLE public.%I ADD COLUMN IF NOT EXISTS archive_origin_id bigint NULL', t);
            EXECUTE format('ALTER TABLE public.%I ADD COLUMN IF NOT EXISTS archive_reason text NULL', t);

            BEGIN
                EXECUTE format(
                    'ALTER TABLE public.%I ADD CONSTRAINT %I FOREIGN KEY (archive_batch_id) REFERENCES public.archive_batches(id) ON DELETE SET NULL NOT VALID',
                    t,
                    t || '_archive_batch_id_fkey'
                );
            EXCEPTION WHEN duplicate_object THEN
                NULL;
            END;

            EXECUTE format(
                'CREATE INDEX IF NOT EXISTS %I ON public.%I (archive_batch_id, archive_origin_entity, archive_origin_id)',
                'idx_' || t || '_archive_cause',
                t
            );
        END IF;
    END LOOP;
END $$;

-- Actors used archived_at as the semantic archive flag. From this migration on,
-- deleted_at is the lifecycle source of truth and archived_at is compatibility.
UPDATE public.actors
SET deleted_at = archived_at
WHERE archived_at IS NOT NULL
  AND deleted_at IS NULL;

-- Allow recreating/relinking a customer for an actor after the old customer is
-- archived. Active uniqueness remains enforced.
DROP INDEX IF EXISTS public.ux_customers_tenant_actor_id;

CREATE UNIQUE INDEX IF NOT EXISTS ux_customers_tenant_actor_id
    ON public.customers (tenant_id, actor_id)
    WHERE actor_id IS NOT NULL AND deleted_at IS NULL;

COMMIT;
