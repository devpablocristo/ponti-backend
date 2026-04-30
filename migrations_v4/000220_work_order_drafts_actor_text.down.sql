BEGIN;

ALTER TABLE IF EXISTS public.work_order_drafts
    ALTER COLUMN created_by TYPE bigint USING CASE WHEN created_by ~ '^[0-9]+$' THEN created_by::bigint ELSE NULL END;

ALTER TABLE IF EXISTS public.work_order_drafts
    ALTER COLUMN updated_by TYPE bigint USING CASE WHEN updated_by ~ '^[0-9]+$' THEN updated_by::bigint ELSE NULL END;

ALTER TABLE IF EXISTS public.work_order_drafts
    ALTER COLUMN deleted_by TYPE bigint USING CASE WHEN deleted_by ~ '^[0-9]+$' THEN deleted_by::bigint ELSE NULL END;

COMMIT;
