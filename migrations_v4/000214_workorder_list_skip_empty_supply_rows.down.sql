BEGIN;

CREATE OR REPLACE VIEW v4_report.workorder_list AS
-- Filas de insumos de workorders publicadas/manuales
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

-- Fila de costo de labor de workorders publicadas/manuales
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

-- Filas de insumos de drafts digitales abiertas
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

-- Fila de costo de labor de drafts digitales abiertas
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

COMMIT;
