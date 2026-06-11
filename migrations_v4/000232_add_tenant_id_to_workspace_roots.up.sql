BEGIN;

-- T1.e: introducir tenant_id físico en las RAÍCES de workspace (customers,
-- campaigns, projects) para poder validar workspace-ownership contra el tenant.
-- Modelo: 1 tenant : N customers (tenant_id referencia lógica a auth_tenants).
-- Aditivo y reversible: columna NULLABLE + backfill al tenant 'default' + índice.
-- El endurecimiento (NOT NULL / FK VALIDATE / unicidad / RLS) pertenece a T2/T3.

-- Resolver (o crear) el tenant 'default' para el backfill de datos
-- pre-multitenancy (hoy el sistema opera con un único tenant lógico).
-- (auth_tenants.id tiene default en DB; se omite igual que en ensureTenantByName.)
INSERT INTO public.auth_tenants (name, created_at, updated_at)
SELECT 'default', now(), now()
WHERE NOT EXISTS (SELECT 1 FROM public.auth_tenants WHERE name = 'default');

-- Columnas aditivas (nullable, sin default volátil => ADD COLUMN casi instantáneo).
ALTER TABLE public.customers ADD COLUMN IF NOT EXISTS tenant_id uuid;
ALTER TABLE public.campaigns ADD COLUMN IF NOT EXISTS tenant_id uuid;
ALTER TABLE public.projects  ADD COLUMN IF NOT EXISTS tenant_id uuid;

-- Backfill: customers y campaigns -> tenant 'default'.
UPDATE public.customers
   SET tenant_id = (SELECT id FROM public.auth_tenants WHERE name = 'default' LIMIT 1)
 WHERE tenant_id IS NULL;

UPDATE public.campaigns
   SET tenant_id = (SELECT id FROM public.auth_tenants WHERE name = 'default' LIMIT 1)
 WHERE tenant_id IS NULL;

-- projects: derivar el tenant de su customer (raíz de propiedad); fallback a 'default'.
UPDATE public.projects p
   SET tenant_id = c.tenant_id
  FROM public.customers c
 WHERE p.customer_id = c.id
   AND p.tenant_id IS NULL;

UPDATE public.projects
   SET tenant_id = (SELECT id FROM public.auth_tenants WHERE name = 'default' LIMIT 1)
 WHERE tenant_id IS NULL;

-- Índices con prefijo tenant_id. No CONCURRENTLY: el runner ejecuta cada archivo
-- dentro de una transacción (CONCURRENTLY no puede correr en un bloque tx).
CREATE INDEX IF NOT EXISTS idx_customers_tenant_id ON public.customers (tenant_id);
CREATE INDEX IF NOT EXISTS idx_campaigns_tenant_id ON public.campaigns (tenant_id);
CREATE INDEX IF NOT EXISTS idx_projects_tenant_id  ON public.projects  (tenant_id);

COMMIT;
