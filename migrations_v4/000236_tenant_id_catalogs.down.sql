BEGIN;

-- Restaura unicidad global de nombre (best-effort: falla si hay homónimos cross-tenant).

DROP INDEX IF EXISTS uq_crops_tenant_name;
DROP INDEX IF EXISTS idx_crops_tenant_id;
ALTER TABLE public.crops DROP COLUMN IF EXISTS tenant_id;
CREATE UNIQUE INDEX IF NOT EXISTS uq_crops_name ON public.crops (name);

DROP INDEX IF EXISTS uq_types_tenant_name;
DROP INDEX IF EXISTS idx_types_tenant_id;
ALTER TABLE public.types DROP COLUMN IF EXISTS tenant_id;
CREATE UNIQUE INDEX IF NOT EXISTS uq_types_name ON public.types (name);

DROP INDEX IF EXISTS uq_lease_types_tenant_name;
DROP INDEX IF EXISTS idx_lease_types_tenant_id;
ALTER TABLE public.lease_types DROP COLUMN IF EXISTS tenant_id;
CREATE UNIQUE INDEX IF NOT EXISTS uq_lease_types_name ON public.lease_types (name);

DROP INDEX IF EXISTS uq_business_parameters_tenant_key;
DROP INDEX IF EXISTS idx_business_parameters_tenant_id;
ALTER TABLE public.business_parameters DROP COLUMN IF EXISTS tenant_id;
CREATE UNIQUE INDEX IF NOT EXISTS uq_business_parameters_key ON public.business_parameters (key);

DROP INDEX IF EXISTS idx_categories_tenant_id;
ALTER TABLE public.categories DROP COLUMN IF EXISTS tenant_id;

COMMIT;
