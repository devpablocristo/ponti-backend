-- ========================================
-- MIGRACIÓN 000071: FIX LABORS LIST
-- Entidad: labor (Labores)
-- Funcionalidad: Crear vista fix_labors_list con cálculos correctos
-- ========================================

-- Eliminar la vista si existe
DROP VIEW IF EXISTS fix_labors_list;

-- Crear la vista fix_labors_list con cálculos correctos
CREATE VIEW fix_labors_list AS
SELECT 
    w.id AS workorder_id,
    w.number AS workorder_number,
    w.date,
    p.name AS project_name,
    f.name AS field_name,
    c.name AS crop_name,
    lb.name AS labor_name,
    lc.name AS category_name,
    w.contractor,
    w.effective_area AS surface_ha,
    lb.price AS cost_ha,
    lb.contractor_name,
    inv.name AS investor_name,
    pdv.average_value AS usd_avg_value,
    
    -- Cálculos correctos implementados
    -- 1. Total neto en pesos (costo por ha * superficie)
    (lb.price * w.effective_area) AS net_total,
    
    -- 2. IVA correcto (10.5%)
    (lb.price * w.effective_area * 0.105) AS total_iva,
    
    -- 3. Costo U$/Ha en ARS (costo por ha * dólar promedio)
    (lb.price * pdv.average_value) AS usd_cost_ha,
    
    -- 4. Total U Neto en ARS (costo por ha * dólar promedio * superficie)
    (lb.price * pdv.average_value * w.effective_area) AS usd_net_total,
    
    -- 5. Total con IVA en pesos
    (lb.price * w.effective_area * 1.105) AS total_with_iva,
    
    -- 6. Total con IVA en ARS
    (lb.price * pdv.average_value * w.effective_area * 1.105) AS total_with_iva_ars,
    
    -- Campos adicionales para referencia
    w.project_id,
    w.field_id,
    w.created_at,
    w.updated_at

FROM workorders w
    INNER JOIN projects p ON w.project_id = p.id AND p.deleted_at IS NULL
    INNER JOIN fields f ON w.field_id = f.id AND f.deleted_at IS NULL
    INNER JOIN crops c ON w.crop_id = c.id AND c.deleted_at IS NULL
    INNER JOIN labors lb ON w.labor_id = lb.id AND lb.deleted_at IS NULL
    INNER JOIN categories lc ON lb.category_id = lc.id AND lc.deleted_at IS NULL
    INNER JOIN investors inv ON w.investor_id = inv.id AND inv.deleted_at IS NULL
    LEFT JOIN project_dollar_values pdv ON pdv.project_id = w.project_id 
        AND pdv.month = TO_CHAR(w.date, 'YYYY-MM') 
        AND pdv.deleted_at IS NULL
WHERE w.deleted_at IS NULL
    AND w.effective_area IS NOT NULL
    AND w.effective_area > 0
    AND lb.price IS NOT NULL;

-- Comentarios en español
COMMENT ON VIEW fix_labors_list IS 'Vista con cálculos correctos de labor: IVA 10.5%, ARS correcto, totales correctos';
COMMENT ON COLUMN fix_labors_list.net_total IS 'Total neto en pesos (costo por ha × superficie)';
COMMENT ON COLUMN fix_labors_list.total_iva IS 'IVA correcto (10.5% del total neto)';
COMMENT ON COLUMN fix_labors_list.usd_cost_ha IS 'Costo por hectárea en ARS (costo × dólar promedio)';
COMMENT ON COLUMN fix_labors_list.usd_net_total IS 'Total neto en ARS (costo × dólar × superficie)';
COMMENT ON COLUMN fix_labors_list.total_with_iva IS 'Total con IVA en pesos';
COMMENT ON COLUMN fix_labors_list.total_with_iva_ars IS 'Total con IVA en ARS';
