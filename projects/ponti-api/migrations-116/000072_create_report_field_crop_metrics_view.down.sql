-- ========================================
-- ROLLBACK: report_field_crop_metrics_view
-- ========================================

-- Eliminar índices
DROP INDEX IF EXISTS idx_report_lots_notdel;
DROP INDEX IF EXISTS idx_report_fields_notdel;
DROP INDEX IF EXISTS idx_report_projects_notdel;
DROP INDEX IF EXISTS idx_report_crops_notdel;
DROP INDEX IF EXISTS idx_report_crop_commercializations_notdel;
DROP INDEX IF EXISTS idx_report_workorders_notdel;
DROP INDEX IF EXISTS idx_report_labors_notdel;
DROP INDEX IF EXISTS idx_report_workorder_items_notdel;
DROP INDEX IF EXISTS idx_report_supplies_notdel;
DROP INDEX IF EXISTS idx_report_workorders_composite;
DROP INDEX IF EXISTS idx_report_lots_composite;
DROP INDEX IF EXISTS idx_report_labors_sowing;
DROP INDEX IF EXISTS idx_report_labors_harvest;

-- Eliminar vista
DROP VIEW IF EXISTS report_field_crop_metrics_view;
