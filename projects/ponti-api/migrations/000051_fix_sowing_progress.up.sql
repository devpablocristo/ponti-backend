-- Migración 000051: Corregir SOLO el avance de siembra
-- Enfoque: Vista simplificada que calcula correctamente el porcentaje de siembra

DROP VIEW IF EXISTS dashboard_view;

CREATE OR REPLACE VIEW dashboard_view AS
SELECT
    p.customer_id,
    p.id AS project_id,
    p.campaign_id,
    f.id AS field_id,
    COALESCE(total.total_hectares, 0) AS sowing_total_hectares,
    COALESCE(sown.sown_hectares, 0) AS sowing_hectares,
    CASE 
        WHEN COALESCE(total.total_hectares, 0) > 0 THEN
            (COALESCE(sown.sown_hectares, 0)::numeric / total.total_hectares * 100)
        ELSE 0 
    END AS sowing_progress_percent
FROM projects p
JOIN fields f ON f.project_id = p.id
LEFT JOIN (
    SELECT field_id, SUM(hectares) AS total_hectares
    FROM lots
    GROUP BY field_id
) total ON total.field_id = f.id
LEFT JOIN (
    SELECT field_id, SUM(hectares) AS sown_hectares
    FROM lots
    WHERE sowing_date IS NOT NULL
    GROUP BY field_id
) sown ON sown.field_id = f.id;