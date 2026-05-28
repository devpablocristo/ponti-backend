-- Foundation for Ponti multi-tenant hardening.
-- This migration is intentionally additive and reversible:
-- 1) add nullable tenant_id columns to tenant-owned tables;
-- 2) backfill current data to the existing default tenant;
-- 3) add tenant indexes/FKs as NOT VALID;
-- 4) seed enterprise permissions/roles and audit/invite tables.

CREATE EXTENSION IF NOT EXISTS pgcrypto;

DO $$
DECLARE
    default_tenant uuid;
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
    SELECT id INTO default_tenant
    FROM public.auth_tenants
    WHERE name = 'default'
    ORDER BY created_at NULLS LAST
    LIMIT 1;

    IF default_tenant IS NULL THEN
        INSERT INTO public.auth_tenants (name, created_at, updated_at)
        VALUES ('default', now(), now())
        RETURNING id INTO default_tenant;
    END IF;

    FOREACH t IN ARRAY tenant_tables LOOP
        IF to_regclass('public.' || t) IS NOT NULL THEN
            EXECUTE format('ALTER TABLE public.%I ADD COLUMN IF NOT EXISTS tenant_id uuid', t);
            EXECUTE format('UPDATE public.%I SET tenant_id = $1 WHERE tenant_id IS NULL', t) USING default_tenant;
            EXECUTE format('CREATE INDEX IF NOT EXISTS idx_%I_tenant_id ON public.%I (tenant_id)', t, t);
            IF EXISTS (
                SELECT 1
                FROM information_schema.columns
                WHERE table_schema = 'public' AND table_name = t AND column_name = 'id'
            ) THEN
                EXECUTE format('CREATE INDEX IF NOT EXISTS idx_%I_tenant_id_id ON public.%I (tenant_id, id)', t, t);
            END IF;

            BEGIN
                EXECUTE format(
                    'ALTER TABLE public.%I ADD CONSTRAINT %I FOREIGN KEY (tenant_id) REFERENCES public.auth_tenants(id) ON DELETE RESTRICT NOT VALID',
                    t,
                    t || '_tenant_id_fkey'
                );
            EXCEPTION WHEN duplicate_object THEN
                NULL;
            END;

            IF EXISTS (
                SELECT 1
                FROM information_schema.columns
                WHERE table_schema = 'public' AND table_name = t AND column_name = 'name'
            ) THEN
                EXECUTE format('CREATE INDEX IF NOT EXISTS idx_%I_tenant_name ON public.%I (tenant_id, name)', t, t);
            END IF;
        END IF;
    END LOOP;

    IF to_regclass('public.project_managers') IS NOT NULL AND to_regclass('public.projects') IS NOT NULL THEN
        UPDATE public.project_managers pm
        SET tenant_id = p.tenant_id
        FROM public.projects p
        WHERE pm.project_id = p.id
          AND pm.tenant_id IS DISTINCT FROM p.tenant_id;
    END IF;

    IF to_regclass('public.project_investors') IS NOT NULL AND to_regclass('public.projects') IS NOT NULL THEN
        UPDATE public.project_investors pi
        SET tenant_id = p.tenant_id
        FROM public.projects p
        WHERE pi.project_id = p.id
          AND pi.tenant_id IS DISTINCT FROM p.tenant_id;
    END IF;

    IF to_regclass('public.admin_cost_investors') IS NOT NULL AND to_regclass('public.projects') IS NOT NULL THEN
        UPDATE public.admin_cost_investors aci
        SET tenant_id = p.tenant_id
        FROM public.projects p
        WHERE aci.project_id = p.id
          AND aci.tenant_id IS DISTINCT FROM p.tenant_id;
    END IF;

    IF to_regclass('public.field_investors') IS NOT NULL AND to_regclass('public.fields') IS NOT NULL THEN
        UPDATE public.field_investors fi
        SET tenant_id = f.tenant_id
        FROM public.fields f
        WHERE fi.field_id = f.id
          AND fi.tenant_id IS DISTINCT FROM f.tenant_id;
    END IF;
END $$;

INSERT INTO public.auth_roles (name)
VALUES
    ('saas_superadmin'),
    ('tenant_owner'),
    ('tenant_admin'),
    ('tenant_manager'),
    ('tenant_viewer')
ON CONFLICT (name) DO NOTHING;

