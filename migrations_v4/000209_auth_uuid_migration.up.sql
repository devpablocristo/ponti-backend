-- =============================================================================
-- Migration 000199: auth tables bigint -> UUID + actor columns bigint/varchar -> text
-- Prepares ponti-backend for core/saas/go, while preserving legacy IDs for rollback.
-- =============================================================================
-- NOTE: no explicit BEGIN/COMMIT. golang-migrate wraps the migration in a transaction.

CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- =============================================================================
-- 1. Drop audit FKs pointing to public.users(id)
-- =============================================================================
ALTER TABLE public.customers DROP CONSTRAINT IF EXISTS fk_customers_created_by;
ALTER TABLE public.customers DROP CONSTRAINT IF EXISTS fk_customers_updated_by;
ALTER TABLE public.customers DROP CONSTRAINT IF EXISTS fk_customers_deleted_by;
ALTER TABLE public.campaigns DROP CONSTRAINT IF EXISTS fk_campaigns_created_by;
ALTER TABLE public.campaigns DROP CONSTRAINT IF EXISTS fk_campaigns_updated_by;
ALTER TABLE public.campaigns DROP CONSTRAINT IF EXISTS fk_campaigns_deleted_by;
ALTER TABLE public.projects DROP CONSTRAINT IF EXISTS fk_projects_created_by;
ALTER TABLE public.projects DROP CONSTRAINT IF EXISTS fk_projects_updated_by;
ALTER TABLE public.projects DROP CONSTRAINT IF EXISTS fk_projects_deleted_by;
ALTER TABLE public.managers DROP CONSTRAINT IF EXISTS fk_managers_created_by;
ALTER TABLE public.managers DROP CONSTRAINT IF EXISTS fk_managers_updated_by;
ALTER TABLE public.managers DROP CONSTRAINT IF EXISTS fk_managers_deleted_by;
ALTER TABLE public.project_managers DROP CONSTRAINT IF EXISTS fk_project_managers_created_by;
ALTER TABLE public.project_managers DROP CONSTRAINT IF EXISTS fk_project_managers_updated_by;
ALTER TABLE public.project_managers DROP CONSTRAINT IF EXISTS fk_project_managers_deleted_by;
ALTER TABLE public.lease_types DROP CONSTRAINT IF EXISTS fk_lease_types_created_by;
ALTER TABLE public.lease_types DROP CONSTRAINT IF EXISTS fk_lease_types_updated_by;
ALTER TABLE public.lease_types DROP CONSTRAINT IF EXISTS fk_lease_types_deleted_by;
ALTER TABLE public.fields DROP CONSTRAINT IF EXISTS fk_fields_created_by;
ALTER TABLE public.fields DROP CONSTRAINT IF EXISTS fk_fields_updated_by;
ALTER TABLE public.fields DROP CONSTRAINT IF EXISTS fk_fields_deleted_by;
ALTER TABLE public.lots DROP CONSTRAINT IF EXISTS fk_lots_created_by;
ALTER TABLE public.lots DROP CONSTRAINT IF EXISTS fk_lots_updated_by;
ALTER TABLE public.lots DROP CONSTRAINT IF EXISTS fk_lots_deleted_by;
ALTER TABLE public.lot_dates DROP CONSTRAINT IF EXISTS fk_lot_dates_created_by;
ALTER TABLE public.lot_dates DROP CONSTRAINT IF EXISTS fk_lot_dates_updated_by;
ALTER TABLE public.lot_dates DROP CONSTRAINT IF EXISTS fk_lot_dates_deleted_by;
ALTER TABLE public.crops DROP CONSTRAINT IF EXISTS fk_crops_created_by;
ALTER TABLE public.crops DROP CONSTRAINT IF EXISTS fk_crops_updated_by;
ALTER TABLE public.crops DROP CONSTRAINT IF EXISTS fk_crops_deleted_by;
ALTER TABLE public.investors DROP CONSTRAINT IF EXISTS fk_investors_created_by;
ALTER TABLE public.investors DROP CONSTRAINT IF EXISTS fk_investors_updated_by;
ALTER TABLE public.investors DROP CONSTRAINT IF EXISTS fk_investors_deleted_by;
ALTER TABLE public.project_investors DROP CONSTRAINT IF EXISTS fk_project_investors_created_by;
ALTER TABLE public.project_investors DROP CONSTRAINT IF EXISTS fk_project_investors_updated_by;
ALTER TABLE public.project_investors DROP CONSTRAINT IF EXISTS fk_project_investors_deleted_by;

