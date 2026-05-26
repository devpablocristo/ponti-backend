BEGIN;

DO $$
DECLARE
    t text;
    mutable_tables text[] := ARRAY[
        'campaigns',
        'supplies',
        'managers',
        'investors',
        'providers',
        'categories',
        'crops',
        'types',
        'lease_types',
        'business_parameters',
        'field_investors',
        'lot_dates',
        'workorder_items',
        'workorder_investor_splits',
        'invoices',
        'work_order_drafts',
        'work_order_draft_items',
        'work_order_draft_investor_splits'
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

COMMIT;
