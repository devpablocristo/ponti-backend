-- ========================================
-- MIGRATION 000215 WORKORDER LEGACY TRANSITION (DOWN)
-- ========================================

BEGIN;

CREATE OR REPLACE FUNCTION v4_ssot.first_workorder_number_for_project(p_project_id bigint)
RETURNS text
LANGUAGE sql STABLE AS $$
  SELECT number::text
  FROM public.workorders
  WHERE project_id = p_project_id
    AND deleted_at IS NULL
  ORDER BY date ASC, id ASC
  LIMIT 1
$$;

CREATE OR REPLACE FUNCTION v4_ssot.last_workorder_number_for_project(p_project_id bigint)
RETURNS text
LANGUAGE sql STABLE AS $$
  SELECT number::text
  FROM public.workorders
  WHERE project_id = p_project_id
    AND deleted_at IS NULL
  ORDER BY date DESC, id DESC
  LIMIT 1
$$;

CREATE OR REPLACE VIEW v4_report.dashboard_operational_indicators AS
SELECT
  p.id AS project_id,
  v4_ssot.first_workorder_date_for_project(p.id) AS start_date,
  v4_ssot.last_workorder_date_for_project(p.id) AS end_date,
  v4_core.calculate_campaign_closing_date(
    v4_ssot.last_workorder_date_for_project(p.id)
  ) AS campaign_closing_date,
  v4_ssot.first_workorder_number_for_project(p.id) AS first_workorder_id,
  v4_ssot.last_workorder_number_for_project(p.id) AS last_workorder_id,
  v4_ssot.last_stock_count_date_for_project(p.id) AS last_stock_count_date
FROM public.projects p
WHERE p.deleted_at IS NULL;

CREATE OR REPLACE VIEW v4_report.dashboard_operational_indicators_field AS
SELECT
  p.id AS project_id,
  p.customer_id,
  p.campaign_id,
  f.id AS field_id,
  (SELECT w2.date
   FROM public.workorders w2
   WHERE w2.field_id = f.id
     AND w2.deleted_at IS NULL
   ORDER BY w2.date ASC, w2.id ASC
   LIMIT 1) AS start_date,
  (SELECT w2.date
   FROM public.workorders w2
   WHERE w2.field_id = f.id
     AND w2.deleted_at IS NULL
   ORDER BY w2.date DESC, w2.id DESC
   LIMIT 1) AS end_date,
  v4_core.calculate_campaign_closing_date(
    (SELECT w2.date
     FROM public.workorders w2
     WHERE w2.field_id = f.id
       AND w2.deleted_at IS NULL
     ORDER BY w2.date DESC, w2.id DESC
     LIMIT 1)
  ) AS campaign_closing_date,
  (SELECT w2.number::text
   FROM public.workorders w2
   WHERE w2.field_id = f.id
     AND w2.deleted_at IS NULL
   ORDER BY w2.date ASC, w2.id ASC
   LIMIT 1) AS first_workorder_id,
  (SELECT w2.number::text
   FROM public.workorders w2
   WHERE w2.field_id = f.id
     AND w2.deleted_at IS NULL
   ORDER BY w2.date DESC, w2.id DESC
   LIMIT 1) AS last_workorder_id,
  v4_ssot.last_stock_count_date_for_project(p.id) AS last_stock_count_date
FROM public.projects p
JOIN public.fields f ON f.project_id = p.id AND f.deleted_at IS NULL
WHERE p.deleted_at IS NULL;

