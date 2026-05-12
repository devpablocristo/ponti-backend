-- Golden master inicial para la migracion de Actores.
-- Debe ejecutarse despues de aplicar 000223 y antes de activar actors_live.
-- Criterio: todas las filas deben devolver diff = 0.

\set ON_ERROR_STOP on

DROP TABLE IF EXISTS pg_temp.actor_golden_master_results;

CREATE TEMP TABLE actor_golden_master_results AS
WITH
project_investors_legacy AS (
  SELECT pi.project_id, pi.investor_id, m.actor_id, SUM(pi.percentage)::numeric AS value
  FROM project_investors pi
  JOIN legacy_actor_map m ON m.source_table = 'investors' AND m.source_id = pi.investor_id
  WHERE pi.deleted_at IS NULL
  GROUP BY pi.project_id, pi.investor_id, m.actor_id
),
project_investors_actor AS (
  SELECT project_id, actor_id, SUM(percentage)::numeric AS value
  FROM project_investor_allocations
  WHERE deleted_at IS NULL
  GROUP BY project_id, actor_id
),
admin_cost_legacy AS (
  SELECT aci.project_id, aci.investor_id, m.actor_id, SUM(aci.percentage)::numeric AS value
  FROM admin_cost_investors aci
  JOIN legacy_actor_map m ON m.source_table = 'investors' AND m.source_id = aci.investor_id
  WHERE aci.deleted_at IS NULL
  GROUP BY aci.project_id, aci.investor_id, m.actor_id
),
admin_cost_actor AS (
  SELECT project_id, actor_id, SUM(percentage)::numeric AS value
  FROM project_admin_cost_allocations
  WHERE deleted_at IS NULL
  GROUP BY project_id, actor_id
),
field_investors_legacy AS (
  SELECT f.project_id, fi.field_id, fi.investor_id, m.actor_id, SUM(fi.percentage)::numeric AS value
  FROM field_investors fi
  JOIN fields f ON f.id = fi.field_id
  JOIN legacy_actor_map m ON m.source_table = 'investors' AND m.source_id = fi.investor_id
  WHERE fi.deleted_at IS NULL
  GROUP BY f.project_id, fi.field_id, fi.investor_id, m.actor_id
),
field_investors_actor AS (
  SELECT f.project_id, flp.field_id, flp.actor_id, SUM(flp.percentage)::numeric AS value
  FROM field_lease_participants flp
  JOIN fields f ON f.id = flp.field_id
  WHERE flp.deleted_at IS NULL
  GROUP BY f.project_id, flp.field_id, flp.actor_id
),
workorders_legacy AS (
  SELECT w.project_id, w.field_id, w.crop_id, w.investor_id, m.actor_id,
         SUM(w.effective_area)::numeric AS value
  FROM workorders w
  JOIN legacy_actor_map m ON m.source_table = 'investors' AND m.source_id = w.investor_id
  WHERE w.deleted_at IS NULL
  GROUP BY w.project_id, w.field_id, w.crop_id, w.investor_id, m.actor_id
),
workorders_actor AS (
  SELECT project_id, field_id, crop_id, investor_actor_id AS actor_id,
         SUM(effective_area)::numeric AS value
  FROM workorders
  WHERE deleted_at IS NULL
  GROUP BY project_id, field_id, crop_id, investor_actor_id
),
supply_movements_investor_legacy AS (
  SELECT sm.project_id, sm.supply_id, sm.investor_id, m.actor_id,
         SUM(sm.quantity)::numeric AS value
  FROM supply_movements sm
  JOIN legacy_actor_map m ON m.source_table = 'investors' AND m.source_id = sm.investor_id
  WHERE sm.deleted_at IS NULL
  GROUP BY sm.project_id, sm.supply_id, sm.investor_id, m.actor_id
),
supply_movements_investor_actor AS (
  SELECT project_id, supply_id, investor_actor_id AS actor_id,
         SUM(quantity)::numeric AS value
  FROM supply_movements
  WHERE deleted_at IS NULL
  GROUP BY project_id, supply_id, investor_actor_id
),
supply_movements_provider_legacy AS (
  SELECT sm.project_id, sm.supply_id, sm.provider_id, m.actor_id,
         SUM(sm.quantity)::numeric AS value
  FROM supply_movements sm
  JOIN legacy_actor_map m ON m.source_table = 'providers' AND m.source_id = sm.provider_id
  WHERE sm.deleted_at IS NULL
  GROUP BY sm.project_id, sm.supply_id, sm.provider_id, m.actor_id
),
supply_movements_provider_actor AS (
  SELECT project_id, supply_id, provider_actor_id AS actor_id,
         SUM(quantity)::numeric AS value
  FROM supply_movements
  WHERE deleted_at IS NULL
  GROUP BY project_id, supply_id, provider_actor_id
),
stocks_legacy AS (
  SELECT s.project_id, s.supply_id, s.investor_id, m.actor_id,
         SUM(s.initial_units + s.units_entered - s.units_consumed)::numeric AS value
  FROM stocks s
  JOIN legacy_actor_map m ON m.source_table = 'investors' AND m.source_id = s.investor_id
  WHERE s.deleted_at IS NULL
  GROUP BY s.project_id, s.supply_id, s.investor_id, m.actor_id
),
stocks_actor AS (
  SELECT project_id, supply_id, investor_actor_id AS actor_id,
         SUM(initial_units + units_entered - units_consumed)::numeric AS value
  FROM stocks
  WHERE deleted_at IS NULL
  GROUP BY project_id, supply_id, investor_actor_id
),
invoices_legacy AS (
  SELECT i.investor_id, m.actor_id, COUNT(*)::numeric AS value
  FROM invoices i
  JOIN legacy_actor_map m ON m.source_table = 'investors' AND m.source_id = i.investor_id
  WHERE i.deleted_at IS NULL
  GROUP BY i.investor_id, m.actor_id
),
invoices_actor AS (
  SELECT investor_actor_id AS actor_id, COUNT(*)::numeric AS value
  FROM invoices
  WHERE deleted_at IS NULL
  GROUP BY investor_actor_id
)
SELECT 'project_investors.percentage' AS check_name,
       COALESCE(SUM(l.value), 0) AS legacy_total,
       COALESCE(SUM(a.value), 0) AS actor_total,
       COALESCE(SUM(l.value), 0) - COALESCE(SUM(a.value), 0) AS diff
