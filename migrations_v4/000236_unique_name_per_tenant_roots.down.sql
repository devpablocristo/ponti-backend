BEGIN;

DROP INDEX IF EXISTS uq_customers_tenant_name;
DROP INDEX IF EXISTS uq_campaigns_tenant_name;

-- Restaura la unicidad global de nombre (best-effort).
CREATE UNIQUE INDEX IF NOT EXISTS uq_customers_name ON public.customers (name);
CREATE UNIQUE INDEX IF NOT EXISTS uq_campaigns_name ON public.campaigns (name);

COMMIT;