INSERT INTO public.auth_permissions (name)
VALUES
    ('customers.read'), ('customers.write'), ('customers.archive'),
    ('projects.read'), ('projects.write'), ('projects.archive'),
    ('lots.read'), ('lots.write'), ('lots.archive'),
    ('workorders.read'), ('workorders.write'), ('workorders.archive'),
    ('labors.read'), ('labors.write'), ('labors.archive'),
    ('supplies.read'), ('supplies.write'), ('supplies.archive'),
    ('stock.read'), ('stock.write'), ('stock.archive'),
    ('actors.read'), ('actors.write'), ('actors.archive'), ('actors.merge'),
    ('admin.tenants'), ('admin.users'), ('admin.memberships'),
    ('exports.run'), ('imports.run'), ('ai.use')
ON CONFLICT (name) DO NOTHING;

DELETE FROM public.auth_role_permissions rp
USING public.auth_roles r, public.auth_permissions p
WHERE rp.role_id = r.id
  AND rp.permission_id = p.id
  AND r.name = 'tenant_owner'
  AND p.name = 'admin.tenants';

INSERT INTO public.auth_role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM public.auth_roles r
CROSS JOIN public.auth_permissions p
WHERE r.name = 'saas_superadmin'
ON CONFLICT (role_id, permission_id) DO NOTHING;

INSERT INTO public.auth_role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM public.auth_roles r
JOIN public.auth_permissions p ON p.name IN (
    'api.read', 'api.write',
    'customers.read', 'customers.write', 'customers.archive',
    'projects.read', 'projects.write', 'projects.archive',
    'lots.read', 'lots.write', 'lots.archive',
    'workorders.read', 'workorders.write', 'workorders.archive',
    'labors.read', 'labors.write', 'labors.archive',
    'supplies.read', 'supplies.write', 'supplies.archive',
    'stock.read', 'stock.write', 'stock.archive',
    'actors.read', 'actors.write', 'actors.archive',
    'admin.users', 'admin.memberships',
    'exports.run', 'imports.run', 'ai.use'
)
WHERE r.name IN ('admin', 'manager', 'tenant_owner', 'tenant_admin', 'tenant_manager')
ON CONFLICT (role_id, permission_id) DO NOTHING;

INSERT INTO public.auth_role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM public.auth_roles r
JOIN public.auth_permissions p ON p.name IN (
    'api.read',
    'customers.read', 'projects.read', 'lots.read', 'workorders.read',
    'labors.read', 'supplies.read', 'stock.read', 'actors.read',
    'exports.run', 'ai.use'
)
WHERE r.name IN ('viewer', 'tenant_viewer')
ON CONFLICT (role_id, permission_id) DO NOTHING;

CREATE TABLE IF NOT EXISTS public.tenant_invites (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id uuid NOT NULL REFERENCES public.auth_tenants(id) ON DELETE CASCADE,
    email text NOT NULL,
    role_id uuid NOT NULL REFERENCES public.auth_roles(id) ON DELETE RESTRICT,
    token_hash text NOT NULL UNIQUE,
    expires_at timestamptz NOT NULL,
    accepted_at timestamptz,
    revoked_at timestamptz,
    invited_by uuid REFERENCES public.users(id) ON DELETE SET NULL,
    accepted_by uuid REFERENCES public.users(id) ON DELETE SET NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_tenant_invites_tenant_status
    ON public.tenant_invites (tenant_id, accepted_at, revoked_at, expires_at);

CREATE TABLE IF NOT EXISTS public.security_audit_events (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id uuid REFERENCES public.auth_tenants(id) ON DELETE SET NULL,
    user_id uuid REFERENCES public.users(id) ON DELETE SET NULL,
    actor text,
    session_id text,
    event_type text NOT NULL,
    severity text NOT NULL DEFAULT 'info',
    resource_type text,
    resource_id text,
    ip_address inet,
    user_agent text,
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_security_audit_events_tenant_created
    ON public.security_audit_events (tenant_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_security_audit_events_type_created
    ON public.security_audit_events (event_type, created_at DESC);

CREATE TABLE IF NOT EXISTS public.auth_session_events (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id uuid REFERENCES public.auth_tenants(id) ON DELETE SET NULL,
    user_id uuid REFERENCES public.users(id) ON DELETE SET NULL,
    idp_sub text,
    session_id text,
    event_type text NOT NULL,
    ip_address inet,
    user_agent text,
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_auth_session_events_user_created
    ON public.auth_session_events (user_id, created_at DESC);