FROM project_investors_legacy l
FULL JOIN project_investors_actor a ON a.project_id = l.project_id AND a.actor_id = l.actor_id
UNION ALL
SELECT 'admin_cost_investors.percentage', COALESCE(SUM(l.value), 0), COALESCE(SUM(a.value), 0), COALESCE(SUM(l.value), 0) - COALESCE(SUM(a.value), 0)
FROM admin_cost_legacy l
FULL JOIN admin_cost_actor a ON a.project_id = l.project_id AND a.actor_id = l.actor_id
UNION ALL
SELECT 'field_investors.percentage', COALESCE(SUM(l.value), 0), COALESCE(SUM(a.value), 0), COALESCE(SUM(l.value), 0) - COALESCE(SUM(a.value), 0)
FROM field_investors_legacy l
FULL JOIN field_investors_actor a ON a.project_id = l.project_id AND a.field_id = l.field_id AND a.actor_id = l.actor_id
UNION ALL
SELECT 'workorders.effective_area', COALESCE(SUM(l.value), 0), COALESCE(SUM(a.value), 0), COALESCE(SUM(l.value), 0) - COALESCE(SUM(a.value), 0)
FROM workorders_legacy l
FULL JOIN workorders_actor a ON a.project_id = l.project_id AND a.field_id = l.field_id AND a.crop_id = l.crop_id AND a.actor_id = l.actor_id
UNION ALL
SELECT 'supply_movements.quantity_by_investor', COALESCE(SUM(l.value), 0), COALESCE(SUM(a.value), 0), COALESCE(SUM(l.value), 0) - COALESCE(SUM(a.value), 0)
FROM supply_movements_investor_legacy l
FULL JOIN supply_movements_investor_actor a ON a.project_id = l.project_id AND a.supply_id = l.supply_id AND a.actor_id = l.actor_id
UNION ALL
SELECT 'supply_movements.quantity_by_provider', COALESCE(SUM(l.value), 0), COALESCE(SUM(a.value), 0), COALESCE(SUM(l.value), 0) - COALESCE(SUM(a.value), 0)
FROM supply_movements_provider_legacy l
FULL JOIN supply_movements_provider_actor a ON a.project_id = l.project_id AND a.supply_id = l.supply_id AND a.actor_id = l.actor_id
UNION ALL
SELECT 'stocks.current_units_by_investor', COALESCE(SUM(l.value), 0), COALESCE(SUM(a.value), 0), COALESCE(SUM(l.value), 0) - COALESCE(SUM(a.value), 0)
FROM stocks_legacy l
FULL JOIN stocks_actor a ON a.project_id = l.project_id AND a.supply_id = l.supply_id AND a.actor_id = l.actor_id
UNION ALL
SELECT 'invoices.count_by_investor', COALESCE(SUM(l.value), 0), COALESCE(SUM(a.value), 0), COALESCE(SUM(l.value), 0) - COALESCE(SUM(a.value), 0)
FROM invoices_legacy l
FULL JOIN invoices_actor a ON a.actor_id = l.actor_id
ORDER BY check_name;

