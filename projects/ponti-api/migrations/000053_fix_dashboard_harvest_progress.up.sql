-- Migración 000053: Corregir SOLO el avance de cosecha
-- Enfoque: Vista que calcula correctamente el porcentaje de avance de cosecha
-- Partir de la migración 000050 (no tocar la 000051 ni 000052)

DROP VIEW IF EXISTS dashboard_harvest_progress_view;

CREATE OR REPLACE VIEW dashboard_harvest_progress_view AS
SELECT
    p.customer_id,
    p.id AS project_id,
    p.campaign_id,
    COALESCE(SUM(CASE WHEN l.tons > 0 THEN l.hectares ELSE 0 END),0) AS harvest_hectares,
    COALESCE(SUM(l.hectares),0) AS harvest_total_hectares,
    CASE WHEN SUM(l.hectares)>0
         THEN (SUM(CASE WHEN l.tons > 0 THEN l.hectares ELSE 0 END)::numeric / SUM(l.hectares)::numeric * 100)
    ELSE 0 END AS harvest_progress_pct
FROM projects p
JOIN fields f ON f.project_id=p.id
JOIN lots l   ON l.field_id=f.id
GROUP BY p.customer_id,p.id,p.campaign_id;