ALTER TABLE public.work_order_drafts DROP CONSTRAINT IF EXISTS fk_work_order_drafts_reviewed_by;

-- =============================================================================
-- 2. Convert actor columns to text
-- =============================================================================
ALTER TABLE public.users ALTER COLUMN created_by TYPE text USING created_by::text;
ALTER TABLE public.users ALTER COLUMN updated_by TYPE text USING updated_by::text;
ALTER TABLE public.users ALTER COLUMN deleted_by TYPE text USING deleted_by::text;

ALTER TABLE public.customers ALTER COLUMN created_by TYPE text USING created_by::text;
ALTER TABLE public.customers ALTER COLUMN updated_by TYPE text USING updated_by::text;
ALTER TABLE public.customers ALTER COLUMN deleted_by TYPE text USING deleted_by::text;

ALTER TABLE public.campaigns ALTER COLUMN created_by TYPE text USING created_by::text;
ALTER TABLE public.campaigns ALTER COLUMN updated_by TYPE text USING updated_by::text;
ALTER TABLE public.campaigns ALTER COLUMN deleted_by TYPE text USING deleted_by::text;

ALTER TABLE public.projects ALTER COLUMN created_by TYPE text USING created_by::text;
ALTER TABLE public.projects ALTER COLUMN updated_by TYPE text USING updated_by::text;
ALTER TABLE public.projects ALTER COLUMN deleted_by TYPE text USING deleted_by::text;

ALTER TABLE public.managers ALTER COLUMN created_by TYPE text USING created_by::text;
ALTER TABLE public.managers ALTER COLUMN updated_by TYPE text USING updated_by::text;
ALTER TABLE public.managers ALTER COLUMN deleted_by TYPE text USING deleted_by::text;

ALTER TABLE public.project_managers ALTER COLUMN created_by TYPE text USING created_by::text;
ALTER TABLE public.project_managers ALTER COLUMN updated_by TYPE text USING updated_by::text;
ALTER TABLE public.project_managers ALTER COLUMN deleted_by TYPE text USING deleted_by::text;

ALTER TABLE public.lease_types ALTER COLUMN created_by TYPE text USING created_by::text;
ALTER TABLE public.lease_types ALTER COLUMN updated_by TYPE text USING updated_by::text;
ALTER TABLE public.lease_types ALTER COLUMN deleted_by TYPE text USING deleted_by::text;

ALTER TABLE public.fields ALTER COLUMN created_by TYPE text USING created_by::text;
ALTER TABLE public.fields ALTER COLUMN updated_by TYPE text USING updated_by::text;
ALTER TABLE public.fields ALTER COLUMN deleted_by TYPE text USING deleted_by::text;

ALTER TABLE public.lots ALTER COLUMN created_by TYPE text USING created_by::text;
ALTER TABLE public.lots ALTER COLUMN updated_by TYPE text USING updated_by::text;
ALTER TABLE public.lots ALTER COLUMN deleted_by TYPE text USING deleted_by::text;

ALTER TABLE public.lot_dates ALTER COLUMN created_by TYPE text USING created_by::text;
ALTER TABLE public.lot_dates ALTER COLUMN updated_by TYPE text USING updated_by::text;
ALTER TABLE public.lot_dates ALTER COLUMN deleted_by TYPE text USING deleted_by::text;

ALTER TABLE public.crops ALTER COLUMN created_by TYPE text USING created_by::text;
ALTER TABLE public.crops ALTER COLUMN updated_by TYPE text USING updated_by::text;
ALTER TABLE public.crops ALTER COLUMN deleted_by TYPE text USING deleted_by::text;

