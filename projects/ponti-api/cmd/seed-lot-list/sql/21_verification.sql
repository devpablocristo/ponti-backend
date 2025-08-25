-- =======================
-- VERIFICACIÓN FINAL PARA LIST LOT
-- =======================
-- Verificar que todos los datos se crearon correctamente para list lot

-- Resumen general
SELECT '📊 RESUMEN GENERAL' as info;
SELECT 
    'Proyecto' as entidad,
    COUNT(*) as total
FROM projects 
WHERE deleted_at IS NULL
UNION ALL
SELECT 
    'Campos' as entidad,
    COUNT(*) as total
FROM fields 
WHERE deleted_at IS NULL
UNION ALL
SELECT 
    'Lotes' as entidad,
    COUNT(*) as total
FROM lots 
WHERE deleted_at IS NULL
UNION ALL
SELECT 
    'Labores' as entidad,
    COUNT(*) as total
FROM labors 
WHERE deleted_at IS NULL
UNION ALL
SELECT 
    'Workorders' as entidad,
    COUNT(*) as total
FROM workorders 
WHERE deleted_at IS NULL
UNION ALL
SELECT 
    'Comercialización' as entidad,
    COUNT(*) as total
FROM crop_commercializations 
WHERE deleted_at IS NULL;

-- Verificar datos de lotes
SELECT '📋 DATOS DE LOTES' as info;
SELECT 
    l.id,
    l.name,
    l.hectares,
    l.tons,
    (l.tons/l.hectares)::numeric(10,2) as rendimiento_tn_ha,
    f.name as campo,
    f.lease_type_id,
    CASE 
        WHEN f.lease_type_id = 1 THEN 'Fijo: $' || f.lease_type_value || '/ha'
        WHEN f.lease_type_id = 2 THEN 'Porcentaje: ' || f.lease_type_percent || '%'
        ELSE 'Sin arriendo'
    END as tipo_arriendo
FROM lots l
JOIN fields f ON l.field_id = f.id
WHERE l.deleted_at IS NULL
ORDER BY l.id;

-- Verificar workorders por categoría
SELECT '🔧 WORKORDERS POR CATEGORÍA' as info;
SELECT 
    lb.category_id,
    lb.name as labor,
    COUNT(w.id) as total_workorders,
    SUM(w.effective_area) as area_total
FROM workorders w
JOIN labors lb ON w.labor_id = lb.id
WHERE w.deleted_at IS NULL
GROUP BY lb.category_id, lb.name
ORDER BY lb.category_id;

-- Verificar precios de comercialización
SELECT '💰 PRECIOS DE COMERCIALIZACIÓN' as info;
SELECT 
    cc.id,
    c.name as cultivo,
    cc.board_price as precio_tablero,
    cc.freight_cost as costo_flete,
    cc.commercial_cost as costo_comercial,
    cc.net_price as precio_neto
FROM crop_commercializations cc
JOIN crops c ON cc.crop_id = c.id
WHERE cc.deleted_at IS NULL
ORDER BY cc.id;

-- Verificar que la vista lot_table_view funcione
SELECT '👁️ VERIFICAR VISTA LOT_TABLE_VIEW' as info;
SELECT 
    id,
    lot_name,
    sowed_area,
    harvested_area,
    tons,
    income_net_total,
    direct_cost_total,
    rent_total,
    admin_total,
    active_total,
    operating_result_per_ha
FROM lot_table_view 
ORDER BY id;
