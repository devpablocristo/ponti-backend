-- =====================================================
-- 000067: LOTS - Vistas de Cálculo
-- =====================================================
-- Entidad: Lots (Lotes)
-- Funcionalidad: Rendimiento y economía por lote
-- =====================================================

-- Vista de cálculo para lots (rendimiento y economía)
CREATE OR REPLACE VIEW v_calc_lots AS
SELECT 
    l.id AS lot_id,
    l.field_id,
    f.project_id,
    l.current_crop_id AS crop_id,
    l.hectares AS lot_hectares,
    -- Yield (ton/ha): fuente primaria = harvests (suma tons por lote); fallback = lots.tons
    COALESCE(h.yield_tonha, 0) AS yield_tonha,
    -- Precio neto selección (USD): último por fecha de crop_commercializations para (field_id, crop_id)
    -- fallback a project+crop si falta
    COALESCE(lnp.net_price, 0) AS net_price_usd,
    -- Ingreso neto por ha (USD)
    COALESCE(h.yield_tonha, 0) * COALESCE(lnp.net_price, 0) AS net_income_per_ha,
    -- Costo por ha (USD): suma de workorders por lote
    COALESCE(SUM(wc.workorder_total_usd), 0) / NULLIF(l.hectares, 0) AS cost_per_ha,
    -- Costo admin por ha (USD): cantidad fija almacenada por proyecto/cliente
    COALESCE(p.admin_cost, 0) AS admin_cost_per_ha,
    -- Lease (arriendo) por ha (USD): tres modos
    f.lease_type_id,
    CASE 
        WHEN f.lease_type_id = 1 THEN -- Fixed: constante por ha
            COALESCE(f.lease_type_value, 0)
        WHEN f.lease_type_id = 2 THEN -- % Net income
            COALESCE(f.lease_type_percent, 0) * (COALESCE(h.yield_tonha, 0) * COALESCE(lnp.net_price, 0))
        WHEN f.lease_type_id = 3 THEN -- % Utility (profit)
            COALESCE(f.lease_type_percent, 0) * (
                (COALESCE(h.yield_tonha, 0) * COALESCE(lnp.net_price, 0)) - 
                (COALESCE(SUM(wc.workorder_total_usd), 0) / NULLIF(l.hectares, 0)) - 
                COALESCE(p.admin_cost, 0)
            )
        WHEN f.lease_type_id = 4 THEN -- Mixed (fixed + % net income)
            COALESCE(f.lease_type_value, 0) + 
            COALESCE(f.lease_type_percent, 0) * (COALESCE(h.yield_tonha, 0) * COALESCE(lnp.net_price, 0))
        ELSE 0
    END AS lease_per_ha,
    -- Total activo por ha (USD)
    (COALESCE(SUM(wc.workorder_total_usd), 0) / NULLIF(l.hectares, 0)) + 
    CASE 
        WHEN f.lease_type_id = 1 THEN COALESCE(f.lease_type_value, 0)
        WHEN f.lease_type_id = 2 THEN COALESCE(f.lease_type_percent, 0) * (COALESCE(h.yield_tonha, 0) * COALESCE(lnp.net_price, 0))
        WHEN f.lease_type_id = 3 THEN COALESCE(f.lease_type_percent, 0) * (
            (COALESCE(h.yield_tonha, 0) * COALESCE(lnp.net_price, 0)) - 
            (COALESCE(SUM(wc.workorder_total_usd), 0) / NULLIF(l.hectares, 0)) - 
            COALESCE(p.admin_cost, 0)
        )
        WHEN f.lease_type_id = 4 THEN COALESCE(f.lease_type_value, 0) + 
            COALESCE(f.lease_type_percent, 0) * (COALESCE(h.yield_tonha, 0) * COALESCE(lnp.net_price, 0))
        ELSE 0
    END + COALESCE(p.admin_cost, 0) AS active_total_per_ha,
    -- Resultado operativo por ha (USD)
    (COALESCE(h.yield_tonha, 0) * COALESCE(lnp.net_price, 0)) - 
    ((COALESCE(SUM(wc.workorder_total_usd), 0) / NULLIF(l.hectares, 0)) + 
    CASE 
        WHEN f.lease_type_id = 1 THEN COALESCE(f.lease_type_value, 0)
        WHEN f.lease_type_id = 2 THEN COALESCE(f.lease_type_percent, 0) * (COALESCE(h.yield_tonha, 0) * COALESCE(lnp.net_price, 0))
        WHEN f.lease_type_id = 3 THEN COALESCE(f.lease_type_percent, 0) * (
            (COALESCE(h.yield_tonha, 0) * COALESCE(lnp.net_price, 0)) - 
            (COALESCE(SUM(wc.workorder_total_usd), 0) / NULLIF(l.hectares, 0)) - 
            COALESCE(p.admin_cost, 0)
        )
        WHEN f.lease_type_id = 4 THEN COALESCE(f.lease_type_value, 0) + 
            COALESCE(f.lease_type_percent, 0) * (COALESCE(h.yield_tonha, 0) * COALESCE(lnp.net_price, 0))
        ELSE 0
    END + COALESCE(p.admin_cost, 0)) AS operating_result_per_ha
FROM lots l
INNER JOIN fields f ON l.field_id = f.id
INNER JOIN projects p ON f.project_id = p.id
LEFT JOIN v_helper_harvests h ON l.id = h.lot_id
LEFT JOIN v_helper_last_net_price lnp ON f.id = lnp.field_id AND l.current_crop_id = lnp.crop_id
LEFT JOIN v_calc_workorders wc ON l.id = wc.lot_id
WHERE l.deleted_at IS NULL 
    AND f.deleted_at IS NULL 
    AND p.deleted_at IS NULL
GROUP BY 
    l.id, l.field_id, f.project_id, l.current_crop_id, l.hectares, 
    h.yield_tonha, lnp.net_price, p.admin_cost, f.lease_type_id, 
    f.lease_type_value, f.lease_type_percent;