CREATE OR REPLACE VIEW v4_report.labor_list AS
WITH workorder_alloc AS (
  SELECT
    w.id AS workorder_id,
    w.investor_id,
    1::numeric AS factor,
    NULL::text AS investor_payment_status,
    false AS investor_payment_enabled
  FROM public.workorders w
  WHERE w.deleted_at IS NULL
    AND NOT EXISTS (
      SELECT 1
      FROM public.workorder_investor_splits wis
      WHERE wis.workorder_id = w.id
        AND wis.deleted_at IS NULL
    )
  UNION ALL
  SELECT
    w.id AS workorder_id,
    wis.investor_id,
    (wis.percentage::numeric / 100)::numeric AS factor,
    wis.payment_status AS investor_payment_status,
    true AS investor_payment_enabled
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
  (w.effective_area * a.factor)::numeric(18,6) AS surface_ha,
  lb.price AS cost_per_ha,
  (lb.price * w.effective_area * a.factor)::numeric AS total_labor_cost,
  v4_core.dollar_average_for_month(w.project_id, w.date) AS dollar_average_month,
  lb.price::numeric AS usd_cost_ha,
  (lb.price * w.effective_area * a.factor)::numeric AS usd_net_total,
  a.investor_id,
  i.name AS investor_name,
  a.investor_payment_status,
  a.investor_payment_enabled
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

CREATE OR REPLACE VIEW v4_report.workorder_list AS
SELECT
  w.id,
  w.number,
  w.project_id,
  w.field_id,
  p.name AS project_name,
  f.name AS field_name,
  l.name AS lot_name,
  w.date,
  c.name AS crop_name,
  lb.name AS labor_name,
  cat_lb.name AS labor_category_name,
  t.name AS type_name,
  w.contractor,
  w.effective_area AS surface_ha,
  s.name AS supply_name,
  wi.total_used AS consumption,
  cat.name AS category_name,
  wi.final_dose AS dose_per_ha,
  s.price AS unit_price,
  CASE
    WHEN wi.final_dose IS NOT NULL AND s.price IS NOT NULL
    THEN v4_core.cost_per_ha((wi.final_dose::numeric * s.price)::numeric, 1::numeric)
    ELSE 0
  END AS supply_cost_per_ha,
  v4_core.supply_cost(
    wi.final_dose::numeric,
    s.price::numeric,
    w.effective_area::numeric
  ) AS supply_total_cost,
  false::boolean AS is_digital,
  'published'::varchar(30) AS status
FROM public.workorders w
JOIN public.projects p ON p.id = w.project_id AND p.deleted_at IS NULL
JOIN public.fields f ON f.id = w.field_id AND f.deleted_at IS NULL
JOIN public.lots l ON l.id = w.lot_id AND l.deleted_at IS NULL
JOIN public.crops c ON c.id = w.crop_id AND c.deleted_at IS NULL
JOIN public.labors lb ON lb.id = w.labor_id
LEFT JOIN public.categories cat_lb ON cat_lb.id = lb.category_id AND cat_lb.deleted_at IS NULL
LEFT JOIN public.workorder_items wi ON wi.workorder_id = w.id AND wi.deleted_at IS NULL
LEFT JOIN public.supplies s ON s.id = wi.supply_id
LEFT JOIN public.types t ON t.id = s.type_id AND t.deleted_at IS NULL
LEFT JOIN public.categories cat ON cat.id = s.category_id AND cat.deleted_at IS NULL
WHERE w.deleted_at IS NULL

UNION ALL

SELECT
  w.id,
  w.number,
  w.project_id,
  w.field_id,
  p.name AS project_name,
  f.name AS field_name,
  l.name AS lot_name,
  w.date,
  c.name AS crop_name,
  lb.name AS labor_name,
  cat_lb.name AS labor_category_name,
  'Labor'::varchar(250) AS type_name,
  w.contractor,
  w.effective_area AS surface_ha,
  lb.name::varchar(100) AS supply_name,
  0::numeric(18,6) AS consumption,
  cat_lb.name::varchar(250) AS category_name,
  0::numeric(18,6) AS dose_per_ha,
  lb.price::numeric(18,6) AS unit_price,
  lb.price::numeric AS supply_cost_per_ha,
  (lb.price * w.effective_area)::numeric AS supply_total_cost,
  false::boolean AS is_digital,
  'published'::varchar(30) AS status
FROM public.workorders w
JOIN public.projects p ON p.id = w.project_id AND p.deleted_at IS NULL
JOIN public.fields f ON f.id = w.field_id AND f.deleted_at IS NULL
JOIN public.lots l ON l.id = w.lot_id AND l.deleted_at IS NULL
JOIN public.crops c ON c.id = w.crop_id AND c.deleted_at IS NULL
JOIN public.labors lb ON lb.id = w.labor_id
LEFT JOIN public.categories cat_lb ON cat_lb.id = lb.category_id AND cat_lb.deleted_at IS NULL
WHERE w.deleted_at IS NULL
  AND w.effective_area IS NOT NULL
  AND w.effective_area > 0
  AND lb.price IS NOT NULL

UNION ALL

SELECT
  -wod.id AS id,
  wod.number,
  wod.project_id,
  wod.field_id,
  p.name AS project_name,
  f.name AS field_name,
  l.name AS lot_name,
  wod.date,
  c.name AS crop_name,
  lb.name AS labor_name,
  cat_lb.name AS labor_category_name,
  t.name AS type_name,
  wod.contractor,
  wod.effective_area AS surface_ha,
  s.name AS supply_name,
  wodi.total_used AS consumption,
  cat.name AS category_name,
  wodi.final_dose AS dose_per_ha,
  s.price AS unit_price,
  CASE
    WHEN wodi.final_dose IS NOT NULL AND s.price IS NOT NULL
    THEN v4_core.cost_per_ha((wodi.final_dose::numeric * s.price)::numeric, 1::numeric)
    ELSE 0
  END AS supply_cost_per_ha,
  v4_core.supply_cost(
    wodi.final_dose::numeric,
    s.price::numeric,
    wod.effective_area::numeric
  ) AS supply_total_cost,
  true::boolean AS is_digital,
  wod.status::varchar(30) AS status
FROM public.work_order_drafts wod
JOIN public.projects p ON p.id = wod.project_id AND p.deleted_at IS NULL
JOIN public.fields f ON f.id = wod.field_id AND f.deleted_at IS NULL
JOIN public.lots l ON l.id = wod.lot_id AND l.deleted_at IS NULL
JOIN public.crops c ON c.id = wod.crop_id AND c.deleted_at IS NULL
JOIN public.labors lb ON lb.id = wod.labor_id
LEFT JOIN public.categories cat_lb ON cat_lb.id = lb.category_id AND cat_lb.deleted_at IS NULL
LEFT JOIN public.work_order_draft_items wodi ON wodi.draft_id = wod.id
LEFT JOIN public.supplies s ON s.id = wodi.supply_id
LEFT JOIN public.types t ON t.id = s.type_id AND t.deleted_at IS NULL
LEFT JOIN public.categories cat ON cat.id = s.category_id AND cat.deleted_at IS NULL
WHERE wod.deleted_at IS NULL
  AND wod.is_digital = true
  AND wod.status = 'draft'

UNION ALL

SELECT
  -wod.id AS id,
  wod.number,
  wod.project_id,
  wod.field_id,
  p.name AS project_name,
  f.name AS field_name,
  l.name AS lot_name,
  wod.date,
  c.name AS crop_name,
  lb.name AS labor_name,
  cat_lb.name AS labor_category_name,
  'Labor'::varchar(250) AS type_name,
  wod.contractor,
  wod.effective_area AS surface_ha,
  lb.name::varchar(100) AS supply_name,
  0::numeric(18,6) AS consumption,
  cat_lb.name::varchar(250) AS category_name,
  0::numeric(18,6) AS dose_per_ha,
  lb.price::numeric(18,6) AS unit_price,
  lb.price::numeric AS supply_cost_per_ha,
  (lb.price * wod.effective_area)::numeric AS supply_total_cost,
  true::boolean AS is_digital,
  wod.status::varchar(30) AS status
FROM public.work_order_drafts wod
JOIN public.projects p ON p.id = wod.project_id AND p.deleted_at IS NULL
JOIN public.fields f ON f.id = wod.field_id AND f.deleted_at IS NULL
JOIN public.lots l ON l.id = wod.lot_id AND l.deleted_at IS NULL
JOIN public.crops c ON c.id = wod.crop_id AND c.deleted_at IS NULL
JOIN public.labors lb ON lb.id = wod.labor_id
LEFT JOIN public.categories cat_lb ON cat_lb.id = lb.category_id AND cat_lb.deleted_at IS NULL
WHERE wod.deleted_at IS NULL
  AND wod.is_digital = true
  AND wod.status = 'draft'
  AND wod.effective_area IS NOT NULL
  AND wod.effective_area > 0
  AND lb.price IS NOT NULL;

DROP TRIGGER IF EXISTS trg_enforce_workorders_number_transition ON public.workorders;
DROP FUNCTION IF EXISTS public.enforce_workorder_number_transition();
DROP INDEX IF EXISTS ux_workorders_project_number_active;

ALTER TABLE public.workorders
  DROP COLUMN IF EXISTS legacy_number;

COMMIT;
