BEGIN;

DROP INDEX IF EXISTS idx_ai_action_drafts_tenant_status;
DROP TABLE IF EXISTS public.ai_action_drafts;

COMMIT;
