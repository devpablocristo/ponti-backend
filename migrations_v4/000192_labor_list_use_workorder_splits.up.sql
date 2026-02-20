-- ========================================
-- MIGRATION 000192 LABOR_LIST USE WORKORDER SPLITS (UP)
-- ========================================
-- Actualiza la vista labor_list para expandir workorder_investor_splits
-- en filas separadas por inversor (mismo patrón que 000191).

BEGIN;

CREATE OR REPLACE VIEW v4_report.labor_list AS
WITH workorder_alloc AS (
  -- Sin splits: usar investor_id original con factor 1
  SELECT
    w.id AS workorder_id,
    w.investor_id,
    1::numeric AS factor
  FROM public.workorders w
  WHERE w.deleted_at IS NULL
    AND NOT EXISTS (
      SELECT 1
      FROM public.workorder_investor_splits wis
      WHERE wis.workorder_id = w.id
        AND wis.deleted_at IS NULL
    )
  UNION ALL
  -- Con splits: una fila por inversor con su porcentaje
  SELECT
    w.id AS workorder_id,
    wis.investor_id,
    (wis.percentage::numeric / 100)::numeric AS factor
  FROM public.workorders w
  JOIN public.workorder_investor_splits wis
    ON wis.workorder_id = w.id
   AND wis.deleted_at IS NULL
  WHERE w.deleted_at IS NULL
)
SELECT
  w.id AS workorder_id,
  w.number AS workorder_number,
  w.date,
  w.project_id,
  p.name AS project_name,
  w.field_id,
  f.name AS field_name,
  w.lot_id,
  l.name AS lot_name,
  w.crop_id,
  c.name AS crop_name,
  w.labor_id,
  lb.name AS labor_name,
  lb.category_id AS labor_category_id,
  cat.name AS labor_category_name,
  w.contractor,
  lb.contractor_name,
  -- Mantener el tipo original de la vista (w.effective_area es numeric(18,6)).
  -- Si cambiamos a numeric "genérico", Postgres no permite CREATE OR REPLACE por cambio de tipo.
  (w.effective_area * a.factor)::numeric(18,6) AS surface_ha,
  lb.price AS cost_per_ha,
  (lb.price * w.effective_area * a.factor)::numeric AS total_labor_cost,
  v4_core.dollar_average_for_month(w.project_id, w.date) AS dollar_average_month,
  lb.price::numeric AS usd_cost_ha,
  (lb.price * w.effective_area * a.factor)::numeric AS usd_net_total,
  a.investor_id,
  i.name AS investor_name
FROM public.workorders w
JOIN workorder_alloc a ON a.workorder_id = w.id
JOIN public.projects p ON p.id = w.project_id AND p.deleted_at IS NULL
JOIN public.fields f ON f.id = w.field_id AND f.deleted_at IS NULL
LEFT JOIN public.lots l ON l.id = w.lot_id AND l.deleted_at IS NULL
LEFT JOIN public.crops c ON c.id = w.crop_id AND c.deleted_at IS NULL
JOIN public.labors lb ON lb.id = w.labor_id AND lb.deleted_at IS NULL
LEFT JOIN public.categories cat ON cat.id = lb.category_id AND cat.deleted_at IS NULL
LEFT JOIN public.investors i ON i.id = a.investor_id AND i.deleted_at IS NULL
WHERE w.deleted_at IS NULL
  AND w.effective_area IS NOT NULL
  AND w.effective_area > 0
  AND lb.price IS NOT NULL;

COMMIT;
