BEGIN;

ALTER TABLE IF EXISTS public.work_order_drafts
    ALTER COLUMN created_by TYPE bigint USING NULLIF(created_by, '')::bigint;

ALTER TABLE IF EXISTS public.work_order_drafts
    ALTER COLUMN updated_by TYPE bigint USING NULLIF(updated_by, '')::bigint;

ALTER TABLE IF EXISTS public.work_order_drafts
    ALTER COLUMN deleted_by TYPE bigint USING NULLIF(deleted_by, '')::bigint;

COMMIT;
