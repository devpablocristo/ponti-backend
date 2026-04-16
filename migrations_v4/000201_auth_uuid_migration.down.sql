-- =============================================================================
-- Migration 000199 rollback: UUID auth tables -> bigint auth tables
-- Reverts actor columns text -> original bigint/varchar types.
-- =============================================================================
-- NOTE: no explicit BEGIN/COMMIT. golang-migrate wraps the migration in a transaction.

CREATE OR REPLACE FUNCTION public._migration_000199_actor_text_to_legacy_id(actor text)
RETURNS bigint
LANGUAGE plpgsql
AS $$
DECLARE
    normalized text;
    mapped_id bigint;
BEGIN
    normalized := NULLIF(btrim(actor), '');
    IF normalized IS NULL THEN
        RETURN NULL;
    END IF;

    BEGIN
        RETURN normalized::bigint;
    EXCEPTION WHEN invalid_text_representation THEN
        SELECT u.legacy_id
        INTO mapped_id
        FROM public.users u
        WHERE u.id::text = normalized;

        IF mapped_id IS NOT NULL THEN
            RETURN mapped_id;
        END IF;

        RAISE EXCEPTION
            '000199 down cannot map actor "%" back to legacy bigint user id',
            actor;
    END;
END;
$$;

-- =============================================================================
-- 1. Convert actor columns back to original types
-- =============================================================================
ALTER TABLE public.users ALTER COLUMN created_by TYPE bigint USING public._migration_000199_actor_text_to_legacy_id(created_by);
ALTER TABLE public.users ALTER COLUMN updated_by TYPE bigint USING public._migration_000199_actor_text_to_legacy_id(updated_by);
ALTER TABLE public.users ALTER COLUMN deleted_by TYPE bigint USING public._migration_000199_actor_text_to_legacy_id(deleted_by);

ALTER TABLE public.customers ALTER COLUMN created_by TYPE bigint USING public._migration_000199_actor_text_to_legacy_id(created_by);
ALTER TABLE public.customers ALTER COLUMN updated_by TYPE bigint USING public._migration_000199_actor_text_to_legacy_id(updated_by);
ALTER TABLE public.customers ALTER COLUMN deleted_by TYPE bigint USING public._migration_000199_actor_text_to_legacy_id(deleted_by);

ALTER TABLE public.campaigns ALTER COLUMN created_by TYPE bigint USING public._migration_000199_actor_text_to_legacy_id(created_by);
ALTER TABLE public.campaigns ALTER COLUMN updated_by TYPE bigint USING public._migration_000199_actor_text_to_legacy_id(updated_by);
ALTER TABLE public.campaigns ALTER COLUMN deleted_by TYPE bigint USING public._migration_000199_actor_text_to_legacy_id(deleted_by);

ALTER TABLE public.projects ALTER COLUMN created_by TYPE bigint USING public._migration_000199_actor_text_to_legacy_id(created_by);
ALTER TABLE public.projects ALTER COLUMN updated_by TYPE bigint USING public._migration_000199_actor_text_to_legacy_id(updated_by);
ALTER TABLE public.projects ALTER COLUMN deleted_by TYPE bigint USING public._migration_000199_actor_text_to_legacy_id(deleted_by);

ALTER TABLE public.managers ALTER COLUMN created_by TYPE bigint USING public._migration_000199_actor_text_to_legacy_id(created_by);
ALTER TABLE public.managers ALTER COLUMN updated_by TYPE bigint USING public._migration_000199_actor_text_to_legacy_id(updated_by);
ALTER TABLE public.managers ALTER COLUMN deleted_by TYPE bigint USING public._migration_000199_actor_text_to_legacy_id(deleted_by);

ALTER TABLE public.project_managers ALTER COLUMN created_by TYPE bigint USING public._migration_000199_actor_text_to_legacy_id(created_by);
ALTER TABLE public.project_managers ALTER COLUMN updated_by TYPE bigint USING public._migration_000199_actor_text_to_legacy_id(updated_by);
ALTER TABLE public.project_managers ALTER COLUMN deleted_by TYPE bigint USING public._migration_000199_actor_text_to_legacy_id(deleted_by);

