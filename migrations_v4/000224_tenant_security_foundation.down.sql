DROP TABLE IF EXISTS public.auth_session_events;
DROP TABLE IF EXISTS public.security_audit_events;
DROP TABLE IF EXISTS public.tenant_invites;

DELETE FROM public.auth_role_permissions rp
USING public.auth_permissions p
WHERE rp.permission_id = p.id
  AND p.name IN (
    'customers.read', 'customers.write', 'customers.archive',
    'projects.read', 'projects.write', 'projects.archive',
    'lots.read', 'lots.write', 'lots.archive',
    'workorders.read', 'workorders.write', 'workorders.archive',
    'labors.read', 'labors.write', 'labors.archive',
    'supplies.read', 'supplies.write', 'supplies.archive',
    'stock.read', 'stock.write', 'stock.archive',
    'actors.read', 'actors.write', 'actors.archive', 'actors.merge',
    'admin.tenants', 'admin.users', 'admin.memberships',
    'exports.run', 'imports.run', 'ai.use'
  );

DELETE FROM public.auth_permissions
WHERE name IN (
    'customers.read', 'customers.write', 'customers.archive',
    'projects.read', 'projects.write', 'projects.archive',
    'lots.read', 'lots.write', 'lots.archive',
    'workorders.read', 'workorders.write', 'workorders.archive',
    'labors.read', 'labors.write', 'labors.archive',
    'supplies.read', 'supplies.write', 'supplies.archive',
    'stock.read', 'stock.write', 'stock.archive',
    'actors.read', 'actors.write', 'actors.archive', 'actors.merge',
    'admin.tenants', 'admin.users', 'admin.memberships',
    'exports.run', 'imports.run', 'ai.use'
);

DELETE FROM public.auth_roles
WHERE name IN ('saas_superadmin', 'tenant_owner', 'tenant_admin', 'tenant_manager', 'tenant_viewer');

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
        IF to_regclass('public.' || t) IS NOT NULL THEN
            EXECUTE format('ALTER TABLE public.%I DROP CONSTRAINT IF EXISTS %I', t, t || '_tenant_id_fkey');
            EXECUTE format('DROP INDEX IF EXISTS public.idx_%I_tenant_name', t);
            EXECUTE format('DROP INDEX IF EXISTS public.idx_%I_tenant_id_id', t);
            EXECUTE format('DROP INDEX IF EXISTS public.idx_%I_tenant_id', t);
            EXECUTE format('ALTER TABLE public.%I DROP COLUMN IF EXISTS tenant_id', t);
        END IF;
    END LOOP;
END $$;
