-- ========================================
-- MIGRATION 000234 ACTOR UNIQUE NORMALIZED NAME (UP)
-- ========================================
-- Un actor activo es una identidad unica por tenant. El nombre normalizado no
-- puede repetirse aunque cambien tipo, roles, perfiles o datos de contacto.

BEGIN;

DROP INDEX IF EXISTS public.idx_actors_tenant_normalized_name;

CREATE TEMP TABLE _actor_name_merge_sources ON COMMIT DROP AS
WITH active_actors AS (
    SELECT
        id,
        tenant_id,
        normalized_name,
        first_value(id) OVER (
            PARTITION BY tenant_id, normalized_name
            ORDER BY id
        ) AS target_actor_id,
        row_number() OVER (
            PARTITION BY tenant_id, normalized_name
            ORDER BY id
        ) AS rn
    FROM public.actors
    WHERE deleted_at IS NULL
      AND merged_into_actor_id IS NULL
      AND normalized_name <> ''
),
duplicates AS (
    SELECT id AS source_actor_id, target_actor_id
    FROM active_actors
    WHERE rn > 1
)
SELECT source_actor_id, target_actor_id
FROM duplicates;

INSERT INTO public.actor_roles (actor_id, role, created_at, deleted_at)
SELECT
    m.target_actor_id,
    ar.role,
    now(),
    CASE
        WHEN COUNT(*) FILTER (WHERE ar.deleted_at IS NULL) > 0 THEN NULL
        ELSE MIN(ar.deleted_at)
    END AS deleted_at
FROM _actor_name_merge_sources m
JOIN public.actor_roles ar ON ar.actor_id = m.source_actor_id
GROUP BY m.target_actor_id, ar.role
ON CONFLICT (actor_id, role) DO UPDATE
SET deleted_at = CASE
    WHEN public.actor_roles.deleted_at IS NULL OR EXCLUDED.deleted_at IS NULL THEN NULL
    ELSE LEAST(public.actor_roles.deleted_at, EXCLUDED.deleted_at)
END;

INSERT INTO public.actor_aliases (tenant_id, actor_id, alias, normalized_alias, source, created_at, deleted_at)
SELECT aa.tenant_id, m.target_actor_id, aa.alias, aa.normalized_alias, aa.source, now(), aa.deleted_at
FROM _actor_name_merge_sources m
JOIN public.actor_aliases aa ON aa.actor_id = m.source_actor_id
WHERE aa.deleted_at IS NULL
ON CONFLICT DO NOTHING;

UPDATE public.actor_identifiers ai
SET actor_id = m.target_actor_id
FROM _actor_name_merge_sources m
WHERE ai.actor_id = m.source_actor_id
  AND NOT EXISTS (
      SELECT 1
      FROM public.actor_identifiers target
      WHERE target.tenant_id = ai.tenant_id
        AND target.country = ai.country
        AND target.identifier_type = ai.identifier_type
        AND target.normalized_identifier_value = ai.normalized_identifier_value
        AND target.actor_id <> ai.actor_id
  );

UPDATE public.legacy_actor_map lam
SET actor_id = m.target_actor_id
FROM _actor_name_merge_sources m
WHERE lam.actor_id = m.source_actor_id;

UPDATE public.projects p
SET customer_actor_id = m.target_actor_id
FROM _actor_name_merge_sources m
WHERE p.customer_actor_id = m.source_actor_id;

UPDATE public.workorders w
SET investor_actor_id = m.target_actor_id
FROM _actor_name_merge_sources m
WHERE w.investor_actor_id = m.source_actor_id;

UPDATE public.workorders w
SET contractor_actor_id = m.target_actor_id
FROM _actor_name_merge_sources m
WHERE w.contractor_actor_id = m.source_actor_id;

UPDATE public.workorder_investor_splits wis
SET actor_id = m.target_actor_id
FROM _actor_name_merge_sources m
WHERE wis.actor_id = m.source_actor_id;

UPDATE public.stocks s
SET investor_actor_id = m.target_actor_id
FROM _actor_name_merge_sources m
WHERE s.investor_actor_id = m.source_actor_id;

UPDATE public.supply_movements sm
SET investor_actor_id = m.target_actor_id
FROM _actor_name_merge_sources m
WHERE sm.investor_actor_id = m.source_actor_id;

UPDATE public.supply_movements sm
SET provider_actor_id = m.target_actor_id
FROM _actor_name_merge_sources m
WHERE sm.provider_actor_id = m.source_actor_id;

UPDATE public.labors l
SET contractor_actor_id = m.target_actor_id
FROM _actor_name_merge_sources m
WHERE l.contractor_actor_id = m.source_actor_id;

