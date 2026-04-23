-- ========================================
-- MIGRATION 000215 DASHBOARD OPERATIONAL INDICATORS SIMPLE DATES (UP)
-- ========================================
-- Simplifica las cards operativas:
-- - Primera OT: fecha minima y numero asociado por date ASC, id ASC
-- - Ultima OT: fecha maxima y numero asociado por date DESC, id DESC
-- - Arqueo de stock: ultimo movimiento manual tipo Stock (project-level)
-- - Cierre de campana: sin definir (NULL)

BEGIN;

CREATE OR REPLACE FUNCTION v4_ssot.last_manual_stock_entry_date_for_project(p_project_id bigint)
RETURNS date
LANGUAGE sql STABLE AS $$
  SELECT MAX(sm.movement_date::date)
  FROM public.supply_movements sm
  WHERE sm.project_id = p_project_id
    AND sm.deleted_at IS NULL
    AND sm.movement_type = 'Stock'
$$;

CREATE OR REPLACE VIEW v4_report.dashboard_operational_indicators AS
SELECT
  p.id AS project_id,
  v4_ssot.first_workorder_date_for_project(p.id) AS start_date,
  v4_ssot.last_workorder_date_for_project(p.id) AS end_date,
  NULL::date AS campaign_closing_date,
  v4_ssot.first_workorder_number_for_project(p.id) AS first_workorder_id,
  v4_ssot.last_workorder_number_for_project(p.id) AS last_workorder_id,
  v4_ssot.last_manual_stock_entry_date_for_project(p.id) AS last_stock_count_date
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
   LIMIT 1
  ) AS start_date,
  (SELECT w2.date
   FROM public.workorders w2
   WHERE w2.field_id = f.id
     AND w2.deleted_at IS NULL
   ORDER BY w2.date DESC, w2.id DESC
   LIMIT 1
  ) AS end_date,
  NULL::date AS campaign_closing_date,
  (SELECT w2.number::text
   FROM public.workorders w2
   WHERE w2.field_id = f.id
     AND w2.deleted_at IS NULL
   ORDER BY w2.date ASC, w2.id ASC
   LIMIT 1
  ) AS first_workorder_id,
  (SELECT w2.number::text
   FROM public.workorders w2
   WHERE w2.field_id = f.id
     AND w2.deleted_at IS NULL
   ORDER BY w2.date DESC, w2.id DESC
   LIMIT 1
  ) AS last_workorder_id,
  v4_ssot.last_manual_stock_entry_date_for_project(p.id) AS last_stock_count_date
FROM public.projects p
JOIN public.fields f ON f.project_id = p.id AND f.deleted_at IS NULL
WHERE p.deleted_at IS NULL;

COMMIT;
