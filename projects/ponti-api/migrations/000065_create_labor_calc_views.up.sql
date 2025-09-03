-- =====================================================
-- 000065: LABOR - Vistas de Cálculo de Labores
-- =====================================================
-- Entidad: labor (Labores)
-- Funcionalidad: Crear vistas para cálculos de labors
-- =====================================================

-- Vista para cálculos de labors con conversiones USD/ARS
CREATE OR REPLACE VIEW v_calc_labors AS
SELECT 
    l.id,
    l.project_id,
    l.category_id,
    l.price,
    l.contractor_name,
    
    -- Cálculo del total neto en USD (asumiendo precio por hectárea)
    COALESCE(l.price, 0) AS total_usd_net,
    
    -- Cálculo del IVA (10.5%)
    COALESCE(l.price * 0.105, 0) AS iva_amount,
    
    -- Total con IVA
    COALESCE(l.price * 1.105, 0) AS total_usd_gross,
    
    -- Conversión a ARS usando la tasa FX más reciente
    COALESCE(l.price * fx.rate, 0) AS cost_ars_per_ha,
    
    -- Tasa de cambio utilizada
    COALESCE(fx.rate, 1.0) AS usd_ars_rate,
    
    -- Campos adicionales para referencia
    p.name AS project_name,
    c.name AS category_name,
    
    l.created_at,
    l.updated_at
FROM labors l
LEFT JOIN projects p ON l.project_id = p.id AND p.deleted_at IS NULL
LEFT JOIN categories c ON l.category_id = c.id AND c.deleted_at IS NULL
LEFT JOIN LATERAL (
    SELECT rate 
    FROM fx_rates 
    WHERE currency_pair = 'USDARS' 
    AND effective_date <= CURRENT_DATE
    ORDER BY effective_date DESC 
    LIMIT 1
) fx ON true
WHERE l.deleted_at IS NULL;

-- Comentarios en español
COMMENT ON VIEW v_calc_labors IS 'Vista para cálculos de labors con conversiones USD/ARS';
COMMENT ON COLUMN v_calc_labors.total_usd_net IS 'Total neto en USD (sin IVA)';
COMMENT ON COLUMN v_calc_labors.iva_amount IS 'Monto del IVA (10.5%)';
COMMENT ON COLUMN v_calc_labors.total_usd_gross IS 'Total bruto en USD (con IVA)';
COMMENT ON COLUMN v_calc_labors.cost_ars_per_ha IS 'Costo por hectárea en ARS';
COMMENT ON COLUMN v_calc_labors.usd_ars_rate IS 'Tasa de cambio USD/ARS utilizada';