ALTER TABLE public.lease_types ALTER COLUMN created_by TYPE bigint USING public._migration_000199_actor_text_to_legacy_id(created_by);
ALTER TABLE public.lease_types ALTER COLUMN updated_by TYPE bigint USING public._migration_000199_actor_text_to_legacy_id(updated_by);
ALTER TABLE public.lease_types ALTER COLUMN deleted_by TYPE bigint USING public._migration_000199_actor_text_to_legacy_id(deleted_by);

ALTER TABLE public.fields ALTER COLUMN created_by TYPE bigint USING public._migration_000199_actor_text_to_legacy_id(created_by);
ALTER TABLE public.fields ALTER COLUMN updated_by TYPE bigint USING public._migration_000199_actor_text_to_legacy_id(updated_by);
ALTER TABLE public.fields ALTER COLUMN deleted_by TYPE bigint USING public._migration_000199_actor_text_to_legacy_id(deleted_by);

ALTER TABLE public.lots ALTER COLUMN created_by TYPE bigint USING public._migration_000199_actor_text_to_legacy_id(created_by);
ALTER TABLE public.lots ALTER COLUMN updated_by TYPE bigint USING public._migration_000199_actor_text_to_legacy_id(updated_by);
ALTER TABLE public.lots ALTER COLUMN deleted_by TYPE bigint USING public._migration_000199_actor_text_to_legacy_id(deleted_by);

ALTER TABLE public.lot_dates ALTER COLUMN created_by TYPE bigint USING public._migration_000199_actor_text_to_legacy_id(created_by);
ALTER TABLE public.lot_dates ALTER COLUMN updated_by TYPE bigint USING public._migration_000199_actor_text_to_legacy_id(updated_by);
ALTER TABLE public.lot_dates ALTER COLUMN deleted_by TYPE bigint USING public._migration_000199_actor_text_to_legacy_id(deleted_by);

ALTER TABLE public.crops ALTER COLUMN created_by TYPE bigint USING public._migration_000199_actor_text_to_legacy_id(created_by);
ALTER TABLE public.crops ALTER COLUMN updated_by TYPE bigint USING public._migration_000199_actor_text_to_legacy_id(updated_by);
ALTER TABLE public.crops ALTER COLUMN deleted_by TYPE bigint USING public._migration_000199_actor_text_to_legacy_id(deleted_by);

ALTER TABLE public.investors ALTER COLUMN created_by TYPE bigint USING public._migration_000199_actor_text_to_legacy_id(created_by);
ALTER TABLE public.investors ALTER COLUMN updated_by TYPE bigint USING public._migration_000199_actor_text_to_legacy_id(updated_by);
ALTER TABLE public.investors ALTER COLUMN deleted_by TYPE bigint USING public._migration_000199_actor_text_to_legacy_id(deleted_by);

ALTER TABLE public.project_investors ALTER COLUMN created_by TYPE bigint USING public._migration_000199_actor_text_to_legacy_id(created_by);
ALTER TABLE public.project_investors ALTER COLUMN updated_by TYPE bigint USING public._migration_000199_actor_text_to_legacy_id(updated_by);
ALTER TABLE public.project_investors ALTER COLUMN deleted_by TYPE bigint USING public._migration_000199_actor_text_to_legacy_id(deleted_by);

ALTER TABLE public.supplies ALTER COLUMN created_by TYPE bigint USING public._migration_000199_actor_text_to_legacy_id(created_by);
ALTER TABLE public.supplies ALTER COLUMN updated_by TYPE bigint USING public._migration_000199_actor_text_to_legacy_id(updated_by);
ALTER TABLE public.supplies ALTER COLUMN deleted_by TYPE bigint USING public._migration_000199_actor_text_to_legacy_id(deleted_by);

ALTER TABLE public.supply_movements ALTER COLUMN created_by TYPE bigint USING public._migration_000199_actor_text_to_legacy_id(created_by);
ALTER TABLE public.supply_movements ALTER COLUMN updated_by TYPE bigint USING public._migration_000199_actor_text_to_legacy_id(updated_by);
ALTER TABLE public.supply_movements ALTER COLUMN deleted_by TYPE bigint USING public._migration_000199_actor_text_to_legacy_id(deleted_by);

