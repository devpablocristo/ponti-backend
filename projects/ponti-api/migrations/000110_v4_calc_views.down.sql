-- ========================================
-- MIGRATION 000110 V4 CALC VIEWS.DOWN (DOWN)
-- ========================================
-- Nota: SQL en inglés, comentarios en español

BEGIN;

DROP VIEW IF EXISTS v4_calc.investor_real_contributions;
DROP VIEW IF EXISTS v4_calc.investor_contribution_categories;
DROP VIEW IF EXISTS v4_calc.dashboard_supply_costs_by_project;
DROP VIEW IF EXISTS v4_calc.dashboard_fertilizers_invested_by_project;
DROP VIEW IF EXISTS v4_calc.field_crop_metrics_aggregated;
DROP VIEW IF EXISTS v4_calc.field_crop_metrics_lot_base;
DROP VIEW IF EXISTS v4_calc.field_crop_aggregated;
DROP VIEW IF EXISTS v4_calc.field_crop_labor_costs_by_lot;
DROP VIEW IF EXISTS v4_calc.field_crop_supply_costs_by_lot;
DROP VIEW IF EXISTS v4_calc.field_crop_lot_base;
DROP VIEW IF EXISTS v4_calc.lot_base_income;
DROP VIEW IF EXISTS v4_calc.lot_base_costs;
DROP VIEW IF EXISTS v4_calc.workorder_metrics;

COMMIT;
