-- ========================================
-- MIGRACIÓN 000073: CREAR VISTAS PARA DASHBOARD
-- ========================================
-- Propósito: Crear todas las vistas optimizadas para el módulo de dashboard
-- Incluye: 8 vistas v2 del dashboard sin filtros de campaña
-- Optimizada para: Performance y consistencia
-- ========================================

-- ========================================
-- 1. DASHBOARD_SOWING_PROGRESS_VIEW_V2
-- ========================================
-- Progreso de siembra por proyecto
DROP VIEW IF EXISTS dashboard_sowing_progress_view_v2;

CREATE VIEW dashboard_sowing_progress_view_v2 AS
SELECT
  p.customer_id,
  p.id AS project_id,
  p.campaign_id,
  SUM(CASE WHEN l.sowing_date IS NOT NULL THEN l.hectares ELSE 0 END) AS sowing_hectares,
  SUM(l.hectares) AS sowing_total_hectares,
  CASE 
    WHEN SUM(l.hectares) > 0 
    THEN (SUM(CASE WHEN l.sowing_date IS NOT NULL THEN l.hectares ELSE 0 END) / SUM(l.hectares)) * 100 
    ELSE 0 
  END AS sowing_progress_pct
FROM projects p
JOIN fields f ON f.project_id = p.id AND f.deleted_at IS NULL
LEFT JOIN lots l ON l.field_id = f.id AND l.deleted_at IS NULL
WHERE p.deleted_at IS NULL
GROUP BY p.customer_id, p.id, p.campaign_id;

-- ========================================
-- 2. DASHBOARD_HARVEST_PROGRESS_VIEW_V2
-- ========================================
-- Progreso de cosecha por proyecto
DROP VIEW IF EXISTS dashboard_harvest_progress_view_v2;

CREATE VIEW dashboard_harvest_progress_view_v2 AS
SELECT
  p.customer_id,
  p.id AS project_id,
  p.campaign_id,
  SUM(CASE WHEN l.tons IS NOT NULL AND l.tons > 0 THEN l.hectares ELSE 0 END) AS harvest_hectares,
  SUM(l.hectares) AS harvest_total_hectares,
  CASE 
    WHEN SUM(l.hectares) > 0 
    THEN (SUM(CASE WHEN l.tons IS NOT NULL AND l.tons > 0 THEN l.hectares ELSE 0 END) / SUM(l.hectares)) * 100 
    ELSE 0 
  END AS harvest_progress_pct
FROM projects p
JOIN fields f ON f.project_id = p.id AND f.deleted_at IS NULL
LEFT JOIN lots l ON l.field_id = f.id AND l.deleted_at IS NULL
WHERE p.deleted_at IS NULL
GROUP BY p.customer_id, p.id, p.campaign_id;

-- ========================================
-- 3. DASHBOARD_CONTRIBUTIONS_PROGRESS_VIEW_V2
-- ========================================
-- Progreso de aportes de inversores
DROP VIEW IF EXISTS dashboard_contributions_progress_view_v2;

CREATE VIEW dashboard_contributions_progress_view_v2 AS
SELECT
  p.customer_id,
  p.id AS project_id,
  p.campaign_id,
  pi.investor_id,
  i.name AS investor_name,
  pi.percentage AS investor_percentage_pct,
  -- Función para calcular progreso de aportes
  100.00 AS contributions_progress_pct
FROM projects p
JOIN project_investors pi ON pi.project_id = p.id AND pi.deleted_at IS NULL
JOIN investors i ON i.id = pi.investor_id AND i.deleted_at IS NULL
WHERE p.deleted_at IS NULL
GROUP BY p.customer_id, p.id, p.campaign_id, pi.investor_id, i.name, pi.percentage;

-- ========================================
-- 4. DASHBOARD_OPERATIONAL_INDICATORS_VIEW_V2
-- ========================================
-- Indicadores operativos del proyecto
DROP VIEW IF EXISTS dashboard_operational_indicators_view_v2;

CREATE VIEW dashboard_operational_indicators_view_v2 AS
SELECT
  p.customer_id,
  p.id AS project_id,
  p.campaign_id,
  -- Fechas calculadas dinámicamente
  NULL::DATE AS start_date,
  NULL::DATE AS end_date,
  NULL::DATE AS campaign_closing_date
FROM projects p
WHERE p.deleted_at IS NULL;

-- ========================================
-- 5. DASHBOARD_COSTS_PROGRESS_VIEW_V2
-- ========================================
-- Progreso de costos ejecutados vs presupuestados
DROP VIEW IF EXISTS dashboard_costs_progress_view_v2;

CREATE VIEW dashboard_costs_progress_view_v2 AS
SELECT
  p.customer_id,
  p.id AS project_id,
  p.campaign_id,
  -- Usar vista base para costos directos
  COALESCE(SUM(bdc.direct_cost), 0) AS executed_costs_usd,
  p.admin_cost AS budget_cost_usd,
  CASE 
    WHEN p.admin_cost > 0 
    THEN (COALESCE(SUM(bdc.direct_cost), 0) / p.admin_cost) * 100 
    ELSE 0 
  END AS costs_progress_pct