ALTER TABLE public.investors ALTER COLUMN created_by TYPE text USING created_by::text;
ALTER TABLE public.investors ALTER COLUMN updated_by TYPE text USING updated_by::text;
ALTER TABLE public.investors ALTER COLUMN deleted_by TYPE text USING deleted_by::text;

ALTER TABLE public.project_investors ALTER COLUMN created_by TYPE text USING created_by::text;
ALTER TABLE public.project_investors ALTER COLUMN updated_by TYPE text USING updated_by::text;
ALTER TABLE public.project_investors ALTER COLUMN deleted_by TYPE text USING deleted_by::text;

ALTER TABLE public.supplies ALTER COLUMN created_by TYPE text USING created_by::text;
ALTER TABLE public.supplies ALTER COLUMN updated_by TYPE text USING updated_by::text;
ALTER TABLE public.supplies ALTER COLUMN deleted_by TYPE text USING deleted_by::text;

ALTER TABLE public.supply_movements ALTER COLUMN created_by TYPE text USING created_by::text;
ALTER TABLE public.supply_movements ALTER COLUMN updated_by TYPE text USING updated_by::text;
ALTER TABLE public.supply_movements ALTER COLUMN deleted_by TYPE text USING deleted_by::text;

ALTER TABLE public.stocks ALTER COLUMN created_by TYPE text USING created_by::text;
ALTER TABLE public.stocks ALTER COLUMN updated_by TYPE text USING updated_by::text;
ALTER TABLE public.stocks ALTER COLUMN deleted_by TYPE text USING deleted_by::text;

ALTER TABLE public.workorders ALTER COLUMN created_by TYPE text USING created_by::text;
ALTER TABLE public.workorders ALTER COLUMN updated_by TYPE text USING updated_by::text;
ALTER TABLE public.workorders ALTER COLUMN deleted_by TYPE text USING deleted_by::text;

ALTER TABLE public.labors ALTER COLUMN created_by TYPE text USING created_by::text;
ALTER TABLE public.labors ALTER COLUMN updated_by TYPE text USING updated_by::text;
ALTER TABLE public.labors ALTER COLUMN deleted_by TYPE text USING deleted_by::text;

ALTER TABLE public.categories ALTER COLUMN created_by TYPE text USING created_by::text;
ALTER TABLE public.categories ALTER COLUMN updated_by TYPE text USING updated_by::text;
ALTER TABLE public.categories ALTER COLUMN deleted_by TYPE text USING deleted_by::text;

ALTER TABLE public.types ALTER COLUMN created_by TYPE text USING created_by::text;
ALTER TABLE public.types ALTER COLUMN updated_by TYPE text USING updated_by::text;
ALTER TABLE public.types ALTER COLUMN deleted_by TYPE text USING deleted_by::text;

ALTER TABLE public.business_parameters ALTER COLUMN created_by TYPE text USING created_by::text;
ALTER TABLE public.business_parameters ALTER COLUMN updated_by TYPE text USING updated_by::text;
ALTER TABLE public.business_parameters ALTER COLUMN deleted_by TYPE text USING deleted_by::text;

ALTER TABLE public.invoices ALTER COLUMN created_by TYPE text USING created_by::text;
ALTER TABLE public.invoices ALTER COLUMN updated_by TYPE text USING updated_by::text;
ALTER TABLE public.invoices ALTER COLUMN deleted_by TYPE text USING deleted_by::text;

ALTER TABLE public.work_order_drafts ALTER COLUMN reviewed_by TYPE text USING reviewed_by::text;

ALTER TABLE public.crop_commercializations ALTER COLUMN created_by TYPE text USING created_by::text;
ALTER TABLE public.crop_commercializations ALTER COLUMN updated_by TYPE text USING updated_by::text;
ALTER TABLE public.crop_commercializations ALTER COLUMN deleted_by TYPE text USING deleted_by::text;

ALTER TABLE public.admin_cost_investors ALTER COLUMN created_by TYPE text USING created_by::text;
ALTER TABLE public.admin_cost_investors ALTER COLUMN updated_by TYPE text USING updated_by::text;
ALTER TABLE public.admin_cost_investors ALTER COLUMN deleted_by TYPE text USING deleted_by::text;

