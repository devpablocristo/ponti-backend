-- Multi-tenant golden master snapshot.
--
-- Usage:
--   1. Run this against the restored dev dump before changing read paths.
--   2. Save the output as the baseline.
--   3. Run it again after tenant hardening/backfill.
--   4. Diff the numeric columns. The accepted diff for this closure is 0.
--
-- The script is intentionally read-only and groups by tenant/project/campaign/field
-- so total equality cannot hide a cross-tenant or cross-project shift.

SELECT
    'project_surface' AS section,
    p.tenant_id,
    p.id AS project_id,
    p.customer_id,
    p.campaign_id,
    NULL::bigint AS field_id,
    NULL::bigint AS actor_legacy_id,
    COALESCE(SUM(l.hectares), 0)::numeric AS value_1,
    COUNT(DISTINCT f.id)::numeric AS value_2,
    COUNT(DISTINCT l.id)::numeric AS value_3
FROM public.projects p
LEFT JOIN public.fields f ON f.project_id = p.id AND f.tenant_id = p.tenant_id AND f.deleted_at IS NULL
LEFT JOIN public.lots l ON l.field_id = f.id AND l.tenant_id = p.tenant_id AND l.deleted_at IS NULL
WHERE p.deleted_at IS NULL
GROUP BY p.tenant_id, p.id, p.customer_id, p.campaign_id
ORDER BY section, tenant_id, project_id;

SELECT
    'investor_allocations' AS section,
    pi.tenant_id,
    pi.project_id,
    NULL::bigint AS customer_id,
    NULL::bigint AS campaign_id,
    NULL::bigint AS field_id,
    pi.investor_id AS actor_legacy_id,
    COALESCE(SUM(pi.percentage), 0)::numeric AS value_1,
    COUNT(*)::numeric AS value_2,
    0::numeric AS value_3
FROM public.project_investors pi
WHERE pi.deleted_at IS NULL
GROUP BY pi.tenant_id, pi.project_id, pi.investor_id
ORDER BY section, tenant_id, project_id, actor_legacy_id;

SELECT
    'admin_cost_allocations' AS section,
    aci.tenant_id,
    aci.project_id,
    NULL::bigint AS customer_id,
    NULL::bigint AS campaign_id,
    NULL::bigint AS field_id,
    aci.investor_id AS actor_legacy_id,
    COALESCE(SUM(aci.percentage), 0)::numeric AS value_1,
    COUNT(*)::numeric AS value_2,
    0::numeric AS value_3
FROM public.admin_cost_investors aci
WHERE aci.deleted_at IS NULL
GROUP BY aci.tenant_id, aci.project_id, aci.investor_id
ORDER BY section, tenant_id, project_id, actor_legacy_id;

SELECT
    'workorders' AS section,
    w.tenant_id,
    w.project_id,
    NULL::bigint AS customer_id,
    NULL::bigint AS campaign_id,
    NULL::bigint AS field_id,
    COALESCE(w.investor_id, 0) AS actor_legacy_id,
    COALESCE(SUM(w.effective_area), 0)::numeric AS value_1,
    COUNT(DISTINCT w.lot_id)::numeric AS value_2,
    COUNT(*)::numeric AS value_3
FROM public.workorders w
WHERE w.deleted_at IS NULL
GROUP BY w.tenant_id, w.project_id, COALESCE(w.investor_id, 0)
ORDER BY section, tenant_id, project_id, actor_legacy_id;

SELECT
    'supply_movements' AS section,
    sm.tenant_id,
    sm.project_id,
    NULL::bigint AS customer_id,
    NULL::bigint AS campaign_id,
    NULL::bigint AS field_id,
    COALESCE(sm.investor_id, 0) AS actor_legacy_id,
    COALESCE(SUM(sm.quantity), 0)::numeric AS value_1,
    COUNT(DISTINCT sm.supply_id)::numeric AS value_2,
    COUNT(*)::numeric AS value_3
FROM public.supply_movements sm
WHERE sm.deleted_at IS NULL
GROUP BY sm.tenant_id, sm.project_id, COALESCE(sm.investor_id, 0)
ORDER BY section, tenant_id, project_id, actor_legacy_id;

SELECT
    'stock' AS section,
    s.tenant_id,
    s.project_id,
    NULL::bigint AS customer_id,
    NULL::bigint AS campaign_id,
    NULL::bigint AS field_id,
    COALESCE(s.investor_id, 0) AS actor_legacy_id,
    COALESCE(SUM(s.real_stock_units), 0)::numeric AS value_1,
    COALESCE(SUM(s.units_entered - s.units_consumed), 0)::numeric AS value_2,
    COUNT(*)::numeric AS value_3
FROM public.stocks s
WHERE s.deleted_at IS NULL
GROUP BY s.tenant_id, s.project_id, COALESCE(s.investor_id, 0)
ORDER BY section, tenant_id, project_id, actor_legacy_id;

