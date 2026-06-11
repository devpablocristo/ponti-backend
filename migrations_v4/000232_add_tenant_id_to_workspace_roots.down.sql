BEGIN;

-- Revierte T1.e (aditivo). No borra datos de dominio ni el tenant 'default'
-- (puede estar referenciado por auth_memberships / business_insight_candidates).
DROP INDEX IF EXISTS idx_projects_tenant_id;
DROP INDEX IF EXISTS idx_campaigns_tenant_id;
DROP INDEX IF EXISTS idx_customers_tenant_id;

ALTER TABLE public.projects  DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE public.campaigns DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE public.customers DROP COLUMN IF EXISTS tenant_id;

COMMIT;
