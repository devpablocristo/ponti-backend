BEGIN;

-- T3 / Modelo 2: tenant ownership de master entities claramente tenant-owned:
-- managers, investors, providers. ADD tenant_id + backfill al default + índice +
-- unicidad de nombre POR TENANT (antes global).
-- PENDIENTE (decisión global-vs-por-tenant del catálogo, NO incluidas acá):
--   crops, categories, types, lease_types, business_parameters.

-- managers
ALTER TABLE public.managers ADD COLUMN IF NOT EXISTS tenant_id uuid;
UPDATE public.managers SET tenant_id = (SELECT id FROM public.auth_tenants WHERE name='default' LIMIT 1) WHERE tenant_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_managers_tenant_id ON public.managers (tenant_id);
ALTER TABLE public.managers DROP CONSTRAINT IF EXISTS uq_managers_name;
DROP INDEX IF EXISTS uq_managers_name;
CREATE UNIQUE INDEX IF NOT EXISTS uq_managers_tenant_name ON public.managers (tenant_id, name) WHERE deleted_at IS NULL;

-- investors
ALTER TABLE public.investors ADD COLUMN IF NOT EXISTS tenant_id uuid;
UPDATE public.investors SET tenant_id = (SELECT id FROM public.auth_tenants WHERE name='default' LIMIT 1) WHERE tenant_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_investors_tenant_id ON public.investors (tenant_id);
ALTER TABLE public.investors DROP CONSTRAINT IF EXISTS uq_investors_name;
DROP INDEX IF EXISTS uq_investors_name;
CREATE UNIQUE INDEX IF NOT EXISTS uq_investors_tenant_name ON public.investors (tenant_id, name) WHERE deleted_at IS NULL;

-- providers
ALTER TABLE public.providers ADD COLUMN IF NOT EXISTS tenant_id uuid;
UPDATE public.providers SET tenant_id = (SELECT id FROM public.auth_tenants WHERE name='default' LIMIT 1) WHERE tenant_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_providers_tenant_id ON public.providers (tenant_id);
ALTER TABLE public.providers DROP CONSTRAINT IF EXISTS uq_providers_name;
DROP INDEX IF EXISTS uq_providers_name;
CREATE UNIQUE INDEX IF NOT EXISTS uq_providers_tenant_name ON public.providers (tenant_id, name) WHERE deleted_at IS NULL;

COMMIT;