SELECT
    'invoices' AS section,
    i.tenant_id,
    w.project_id,
    NULL::bigint AS customer_id,
    NULL::bigint AS campaign_id,
    NULL::bigint AS field_id,
    COALESCE(w.investor_id, 0) AS actor_legacy_id,
    COUNT(*)::numeric AS value_1,
    COUNT(DISTINCT i.company)::numeric AS value_2,
    0::numeric AS value_3
FROM public.invoices i
JOIN public.workorders w ON w.id = i.work_order_id AND w.tenant_id = i.tenant_id
WHERE i.deleted_at IS NULL
GROUP BY i.tenant_id, w.project_id, COALESCE(w.investor_id, 0)
ORDER BY section, tenant_id, project_id, actor_legacy_id;

DROP TABLE IF EXISTS pg_temp.multi_tenant_golden_master_diffs;

CREATE TEMP TABLE multi_tenant_golden_master_diffs (
    check_name text NOT NULL,
    legacy_total numeric NOT NULL,
    tenant_scoped_total numeric NOT NULL,
    diff numeric NOT NULL
);

WITH legacy_project_surface AS (
    SELECT COALESCE(SUM(l.hectares), 0)::numeric AS value
    FROM public.projects p
    LEFT JOIN public.fields f ON f.project_id = p.id AND f.deleted_at IS NULL
    LEFT JOIN public.lots l ON l.field_id = f.id AND l.deleted_at IS NULL
    WHERE p.deleted_at IS NULL
),
tenant_scoped_project_surface AS (
    SELECT COALESCE(SUM(l.hectares), 0)::numeric AS value
    FROM public.projects p
    LEFT JOIN public.fields f ON f.project_id = p.id AND f.tenant_id = p.tenant_id AND f.deleted_at IS NULL
    LEFT JOIN public.lots l ON l.field_id = f.id AND l.tenant_id = p.tenant_id AND l.deleted_at IS NULL
    WHERE p.deleted_at IS NULL
),
legacy_invoices AS (
    SELECT COUNT(*)::numeric AS value
    FROM public.invoices i
    JOIN public.workorders w ON w.id = i.work_order_id
    WHERE i.deleted_at IS NULL
),
tenant_scoped_invoices AS (
    SELECT COUNT(*)::numeric AS value
    FROM public.invoices i
    JOIN public.workorders w ON w.id = i.work_order_id AND w.tenant_id = i.tenant_id
    WHERE i.deleted_at IS NULL
),
legacy_supply_movements AS (
    SELECT COALESCE(SUM(sm.quantity), 0)::numeric AS value
    FROM public.supply_movements sm
    JOIN public.projects p ON p.id = sm.project_id
    WHERE sm.deleted_at IS NULL
),
tenant_scoped_supply_movements AS (
    SELECT COALESCE(SUM(sm.quantity), 0)::numeric AS value
    FROM public.supply_movements sm
    JOIN public.projects p ON p.id = sm.project_id AND p.tenant_id = sm.tenant_id
    WHERE sm.deleted_at IS NULL
),
legacy_workorders AS (
    SELECT COALESCE(SUM(w.effective_area), 0)::numeric AS value
    FROM public.workorders w
    JOIN public.projects p ON p.id = w.project_id
    WHERE w.deleted_at IS NULL
),
tenant_scoped_workorders AS (
    SELECT COALESCE(SUM(w.effective_area), 0)::numeric AS value
    FROM public.workorders w
    JOIN public.projects p ON p.id = w.project_id AND p.tenant_id = w.tenant_id
    WHERE w.deleted_at IS NULL
)
INSERT INTO multi_tenant_golden_master_diffs(check_name, legacy_total, tenant_scoped_total, diff)
SELECT
    'project_surface.tenant_scoped_join',
    legacy_project_surface.value,
    tenant_scoped_project_surface.value,
    legacy_project_surface.value - tenant_scoped_project_surface.value
FROM legacy_project_surface, tenant_scoped_project_surface
UNION ALL
SELECT
    'invoices.tenant_scoped_join',
    legacy_invoices.value,
    tenant_scoped_invoices.value,
    legacy_invoices.value - tenant_scoped_invoices.value
FROM legacy_invoices, tenant_scoped_invoices
UNION ALL
SELECT
    'supply_movements.tenant_scoped_project_join',
    legacy_supply_movements.value,
    tenant_scoped_supply_movements.value,
    legacy_supply_movements.value - tenant_scoped_supply_movements.value
FROM legacy_supply_movements, tenant_scoped_supply_movements
UNION ALL
SELECT
    'workorders.tenant_scoped_project_join',
    legacy_workorders.value,
    tenant_scoped_workorders.value,
    legacy_workorders.value - tenant_scoped_workorders.value
FROM legacy_workorders, tenant_scoped_workorders;

SELECT *
FROM multi_tenant_golden_master_diffs
ORDER BY check_name;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM multi_tenant_golden_master_diffs
        WHERE diff <> 0
    ) THEN
        RAISE EXCEPTION 'multi_tenant_golden_master_diffs contains non-zero diffs';
    END IF;
END $$;
