-- =========================================================
-- Simplificar vistas del dashboard para mejor rendimiento
-- =========================================================

-- Eliminar vistas problemáticas
DROP VIEW IF EXISTS dashboard_card_sowing_view;
DROP VIEW IF EXISTS dashboard_card_harvest_view;
DROP VIEW IF EXISTS dashboard_card_costs_progress_view;
DROP VIEW IF EXISTS dashboard_card_contributions_view;
DROP VIEW IF EXISTS dashboard_card_operating_result_view;

-- Vista simplificada para métricas de siembra
CREATE VIEW dashboard_card_sowing_view AS
SELECT 
    p.customer_id,
    p.id as project_id,
    p.campaign_id,
    f.id as field_id,
    -- Calcular área sembrada basándose en sowing_date
    COALESCE(SUM(CASE WHEN l.sowing_date IS NOT NULL THEN l.hectares ELSE 0 END), 0)::numeric(14,2) as sowed_area,
    -- Total de hectáreas del proyecto
    COALESCE(SUM(l.hectares), 0)::double precision as total_hectares
FROM projects p
JOIN fields f ON f.project_id = p.id
JOIN lots l ON l.field_id = f.id
GROUP BY p.customer_id, p.id, p.campaign_id, f.id;

-- Vista simplificada para métricas de cosecha
CREATE VIEW dashboard_card_harvest_view AS
SELECT 
    p.customer_id,
    p.id as project_id,
    p.campaign_id,
    f.id as field_id,
    -- Calcular área cosechada basándose en tons
    COALESCE(SUM(CASE WHEN l.tons IS NOT NULL AND l.tons > 0 THEN l.hectares ELSE 0 END), 0)::numeric(14,2) as harvested_area,
    -- Total de hectáreas del proyecto
    COALESCE(SUM(l.hectares), 0)::double precision as total_hectares
FROM projects p
JOIN fields f ON f.project_id = p.id
JOIN lots l ON l.field_id = f.id
GROUP BY p.customer_id, p.id, p.campaign_id, f.id;

-- Vista simplificada para progreso de costos
CREATE VIEW dashboard_card_costs_progress_view AS
SELECT 
    p.customer_id,
    p.id as project_id,
    p.campaign_id,
    f.id as field_id,
    -- Costo presupuestado del proyecto
    COALESCE(p.admin_cost, 0)::numeric(14,2) as budget_cost_usd,
    -- Costos ejecutados (solo labors por ahora)
    COALESCE(
        (SELECT COALESCE(SUM(l.price), 0) FROM labors l WHERE l.project_id = p.id),
        0
    )::numeric(14,2) as executed_costs_usd
FROM projects p
JOIN fields f ON f.project_id = p.id
GROUP BY p.customer_id, p.id, p.campaign_id, f.id, p.admin_cost;

-- Vista simplificada para contribuciones de inversores
CREATE VIEW dashboard_card_contributions_view AS
SELECT 
    p.customer_id,
    p.id as project_id,
    p.campaign_id,
    f.id as field_id,
    -- Progreso de contribuciones (simplificado)
    100.0::numeric(14,2) as contributions_progress_pct,
    -- Breakdown de inversores
    jsonb_build_object(
        'investor_id', pi.investor_id,
        'investor_name', i.name,
        'percentage', pi.percentage
    ) as investors_breakdown
FROM projects p
JOIN fields f ON f.project_id = p.id
JOIN project_investors pi ON pi.project_id = p.id
JOIN investors i ON i.id = pi.investor_id;

-- Vista simplificada para resultado operativo
CREATE VIEW dashboard_card_operating_result_view AS
SELECT 
    p.customer_id,
    p.id as project_id,
    p.campaign_id,
    f.id as field_id,
    -- Ingresos (simplificado)
    COALESCE(
        (SELECT SUM(l.tons * 200) FROM lots l WHERE l.field_id = f.id AND l.tons IS NOT NULL AND l.tons > 0),
        0
    )::numeric(14,2) as income_usd,
    -- Costos directos ejecutados
    COALESCE(
        (SELECT COALESCE(SUM(l.price), 0) FROM labors l WHERE l.project_id = p.id),
        0
    )::numeric(14,2) as direct_costs_executed_usd,
    -- Resultado operativo
    COALESCE(
        (SELECT SUM(l.tons * 200) FROM lots l WHERE l.field_id = f.id AND l.tons IS NOT NULL AND l.tons > 0),
        0
    ) - COALESCE(
        (SELECT COALESCE(SUM(l.price), 0) FROM labors l WHERE l.project_id = p.id),
        0
    ) as operating_result_usd,
    -- Porcentaje de resultado
    CASE 
        WHEN COALESCE((SELECT COALESCE(SUM(l.price), 0) FROM labors l WHERE l.project_id = p.id), 0) > 0 THEN
            ROUND((
                (COALESCE((SELECT SUM(l.tons * 200) FROM lots l WHERE l.field_id = f.id AND l.tons IS NOT NULL AND l.tons > 0), 0) -
                 COALESCE((SELECT COALESCE(SUM(l.price), 0) FROM labors l WHERE l.project_id = p.id), 0)
                ) / COALESCE((SELECT COALESCE(SUM(l.price), 0) FROM labors l WHERE l.project_id = p.id), 0) * 100
            )::numeric, 2)
        ELSE 0 
    END as operating_result_pct
FROM projects p
JOIN fields f ON f.project_id = p.id
GROUP BY p.customer_id, p.id, p.campaign_id, f.id;
