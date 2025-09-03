-- =====================================================
-- 000063: LABORS - Vistas de Cálculo
-- =====================================================
-- Entidad: Labors (Labores)
-- Funcionalidad: Métricas de labor con IVA 10.5% y conversión ARS
-- =====================================================

-- Vista de cálculo para labors (métricas de tabla usadas por órdenes)
CREATE OR REPLACE VIEW v_calc_labors AS
SELECT 
    w.id AS workorder_id,
    w.project_id,
    w.field_id,
    w.lot_id,
    w.labor_id,
    w.effective_area,
    lb.price AS labor_price_per_ha,
    lb.category_id,
    -- Total USD neto (por fila de labor usada en órdenes)
    lb.price * w.effective_area AS total_usd_net,
    -- IVA para labors: usar 10.5% (reemplazar viejo 21%)
    (lb.price * w.effective_area) * 0.105 AS iva_amount,
    -- USD/Ha a ARS (mostrar en pesos): usar último FX (USD→ARS)
    -- Si no hay FX, default a 1 (mostrando USD)
    COALESCE(fx.latest_rate, 1) AS usd_ars_rate,
    lb.price * COALESCE(fx.latest_rate, 1) AS cost_ars_per_ha,
    -- Total ARS (por orden)
    (lb.price * COALESCE(fx.latest_rate, 1)) * w.effective_area AS total_ars
FROM workorders w
INNER JOIN labors lb ON w.labor_id = lb.id
-- Left join con la tabla de tipos de cambio (se creará en migración 000067)
LEFT JOIN LATERAL (
    SELECT rate AS latest_rate
    FROM fx_rates 
    WHERE code = 'USDARS' 
        AND deleted_at IS NULL
    ORDER BY as_of_date DESC 
    LIMIT 1
) fx ON true
WHERE w.deleted_at IS NULL AND lb.deleted_at IS NULL;
