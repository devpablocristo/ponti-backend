BEGIN;

-- Digital multi-lot work orders are persisted as one physical row per lot
-- (D-n.1, D-n.2, ...). Before this migration some rows stored the full group
-- consumption in every lot. The canonical persisted value per lot is:
-- total_used = final_dose * effective_area.

UPDATE public.work_order_draft_items wodi
SET total_used = ROUND((wodi.final_dose::numeric * wod.effective_area::numeric), 6)
FROM public.work_order_drafts wod
WHERE wodi.draft_id = wod.id
  AND wodi.deleted_at IS NULL
  AND wod.deleted_at IS NULL
  AND wod.is_digital = TRUE
  AND wod.number ~ '^D-[0-9]+\.[0-9]+$'
  AND wodi.final_dose IS NOT NULL
  AND wod.effective_area IS NOT NULL
  AND wod.effective_area > 0;

UPDATE public.workorder_items wi
SET total_used = ROUND((wi.final_dose::numeric * wo.effective_area::numeric), 6)
FROM public.workorders wo
WHERE wi.workorder_id = wo.id
  AND wi.deleted_at IS NULL
  AND wo.deleted_at IS NULL
  AND wo.is_digital = TRUE
  AND wo.number ~ '^D-[0-9]+\.[0-9]+$'
  AND wi.final_dose IS NOT NULL
  AND wo.effective_area IS NOT NULL
  AND wo.effective_area > 0;

COMMIT;
