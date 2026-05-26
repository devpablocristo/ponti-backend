-- archived_invariants.sql — read-only data audit for the hierarchical
-- integrity invariant: no child can be active under an archived parent.
--
-- Run against staging first, then production (read-only). Each query
-- reports a check name and the count + sample ids of rows that violate
-- the invariant. A zero row count means the invariant holds; any row > 0
-- needs a remediation decision (cascade-archive the affected children,
-- or restore the parent if it was archived by mistake).
--
-- Postgres-compatible. Schema assumes the standard `deleted_at` GORM
-- soft-delete column on every table listed below.

-- IA-1: active projects under archived customer
SELECT 'IA-1 projects_under_archived_customer' AS check_id,
       COUNT(*) AS rows,
       STRING_AGG(p.id::text, ',') AS sample_ids
FROM projects p
JOIN customers c ON p.customer_id = c.id
WHERE p.deleted_at IS NULL AND c.deleted_at IS NOT NULL;

-- IA-2: active fields under archived project
SELECT 'IA-2 fields_under_archived_project',
       COUNT(*),
       STRING_AGG(f.id::text, ',')
FROM fields f
JOIN projects p ON f.project_id = p.id
WHERE f.deleted_at IS NULL AND p.deleted_at IS NOT NULL;

-- IA-3: active lots under archived field
SELECT 'IA-3 lots_under_archived_field',
       COUNT(*),
       STRING_AGG(l.id::text, ',')
FROM lots l
JOIN fields f ON l.field_id = f.id
WHERE l.deleted_at IS NULL AND f.deleted_at IS NOT NULL;

-- IA-4: active work orders under any archived parent (project/field/lot)
SELECT 'IA-4 workorders_under_archived_parent',
       COUNT(*),
       STRING_AGG(wo.id::text, ',')
FROM workorders wo
LEFT JOIN projects p ON wo.project_id = p.id
LEFT JOIN fields f ON wo.field_id = f.id
LEFT JOIN lots l ON wo.lot_id = l.id
WHERE wo.deleted_at IS NULL
  AND (p.deleted_at IS NOT NULL
       OR f.deleted_at IS NOT NULL
       OR l.deleted_at IS NOT NULL);

-- IA-5: active labors under archived project
SELECT 'IA-5 labors_under_archived_project',
       COUNT(*),
       STRING_AGG(l.id::text, ',')
FROM labors l
JOIN projects p ON l.project_id = p.id
WHERE l.deleted_at IS NULL AND p.deleted_at IS NOT NULL;

-- IA-6: active supplies under archived project
SELECT 'IA-6 supplies_under_archived_project',
       COUNT(*),
       STRING_AGG(s.id::text, ',')
FROM supplies s
JOIN projects p ON s.project_id = p.id
WHERE s.deleted_at IS NULL AND p.deleted_at IS NOT NULL;

-- IA-7: active supply movements under archived parent (project or supply)
SELECT 'IA-7 movements_under_archived_parent',
       COUNT(*),
       STRING_AGG(sm.id::text, ',')
FROM supply_movements sm
LEFT JOIN projects p ON sm.project_id = p.id
LEFT JOIN supplies s ON sm.supply_id = s.id
WHERE sm.deleted_at IS NULL
  AND (p.deleted_at IS NOT NULL OR s.deleted_at IS NOT NULL);

-- IA-8: active stocks under archived parent (project or supply)
SELECT 'IA-8 stocks_under_archived_parent',
       COUNT(*),
       STRING_AGG(st.id::text, ',')
FROM stocks st
LEFT JOIN projects p ON st.project_id = p.id
LEFT JOIN supplies s ON st.supply_id = s.id
WHERE st.deleted_at IS NULL
  AND (p.deleted_at IS NOT NULL OR s.deleted_at IS NOT NULL);

-- IA-9: active workorder_items under archived workorder
SELECT 'IA-9 wo_items_under_archived_workorder',
       COUNT(*),
       STRING_AGG(woi.id::text, ',')
