-- ========================================
-- MIGRACIÓN 000351: Drop v4 views reescritas (DOWN)
-- ========================================

BEGIN;
DROP VIEW IF EXISTS v4_report.workorder_metrics CASCADE;
DROP VIEW IF EXISTS v4_report.summary_results CASCADE;
DROP VIEW IF EXISTS v4_report.lot_metrics CASCADE;
DROP VIEW IF EXISTS v4_report.lot_list CASCADE;
DROP VIEW IF EXISTS v4_report.labor_metrics CASCADE;
DROP VIEW IF EXISTS v4_report.labor_list CASCADE;
DROP VIEW IF EXISTS v4_report.investor_project_base CASCADE;
DROP VIEW IF EXISTS v4_report.investor_distributions CASCADE;
DROP VIEW IF EXISTS v4_report.investor_contribution_data CASCADE;
DROP VIEW IF EXISTS v4_report.investor_contribution_categories CASCADE;
DROP VIEW IF EXISTS v4_report.field_crop_rentabilidad CASCADE;
DROP VIEW IF EXISTS v4_report.field_crop_metrics CASCADE;
DROP VIEW IF EXISTS v4_report.field_crop_labores CASCADE;
DROP VIEW IF EXISTS v4_report.field_crop_insumos CASCADE;
DROP VIEW IF EXISTS v4_report.field_crop_economicos CASCADE;
DROP VIEW IF EXISTS v4_report.field_crop_cultivos CASCADE;
DROP VIEW IF EXISTS v4_report.dashboard_operational_indicators CASCADE;
DROP VIEW IF EXISTS v4_report.dashboard_metrics CASCADE;
DROP VIEW IF EXISTS v4_report.dashboard_management_balance CASCADE;
DROP VIEW IF EXISTS v4_report.dashboard_crop_incidence CASCADE;
DROP VIEW IF EXISTS v4_report.dashboard_contributions_progress CASCADE;
DROP VIEW IF EXISTS v4_calc.lot_base_income CASCADE;
DROP VIEW IF EXISTS v4_calc.lot_base_costs CASCADE;
DROP VIEW IF EXISTS v4_calc.investor_real_contributions CASCADE;
DROP VIEW IF EXISTS v4_calc.field_crop_supply_costs_by_lot CASCADE;
DROP VIEW IF EXISTS v4_calc.field_crop_metrics_lot_base CASCADE;
DROP VIEW IF EXISTS v4_calc.field_crop_metrics_aggregated CASCADE;
DROP VIEW IF EXISTS v4_calc.field_crop_lot_base CASCADE;
DROP VIEW IF EXISTS v4_calc.field_crop_labor_costs_by_lot CASCADE;
DROP VIEW IF EXISTS v4_calc.field_crop_aggregated CASCADE;
DROP VIEW IF EXISTS v4_calc.dashboard_supply_costs_by_project CASCADE;
DROP VIEW IF EXISTS v4_calc.dashboard_fertilizers_invested_by_project CASCADE;
COMMIT;
