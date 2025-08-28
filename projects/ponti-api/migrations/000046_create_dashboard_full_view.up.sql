-- MIGRATION 000046: CREATE DASHBOARD_FULL_VIEW
-- Simplified view providing only essential dashboard data

CREATE OR REPLACE VIEW dashboard_full_view AS
SELECT
  p.id AS project_id,
  p.customer_id,
  p.campaign_id,
  f.id AS field_id,
  
  -- Basic metrics
  COALESCE(SUM(l.hectares), 0) AS total_hectares,
  COALESCE(SUM(CASE WHEN c.name = 'Siembra' THEN w.effective_area ELSE 0 END), 0) AS sowed_area,
  COALESCE(SUM(CASE WHEN c.name = 'Cosecha' THEN w.effective_area ELSE 0 END), 0) AS harvested_area,
  
  -- Progress percentages
  CASE WHEN SUM(l.hectares) > 0 THEN 
    ROUND((SUM(CASE WHEN c.name = 'Siembra' THEN w.effective_area ELSE 0 END) / SUM(l.hectares) * 100)::numeric, 2) 
  ELSE 0 END AS sowing_progress_pct,
  
  CASE WHEN SUM(l.hectares) > 0 THEN 
    ROUND((SUM(CASE WHEN c.name = 'Cosecha' THEN w.effective_area ELSE 0 END) / SUM(l.hectares) * 100)::numeric, 2) 
  ELSE 0 END AS harvest_progress_pct,
  
  -- Labor costs
  COALESCE(SUM(CASE WHEN c.name = 'Siembra' THEN w.effective_area * lb.price ELSE 0 END), 0) AS labors_executed_usd,
  COALESCE(SUM(CASE WHEN c.name = 'Cosecha' THEN w.effective_area * lb.price ELSE 0 END), 0) AS labors_invested_usd,
  
  -- Supply costs (simplified)
  0 AS supplies_executed_usd,
  0 AS seed_executed_usd,
  COALESCE(SUM(CASE WHEN c.name = 'Siembra' THEN w.effective_area * lb.price ELSE 0 END), 0) AS direct_costs_executed_usd,
  0 AS supplies_invested_usd,
  0 AS seed_invested_usd,
  COALESCE(SUM(CASE WHEN c.name = 'Cosecha' THEN w.effective_area * lb.price ELSE 0 END), 0) AS direct_costs_invested_usd,
  
  -- Stock and budget
  0 AS stock_usd,
  2000.0 AS budget_cost_usd,
  CASE WHEN 2000.0 > 0 THEN 
    ROUND((COALESCE(SUM(CASE WHEN c.name = 'Siembra' THEN w.effective_area * lb.price ELSE 0 END), 0) / 2000.0 * 100)::numeric, 2) 
  ELSE 0 END AS costs_progress_pct,
  
  -- Income and structure (simplified)
  0 AS income_usd,
  0 AS rent_usd,
  0 AS structure_usd,
  
  -- Operating result (simplified)
  0 AS operating_result_usd,
  0 AS operating_result_pct,
  
  -- Cost per hectare
  CASE WHEN SUM(l.hectares) > 0 THEN 
    ROUND((COALESCE(SUM(CASE WHEN c.name = 'Siembra' THEN w.effective_area * lb.price ELSE 0 END), 0) / SUM(l.hectares))::numeric, 2) 
  ELSE 0 END AS total_cost_per_hectare,
  
  -- Crops breakdown (simplified)
  '{}'::jsonb AS crops_breakdown

FROM projects p
JOIN fields f ON f.project_id = p.id
LEFT JOIN lots l ON l.field_id = f.id AND l.deleted_at IS NULL
LEFT JOIN workorders w ON w.lot_id = l.id AND w.deleted_at IS NULL
LEFT JOIN labors lb ON lb.id = w.labor_id AND lb.deleted_at IS NULL
LEFT JOIN categories c ON c.id = lb.category_id AND c.deleted_at IS NULL

WHERE p.deleted_at IS NULL AND f.deleted_at IS NULL

GROUP BY p.id, p.customer_id, p.campaign_id, f.id
ORDER BY p.campaign_id, p.id, p.customer_id, f.id;