ALTER TABLE public.stocks ALTER COLUMN created_by TYPE bigint USING public._migration_000199_actor_text_to_legacy_id(created_by);
ALTER TABLE public.stocks ALTER COLUMN updated_by TYPE bigint USING public._migration_000199_actor_text_to_legacy_id(updated_by);
ALTER TABLE public.stocks ALTER COLUMN deleted_by TYPE bigint USING public._migration_000199_actor_text_to_legacy_id(deleted_by);

ALTER TABLE public.workorders ALTER COLUMN created_by TYPE bigint USING public._migration_000199_actor_text_to_legacy_id(created_by);
ALTER TABLE public.workorders ALTER COLUMN updated_by TYPE bigint USING public._migration_000199_actor_text_to_legacy_id(updated_by);
ALTER TABLE public.workorders ALTER COLUMN deleted_by TYPE bigint USING public._migration_000199_actor_text_to_legacy_id(deleted_by);

ALTER TABLE public.labors ALTER COLUMN created_by TYPE bigint USING public._migration_000199_actor_text_to_legacy_id(created_by);
ALTER TABLE public.labors ALTER COLUMN updated_by TYPE bigint USING public._migration_000199_actor_text_to_legacy_id(updated_by);
ALTER TABLE public.labors ALTER COLUMN deleted_by TYPE bigint USING public._migration_000199_actor_text_to_legacy_id(deleted_by);

ALTER TABLE public.categories ALTER COLUMN created_by TYPE bigint USING public._migration_000199_actor_text_to_legacy_id(created_by);
ALTER TABLE public.categories ALTER COLUMN updated_by TYPE bigint USING public._migration_000199_actor_text_to_legacy_id(updated_by);
ALTER TABLE public.categories ALTER COLUMN deleted_by TYPE bigint USING public._migration_000199_actor_text_to_legacy_id(deleted_by);

ALTER TABLE public.types ALTER COLUMN created_by TYPE bigint USING public._migration_000199_actor_text_to_legacy_id(created_by);
ALTER TABLE public.types ALTER COLUMN updated_by TYPE bigint USING public._migration_000199_actor_text_to_legacy_id(updated_by);
ALTER TABLE public.types ALTER COLUMN deleted_by TYPE bigint USING public._migration_000199_actor_text_to_legacy_id(deleted_by);

ALTER TABLE public.business_parameters ALTER COLUMN created_by TYPE bigint USING public._migration_000199_actor_text_to_legacy_id(created_by);
ALTER TABLE public.business_parameters ALTER COLUMN updated_by TYPE bigint USING public._migration_000199_actor_text_to_legacy_id(updated_by);
ALTER TABLE public.business_parameters ALTER COLUMN deleted_by TYPE bigint USING public._migration_000199_actor_text_to_legacy_id(deleted_by);

ALTER TABLE public.invoices ALTER COLUMN created_by TYPE bigint USING public._migration_000199_actor_text_to_legacy_id(created_by);
ALTER TABLE public.invoices ALTER COLUMN updated_by TYPE bigint USING public._migration_000199_actor_text_to_legacy_id(updated_by);
ALTER TABLE public.invoices ALTER COLUMN deleted_by TYPE bigint USING public._migration_000199_actor_text_to_legacy_id(deleted_by);

ALTER TABLE public.crop_commercializations ALTER COLUMN created_by TYPE bigint USING public._migration_000199_actor_text_to_legacy_id(created_by);
ALTER TABLE public.crop_commercializations ALTER COLUMN updated_by TYPE bigint USING public._migration_000199_actor_text_to_legacy_id(updated_by);
ALTER TABLE public.crop_commercializations ALTER COLUMN deleted_by TYPE bigint USING public._migration_000199_actor_text_to_legacy_id(deleted_by);

ALTER TABLE public.admin_cost_investors ALTER COLUMN created_by TYPE bigint USING public._migration_000199_actor_text_to_legacy_id(created_by);
ALTER TABLE public.admin_cost_investors ALTER COLUMN updated_by TYPE bigint USING public._migration_000199_actor_text_to_legacy_id(updated_by);
ALTER TABLE public.admin_cost_investors ALTER COLUMN deleted_by TYPE bigint USING public._migration_000199_actor_text_to_legacy_id(deleted_by);

