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
            EXECUTE format('DROP INDEX IF EXISTS %I', 'idx_' || t || '_archive_cause');
            EXECUTE format('ALTER TABLE public.%I DROP CONSTRAINT IF EXISTS %I', t, t || '_archive_batch_id_fkey');
            EXECUTE format('ALTER TABLE public.%I DROP COLUMN IF EXISTS archive_reason', t);
            EXECUTE format('ALTER TABLE public.%I DROP COLUMN IF EXISTS archive_origin_id', t);
            EXECUTE format('ALTER TABLE public.%I DROP COLUMN IF EXISTS archive_origin_entity', t);
            EXECUTE format('ALTER TABLE public.%I DROP COLUMN IF EXISTS archive_batch_id', t);
        END IF;
    END LOOP;
END $$;

COMMIT;
