-- =====================================================
-- 000066: LOT - Vistas de Cálculo de Lotes
-- =====================================================
-- Entidad: lot (Lotes)
-- Funcionalidad: Crear vistas para cálculos de lots
-- =====================================================

-- Vista para cálculos de lots con métricas económicas
CREATE OR REPLACE VIEW v_calc_lots AS
SELECT 
    l.id,
    f.project_id,
    l.field_id,
    l.current_crop_id,
    l.hectares,
    
    -- Cálculo de rendimiento (usar lots.tons como fallback)
    COALESCE(l.tons / NULLIF(l.hectares, 0), 0) AS yield_tonha,
    
    -- Cálculo de precio neto (último precio por fecha para project_id + crop_id)
    COALESCE(np.net_price, 0) AS net_price_usd,
    
    -- Cálculo de ingreso neto por hectárea
    COALESCE((l.tons / NULLIF(l.hectares, 0)) * np.net_price, 0) AS income_net_per_ha,
    
    -- Arriendo por hectárea (simplificado - usar 0 por ahora)
    0 AS lease_per_ha,
    
    -- Total activo por hectárea (simplificado - usar 0 por ahora)
    0 AS active_total_per_ha,
    
    -- Resultado operativo por hectárea (ingreso - total activo)
    COALESCE((l.tons / NULLIF(l.hectares, 0)) * np.net_price, 0) AS operating_result_per_ha,
    
    -- Campos adicionales para referencia
    p.name AS project_name,
    f.name AS field_name,
    l.name AS lot_name,
    c.name AS crop_name,
    lt.name AS lease_type_name,
    
    l.created_at,
    l.updated_at
FROM lots l
LEFT JOIN fields f ON l.field_id = f.id AND f.deleted_at IS NULL
LEFT JOIN projects p ON f.project_id = p.id AND p.deleted_at IS NULL
LEFT JOIN crops c ON l.current_crop_id = c.id AND c.deleted_at IS NULL
LEFT JOIN lease_types lt ON f.lease_type_id = lt.id AND lt.deleted_at IS NULL
LEFT JOIN LATERAL (
    SELECT net_price
    FROM crop_commercializations cc
    WHERE cc.project_id = f.project_id 
    AND cc.crop_id = l.current_crop_id
    AND cc.deleted_at IS NULL
    ORDER BY cc.created_at DESC
    LIMIT 1
) np ON true
WHERE l.deleted_at IS NULL;

-- Comentarios en español
COMMENT ON VIEW v_calc_lots IS 'Vista para cálculos de lots con métricas económicas';
COMMENT ON COLUMN v_calc_lots.yield_tonha IS 'Rendimiento en toneladas por hectárea';
COMMENT ON COLUMN v_calc_lots.net_price_usd IS 'Precio neto en USD por tonelada';
COMMENT ON COLUMN v_calc_lots.income_net_per_ha IS 'Ingreso neto por hectárea en USD';
COMMENT ON COLUMN v_calc_lots.lease_per_ha IS 'Arriendo por hectárea (simplificado)';
COMMENT ON COLUMN v_calc_lots.active_total_per_ha IS 'Total activo por hectárea (simplificado)';
COMMENT ON COLUMN v_calc_lots.operating_result_per_ha IS 'Resultado operativo por hectárea';
