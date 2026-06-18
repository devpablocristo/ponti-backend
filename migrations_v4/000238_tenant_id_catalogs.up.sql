BEGIN;

-- T3 / Modelo 2 (decisión IV.6): catálogos POR-TENANT. Cada tenant maneja su propia
-- data aislada de catálogo (mismas entidades, info aislada). ADD tenant_id + backfill
-- al tenant 'default' + índice + unicidad de nombre POR TENANT (antes global) donde
-- existía unique.
-- fx_rates EXCLUIDO a propósito: es dato de mercado (par/fecha), sin deleted_at →
-- tratado como referencia global; decisión aparte.

-- crops (name + uq_crops_name global)
ALTER TABLE public.crops ADD COLUMN IF NOT EXISTS tenant_id uuid;
UPDATE public.crops SET tenant_id = (SELECT id FROM public.auth_tenants WHERE name='default' LIMIT 1) WHERE tenant_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_crops_tenant_id ON public.crops (tenant_id);
ALTER TABLE public.crops DROP CONSTRAINT IF EXISTS uq_crops_name;
DROP INDEX IF EXISTS uq_crops_name;
CREATE UNIQUE INDEX IF NOT EXISTS uq_crops_tenant_name ON public.crops (tenant_id, name) WHERE deleted_at IS NULL;

-- types (módulo class-type; name + uq_types_name global)
ALTER TABLE public.types ADD COLUMN IF NOT EXISTS tenant_id uuid;
UPDATE public.types SET tenant_id = (SELECT id FROM public.auth_tenants WHERE name='default' LIMIT 1) WHERE tenant_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_types_tenant_id ON public.types (tenant_id);
ALTER TABLE public.types DROP CONSTRAINT IF EXISTS uq_types_name;
DROP INDEX IF EXISTS uq_types_name;
CREATE UNIQUE INDEX IF NOT EXISTS uq_types_tenant_name ON public.types (tenant_id, name) WHERE deleted_at IS NULL;

-- lease_types (name + uq_lease_types_name global)
ALTER TABLE public.lease_types ADD COLUMN IF NOT EXISTS tenant_id uuid;
UPDATE public.lease_types SET tenant_id = (SELECT id FROM public.auth_tenants WHERE name='default' LIMIT 1) WHERE tenant_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_lease_types_tenant_id ON public.lease_types (tenant_id);
ALTER TABLE public.lease_types DROP CONSTRAINT IF EXISTS uq_lease_types_name;
DROP INDEX IF EXISTS uq_lease_types_name;
CREATE UNIQUE INDEX IF NOT EXISTS uq_lease_types_tenant_name ON public.lease_types (tenant_id, name) WHERE deleted_at IS NULL;

-- business_parameters (keyed por `key` + uq_business_parameters_key global; sin name)
ALTER TABLE public.business_parameters ADD COLUMN IF NOT EXISTS tenant_id uuid;
UPDATE public.business_parameters SET tenant_id = (SELECT id FROM public.auth_tenants WHERE name='default' LIMIT 1) WHERE tenant_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_business_parameters_tenant_id ON public.business_parameters (tenant_id);
ALTER TABLE public.business_parameters DROP CONSTRAINT IF EXISTS uq_business_parameters_key;
DROP INDEX IF EXISTS uq_business_parameters_key;
CREATE UNIQUE INDEX IF NOT EXISTS uq_business_parameters_tenant_key ON public.business_parameters (tenant_id, key) WHERE deleted_at IS NULL;

-- categories (sin unique global previo → solo tenant_id + índice)
ALTER TABLE public.categories ADD COLUMN IF NOT EXISTS tenant_id uuid;
UPDATE public.categories SET tenant_id = (SELECT id FROM public.auth_tenants WHERE name='default' LIMIT 1) WHERE tenant_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_categories_tenant_id ON public.categories (tenant_id);

COMMIT;