ALTER TABLE public.project_dollar_values ALTER COLUMN created_by TYPE text USING created_by::text;
ALTER TABLE public.project_dollar_values ALTER COLUMN updated_by TYPE text USING updated_by::text;
ALTER TABLE public.project_dollar_values ALTER COLUMN deleted_by TYPE text USING deleted_by::text;

ALTER TABLE public.field_investors ALTER COLUMN created_by TYPE text USING created_by::text;
ALTER TABLE public.field_investors ALTER COLUMN updated_by TYPE text USING updated_by::text;
ALTER TABLE public.field_investors ALTER COLUMN deleted_by TYPE text USING deleted_by::text;

ALTER TABLE public.labor_categories ALTER COLUMN created_by TYPE text USING created_by::text;
ALTER TABLE public.labor_categories ALTER COLUMN updated_by TYPE text USING updated_by::text;
ALTER TABLE public.labor_categories ALTER COLUMN deleted_by TYPE text USING deleted_by::text;

ALTER TABLE public.labor_types ALTER COLUMN created_by TYPE text USING created_by::text;
ALTER TABLE public.labor_types ALTER COLUMN updated_by TYPE text USING updated_by::text;
ALTER TABLE public.labor_types ALTER COLUMN deleted_by TYPE text USING deleted_by::text;

ALTER TABLE public.providers ALTER COLUMN created_by TYPE text USING created_by::text;
ALTER TABLE public.providers ALTER COLUMN updated_by TYPE text USING updated_by::text;
ALTER TABLE public.providers ALTER COLUMN deleted_by TYPE text USING deleted_by::text;

-- =============================================================================
-- 3. auth_tenants bigint -> uuid, preserving legacy_id
-- =============================================================================
ALTER TABLE public.auth_memberships DROP CONSTRAINT IF EXISTS auth_memberships_tenant_id_fkey;

ALTER TABLE public.auth_tenants ADD COLUMN new_id uuid NOT NULL DEFAULT gen_random_uuid();
ALTER TABLE public.auth_tenants DROP CONSTRAINT IF EXISTS auth_tenants_pkey;
ALTER TABLE public.auth_tenants RENAME COLUMN id TO legacy_id;
ALTER TABLE public.auth_tenants RENAME COLUMN new_id TO id;
ALTER TABLE public.auth_tenants ALTER COLUMN id SET DEFAULT gen_random_uuid();
ALTER TABLE public.auth_tenants ALTER COLUMN legacy_id SET DEFAULT nextval('public.auth_tenants_id_seq'::regclass);
ALTER TABLE public.auth_tenants ADD CONSTRAINT auth_tenants_pkey PRIMARY KEY (id);
ALTER TABLE public.auth_tenants ADD CONSTRAINT uq_auth_tenants_legacy_id UNIQUE (legacy_id);
ALTER SEQUENCE IF EXISTS public.auth_tenants_id_seq OWNED BY public.auth_tenants.legacy_id;

-- =============================================================================
-- 4. auth_roles bigint -> uuid, preserving legacy_id
-- =============================================================================
ALTER TABLE public.auth_role_permissions DROP CONSTRAINT IF EXISTS auth_role_permissions_role_id_fkey;
ALTER TABLE public.auth_memberships DROP CONSTRAINT IF EXISTS auth_memberships_role_id_fkey;

ALTER TABLE public.auth_roles ADD COLUMN new_id uuid NOT NULL DEFAULT gen_random_uuid();
ALTER TABLE public.auth_roles DROP CONSTRAINT IF EXISTS auth_roles_pkey;
ALTER TABLE public.auth_roles RENAME COLUMN id TO legacy_id;
ALTER TABLE public.auth_roles RENAME COLUMN new_id TO id;
ALTER TABLE public.auth_roles ALTER COLUMN id SET DEFAULT gen_random_uuid();
ALTER TABLE public.auth_roles ALTER COLUMN legacy_id SET DEFAULT nextval('public.auth_roles_id_seq'::regclass);
ALTER TABLE public.auth_roles ADD CONSTRAINT auth_roles_pkey PRIMARY KEY (id);
ALTER TABLE public.auth_roles ADD CONSTRAINT uq_auth_roles_legacy_id UNIQUE (legacy_id);
ALTER SEQUENCE IF EXISTS public.auth_roles_id_seq OWNED BY public.auth_roles.legacy_id;

