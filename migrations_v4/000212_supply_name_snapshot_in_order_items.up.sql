BEGIN;

ALTER TABLE public.work_order_draft_items
    ADD COLUMN IF NOT EXISTS supply_name character varying(100);

ALTER TABLE public.workorder_items
    ADD COLUMN IF NOT EXISTS supply_name character varying(100);

UPDATE public.work_order_draft_items wodi
SET supply_name = s.name
FROM public.supplies s
WHERE s.id = wodi.supply_id
  AND wodi.supply_name IS NULL;

UPDATE public.workorder_items wi
SET supply_name = s.name
FROM public.supplies s
WHERE s.id = wi.supply_id
  AND wi.supply_name IS NULL;

ALTER TABLE public.work_order_draft_items
    ALTER COLUMN supply_name SET NOT NULL;

ALTER TABLE public.workorder_items
    ALTER COLUMN supply_name SET NOT NULL;

COMMIT;