ALTER TABLE public.project_dollar_values ALTER COLUMN created_by TYPE bigint USING public._migration_000199_actor_text_to_legacy_id(created_by);
ALTER TABLE public.project_dollar_values ALTER COLUMN updated_by TYPE bigint USING public._migration_000199_actor_text_to_legacy_id(updated_by);
ALTER TABLE public.project_dollar_values ALTER COLUMN deleted_by TYPE bigint USING public._migration_000199_actor_text_to_legacy_id(deleted_by);

ALTER TABLE public.field_investors ALTER COLUMN created_by TYPE bigint USING public._migration_000199_actor_text_to_legacy_id(created_by);
ALTER TABLE public.field_investors ALTER COLUMN updated_by TYPE bigint USING public._migration_000199_actor_text_to_legacy_id(updated_by);
ALTER TABLE public.field_investors ALTER COLUMN deleted_by TYPE bigint USING public._migration_000199_actor_text_to_legacy_id(deleted_by);

ALTER TABLE public.labor_categories ALTER COLUMN created_by TYPE character varying(255) USING created_by::character varying(255);
ALTER TABLE public.labor_categories ALTER COLUMN updated_by TYPE character varying(255) USING updated_by::character varying(255);
ALTER TABLE public.labor_categories ALTER COLUMN deleted_by TYPE character varying(255) USING deleted_by::character varying(255);

ALTER TABLE public.labor_types ALTER COLUMN created_by TYPE character varying(255) USING created_by::character varying(255);
ALTER TABLE public.labor_types ALTER COLUMN updated_by TYPE character varying(255) USING updated_by::character varying(255);
ALTER TABLE public.labor_types ALTER COLUMN deleted_by TYPE character varying(255) USING deleted_by::character varying(255);

ALTER TABLE public.providers ALTER COLUMN created_by TYPE character varying(255) USING created_by::character varying(255);
ALTER TABLE public.providers ALTER COLUMN updated_by TYPE character varying(255) USING updated_by::character varying(255);
ALTER TABLE public.providers ALTER COLUMN deleted_by TYPE character varying(255) USING deleted_by::character varying(255);

-- =============================================================================
-- 2. Rebuild auth tables back to bigint using preserved legacy_id columns
-- =============================================================================
ALTER TABLE public.auth_memberships DROP CONSTRAINT IF EXISTS auth_memberships_user_id_fkey;
ALTER TABLE public.auth_memberships DROP CONSTRAINT IF EXISTS auth_memberships_tenant_id_fkey;
ALTER TABLE public.auth_memberships DROP CONSTRAINT IF EXISTS auth_memberships_role_id_fkey;
ALTER TABLE public.auth_memberships DROP CONSTRAINT IF EXISTS auth_memberships_pkey;
ALTER TABLE public.auth_memberships DROP CONSTRAINT IF EXISTS auth_memberships_user_id_tenant_id_key;
ALTER TABLE public.auth_memberships DROP CONSTRAINT IF EXISTS uq_auth_memberships_legacy_id;

ALTER TABLE public.auth_memberships ADD COLUMN old_user_id bigint;
ALTER TABLE public.auth_memberships ADD COLUMN old_tenant_id bigint;
ALTER TABLE public.auth_memberships ADD COLUMN old_role_id bigint;

UPDATE public.auth_memberships am
SET old_user_id = u.legacy_id
FROM public.users u
WHERE u.id = am.user_id;

UPDATE public.auth_memberships am
SET old_tenant_id = t.legacy_id
FROM public.auth_tenants t
WHERE t.id = am.tenant_id;

UPDATE public.auth_memberships am
SET old_role_id = r.legacy_id
FROM public.auth_roles r
WHERE r.id = am.role_id;

ALTER TABLE public.auth_role_permissions DROP CONSTRAINT IF EXISTS auth_role_permissions_role_id_fkey;
ALTER TABLE public.auth_role_permissions DROP CONSTRAINT IF EXISTS auth_role_permissions_permission_id_fkey;
ALTER TABLE public.auth_role_permissions DROP CONSTRAINT IF EXISTS auth_role_permissions_pkey;

ALTER TABLE public.auth_role_permissions ADD COLUMN old_role_id bigint;
ALTER TABLE public.auth_role_permissions ADD COLUMN old_permission_id bigint;

UPDATE public.auth_role_permissions arp
SET old_role_id = ar.legacy_id
FROM public.auth_roles ar
WHERE ar.id = arp.role_id;

