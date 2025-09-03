-- =====================================================
-- 000067: PROJECT - Vistas de Consolidación de Proyectos
-- =====================================================
-- Entidad: project (Proyectos)
-- Funcionalidad: Crear vistas para consolidaciones de proyectos
-- =====================================================

-- Vista para consolidación de costos por proyecto
CREATE OR REPLACE VIEW v_calc_project_costs AS
SELECT 
    p.id AS project_id,
    p.name AS project_name,
    
    -- Costos totales de workorders
    COALESCE(SUM(wo.workorder_total_usd), 0) AS total_costs_usd,
    
    -- Costos de labor
    COALESCE(SUM(wo.labor_total_usd), 0) AS labor_costs_usd,
    
    -- Costos de supplies
    COALESCE(SUM(wo.supplies_total_usd), 0) AS supplies_costs_usd,
    
    -- Superficie total del proyecto
    COALESCE(SUM(wo.effective_area), 0) AS total_surface_ha,
    
    -- Costo promedio por hectárea
    CASE 
        WHEN SUM(wo.effective_area) > 0 
        THEN COALESCE(SUM(wo.workorder_total_usd), 0) / SUM(wo.effective_area)
        ELSE 0
    END AS avg_cost_per_ha,
    
    p.created_at,
    p.updated_at
FROM projects p
LEFT JOIN v_calc_workorders wo ON p.id = wo.project_id
WHERE p.deleted_at IS NULL
GROUP BY p.id, p.name, p.created_at, p.updated_at;

-- Vista para consolidación de economía por proyecto
CREATE OR REPLACE VIEW v_calc_project_economics AS
SELECT 
    p.id AS project_id,
    p.name AS project_name,
    
    -- Ingresos netos totales
    COALESCE(SUM(l.income_net_per_ha * l.hectares), 0) AS net_income_usd,
    
    -- Total activo (costos + arriendo)
    COALESCE(SUM(l.active_total_per_ha * l.hectares), 0) AS active_total_usd,
    
    -- Resultado operativo total
    COALESCE(SUM(l.operating_result_per_ha * l.hectares), 0) AS operating_result_usd,
    
    -- Superficie total sembrada
    COALESCE(SUM(l.hectares), 0) AS total_sowed_area,
    
    -- Superficie total cosechada (simplificado - usar hectares por ahora)
    COALESCE(SUM(l.hectares), 0) AS total_harvested_area,
    
    -- Rendimiento promedio del proyecto
    CASE 
        WHEN SUM(l.hectares) > 0 
        THEN COALESCE(SUM(l.yield_tonha * l.hectares), 0) / SUM(l.hectares)
        ELSE 0
    END AS avg_yield_tonha,
    
    p.created_at,
    p.updated_at
FROM projects p
LEFT JOIN v_calc_lots l ON p.id = l.project_id
WHERE p.deleted_at IS NULL
GROUP BY p.id, p.name, p.created_at, p.updated_at;

-- Comentarios en español
COMMENT ON VIEW v_calc_project_costs IS 'Vista para consolidación de costos por proyecto';
COMMENT ON VIEW v_calc_project_economics IS 'Vista para consolidación de economía por proyecto';
COMMENT ON COLUMN v_calc_project_costs.total_costs_usd IS 'Costos totales del proyecto en USD';
COMMENT ON COLUMN v_calc_project_costs.avg_cost_per_ha IS 'Costo promedio por hectárea';
COMMENT ON COLUMN v_calc_project_economics.net_income_usd IS 'Ingresos netos totales del proyecto';
COMMENT ON COLUMN v_calc_project_economics.operating_result_usd IS 'Resultado operativo total del proyecto';
