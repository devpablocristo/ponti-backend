-- ========================================
-- MIGRATION 000130 V3 MINIMAL FALLBACK VIEWS (DOWN)
-- ========================================
-- Nota: SQL en inglés, comentarios en español

BEGIN;
DROP VIEW IF EXISTS public.v3_dashboard_contributions_progress;
DROP VIEW IF EXISTS public.v3_dashboard_operational_indicators;
DROP VIEW IF EXISTS public.v3_dashboard_crop_incidence;
DROP VIEW IF EXISTS public.v3_dashboard_management_balance;
DROP VIEW IF EXISTS public.v3_dashboard_metrics;
DROP VIEW IF EXISTS public.v3_investor_contribution_data_view;
DROP VIEW IF EXISTS public.v3_report_investor_distributions;
DROP VIEW IF EXISTS public.v3_report_investor_contribution_categories;
DROP VIEW IF EXISTS public.v3_report_investor_project_base;
DROP VIEW IF EXISTS public.v3_report_summary_results_view;
DROP VIEW IF EXISTS public.v3_report_field_crop_rentabilidad;
DROP VIEW IF EXISTS public.v3_report_field_crop_economicos;
DROP VIEW IF EXISTS public.v3_report_field_crop_insumos;
DROP VIEW IF EXISTS public.v3_report_field_crop_labores;
DROP VIEW IF EXISTS public.v3_report_field_crop_cultivos;
DROP VIEW IF EXISTS public.v3_report_field_crop_metrics;
DROP VIEW IF EXISTS public.v3_labor_list;
DROP VIEW IF EXISTS public.v3_labor_metrics;
DROP VIEW IF EXISTS public.v3_lot_list;
DROP VIEW IF EXISTS public.v3_lot_metrics;
DROP VIEW IF EXISTS public.v3_workorder_metrics;

COMMIT;
