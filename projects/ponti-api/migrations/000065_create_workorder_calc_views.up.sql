-- =====================================================
-- 000065: WORKORDERS - Vistas de Cálculo
-- =====================================================
-- Entidad: Workorders (Órdenes de Trabajo)
-- Funcionalidad: Cálculos de costos por orden de trabajo
-- =====================================================

-- Vista de cálculo para workorders (casos A y B)
CREATE OR REPLACE VIEW v_calc_workorders AS
SELECT 
    w.id AS workorder_id,
    w.project_id,
    w.field_id,
    w.lot_id,
    w.crop_id,
    w.effective_area,
    lb.price AS labor_price_per_ha,
    lb.category_id,
    -- Total de supplies por workorder (ya calculado por orden)
    COALESCE(SUM(wi.total_used * s.price), 0) AS supplies_total_usd,
    -- Caso A: labor-only workorder
    -- Caso B: labor + supplies (supplies total ya es por orden)
    CASE 
        WHEN COALESCE(SUM(wi.total_used * s.price), 0) > 0 THEN
            COALESCE(SUM(wi.total_used * s.price), 0) + (lb.price * w.effective_area)
        ELSE
            lb.price * w.effective_area
    END AS workorder_total_usd
FROM workorders w
INNER JOIN labors lb ON w.labor_id = lb.id
LEFT JOIN workorder_items wi ON w.id = wi.workorder_id AND wi.deleted_at IS NULL
LEFT JOIN supplies s ON wi.supply_id = s.id AND s.deleted_at IS NULL
WHERE w.deleted_at IS NULL AND lb.deleted_at IS NULL
GROUP BY 
    w.id, w.project_id, w.field_id, w.lot_id, w.crop_id, 
    w.effective_area, lb.price, lb.category_id;
