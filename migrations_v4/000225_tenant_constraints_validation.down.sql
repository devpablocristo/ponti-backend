-- Roll back strict tenant validation to the additive nullable phase.

DO $$
DECLARE
    t text;
    tenant_tables text[] := ARRAY[
        'customers',
        'projects',
        'campaigns',
        'fields',
        'lots',
        'lot_dates',
        'workorders',
        'workorder_items',
        'workorder_investor_splits',
        'workorder_supply_items',
        'work_order_drafts',
        'project_managers',
        'project_investors',
        'admin_cost_investors',
        'field_investors',
        'labors',
        'supplies',
        'supply_movements',
        'stock_movements',
        'stocks',
        'invoices',
        'investors',
        'managers',
        'providers',
        'crops',
        'categories',
        'class_types',
        'lease_types',
        'business_parameters',
        'crop_commercializations',
        'project_dollar_values'
    ];
BEGIN
    FOREACH t IN ARRAY tenant_tables LOOP
        IF to_regclass('public.' || t) IS NOT NULL
           AND EXISTS (
                SELECT 1
                FROM information_schema.columns
                WHERE table_schema = 'public' AND table_name = t AND column_name = 'tenant_id'
           )
        THEN
            EXECUTE format('DROP INDEX IF EXISTS public.%I', 'uq_' || t || '_tenant_name_active');
            EXECUTE format('DROP INDEX IF EXISTS public.%I', 'uq_' || t || '_tenant_name');
            EXECUTE format('ALTER TABLE public.%I ALTER COLUMN tenant_id DROP NOT NULL', t);
        END IF;
    END LOOP;
END $$;
