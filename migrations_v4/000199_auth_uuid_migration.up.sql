-- =============================================================================
-- Migración: auth tables int64 → UUID + created_by/updated_by → text
-- Prepara ponti para integración con saas-core (que usa uuid.UUID)
-- =============================================================================
-- NOTA: sin BEGIN/COMMIT explícito — golang-migrate maneja la transacción.

-- 1. Extensión para gen_random_uuid()
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- =============================================================================
-- 2. Drop 36 FK constraints de created_by/updated_by/deleted_by → users(id)
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

-- =============================================================================
-- 3. Convertir created_by/updated_by/deleted_by → text en TODAS las tablas
--    (29 tablas: bigint se castea a text, varchar se castea a text)
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

-- Tablas con varchar (labor_categories, labor_types, providers) ya son text-compatible:
ALTER TABLE public.labor_categories ALTER COLUMN created_by TYPE text USING created_by::text;
ALTER TABLE public.labor_categories ALTER COLUMN updated_by TYPE text USING updated_by::text;
ALTER TABLE public.labor_categories ALTER COLUMN deleted_by TYPE text USING deleted_by::text;

ALTER TABLE public.labor_types ALTER COLUMN created_by TYPE text USING created_by::text;
ALTER TABLE public.labor_types ALTER COLUMN updated_by TYPE text USING updated_by::text;
ALTER TABLE public.labor_types ALTER COLUMN deleted_by TYPE text USING deleted_by::text;

ALTER TABLE public.providers ALTER COLUMN created_by TYPE text USING created_by::text;
ALTER TABLE public.providers ALTER COLUMN updated_by TYPE text USING updated_by::text;
ALTER TABLE public.providers ALTER COLUMN deleted_by TYPE text USING deleted_by::text;

-- stock_movements (si existe)
DO $$ BEGIN
    ALTER TABLE public.stock_movements ALTER COLUMN created_by TYPE text USING created_by::text;
    ALTER TABLE public.stock_movements ALTER COLUMN updated_by TYPE text USING updated_by::text;
    ALTER TABLE public.stock_movements ALTER COLUMN deleted_by TYPE text USING deleted_by::text;
EXCEPTION WHEN undefined_table THEN NULL;
END $$;

-- =============================================================================
-- 4. Migrar auth_tenants.id bigserial → uuid
-- =============================================================================
ALTER TABLE public.auth_memberships DROP CONSTRAINT IF EXISTS auth_memberships_tenant_id_fkey;

ALTER TABLE public.auth_tenants ADD COLUMN new_id uuid DEFAULT gen_random_uuid();
UPDATE public.auth_tenants SET new_id = gen_random_uuid() WHERE new_id IS NULL;

CREATE TEMPORARY TABLE _tenant_map AS SELECT id AS old_id, new_id FROM public.auth_tenants;

ALTER TABLE public.auth_tenants DROP CONSTRAINT IF EXISTS auth_tenants_pkey;
ALTER TABLE public.auth_tenants DROP COLUMN id;
ALTER TABLE public.auth_tenants RENAME COLUMN new_id TO id;
ALTER TABLE public.auth_tenants ADD PRIMARY KEY (id);
ALTER TABLE public.auth_tenants ALTER COLUMN id SET DEFAULT gen_random_uuid();

-- =============================================================================
-- 5. Migrar auth_roles.id bigserial → uuid
-- =============================================================================
ALTER TABLE public.auth_role_permissions DROP CONSTRAINT IF EXISTS auth_role_permissions_role_id_fkey;
ALTER TABLE public.auth_memberships DROP CONSTRAINT IF EXISTS auth_memberships_role_id_fkey;

ALTER TABLE public.auth_roles ADD COLUMN new_id uuid DEFAULT gen_random_uuid();
UPDATE public.auth_roles SET new_id = gen_random_uuid() WHERE new_id IS NULL;

CREATE TEMPORARY TABLE _role_map AS SELECT id AS old_id, new_id FROM public.auth_roles;

ALTER TABLE public.auth_roles DROP CONSTRAINT IF EXISTS auth_roles_pkey;
ALTER TABLE public.auth_roles DROP COLUMN id;
ALTER TABLE public.auth_roles RENAME COLUMN new_id TO id;
ALTER TABLE public.auth_roles ADD PRIMARY KEY (id);
ALTER TABLE public.auth_roles ALTER COLUMN id SET DEFAULT gen_random_uuid();

-- =============================================================================
-- 6. Migrar auth_permissions.id bigserial → uuid
-- =============================================================================
ALTER TABLE public.auth_role_permissions DROP CONSTRAINT IF EXISTS auth_role_permissions_permission_id_fkey;

ALTER TABLE public.auth_permissions ADD COLUMN new_id uuid DEFAULT gen_random_uuid();
UPDATE public.auth_permissions SET new_id = gen_random_uuid() WHERE new_id IS NULL;

CREATE TEMPORARY TABLE _perm_map AS SELECT id AS old_id, new_id FROM public.auth_permissions;

ALTER TABLE public.auth_permissions DROP CONSTRAINT IF EXISTS auth_permissions_pkey;
ALTER TABLE public.auth_permissions DROP COLUMN id;
ALTER TABLE public.auth_permissions RENAME COLUMN new_id TO id;
ALTER TABLE public.auth_permissions ADD PRIMARY KEY (id);
ALTER TABLE public.auth_permissions ALTER COLUMN id SET DEFAULT gen_random_uuid();

-- =============================================================================
-- 7. Reconstruir auth_role_permissions con UUIDs
-- =============================================================================
ALTER TABLE public.auth_role_permissions DROP CONSTRAINT IF EXISTS auth_role_permissions_pkey;

