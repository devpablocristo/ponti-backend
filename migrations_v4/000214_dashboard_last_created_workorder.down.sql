-- ========================================
-- MIGRATION 000214 DASHBOARD LAST CREATED WORKORDER (DOWN)
-- ========================================

BEGIN;

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
  (SELECT w2.date FROM public.workorders w2
   WHERE w2.field_id = f.id AND w2.deleted_at IS NULL
   ORDER BY w2.date ASC, w2.id ASC LIMIT 1
  ) AS start_date,
  (SELECT w2.date FROM public.workorders w2
   WHERE w2.field_id = f.id AND w2.deleted_at IS NULL
   ORDER BY w2.date DESC, w2.id DESC LIMIT 1
  ) AS end_date,
  v4_core.calculate_campaign_closing_date(
    (SELECT w2.date FROM public.workorders w2
     WHERE w2.field_id = f.id AND w2.deleted_at IS NULL
     ORDER BY w2.date DESC, w2.id DESC LIMIT 1)
  ) AS campaign_closing_date,
  (SELECT w2.number::text FROM public.workorders w2
   WHERE w2.field_id = f.id AND w2.deleted_at IS NULL
   ORDER BY w2.date ASC, w2.id ASC LIMIT 1
  ) AS first_workorder_id,
  (SELECT w2.number::text FROM public.workorders w2
   WHERE w2.field_id = f.id AND w2.deleted_at IS NULL
   ORDER BY w2.date DESC, w2.id DESC LIMIT 1
  ) AS last_workorder_id,
  v4_ssot.last_stock_count_date_for_project(p.id) AS last_stock_count_date
FROM public.projects p
JOIN public.fields f ON f.project_id = p.id AND f.deleted_at IS NULL
WHERE p.deleted_at IS NULL;

COMMIT;
