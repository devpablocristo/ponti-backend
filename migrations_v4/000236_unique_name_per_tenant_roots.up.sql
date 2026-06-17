BEGIN;

-- T3 (soporte Modelo 2): unicidad de nombre POR TENANT en customers y campaigns
-- (antes era GLOBAL, lo que impedía que dos tenants tuvieran el mismo nombre).
-- Mantiene la semántica raw-name del unique anterior, solo acotada por tenant,
-- y parcial WHERE deleted_at IS NULL (no bloquea nombres de archivados).
-- NOTA: managers/investors/providers/crops/categories/types/lease_types/
-- business_parameters todavía NO tienen tenant_id -> su unicidad-por-tenant
-- queda pendiente (requiere agregarles tenant_id primero).

ALTER TABLE public.customers DROP CONSTRAINT IF EXISTS uq_customers_name;
DROP INDEX IF EXISTS uq_customers_name;
CREATE UNIQUE INDEX IF NOT EXISTS uq_customers_tenant_name
	ON public.customers (tenant_id, name) WHERE deleted_at IS NULL;

ALTER TABLE public.campaigns DROP CONSTRAINT IF EXISTS uq_campaigns_name;
DROP INDEX IF EXISTS uq_campaigns_name;
CREATE UNIQUE INDEX IF NOT EXISTS uq_campaigns_tenant_name
	ON public.campaigns (tenant_id, name) WHERE deleted_at IS NULL;

COMMIT;
