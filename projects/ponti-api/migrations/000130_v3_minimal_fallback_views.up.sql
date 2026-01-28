-- ========================================
-- MIGRATION 000130 V3 MINIMAL FALLBACK VIEWS (UP)
-- ========================================
-- Nota: SQL en inglés, comentarios en español

BEGIN;
CREATE OR REPLACE VIEW public.v3_workorder_metrics AS SELECT * FROM v4_calc.workorder_metrics;

CREATE OR REPLACE VIEW public.v3_lot_metrics AS SELECT * FROM v4_report.lot_metrics;

CREATE OR REPLACE VIEW public.v3_lot_list AS SELECT * FROM v4_report.lot_list;

CREATE OR REPLACE VIEW public.v3_labor_metrics AS SELECT * FROM v4_report.labor_metrics;

CREATE OR REPLACE VIEW public.v3_labor_list AS SELECT * FROM v4_report.labor_list;

CREATE OR REPLACE VIEW public.v3_report_field_crop_metrics AS SELECT * FROM v4_report.field_crop_metrics;

CREATE OR REPLACE VIEW public.v3_report_field_crop_cultivos AS SELECT * FROM v4_report.field_crop_cultivos;

CREATE OR REPLACE VIEW public.v3_report_field_crop_labores AS SELECT * FROM v4_report.field_crop_labores;

CREATE OR REPLACE VIEW public.v3_report_field_crop_insumos AS SELECT * FROM v4_report.field_crop_insumos;

CREATE OR REPLACE VIEW public.v3_report_field_crop_economicos AS SELECT * FROM v4_report.field_crop_economicos;

CREATE OR REPLACE VIEW public.v3_report_field_crop_rentabilidad AS SELECT * FROM v4_report.field_crop_rentabilidad;

CREATE OR REPLACE VIEW public.v3_report_summary_results_view AS SELECT * FROM v4_report.summary_results;

CREATE OR REPLACE VIEW public.v3_report_investor_project_base AS SELECT * FROM v4_report.investor_project_base;

CREATE OR REPLACE VIEW public.v3_report_investor_contribution_categories AS SELECT * FROM v4_report.investor_contribution_categories;

CREATE OR REPLACE VIEW public.v3_report_investor_distributions AS SELECT * FROM v4_report.investor_distributions;

CREATE OR REPLACE VIEW public.v3_investor_contribution_data_view AS SELECT * FROM v4_report.investor_contribution_data;

CREATE OR REPLACE VIEW public.v3_dashboard_metrics AS SELECT * FROM v4_report.dashboard_metrics;

CREATE OR REPLACE VIEW public.v3_dashboard_management_balance AS SELECT * FROM v4_report.dashboard_management_balance;

CREATE OR REPLACE VIEW public.v3_dashboard_crop_incidence AS SELECT * FROM v4_report.dashboard_crop_incidence;

CREATE OR REPLACE VIEW public.v3_dashboard_operational_indicators AS SELECT * FROM v4_report.dashboard_operational_indicators;

CREATE OR REPLACE VIEW public.v3_dashboard_contributions_progress AS SELECT * FROM v4_report.dashboard_contributions_progress;

COMMIT;
