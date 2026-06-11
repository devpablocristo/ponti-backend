BEGIN;

DROP INDEX IF EXISTS idx_ai_decision_cards_tenant_route;
DROP INDEX IF EXISTS idx_ai_decision_cards_tenant_status_seen;
DROP TABLE IF EXISTS public.ai_decision_cards;

DROP INDEX IF EXISTS idx_ai_decision_runs_tenant_created;
DROP TABLE IF EXISTS public.ai_decision_runs;

COMMIT;
