BEGIN;

DROP INDEX IF EXISTS idx_ai_governance_requests_nexus_request;
DROP INDEX IF EXISTS idx_ai_governance_requests_tenant_status;
DROP TABLE IF EXISTS public.ai_governance_requests;

COMMIT;
