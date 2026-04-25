BEGIN;

ALTER TABLE public.work_order_draft_items
    DROP COLUMN IF EXISTS supply_name;

ALTER TABLE public.workorder_items
    DROP COLUMN IF EXISTS supply_name;

COMMIT;
