-- 000233_archived_invariant_triggers.up.sql
--
-- Defense in depth: BEFORE INSERT/UPDATE triggers that reject the impossible
-- state "child active row references archived parent". Even if a raw SQL
-- or a code path that bypasses the BE `assertXReferencesActive` helpers
-- tries to write an inconsistent row, Postgres refuses at the DB level.
--
-- This is the THIRD line of defense after:
--   1. Go-level `lifecycle.RequireAllActive` in every Create/Update.
--   2. Restore checks (RestoreRequiresActiveParent).
--
-- Prerequisite: data-audit cleanup must have run. If the DB contains rows
-- that already violate the invariant, the CREATE TRIGGER statements
-- themselves don't fail (triggers fire only on new INSERT/UPDATE), but the
-- first UPDATE on such a row will. Run scripts/data-audit/archived_invariants.sql
-- in staging first.

BEGIN;

-- Generic helper: assert that the parent referenced by NEW.<parent_fk> is
-- active (deleted_at IS NULL). Active children referencing archived parents
-- are rejected with a CHECK violation (SQLSTATE 23514) that maps to a
-- Conflict in the BE error layer.
--
-- TG_ARGV: parent_table, parent_fk_column
CREATE OR REPLACE FUNCTION public.assert_parent_active() RETURNS trigger
LANGUAGE plpgsql AS $$
DECLARE
  parent_table   text := TG_ARGV[0];
  parent_fk      text := TG_ARGV[1];
  parent_id      bigint;
  parent_deleted timestamptz;
BEGIN
  -- Archiving the child is always allowed (deleted_at being set on NEW).
  IF NEW.deleted_at IS NOT NULL THEN
    RETURN NEW;
  END IF;

  EXECUTE format('SELECT ($1).%I', parent_fk)
    INTO parent_id
    USING NEW;

  -- Nullable FKs are skipped — the BE-level assertReferences validates
  -- presence; here we only enforce "if there's a parent, it must be active".
  IF parent_id IS NULL OR parent_id <= 0 THEN
    RETURN NEW;
  END IF;

  EXECUTE format('SELECT deleted_at FROM %I WHERE id = $1', parent_table)
    INTO parent_deleted
    USING parent_id;

  IF parent_deleted IS NOT NULL THEN
    RAISE EXCEPTION
      'active % row references archived % (id=%)',
      TG_TABLE_NAME, parent_table, parent_id
      USING ERRCODE = '23514';
  END IF;

  RETURN NEW;
END;
$$;

-- ─────────────────── triggers per parent → child relation ───────────────────
-- projects.customer_id → customers.id
CREATE TRIGGER trg_projects_active_customer
  BEFORE INSERT OR UPDATE OF customer_id, deleted_at ON public.projects
  FOR EACH ROW EXECUTE FUNCTION public.assert_parent_active('customers', 'customer_id');

-- fields.project_id → projects.id
CREATE TRIGGER trg_fields_active_project
  BEFORE INSERT OR UPDATE OF project_id, deleted_at ON public.fields
  FOR EACH ROW EXECUTE FUNCTION public.assert_parent_active('projects', 'project_id');

-- lots.field_id → fields.id
CREATE TRIGGER trg_lots_active_field
  BEFORE INSERT OR UPDATE OF field_id, deleted_at ON public.lots
  FOR EACH ROW EXECUTE FUNCTION public.assert_parent_active('fields', 'field_id');

-- workorders: project + field + lot
CREATE TRIGGER trg_workorders_active_project
  BEFORE INSERT OR UPDATE OF project_id, deleted_at ON public.workorders
  FOR EACH ROW EXECUTE FUNCTION public.assert_parent_active('projects', 'project_id');
CREATE TRIGGER trg_workorders_active_field
  BEFORE INSERT OR UPDATE OF field_id, deleted_at ON public.workorders
  FOR EACH ROW EXECUTE FUNCTION public.assert_parent_active('fields', 'field_id');
CREATE TRIGGER trg_workorders_active_lot
  BEFORE INSERT OR UPDATE OF lot_id, deleted_at ON public.workorders
  FOR EACH ROW EXECUTE FUNCTION public.assert_parent_active('lots', 'lot_id');

-- work_order_drafts: project
CREATE TRIGGER trg_drafts_active_project
  BEFORE INSERT OR UPDATE OF project_id, deleted_at ON public.work_order_drafts
  FOR EACH ROW EXECUTE FUNCTION public.assert_parent_active('projects', 'project_id');

-- labors / supplies / stocks / supply_movements: project
CREATE TRIGGER trg_labors_active_project
  BEFORE INSERT OR UPDATE OF project_id, deleted_at ON public.labors
  FOR EACH ROW EXECUTE FUNCTION public.assert_parent_active('projects', 'project_id');
CREATE TRIGGER trg_supplies_active_project
  BEFORE INSERT OR UPDATE OF project_id, deleted_at ON public.supplies
  FOR EACH ROW EXECUTE FUNCTION public.assert_parent_active('projects', 'project_id');
CREATE TRIGGER trg_supply_movements_active_project
  BEFORE INSERT OR UPDATE OF project_id, deleted_at ON public.supply_movements
  FOR EACH ROW EXECUTE FUNCTION public.assert_parent_active('projects', 'project_id');
CREATE TRIGGER trg_stocks_active_project
  BEFORE INSERT OR UPDATE OF project_id, deleted_at ON public.stocks
  FOR EACH ROW EXECUTE FUNCTION public.assert_parent_active('projects', 'project_id');

-- crop_commercializations: project + crop
CREATE TRIGGER trg_commercializations_active_project
  BEFORE INSERT OR UPDATE OF project_id, deleted_at ON public.crop_commercializations
  FOR EACH ROW EXECUTE FUNCTION public.assert_parent_active('projects', 'project_id');
CREATE TRIGGER trg_commercializations_active_crop
  BEFORE INSERT OR UPDATE OF crop_id, deleted_at ON public.crop_commercializations
  FOR EACH ROW EXECUTE FUNCTION public.assert_parent_active('crops', 'crop_id');

COMMIT;
