BEGIN;

ALTER TABLE IF EXISTS public.work_order_drafts
    ALTER COLUMN created_by TYPE text USING created_by::text;

ALTER TABLE IF EXISTS public.work_order_drafts
    ALTER COLUMN updated_by TYPE text USING updated_by::text;

ALTER TABLE IF EXISTS public.work_order_drafts
    ALTER COLUMN deleted_by TYPE text USING deleted_by::text;

COMMIT;
