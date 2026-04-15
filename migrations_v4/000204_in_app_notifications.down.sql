BEGIN;

DROP INDEX IF EXISTS idx_business_insight_candidates_tenant_entity;
DROP INDEX IF EXISTS idx_business_insight_candidates_tenant_status;
DROP TABLE IF EXISTS public.business_insight_candidates;

DROP INDEX IF EXISTS idx_in_app_notifications_tenant_created;
DROP INDEX IF EXISTS idx_in_app_notifications_user_created;
DROP TABLE IF EXISTS public.in_app_notifications;

COMMIT;
