BEGIN;

DROP INDEX IF EXISTS idx_business_insight_candidates_tenant_entity;
DROP INDEX IF EXISTS idx_business_insight_candidates_tenant_status;
DROP TABLE IF EXISTS public.business_insight_candidates;

COMMIT;