INSERT INTO actor_golden_master_results(check_name, legacy_total, actor_total, diff)
SELECT 'coverage.customers.map',
       COUNT(c.id)::numeric,
       COUNT(m.actor_id)::numeric,
       COUNT(c.id)::numeric - COUNT(m.actor_id)::numeric
FROM customers c
LEFT JOIN legacy_actor_map m ON m.source_table = 'customers' AND m.source_id = c.id
WHERE c.deleted_at IS NULL
UNION ALL
SELECT 'coverage.investors.map', COUNT(i.id)::numeric, COUNT(m.actor_id)::numeric, COUNT(i.id)::numeric - COUNT(m.actor_id)::numeric
FROM investors i
LEFT JOIN legacy_actor_map m ON m.source_table = 'investors' AND m.source_id = i.id
WHERE i.deleted_at IS NULL
UNION ALL
SELECT 'coverage.managers.map', COUNT(mg.id)::numeric, COUNT(m.actor_id)::numeric, COUNT(mg.id)::numeric - COUNT(m.actor_id)::numeric
FROM managers mg
LEFT JOIN legacy_actor_map m ON m.source_table = 'managers' AND m.source_id = mg.id
WHERE mg.deleted_at IS NULL
UNION ALL
SELECT 'coverage.providers.map', COUNT(p.id)::numeric, COUNT(m.actor_id)::numeric, COUNT(p.id)::numeric - COUNT(m.actor_id)::numeric
FROM providers p
LEFT JOIN legacy_actor_map m ON m.source_table = 'providers' AND m.source_id = p.id
WHERE p.deleted_at IS NULL
UNION ALL
SELECT 'coverage.customers.role_cliente',
       COUNT(DISTINCT m.actor_id)::numeric,
       COUNT(DISTINCT ar.actor_id)::numeric,
       COUNT(DISTINCT m.actor_id)::numeric - COUNT(DISTINCT ar.actor_id)::numeric
