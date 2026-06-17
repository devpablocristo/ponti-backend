BEGIN;

DROP INDEX IF EXISTS uq_managers_tenant_name;
DROP INDEX IF EXISTS idx_managers_tenant_id;
ALTER TABLE public.managers DROP COLUMN IF EXISTS tenant_id;
CREATE UNIQUE INDEX IF NOT EXISTS uq_managers_name ON public.managers (name);

DROP INDEX IF EXISTS uq_investors_tenant_name;
DROP INDEX IF EXISTS idx_investors_tenant_id;
ALTER TABLE public.investors DROP COLUMN IF EXISTS tenant_id;
CREATE UNIQUE INDEX IF NOT EXISTS uq_investors_name ON public.investors (name);

DROP INDEX IF EXISTS uq_providers_tenant_name;
DROP INDEX IF EXISTS idx_providers_tenant_id;
ALTER TABLE public.providers DROP COLUMN IF EXISTS tenant_id;
CREATE UNIQUE INDEX IF NOT EXISTS uq_providers_name ON public.providers (name);

COMMIT;