-- =============================================================================
-- 5. auth_permissions bigint -> uuid, preserving legacy_id
-- =============================================================================
ALTER TABLE public.auth_role_permissions DROP CONSTRAINT IF EXISTS auth_role_permissions_permission_id_fkey;

ALTER TABLE public.auth_permissions ADD COLUMN new_id uuid NOT NULL DEFAULT gen_random_uuid();
ALTER TABLE public.auth_permissions DROP CONSTRAINT IF EXISTS auth_permissions_pkey;
ALTER TABLE public.auth_permissions RENAME COLUMN id TO legacy_id;
ALTER TABLE public.auth_permissions RENAME COLUMN new_id TO id;
ALTER TABLE public.auth_permissions ALTER COLUMN id SET DEFAULT gen_random_uuid();
ALTER TABLE public.auth_permissions ALTER COLUMN legacy_id SET DEFAULT nextval('public.auth_permissions_id_seq'::regclass);
ALTER TABLE public.auth_permissions ADD CONSTRAINT auth_permissions_pkey PRIMARY KEY (id);
ALTER TABLE public.auth_permissions ADD CONSTRAINT uq_auth_permissions_legacy_id UNIQUE (legacy_id);
ALTER SEQUENCE IF EXISTS public.auth_permissions_id_seq OWNED BY public.auth_permissions.legacy_id;

-- =============================================================================
-- 6. auth_role_permissions bigint FKs -> uuid FKs
-- =============================================================================
ALTER TABLE public.auth_role_permissions DROP CONSTRAINT IF EXISTS auth_role_permissions_pkey;
ALTER TABLE public.auth_role_permissions ADD COLUMN new_role_id uuid;
ALTER TABLE public.auth_role_permissions ADD COLUMN new_permission_id uuid;

UPDATE public.auth_role_permissions rp
SET new_role_id = ar.id
FROM public.auth_roles ar
WHERE ar.legacy_id = rp.role_id;

UPDATE public.auth_role_permissions rp
SET new_permission_id = ap.id
FROM public.auth_permissions ap
WHERE ap.legacy_id = rp.permission_id;

ALTER TABLE public.auth_role_permissions DROP COLUMN role_id;
ALTER TABLE public.auth_role_permissions DROP COLUMN permission_id;
ALTER TABLE public.auth_role_permissions RENAME COLUMN new_role_id TO role_id;
ALTER TABLE public.auth_role_permissions RENAME COLUMN new_permission_id TO permission_id;
ALTER TABLE public.auth_role_permissions ALTER COLUMN role_id SET NOT NULL;
ALTER TABLE public.auth_role_permissions ALTER COLUMN permission_id SET NOT NULL;
ALTER TABLE public.auth_role_permissions ADD CONSTRAINT auth_role_permissions_pkey PRIMARY KEY (role_id, permission_id);
ALTER TABLE public.auth_role_permissions
    ADD CONSTRAINT auth_role_permissions_role_id_fkey FOREIGN KEY (role_id) REFERENCES public.auth_roles(id) ON DELETE CASCADE;
