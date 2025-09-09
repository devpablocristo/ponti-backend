-- ========================================
-- MIGRACIÓN 000072: CREAR VISTAS BASE DE REPORTES
-- Entidad: report (Crear vistas base para reportes)
-- Funcionalidad: Vista optimizada para reportes de métricas por campo y cultivo
-- ========================================

-- ========================================
-- 1. CREAR VISTA REPORT_FIELD_CROP_METRICS_VIEW_V2
-- ========================================
DROP VIEW IF EXISTS report_field_crop_metrics_view_v2;

CREATE VIEW report_field_crop_metrics_view_v2 AS
SELECT 
    p.id AS project_id,
    f.id AS field_id,
    f.name AS field_name,
    c.id AS current_crop_id,
    c.name AS crop_name,
    
    -- Superficie y áreas
    COALESCE(SUM(l.hectares), 0) AS superficie_ha,
    COALESCE(SUM(l.tons), 0) AS produccion_tn,
    COALESCE(SUM(CASE WHEN l.sowing_date IS NOT NULL THEN l.hectares ELSE 0 END), 0) AS area_sembrada_ha,
    COALESCE(SUM(CASE WHEN l.tons IS NOT NULL AND l.tons > 0 THEN l.hectares ELSE 0 END), 0) AS area_cosechada_ha,
    
    -- Rendimiento
    CASE 
        WHEN COALESCE(SUM(CASE WHEN l.tons IS NOT NULL AND l.tons > 0 THEN l.hectares ELSE 0 END), 0) > 0 
        THEN COALESCE(SUM(l.tons), 0) / COALESCE(SUM(CASE WHEN l.tons IS NOT NULL AND l.tons > 0 THEN l.hectares ELSE 0 END), 0)
        ELSE 0 
    END AS rendimiento_tn_ha,
    
    -- Precios (valores de ejemplo - ajustar según lógica de negocio)
    0.00 AS precio_bruto_usd_tn,
    0.00 AS gasto_flete_usd_tn,
    0.00 AS gasto_comercial_usd_tn,
    0.00 AS precio_neto_usd_tn,
    
    -- Ingresos (valores de ejemplo - ajustar según lógica de negocio)
    0.00 AS ingreso_neto_usd,
    0.00 AS ingreso_neto_usd_ha,
    
    -- Costos (valores de ejemplo - ajustar según lógica de negocio)
    0.00 AS costos_labores_usd,
    0.00 AS costos_insumos_usd,
    0.00 AS total_costos_directos_usd,
    0.00 AS costos_directos_usd_ha,
    
    -- Márgenes (valores de ejemplo - ajustar según lógica de negocio)
    0.00 AS margen_bruto_usd,
    0.00 AS margen_bruto_usd_ha,
    
    -- Arriendo y administración (valores de ejemplo - ajustar según lógica de negocio)
    0.00 AS arriendo_usd,
    0.00 AS arriendo_usd_ha,
    0.00 AS administracion_usd,
    0.00 AS administracion_usd_ha,
    
    -- Resultado operativo (valores de ejemplo - ajustar según lógica de negocio)
    0.00 AS resultado_operativo_usd,
    0.00 AS resultado_operativo_usd_ha,
    0.00 AS total_invertido_usd,
    0.00 AS total_invertido_usd_ha,
    
    -- Indicadores (valores de ejemplo - ajustar según lógica de negocio)
    0.00 AS renta_pct,
    0.00 AS rinde_indiferencia_usd_tn

FROM projects p
JOIN fields f ON f.project_id = p.id AND f.deleted_at IS NULL
LEFT JOIN lots l ON l.field_id = f.id AND l.deleted_at IS NULL
LEFT JOIN crops c ON l.current_crop_id = c.id AND c.deleted_at IS NULL
WHERE p.deleted_at IS NULL
GROUP BY p.id, f.id, f.name, c.id, c.name;