ALTER TABLE public.auth_role_permissions ADD COLUMN new_role_id uuid;
ALTER TABLE public.auth_role_permissions ADD COLUMN new_permission_id uuid;

UPDATE public.auth_role_permissions rp
SET new_role_id = rm.new_id
FROM _role_map rm WHERE rm.old_id = rp.role_id;

UPDATE public.auth_role_permissions rp
SET new_permission_id = pm.new_id
FROM _perm_map pm WHERE pm.old_id = rp.permission_id;

ALTER TABLE public.auth_role_permissions DROP COLUMN role_id;
ALTER TABLE public.auth_role_permissions DROP COLUMN permission_id;
ALTER TABLE public.auth_role_permissions RENAME COLUMN new_role_id TO role_id;
ALTER TABLE public.auth_role_permissions RENAME COLUMN new_permission_id TO permission_id;
ALTER TABLE public.auth_role_permissions ADD PRIMARY KEY (role_id, permission_id);
ALTER TABLE public.auth_role_permissions
    ADD CONSTRAINT fk_auth_rp_role FOREIGN KEY (role_id) REFERENCES public.auth_roles(id) ON DELETE CASCADE;
ALTER TABLE public.auth_role_permissions
    ADD CONSTRAINT fk_auth_rp_perm FOREIGN KEY (permission_id) REFERENCES public.auth_permissions(id) ON DELETE CASCADE;

-- =============================================================================
-- 8. Migrar users.id bigint → uuid
-- =============================================================================
ALTER TABLE public.auth_memberships DROP CONSTRAINT IF EXISTS auth_memberships_user_id_fkey;

-- Drop FK from other tables that reference users(id)
-- (id_rol FK if exists)
ALTER TABLE public.users DROP CONSTRAINT IF EXISTS fk_users_id_rol;

ALTER TABLE public.users ADD COLUMN new_id uuid DEFAULT gen_random_uuid();
UPDATE public.users SET new_id = gen_random_uuid() WHERE new_id IS NULL;

CREATE TEMPORARY TABLE _user_map AS SELECT id AS old_id, new_id FROM public.users;

ALTER TABLE public.users DROP CONSTRAINT IF EXISTS pk_users;
ALTER TABLE public.users DROP COLUMN id;
ALTER TABLE public.users RENAME COLUMN new_id TO id;
ALTER TABLE public.users ADD PRIMARY KEY (id);
ALTER TABLE public.users ALTER COLUMN id SET DEFAULT gen_random_uuid();

-- =============================================================================
-- 9. Reconstruir auth_memberships con UUIDs
-- =============================================================================
ALTER TABLE public.auth_memberships DROP CONSTRAINT IF EXISTS auth_memberships_pkey;
ALTER TABLE public.auth_memberships DROP CONSTRAINT IF EXISTS auth_memberships_user_id_tenant_id_key;

ALTER TABLE public.auth_memberships ADD COLUMN new_id uuid DEFAULT gen_random_uuid();
ALTER TABLE public.auth_memberships ADD COLUMN new_user_id uuid;
ALTER TABLE public.auth_memberships ADD COLUMN new_tenant_id uuid;
ALTER TABLE public.auth_memberships ADD COLUMN new_role_id uuid;

UPDATE public.auth_memberships m
SET new_user_id = um.new_id
FROM _user_map um WHERE um.old_id = m.user_id;

UPDATE public.auth_memberships m
SET new_tenant_id = tm.new_id
FROM _tenant_map tm WHERE tm.old_id = m.tenant_id;

UPDATE public.auth_memberships m
SET new_role_id = rm.new_id
FROM _role_map rm WHERE rm.old_id = m.role_id;

ALTER TABLE public.auth_memberships DROP COLUMN id;
ALTER TABLE public.auth_memberships DROP COLUMN user_id;
ALTER TABLE public.auth_memberships DROP COLUMN tenant_id;
ALTER TABLE public.auth_memberships DROP COLUMN role_id;
ALTER TABLE public.auth_memberships RENAME COLUMN new_id TO id;
ALTER TABLE public.auth_memberships RENAME COLUMN new_user_id TO user_id;
ALTER TABLE public.auth_memberships RENAME COLUMN new_tenant_id TO tenant_id;
ALTER TABLE public.auth_memberships RENAME COLUMN new_role_id TO role_id;
ALTER TABLE public.auth_memberships ADD PRIMARY KEY (id);
ALTER TABLE public.auth_memberships ADD CONSTRAINT uq_membership_user_tenant UNIQUE (user_id, tenant_id);
ALTER TABLE public.auth_memberships
    ADD CONSTRAINT fk_memberships_user FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;
ALTER TABLE public.auth_memberships
    ADD CONSTRAINT fk_memberships_tenant FOREIGN KEY (tenant_id) REFERENCES public.auth_tenants(id) ON DELETE CASCADE;
ALTER TABLE public.auth_memberships
    ADD CONSTRAINT fk_memberships_role FOREIGN KEY (role_id) REFERENCES public.auth_roles(id);

-- =============================================================================
-- 10. Limpiar secuencias huérfanas
-- =============================================================================
DROP SEQUENCE IF EXISTS auth_tenants_id_seq CASCADE;
DROP SEQUENCE IF EXISTS auth_roles_id_seq CASCADE;
DROP SEQUENCE IF EXISTS auth_permissions_id_seq CASCADE;
DROP SEQUENCE IF EXISTS auth_memberships_id_seq CASCADE;