FROM workorder_items woi
JOIN workorders wo ON woi.workorder_id = wo.id
WHERE woi.deleted_at IS NULL AND wo.deleted_at IS NOT NULL;

-- IA-10: active workorder_investor_splits under archived workorder
SELECT 'IA-10 wo_splits_under_archived_workorder',
       COUNT(*),
       STRING_AGG(wis.id::text, ',')
FROM workorder_investor_splits wis
JOIN workorders wo ON wis.workorder_id = wo.id
WHERE wis.deleted_at IS NULL AND wo.deleted_at IS NOT NULL;

-- IA-11a: active project_managers with archived manager
SELECT 'IA-11a project_managers_archived_ref',
       COUNT(*),
       STRING_AGG(pm.manager_id::text, ',')
FROM project_managers pm
JOIN managers m ON pm.manager_id = m.id
WHERE pm.deleted_at IS NULL AND m.deleted_at IS NOT NULL;

-- IA-11b: active project_investors with archived investor
SELECT 'IA-11b project_investors_archived_ref',
       COUNT(*),
       STRING_AGG(pi.investor_id::text, ',')
FROM project_investors pi
JOIN investors i ON pi.investor_id = i.id
WHERE pi.deleted_at IS NULL AND i.deleted_at IS NOT NULL;

-- IA-11c: active field_investors with archived investor
SELECT 'IA-11c field_investors_archived_ref',
       COUNT(*),
       STRING_AGG(fi.field_id::text || ':' || fi.investor_id::text, ',')
FROM field_investors fi
JOIN investors i ON fi.investor_id = i.id
WHERE fi.deleted_at IS NULL AND i.deleted_at IS NOT NULL;

-- IA-11d: active workorder_investor_splits with archived investor
SELECT 'IA-11d wo_splits_archived_investor',
       COUNT(*),
       STRING_AGG(wis.id::text, ',')
FROM workorder_investor_splits wis
JOIN investors i ON wis.investor_id = i.id
WHERE wis.deleted_at IS NULL AND i.deleted_at IS NOT NULL;

-- IA-12: active customers.actor_id pointing to archived actor (the bug
-- originally reported — customer kept referencing an archived actor)
SELECT 'IA-12 customers_with_archived_actor',
       COUNT(*),
       STRING_AGG(c.id::text, ',')
FROM customers c
JOIN actors a ON c.actor_id = a.id
WHERE c.deleted_at IS NULL AND a.deleted_at IS NOT NULL;

-- IA-13: active legacy_actor_map entries pointing to archived actor
SELECT 'IA-13 legacy_map_archived_actor',
       COUNT(*),
       STRING_AGG(m.source_table || ':' || m.source_id::text, ',')
FROM legacy_actor_map m
JOIN actors a ON m.actor_id = a.id
WHERE a.deleted_at IS NOT NULL;

-- IA-14: rows archived without trace metadata (pre-feature data).
-- These can not be restored selectively because Cause is unknown — they
-- are not strictly broken but require manual intervention if they ever
-- need to be restored.
SELECT 'IA-14 untraceable_archives' AS check_id,
       tbl,
       untraceable
FROM (
    SELECT 'projects' AS tbl, COUNT(*) AS untraceable
      FROM projects WHERE deleted_at IS NOT NULL AND archive_batch_id IS NULL
    UNION ALL
    SELECT 'fields',   COUNT(*) FROM fields    WHERE deleted_at IS NOT NULL AND archive_batch_id IS NULL
    UNION ALL
    SELECT 'lots',     COUNT(*) FROM lots      WHERE deleted_at IS NOT NULL AND archive_batch_id IS NULL
    UNION ALL
    SELECT 'workorders', COUNT(*) FROM workorders WHERE deleted_at IS NOT NULL AND archive_batch_id IS NULL
) t
WHERE untraceable > 0;
