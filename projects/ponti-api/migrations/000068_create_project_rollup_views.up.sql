-- =====================================================
-- 000068: PROJECT ROLLUPS - Vistas de Consolidación
-- =====================================================
-- Entidad: Projects (Proyectos)
-- Funcionalidad: Consolidados de costos y economía por proyecto
-- =====================================================

-- Rollup de costos por proyecto
CREATE OR REPLACE VIEW v_calc_project_costs AS
SELECT 
    project_id,
    SUM(CASE WHEN category_id = 13 THEN workorder_total_usd ELSE 0 END) AS harvest_costs_usd,
    SUM(CASE WHEN category_id != 13 THEN workorder_total_usd ELSE 0 END) AS other_costs_usd,
    SUM(workorder_total_usd) AS total_costs_usd
FROM v_calc_workorders
GROUP BY project_id;

-- Rollup de economía por proyecto
CREATE OR REPLACE VIEW v_calc_project_economics AS
SELECT 
    project_id,
    SUM(net_income_per_ha * lot_hectares) AS net_income_usd,
    SUM(active_total_per_ha * lot_hectares) AS active_total_usd,
    SUM(operating_result_per_ha * lot_hectares) AS operating_result_usd
FROM v_calc_lots
GROUP BY project_id;