FROM legacy_actor_map m
JOIN customers c ON c.id = m.source_id AND c.deleted_at IS NULL
LEFT JOIN actor_roles ar ON ar.actor_id = m.actor_id AND ar.role = 'cliente' AND ar.archived_at IS NULL
WHERE m.source_table = 'customers'
UNION ALL
SELECT 'coverage.investors.role_inversor', COUNT(DISTINCT m.actor_id)::numeric, COUNT(DISTINCT ar.actor_id)::numeric, COUNT(DISTINCT m.actor_id)::numeric - COUNT(DISTINCT ar.actor_id)::numeric
FROM legacy_actor_map m
JOIN investors i ON i.id = m.source_id AND i.deleted_at IS NULL
LEFT JOIN actor_roles ar ON ar.actor_id = m.actor_id AND ar.role = 'inversor' AND ar.archived_at IS NULL
WHERE m.source_table = 'investors'
UNION ALL
SELECT 'coverage.managers.role_responsable', COUNT(DISTINCT m.actor_id)::numeric, COUNT(DISTINCT ar.actor_id)::numeric, COUNT(DISTINCT m.actor_id)::numeric - COUNT(DISTINCT ar.actor_id)::numeric
FROM legacy_actor_map m
JOIN managers mg ON mg.id = m.source_id AND mg.deleted_at IS NULL
LEFT JOIN actor_roles ar ON ar.actor_id = m.actor_id AND ar.role = 'responsable' AND ar.archived_at IS NULL
WHERE m.source_table = 'managers'
UNION ALL
SELECT 'coverage.providers.role_proveedor', COUNT(DISTINCT m.actor_id)::numeric, COUNT(DISTINCT ar.actor_id)::numeric, COUNT(DISTINCT m.actor_id)::numeric - COUNT(DISTINCT ar.actor_id)::numeric
FROM legacy_actor_map m
JOIN providers p ON p.id = m.source_id AND p.deleted_at IS NULL
LEFT JOIN actor_roles ar ON ar.actor_id = m.actor_id AND ar.role = 'proveedor' AND ar.archived_at IS NULL
WHERE m.source_table = 'providers'
UNION ALL
SELECT 'coverage.projects.customer_actor_id',
       COUNT(*) FILTER (WHERE customer_id IS NOT NULL)::numeric,
       COUNT(*) FILTER (WHERE customer_actor_id IS NOT NULL)::numeric,
       COUNT(*) FILTER (WHERE customer_id IS NOT NULL)::numeric - COUNT(*) FILTER (WHERE customer_actor_id IS NOT NULL)::numeric
FROM projects
WHERE deleted_at IS NULL
UNION ALL
SELECT 'coverage.workorders.investor_actor_id',
       COUNT(*) FILTER (WHERE investor_id IS NOT NULL)::numeric,
       COUNT(*) FILTER (WHERE investor_actor_id IS NOT NULL)::numeric,
       COUNT(*) FILTER (WHERE investor_id IS NOT NULL)::numeric - COUNT(*) FILTER (WHERE investor_actor_id IS NOT NULL)::numeric
FROM workorders
WHERE deleted_at IS NULL
UNION ALL
SELECT 'coverage.workorder_splits.actor_id',
       COUNT(*) FILTER (WHERE investor_id IS NOT NULL)::numeric,
       COUNT(*) FILTER (WHERE actor_id IS NOT NULL)::numeric,
       COUNT(*) FILTER (WHERE investor_id IS NOT NULL)::numeric - COUNT(*) FILTER (WHERE actor_id IS NOT NULL)::numeric
FROM workorder_investor_splits
WHERE deleted_at IS NULL
UNION ALL
SELECT 'coverage.stocks.investor_actor_id',
       COUNT(*) FILTER (WHERE investor_id IS NOT NULL)::numeric,
       COUNT(*) FILTER (WHERE investor_actor_id IS NOT NULL)::numeric,
       COUNT(*) FILTER (WHERE investor_id IS NOT NULL)::numeric - COUNT(*) FILTER (WHERE investor_actor_id IS NOT NULL)::numeric
FROM stocks
WHERE deleted_at IS NULL
UNION ALL
SELECT 'coverage.supply_movements.investor_actor_id',
       COUNT(*) FILTER (WHERE investor_id IS NOT NULL)::numeric,
       COUNT(*) FILTER (WHERE investor_actor_id IS NOT NULL)::numeric,
       COUNT(*) FILTER (WHERE investor_id IS NOT NULL)::numeric - COUNT(*) FILTER (WHERE investor_actor_id IS NOT NULL)::numeric
