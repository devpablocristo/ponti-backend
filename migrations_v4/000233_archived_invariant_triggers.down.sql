-- 000233_archived_invariant_triggers.down.sql
-- Reverse: drop triggers + the shared function.

BEGIN;

DROP TRIGGER IF EXISTS trg_commercializations_active_crop    ON public.crop_commercializations;
DROP TRIGGER IF EXISTS trg_commercializations_active_project ON public.crop_commercializations;
DROP TRIGGER IF EXISTS trg_stocks_active_project             ON public.stocks;
DROP TRIGGER IF EXISTS trg_supply_movements_active_project   ON public.supply_movements;
DROP TRIGGER IF EXISTS trg_supplies_active_project           ON public.supplies;
DROP TRIGGER IF EXISTS trg_labors_active_project             ON public.labors;
DROP TRIGGER IF EXISTS trg_drafts_active_project             ON public.work_order_drafts;
DROP TRIGGER IF EXISTS trg_workorders_active_lot             ON public.workorders;
DROP TRIGGER IF EXISTS trg_workorders_active_field           ON public.workorders;
DROP TRIGGER IF EXISTS trg_workorders_active_project         ON public.workorders;
DROP TRIGGER IF EXISTS trg_lots_active_field                 ON public.lots;
DROP TRIGGER IF EXISTS trg_fields_active_project             ON public.fields;
DROP TRIGGER IF EXISTS trg_projects_active_customer          ON public.projects;

DROP FUNCTION IF EXISTS public.assert_parent_active();

COMMIT;