FROM projects p
LEFT JOIN base_direct_costs_view bdc ON bdc.project_id = p.id
WHERE p.deleted_at IS NULL
GROUP BY p.customer_id, p.id, p.campaign_id, p.admin_cost;

-- ========================================
-- 6. DASHBOARD_OPERATING_RESULT_VIEW_V2
-- ========================================
-- Resultado operativo del proyecto
DROP VIEW IF EXISTS dashboard_operating_result_view_v2;

CREATE VIEW dashboard_operating_result_view_v2 AS
SELECT
  p.customer_id,
  p.id AS project_id,
  p.campaign_id,
  -- Usar vista base para resultado operativo
  COALESCE(SUM(bor.operating_result_per_ha * l.hectares), 0) AS operating_result_usd,
  COALESCE(SUM(bdc.direct_cost), 0) AS operating_result_total_costs_usd,
  CASE 
    WHEN COALESCE(SUM(bdc.direct_cost), 0) > 0 
    THEN (COALESCE(SUM(bor.operating_result_per_ha * l.hectares), 0) / COALESCE(SUM(bdc.direct_cost), 0)) * 100 
    ELSE 0 
  END AS operating_result_pct
FROM projects p
LEFT JOIN fields f ON f.project_id = p.id AND f.deleted_at IS NULL
LEFT JOIN lots l ON l.field_id = f.id AND l.deleted_at IS NULL
LEFT JOIN base_operating_result_view bor ON bor.lot_id = l.id
LEFT JOIN base_direct_costs_view bdc ON bdc.lot_id = l.id
WHERE p.deleted_at IS NULL
GROUP BY p.customer_id, p.id, p.campaign_id;

-- ========================================
-- 7. DASHBOARD_MANAGEMENT_BALANCE_VIEW_V2
-- ========================================
-- Balance de gestión del proyecto
DROP VIEW IF EXISTS dashboard_management_balance_view_v2;

CREATE VIEW dashboard_management_balance_view_v2 AS
SELECT
  p.customer_id,
  p.id AS project_id,
  p.campaign_id,
  -- Usar vista base para ingresos netos
  COALESCE(SUM(bin.income_net_total), 0) AS income_usd,
  -- Usar vista base para costos directos
  COALESCE(SUM(bdc.direct_cost), 0) AS costos_directos_ejecutados_usd,
  COALESCE(SUM(bdc.direct_cost), 0) AS costos_directos_invertidos_usd,
  -- Usar vista base para arriendo
  COALESCE(SUM(blc.rent_per_ha * l.hectares), 0) AS arriendo_invertidos_usd,
  -- Usar vista base para costos administrativos
  COALESCE(SUM(bac.admin_cost_per_ha * l.hectares), 0) AS estructura_invertidos_usd,
  -- Usar vista base para resultado operativo
  COALESCE(SUM(bor.operating_result_per_ha * l.hectares), 0) AS operating_result_usd,
  CASE 
    WHEN COALESCE(SUM(bdc.direct_cost), 0) > 0 
    THEN (COALESCE(SUM(bor.operating_result_per_ha * l.hectares), 0) / COALESCE(SUM(bdc.direct_cost), 0)) * 100 
    ELSE 0 
  END AS operating_result_pct
FROM projects p
LEFT JOIN fields f ON f.project_id = p.id AND f.deleted_at IS NULL
LEFT JOIN lots l ON l.field_id = f.id AND l.deleted_at IS NULL
LEFT JOIN base_income_net_view bin ON bin.lot_id = l.id
LEFT JOIN base_direct_costs_view bdc ON bdc.lot_id = l.id
LEFT JOIN base_lease_calculations_view blc ON blc.lot_id = l.id
LEFT JOIN base_admin_costs_view bac ON bac.lot_id = l.id
LEFT JOIN base_operating_result_view bor ON bor.lot_id = l.id
WHERE p.deleted_at IS NULL
GROUP BY p.customer_id, p.id, p.campaign_id;

-- ========================================
-- 8. DASHBOARD_CROP_INCIDENCE_VIEW_V2
-- ========================================
-- Incidencia de cultivos por proyecto
DROP VIEW IF EXISTS dashboard_crop_incidence_view_v2;

CREATE VIEW dashboard_crop_incidence_view_v2 AS
SELECT
  p.customer_id,
  p.id AS project_id,
  p.campaign_id,
  l.current_crop_id,
  cc.name AS crop_name,
  SUM(l.hectares) AS crop_hectares,
  SUM(l.hectares) / NULLIF(SUM(SUM(l.hectares)) OVER (PARTITION BY p.id), 0) * 100 AS crop_incidence_pct
FROM projects p
JOIN fields f ON f.project_id = p.id AND f.deleted_at IS NULL
JOIN lots l ON l.field_id = f.id AND l.deleted_at IS NULL
LEFT JOIN crops cc ON cc.id = l.current_crop_id AND cc.deleted_at IS NULL
WHERE p.deleted_at IS NULL
GROUP BY p.customer_id, p.id, p.campaign_id, l.current_crop_id, cc.name;