FROM supply_movements
WHERE deleted_at IS NULL
UNION ALL
SELECT 'coverage.supply_movements.provider_actor_id',
       COUNT(*) FILTER (WHERE provider_id IS NOT NULL)::numeric,
       COUNT(*) FILTER (WHERE provider_actor_id IS NOT NULL)::numeric,
       COUNT(*) FILTER (WHERE provider_id IS NOT NULL)::numeric - COUNT(*) FILTER (WHERE provider_actor_id IS NOT NULL)::numeric
FROM supply_movements
WHERE deleted_at IS NULL
UNION ALL
SELECT 'coverage.invoices.investor_actor_id',
       COUNT(*) FILTER (WHERE investor_id IS NOT NULL)::numeric,
       COUNT(*) FILTER (WHERE investor_actor_id IS NOT NULL)::numeric,
       COUNT(*) FILTER (WHERE investor_id IS NOT NULL)::numeric - COUNT(*) FILTER (WHERE investor_actor_id IS NOT NULL)::numeric
FROM invoices
WHERE deleted_at IS NULL;

INSERT INTO actor_golden_master_results(check_name, legacy_total, actor_total, diff)
SELECT 'mirror.project_investor_allocations.orphans', 0::numeric, COUNT(*)::numeric, COUNT(*)::numeric
FROM project_investor_allocations pia
WHERE pia.deleted_at IS NULL
  AND NOT EXISTS (
    SELECT 1
    FROM project_investors pi
    JOIN legacy_actor_map m ON m.source_table = 'investors' AND m.source_id = pi.investor_id
    WHERE pi.deleted_at IS NULL
      AND pi.project_id = pia.project_id
      AND m.actor_id = pia.actor_id
  )
UNION ALL
SELECT 'mirror.project_admin_cost_allocations.orphans', 0::numeric, COUNT(*)::numeric, COUNT(*)::numeric
FROM project_admin_cost_allocations paca
WHERE paca.deleted_at IS NULL
  AND NOT EXISTS (
    SELECT 1
    FROM admin_cost_investors aci
    JOIN legacy_actor_map m ON m.source_table = 'investors' AND m.source_id = aci.investor_id
    WHERE aci.deleted_at IS NULL
      AND aci.project_id = paca.project_id
      AND m.actor_id = paca.actor_id
  )
UNION ALL
SELECT 'mirror.project_responsibles.orphans', 0::numeric, COUNT(*)::numeric, COUNT(*)::numeric
FROM project_responsibles pr
WHERE pr.deleted_at IS NULL
  AND NOT EXISTS (
    SELECT 1
    FROM project_managers pm
    JOIN legacy_actor_map m ON m.source_table = 'managers' AND m.source_id = pm.manager_id
    WHERE pm.deleted_at IS NULL
      AND pm.project_id = pr.project_id
      AND m.actor_id = pr.actor_id
  )
UNION ALL
SELECT 'mirror.field_lease_participants.orphans', 0::numeric, COUNT(*)::numeric, COUNT(*)::numeric
FROM field_lease_participants flp
WHERE flp.deleted_at IS NULL
  AND NOT EXISTS (
    SELECT 1
    FROM field_investors fi
    JOIN legacy_actor_map m ON m.source_table = 'investors' AND m.source_id = fi.investor_id
    WHERE fi.deleted_at IS NULL
      AND fi.field_id = flp.field_id
      AND m.actor_id = flp.actor_id
  );

SELECT *
FROM actor_golden_master_results
ORDER BY check_name;

DO $$
BEGIN
  IF EXISTS (
    SELECT 1
    FROM actor_golden_master_results
    WHERE COALESCE(diff, 0) <> 0
  ) THEN
    RAISE EXCEPTION 'actors golden master failed: one or more checks have non-zero diff';
  END IF;
END $$;