UPDATE public.auth_role_permissions arp
SET old_permission_id = ap.legacy_id
FROM public.auth_permissions ap
WHERE ap.id = arp.permission_id;

CREATE TABLE public.auth_memberships_rollback (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    tenant_id bigint NOT NULL,
    role_id bigint NOT NULL,
    status text NOT NULL DEFAULT 'active',
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO public.auth_memberships_rollback (
    id,
    user_id,
    tenant_id,
    role_id,
    status,
    created_at,
    updated_at
)
SELECT
    legacy_id,
    old_user_id,
    old_tenant_id,
    old_role_id,
    status,
    created_at,
    updated_at
FROM public.auth_memberships;

ALTER SEQUENCE IF EXISTS public.auth_memberships_id_seq OWNED BY NONE;
DROP TABLE public.auth_memberships;
ALTER TABLE public.auth_memberships_rollback RENAME TO auth_memberships;
ALTER TABLE public.auth_memberships ALTER COLUMN id SET DEFAULT nextval('public.auth_memberships_id_seq'::regclass);
ALTER SEQUENCE IF EXISTS public.auth_memberships_id_seq OWNED BY public.auth_memberships.id;

CREATE TABLE public.auth_role_permissions_rollback (
    role_id bigint NOT NULL,
    permission_id bigint NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO public.auth_role_permissions_rollback (
    role_id,
    permission_id,
    created_at
)
SELECT
    old_role_id,
    old_permission_id,
    created_at
FROM public.auth_role_permissions;

DROP TABLE public.auth_role_permissions;
ALTER TABLE public.auth_role_permissions_rollback RENAME TO auth_role_permissions;

ALTER TABLE public.auth_tenants DROP CONSTRAINT IF EXISTS auth_tenants_pkey;
ALTER TABLE public.auth_tenants DROP CONSTRAINT IF EXISTS uq_auth_tenants_legacy_id;
ALTER TABLE public.auth_tenants DROP COLUMN id;
ALTER TABLE public.auth_tenants RENAME COLUMN legacy_id TO id;
ALTER TABLE public.auth_tenants ALTER COLUMN id SET DEFAULT nextval('public.auth_tenants_id_seq'::regclass);
ALTER TABLE public.auth_tenants ADD CONSTRAINT auth_tenants_pkey PRIMARY KEY (id);
ALTER SEQUENCE IF EXISTS public.auth_tenants_id_seq OWNED BY public.auth_tenants.id;

ALTER TABLE public.auth_roles DROP CONSTRAINT IF EXISTS auth_roles_pkey;
ALTER TABLE public.auth_roles DROP CONSTRAINT IF EXISTS uq_auth_roles_legacy_id;
ALTER TABLE public.auth_roles DROP COLUMN id;
ALTER TABLE public.auth_roles RENAME COLUMN legacy_id TO id;
ALTER TABLE public.auth_roles ALTER COLUMN id SET DEFAULT nextval('public.auth_roles_id_seq'::regclass);
ALTER TABLE public.auth_roles ADD CONSTRAINT auth_roles_pkey PRIMARY KEY (id);
ALTER SEQUENCE IF EXISTS public.auth_roles_id_seq OWNED BY public.auth_roles.id;

ALTER TABLE public.auth_permissions DROP CONSTRAINT IF EXISTS auth_permissions_pkey;
ALTER TABLE public.auth_permissions DROP CONSTRAINT IF EXISTS uq_auth_permissions_legacy_id;
ALTER TABLE public.auth_permissions DROP COLUMN id;
ALTER TABLE public.auth_permissions RENAME COLUMN legacy_id TO id;
ALTER TABLE public.auth_permissions ALTER COLUMN id SET DEFAULT nextval('public.auth_permissions_id_seq'::regclass);
ALTER TABLE public.auth_permissions ADD CONSTRAINT auth_permissions_pkey PRIMARY KEY (id);
ALTER SEQUENCE IF EXISTS public.auth_permissions_id_seq OWNED BY public.auth_permissions.id;

ALTER TABLE public.users DROP CONSTRAINT IF EXISTS pk_users;
ALTER TABLE public.users DROP CONSTRAINT IF EXISTS uq_users_legacy_id;
ALTER TABLE public.users DROP COLUMN id;
ALTER TABLE public.users RENAME COLUMN legacy_id TO id;
ALTER TABLE public.users ALTER COLUMN id SET DEFAULT nextval('public.users_id_seq'::regclass);
ALTER TABLE public.users ADD CONSTRAINT pk_users PRIMARY KEY (id);
ALTER SEQUENCE IF EXISTS public.users_id_seq OWNED BY public.users.id;

ALTER TABLE public.auth_memberships ADD CONSTRAINT auth_memberships_pkey PRIMARY KEY (id);
ALTER TABLE public.auth_memberships ADD CONSTRAINT auth_memberships_user_id_tenant_id_key UNIQUE (user_id, tenant_id);
CREATE INDEX IF NOT EXISTS idx_auth_memberships_tenant_id
    ON public.auth_memberships (tenant_id);
CREATE INDEX IF NOT EXISTS idx_auth_memberships_role_id
    ON public.auth_memberships (role_id);
ALTER TABLE public.auth_memberships
    ADD CONSTRAINT auth_memberships_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;
ALTER TABLE public.auth_memberships
    ADD CONSTRAINT auth_memberships_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES public.auth_tenants(id) ON DELETE CASCADE;
ALTER TABLE public.auth_memberships
    ADD CONSTRAINT auth_memberships_role_id_fkey FOREIGN KEY (role_id) REFERENCES public.auth_roles(id);

ALTER TABLE public.auth_role_permissions ADD CONSTRAINT auth_role_permissions_pkey PRIMARY KEY (role_id, permission_id);
CREATE INDEX IF NOT EXISTS idx_auth_role_permissions_permission_id
    ON public.auth_role_permissions (permission_id);
ALTER TABLE public.auth_role_permissions
    ADD CONSTRAINT auth_role_permissions_role_id_fkey FOREIGN KEY (role_id) REFERENCES public.auth_roles(id) ON DELETE CASCADE;
ALTER TABLE public.auth_role_permissions
    ADD CONSTRAINT auth_role_permissions_permission_id_fkey FOREIGN KEY (permission_id) REFERENCES public.auth_permissions(id) ON DELETE CASCADE;

-- =============================================================================
-- 3. Restore audit FKs to public.users(id)
-- =============================================================================
ALTER TABLE ONLY public.customers
    ADD CONSTRAINT fk_customers_created_by FOREIGN KEY (created_by) REFERENCES public.users(id);
ALTER TABLE ONLY public.customers
    ADD CONSTRAINT fk_customers_updated_by FOREIGN KEY (updated_by) REFERENCES public.users(id);
ALTER TABLE ONLY public.customers
    ADD CONSTRAINT fk_customers_deleted_by FOREIGN KEY (deleted_by) REFERENCES public.users(id);

ALTER TABLE ONLY public.campaigns
    ADD CONSTRAINT fk_campaigns_created_by FOREIGN KEY (created_by) REFERENCES public.users(id);
ALTER TABLE ONLY public.campaigns
    ADD CONSTRAINT fk_campaigns_updated_by FOREIGN KEY (updated_by) REFERENCES public.users(id);
ALTER TABLE ONLY public.campaigns
    ADD CONSTRAINT fk_campaigns_deleted_by FOREIGN KEY (deleted_by) REFERENCES public.users(id);

ALTER TABLE ONLY public.projects
    ADD CONSTRAINT fk_projects_created_by FOREIGN KEY (created_by) REFERENCES public.users(id);
ALTER TABLE ONLY public.projects
    ADD CONSTRAINT fk_projects_updated_by FOREIGN KEY (updated_by) REFERENCES public.users(id);
ALTER TABLE ONLY public.projects
    ADD CONSTRAINT fk_projects_deleted_by FOREIGN KEY (deleted_by) REFERENCES public.users(id);

ALTER TABLE ONLY public.managers
    ADD CONSTRAINT fk_managers_created_by FOREIGN KEY (created_by) REFERENCES public.users(id);
ALTER TABLE ONLY public.managers
    ADD CONSTRAINT fk_managers_updated_by FOREIGN KEY (updated_by) REFERENCES public.users(id);
ALTER TABLE ONLY public.managers
    ADD CONSTRAINT fk_managers_deleted_by FOREIGN KEY (deleted_by) REFERENCES public.users(id);

ALTER TABLE ONLY public.project_managers
    ADD CONSTRAINT fk_project_managers_created_by FOREIGN KEY (created_by) REFERENCES public.users(id);
ALTER TABLE ONLY public.project_managers
    ADD CONSTRAINT fk_project_managers_updated_by FOREIGN KEY (updated_by) REFERENCES public.users(id);
ALTER TABLE ONLY public.project_managers
    ADD CONSTRAINT fk_project_managers_deleted_by FOREIGN KEY (deleted_by) REFERENCES public.users(id);

ALTER TABLE ONLY public.lease_types
    ADD CONSTRAINT fk_lease_types_created_by FOREIGN KEY (created_by) REFERENCES public.users(id);
ALTER TABLE ONLY public.lease_types
    ADD CONSTRAINT fk_lease_types_updated_by FOREIGN KEY (updated_by) REFERENCES public.users(id);
ALTER TABLE ONLY public.lease_types
    ADD CONSTRAINT fk_lease_types_deleted_by FOREIGN KEY (deleted_by) REFERENCES public.users(id);

ALTER TABLE ONLY public.fields
    ADD CONSTRAINT fk_fields_created_by FOREIGN KEY (created_by) REFERENCES public.users(id);
ALTER TABLE ONLY public.fields
    ADD CONSTRAINT fk_fields_updated_by FOREIGN KEY (updated_by) REFERENCES public.users(id);
ALTER TABLE ONLY public.fields
    ADD CONSTRAINT fk_fields_deleted_by FOREIGN KEY (deleted_by) REFERENCES public.users(id);

ALTER TABLE ONLY public.lots
    ADD CONSTRAINT fk_lots_created_by FOREIGN KEY (created_by) REFERENCES public.users(id);
ALTER TABLE ONLY public.lots
    ADD CONSTRAINT fk_lots_updated_by FOREIGN KEY (updated_by) REFERENCES public.users(id);
ALTER TABLE ONLY public.lots
    ADD CONSTRAINT fk_lots_deleted_by FOREIGN KEY (deleted_by) REFERENCES public.users(id);

ALTER TABLE ONLY public.lot_dates
    ADD CONSTRAINT fk_lot_dates_created_by FOREIGN KEY (created_by) REFERENCES public.users(id);
ALTER TABLE ONLY public.lot_dates
    ADD CONSTRAINT fk_lot_dates_updated_by FOREIGN KEY (updated_by) REFERENCES public.users(id);
ALTER TABLE ONLY public.lot_dates
    ADD CONSTRAINT fk_lot_dates_deleted_by FOREIGN KEY (deleted_by) REFERENCES public.users(id);

ALTER TABLE ONLY public.crops
    ADD CONSTRAINT fk_crops_created_by FOREIGN KEY (created_by) REFERENCES public.users(id);
ALTER TABLE ONLY public.crops
    ADD CONSTRAINT fk_crops_updated_by FOREIGN KEY (updated_by) REFERENCES public.users(id);
ALTER TABLE ONLY public.crops
    ADD CONSTRAINT fk_crops_deleted_by FOREIGN KEY (deleted_by) REFERENCES public.users(id);

ALTER TABLE ONLY public.investors
    ADD CONSTRAINT fk_investors_created_by FOREIGN KEY (created_by) REFERENCES public.users(id);
ALTER TABLE ONLY public.investors
    ADD CONSTRAINT fk_investors_updated_by FOREIGN KEY (updated_by) REFERENCES public.users(id);
ALTER TABLE ONLY public.investors
    ADD CONSTRAINT fk_investors_deleted_by FOREIGN KEY (deleted_by) REFERENCES public.users(id);

ALTER TABLE ONLY public.project_investors
    ADD CONSTRAINT fk_project_investors_created_by FOREIGN KEY (created_by) REFERENCES public.users(id);
ALTER TABLE ONLY public.project_investors
    ADD CONSTRAINT fk_project_investors_updated_by FOREIGN KEY (updated_by) REFERENCES public.users(id);
ALTER TABLE ONLY public.project_investors
    ADD CONSTRAINT fk_project_investors_deleted_by FOREIGN KEY (deleted_by) REFERENCES public.users(id);

DROP FUNCTION IF EXISTS public._migration_000199_actor_text_to_legacy_id(text);
DROP EXTENSION IF EXISTS pgcrypto;
