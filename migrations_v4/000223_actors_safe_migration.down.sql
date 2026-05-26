-- ========================================
-- MIGRATION 000223 ACTORS SAFE MIGRATION (DOWN)
-- ========================================

BEGIN;

DROP VIEW IF EXISTS public.actor_migration_coverage;

ALTER TABLE IF EXISTS public.invoices
    DROP CONSTRAINT IF EXISTS fk_invoices_company_actor,
    DROP CONSTRAINT IF EXISTS fk_invoices_investor_actor;
ALTER TABLE IF EXISTS public.labors
    DROP CONSTRAINT IF EXISTS fk_labors_contractor_actor;
ALTER TABLE IF EXISTS public.supply_movements
    DROP CONSTRAINT IF EXISTS fk_supply_movements_provider_actor,
    DROP CONSTRAINT IF EXISTS fk_supply_movements_investor_actor;
ALTER TABLE IF EXISTS public.stocks
    DROP CONSTRAINT IF EXISTS fk_stocks_investor_actor;
ALTER TABLE IF EXISTS public.workorder_investor_splits
    DROP CONSTRAINT IF EXISTS fk_workorder_investor_splits_actor;
ALTER TABLE IF EXISTS public.workorders
    DROP CONSTRAINT IF EXISTS fk_workorders_contractor_actor,
    DROP CONSTRAINT IF EXISTS fk_workorders_investor_actor;
ALTER TABLE IF EXISTS public.projects
    DROP CONSTRAINT IF EXISTS fk_projects_customer_actor;

DROP INDEX IF EXISTS public.idx_invoices_company_actor_id;
DROP INDEX IF EXISTS public.idx_invoices_investor_actor_id;
DROP INDEX IF EXISTS public.idx_labors_contractor_actor_id;
DROP INDEX IF EXISTS public.idx_supply_movements_provider_actor_id;
DROP INDEX IF EXISTS public.idx_supply_movements_investor_actor_id;
DROP INDEX IF EXISTS public.idx_stocks_investor_actor_id;
DROP INDEX IF EXISTS public.idx_workorder_splits_actor_id;
DROP INDEX IF EXISTS public.idx_workorders_contractor_actor_id;
DROP INDEX IF EXISTS public.idx_workorders_investor_actor_id;
DROP INDEX IF EXISTS public.idx_projects_customer_actor_id;

ALTER TABLE IF EXISTS public.invoices
    DROP COLUMN IF EXISTS company_name_snapshot,
    DROP COLUMN IF EXISTS company_actor_id,
    DROP COLUMN IF EXISTS investor_actor_id;
ALTER TABLE IF EXISTS public.labors
    DROP COLUMN IF EXISTS contractor_name_snapshot,
    DROP COLUMN IF EXISTS contractor_actor_id;
ALTER TABLE IF EXISTS public.supply_movements
    DROP COLUMN IF EXISTS provider_actor_id,
    DROP COLUMN IF EXISTS investor_actor_id;
ALTER TABLE IF EXISTS public.stocks
    DROP COLUMN IF EXISTS investor_actor_id;
ALTER TABLE IF EXISTS public.workorder_investor_splits
    DROP COLUMN IF EXISTS actor_id;
ALTER TABLE IF EXISTS public.workorders
    DROP COLUMN IF EXISTS contractor_name_snapshot,
    DROP COLUMN IF EXISTS contractor_actor_id,
    DROP COLUMN IF EXISTS investor_actor_id;
ALTER TABLE IF EXISTS public.projects
    DROP COLUMN IF EXISTS customer_actor_id;

DROP TABLE IF EXISTS public.field_lease_participants;
DROP TABLE IF EXISTS public.project_admin_cost_allocations;
DROP TABLE IF EXISTS public.project_investor_allocations;
DROP TABLE IF EXISTS public.project_responsibles;
DROP TABLE IF EXISTS public.legacy_actor_map;
DROP TABLE IF EXISTS public.actor_merge_log;
DROP TABLE IF EXISTS public.actor_relationships;
DROP TABLE IF EXISTS public.actor_aliases;
DROP TABLE IF EXISTS public.actor_roles;
DROP TABLE IF EXISTS public.actor_identifiers;
DROP TABLE IF EXISTS public.actor_organization_profiles;
DROP TABLE IF EXISTS public.actor_person_profiles;
DROP TABLE IF EXISTS public.actors;

DROP FUNCTION IF EXISTS public.normalize_actor_name(text);

COMMIT;
