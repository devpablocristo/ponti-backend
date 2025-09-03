-- =====================================================
-- 000064: WORKORDER - Vistas de Cálculo de Órdenes de Trabajo
-- =====================================================
-- Entidad: workorder (Órdenes de Trabajo)
-- Funcionalidad: Crear vistas para cálculos de workorders
-- =====================================================

-- Vista para cálculos de workorders con labor y supplies
CREATE OR REPLACE VIEW v_calc_workorders AS
SELECT 
    w.id,
    w.project_id,
    w.field_id,
    w.lot_id,
    w.labor_id,
    w.date,
    w.effective_area,
    
    -- Obtener precio de labor desde la tabla labors
    COALESCE(lb.price, 0) AS labor_price_per_ha,
    
    -- Cálculo de costos de labor
    COALESCE(lb.price * w.effective_area, 0) AS labor_total_usd,
    
    -- Cálculo de costos de supplies (total_used * precio del supply)
    COALESCE(SUM(wi.total_used * s.price), 0) AS supplies_total_usd,
    
    -- Total del workorder (labor + supplies)
    COALESCE(lb.price * w.effective_area, 0) + 
    COALESCE(SUM(wi.total_used * s.price), 0) AS workorder_total_usd,
    
    -- Campos adicionales para referencia
    p.name AS project_name,
    f.name AS field_name,
    l.name AS lot_name,
    c.name AS crop_name,
    lb.name AS labor_name,
    cat.name AS category_name,
    
    w.created_at,
    w.updated_at
FROM workorders w
LEFT JOIN projects p ON w.project_id = p.id AND p.deleted_at IS NULL
LEFT JOIN fields f ON w.field_id = f.id AND f.deleted_at IS NULL
LEFT JOIN lots l ON w.lot_id = l.id AND l.deleted_at IS NULL
LEFT JOIN crops c ON w.crop_id = c.id AND c.deleted_at IS NULL
LEFT JOIN labors lb ON w.labor_id = lb.id AND lb.deleted_at IS NULL
LEFT JOIN categories cat ON lb.category_id = cat.id AND cat.deleted_at IS NULL
LEFT JOIN workorder_items wi ON w.id = wi.workorder_id AND wi.deleted_at IS NULL
LEFT JOIN supplies s ON wi.supply_id = s.id AND s.deleted_at IS NULL
WHERE w.deleted_at IS NULL
GROUP BY 
    w.id, w.project_id, w.field_id, w.lot_id, w.labor_id, w.date, 
    w.effective_area, lb.price, p.name, f.name, l.name, c.name, lb.name, cat.name,
    w.created_at, w.updated_at;

-- Comentarios en español
COMMENT ON VIEW v_calc_workorders IS 'Vista para cálculos de workorders con labor y supplies';
COMMENT ON COLUMN v_calc_workorders.labor_total_usd IS 'Total de costos de labor en USD';
COMMENT ON COLUMN v_calc_workorders.supplies_total_usd IS 'Total de costos de supplies en USD';
COMMENT ON COLUMN v_calc_workorders.workorder_total_usd IS 'Total del workorder (labor + supplies) en USD';
