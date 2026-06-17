BEGIN;

DROP INDEX IF EXISTS idx_auth_tenants_status;
ALTER TABLE public.auth_tenants DROP CONSTRAINT IF EXISTS auth_tenants_status_check;
ALTER TABLE public.auth_tenants DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE public.auth_tenants DROP COLUMN IF EXISTS status;

COMMIT;