ALTER TABLE public.auth_role_permissions
    ADD CONSTRAINT auth_role_permissions_permission_id_fkey FOREIGN KEY (permission_id) REFERENCES public.auth_permissions(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_auth_role_permissions_permission_id
    ON public.auth_role_permissions (permission_id);

-- =============================================================================
-- 7. users bigint -> uuid, preserving legacy_id
-- =============================================================================
ALTER TABLE public.auth_memberships DROP CONSTRAINT IF EXISTS auth_memberships_user_id_fkey;

ALTER TABLE public.users ADD COLUMN new_id uuid NOT NULL DEFAULT gen_random_uuid();
ALTER TABLE public.users DROP CONSTRAINT IF EXISTS pk_users;
ALTER TABLE public.users RENAME COLUMN id TO legacy_id;
ALTER TABLE public.users RENAME COLUMN new_id TO id;
ALTER TABLE public.users ALTER COLUMN id SET DEFAULT gen_random_uuid();
ALTER TABLE public.users ALTER COLUMN legacy_id SET DEFAULT nextval('public.users_id_seq'::regclass);
ALTER TABLE public.users ADD CONSTRAINT pk_users PRIMARY KEY (id);
ALTER TABLE public.users ADD CONSTRAINT uq_users_legacy_id UNIQUE (legacy_id);
ALTER SEQUENCE IF EXISTS public.users_id_seq OWNED BY public.users.legacy_id;

-- =============================================================================
-- 8. auth_memberships bigint PK/FKs -> uuid PK/FKs
-- =============================================================================
ALTER TABLE public.auth_memberships DROP CONSTRAINT IF EXISTS auth_memberships_pkey;
ALTER TABLE public.auth_memberships DROP CONSTRAINT IF EXISTS auth_memberships_user_id_tenant_id_key;
ALTER TABLE public.auth_memberships ADD COLUMN new_id uuid NOT NULL DEFAULT gen_random_uuid();
ALTER TABLE public.auth_memberships ADD COLUMN new_user_id uuid;
ALTER TABLE public.auth_memberships ADD COLUMN new_tenant_id uuid;
ALTER TABLE public.auth_memberships ADD COLUMN new_role_id uuid;

UPDATE public.auth_memberships am
SET new_user_id = u.id
FROM public.users u
WHERE u.legacy_id = am.user_id;

UPDATE public.auth_memberships am
SET new_tenant_id = t.id
FROM public.auth_tenants t
WHERE t.legacy_id = am.tenant_id;

UPDATE public.auth_memberships am
SET new_role_id = r.id
FROM public.auth_roles r
WHERE r.legacy_id = am.role_id;

ALTER TABLE public.auth_memberships RENAME COLUMN id TO legacy_id;
ALTER TABLE public.auth_memberships DROP COLUMN user_id;
ALTER TABLE public.auth_memberships DROP COLUMN tenant_id;
ALTER TABLE public.auth_memberships DROP COLUMN role_id;
ALTER TABLE public.auth_memberships RENAME COLUMN new_id TO id;
ALTER TABLE public.auth_memberships RENAME COLUMN new_user_id TO user_id;
ALTER TABLE public.auth_memberships RENAME COLUMN new_tenant_id TO tenant_id;
ALTER TABLE public.auth_memberships RENAME COLUMN new_role_id TO role_id;
ALTER TABLE public.auth_memberships ALTER COLUMN id SET DEFAULT gen_random_uuid();
ALTER TABLE public.auth_memberships ALTER COLUMN legacy_id SET DEFAULT nextval('public.auth_memberships_id_seq'::regclass);
ALTER TABLE public.auth_memberships ALTER COLUMN user_id SET NOT NULL;
ALTER TABLE public.auth_memberships ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE public.auth_memberships ALTER COLUMN role_id SET NOT NULL;
ALTER TABLE public.auth_memberships ADD CONSTRAINT auth_memberships_pkey PRIMARY KEY (id);
ALTER TABLE public.auth_memberships ADD CONSTRAINT uq_auth_memberships_legacy_id UNIQUE (legacy_id);
ALTER TABLE public.auth_memberships ADD CONSTRAINT auth_memberships_user_id_tenant_id_key UNIQUE (user_id, tenant_id);
ALTER TABLE public.auth_memberships
    ADD CONSTRAINT auth_memberships_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;
ALTER TABLE public.auth_memberships
    ADD CONSTRAINT auth_memberships_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES public.auth_tenants(id) ON DELETE CASCADE;
ALTER TABLE public.auth_memberships
    ADD CONSTRAINT auth_memberships_role_id_fkey FOREIGN KEY (role_id) REFERENCES public.auth_roles(id);
CREATE INDEX IF NOT EXISTS idx_auth_memberships_tenant_id
    ON public.auth_memberships (tenant_id);
CREATE INDEX IF NOT EXISTS idx_auth_memberships_role_id
    ON public.auth_memberships (role_id);
ALTER SEQUENCE IF EXISTS public.auth_memberships_id_seq OWNED BY public.auth_memberships.legacy_id;
