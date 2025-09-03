-- =====================================================
-- 000066: HELPER VIEWS - Vistas Auxiliares Compartidas
-- =====================================================
-- Entidad: Shared (Compartidas)
-- Funcionalidad: Vistas helper para cálculos de todas las entidades
-- =====================================================

-- Vista helper para totales de cosecha por lote
-- Preferencia: harvests (si existe) > lots.tons como fallback
CREATE OR REPLACE VIEW v_helper_harvests AS
WITH harvest_totals AS (
    SELECT 
        w.lot_id,
        SUM(w.effective_area) AS harvested_area,
        COALESCE(l.tons, 0) AS fallback_tons
    FROM workorders w
    INNER JOIN labors lb ON w.labor_id = lb.id
    INNER JOIN lots l ON w.lot_id = l.id
    WHERE w.deleted_at IS NULL 
        AND lb.deleted_at IS NULL 
        AND l.deleted_at IS NULL
        -- Asumimos que category_id = 13 es "Harvest" basado en las migraciones existentes
        AND lb.category_id = 13
    GROUP BY w.lot_id, l.tons
)
SELECT 
    h.lot_id,
    h.harvested_area,
    h.fallback_tons,
    CASE 
        WHEN h.harvested_area > 0 THEN h.fallback_tons / h.harvested_area
        ELSE 0 
    END AS yield_tonha
FROM harvest_totals h;

-- Vista helper para último precio neto por (field_id, crop_id) con fallback a project+crop
CREATE OR REPLACE VIEW v_helper_last_net_price AS
WITH last_prices AS (
    SELECT DISTINCT ON (f.id, l.current_crop_id)
        f.id AS field_id,
        l.current_crop_id AS crop_id,
        f.project_id,
        cc.net_price,
        cc.created_at
    FROM fields f
    INNER JOIN lots l ON f.id = l.field_id
    INNER JOIN crop_commercializations cc ON f.project_id = cc.project_id 
        AND l.current_crop_id = cc.crop_id
    WHERE f.deleted_at IS NULL 
        AND l.deleted_at IS NULL 
        AND cc.deleted_at IS NULL
    ORDER BY f.id, l.current_crop_id, cc.created_at DESC
)
SELECT 
    field_id,
    crop_id,
    project_id,
    net_price
FROM last_prices;