UPDATE public.invoices i
SET investor_actor_id = m.target_actor_id
FROM _actor_name_merge_sources m
WHERE i.investor_actor_id = m.source_actor_id;

UPDATE public.invoices i
SET company_actor_id = m.target_actor_id
FROM _actor_name_merge_sources m
WHERE i.company_actor_id = m.source_actor_id;

UPDATE public.project_responsibles pr
SET deleted_at = now(), updated_at = now(), deleted_by = COALESCE(deleted_by, 'migration:000234')
FROM _actor_name_merge_sources m
WHERE pr.actor_id = m.source_actor_id
  AND pr.deleted_at IS NULL
  AND EXISTS (
      SELECT 1
      FROM public.project_responsibles target
      WHERE target.project_id = pr.project_id
        AND target.actor_id = m.target_actor_id
        AND target.deleted_at IS NULL
  );

UPDATE public.project_responsibles pr
SET actor_id = m.target_actor_id, updated_at = now(), updated_by = COALESCE(updated_by, 'migration:000234')
FROM _actor_name_merge_sources m
WHERE pr.actor_id = m.source_actor_id
  AND pr.deleted_at IS NULL;

UPDATE public.project_investor_allocations pia
SET deleted_at = now(), updated_at = now(), deleted_by = COALESCE(deleted_by, 'migration:000234')
FROM _actor_name_merge_sources m
WHERE pia.actor_id = m.source_actor_id
  AND pia.deleted_at IS NULL
  AND EXISTS (
      SELECT 1
      FROM public.project_investor_allocations target
      WHERE target.project_id = pia.project_id
        AND target.actor_id = m.target_actor_id
        AND target.deleted_at IS NULL
  );

UPDATE public.project_investor_allocations pia
SET actor_id = m.target_actor_id, updated_at = now(), updated_by = COALESCE(updated_by, 'migration:000234')
FROM _actor_name_merge_sources m
WHERE pia.actor_id = m.source_actor_id
  AND pia.deleted_at IS NULL;

UPDATE public.project_admin_cost_allocations paca
SET deleted_at = now(), updated_at = now(), deleted_by = COALESCE(deleted_by, 'migration:000234')
FROM _actor_name_merge_sources m
WHERE paca.actor_id = m.source_actor_id
  AND paca.deleted_at IS NULL
  AND EXISTS (
      SELECT 1
      FROM public.project_admin_cost_allocations target
      WHERE target.project_id = paca.project_id
        AND target.actor_id = m.target_actor_id
        AND target.deleted_at IS NULL
  );

UPDATE public.project_admin_cost_allocations paca
SET actor_id = m.target_actor_id, updated_at = now(), updated_by = COALESCE(updated_by, 'migration:000234')
FROM _actor_name_merge_sources m
WHERE paca.actor_id = m.source_actor_id
  AND paca.deleted_at IS NULL;

UPDATE public.field_lease_participants flp
SET deleted_at = now(), updated_at = now(), deleted_by = COALESCE(deleted_by, 'migration:000234')
FROM _actor_name_merge_sources m
WHERE flp.actor_id = m.source_actor_id
  AND flp.deleted_at IS NULL
  AND EXISTS (
      SELECT 1
      FROM public.field_lease_participants target
      WHERE target.field_id = flp.field_id
        AND target.actor_id = m.target_actor_id
        AND target.deleted_at IS NULL
  );

UPDATE public.field_lease_participants flp
SET actor_id = m.target_actor_id, updated_at = now(), updated_by = COALESCE(updated_by, 'migration:000234')
FROM _actor_name_merge_sources m
WHERE flp.actor_id = m.source_actor_id
  AND flp.deleted_at IS NULL;

INSERT INTO public.actor_merge_log (from_actor_id, to_actor_id, merged_by, reason, impact)
SELECT source_actor_id,
       target_actor_id,
       'migration:000234',
       'auto merge duplicate normalized actor name before unique index',
       jsonb_build_object('source', 'migration:000234')
FROM _actor_name_merge_sources;

UPDATE public.actors a
SET merged_into_actor_id = m.target_actor_id,
    deleted_at = now(),
    updated_at = now(),
    updated_by = COALESCE(updated_by, 'migration:000234'),
    deleted_by = COALESCE(deleted_by, 'migration:000234')
FROM _actor_name_merge_sources m
WHERE a.id = m.source_actor_id;

CREATE UNIQUE INDEX IF NOT EXISTS ux_actors_tenant_normalized_name_active
    ON public.actors (tenant_id, normalized_name)
    WHERE deleted_at IS NULL
      AND merged_into_actor_id IS NULL;

COMMIT;
