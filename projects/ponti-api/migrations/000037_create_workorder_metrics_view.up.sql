CREATE OR REPLACE VIEW workorder_metrics_view AS
WITH
-- =======================
-- BASE DE ÓRDENES DE TRABAJO (pre-agregada para evitar duplicados)
-- =======================
workorder_base AS (
  SELECT
    w.id              AS workorder_id,
    w.project_id,
    w.field_id,
    w.effective_area,
    COALESCE(lb.price, 0) AS labor_price_per_ha,
    -- Pre-calcular costo de labor por workorder
    COALESCE(lb.price, 0) * w.effective_area AS labor_cost_per_wo
  FROM workorders w
  JOIN labors lb ON lb.id = w.labor_id
  WHERE w.deleted_at IS NULL 
    AND lb.deleted_at IS NULL
    AND w.effective_area IS NOT NULL
    AND w.effective_area > 0
),

-- =======================
-- INSUMOS CENTRALIZADOS (evitar duplicados por unidades)
-- =======================
supply_aggregation AS (
  SELECT
    w.id         AS workorder_id,
    w.project_id,
    w.field_id,
    -- Agregación por tipo de unidad (evitar múltiples JOINs)
    SUM(CASE 
      WHEN u.name = 'Lts' THEN wi.total_used 
      ELSE 0 
    END) AS total_liters,
    SUM(CASE 
      WHEN u.name = 'Kg' THEN wi.total_used 
      ELSE 0 
    END) AS total_kilograms,
    SUM(CASE 
      WHEN u.name = 'Ha' THEN wi.total_used 
      ELSE 0 
    END) AS total_hectares,
    -- Costo total de insumos por workorder (dose * price * effective_area)
    SUM(COALESCE(wi.final_dose, 0) * COALESCE(s.price, 0) * w.effective_area) AS total_supplies_cost
  FROM workorders w
  JOIN workorder_items wi ON wi.workorder_id = w.id AND wi.deleted_at IS NULL
  JOIN supplies s         ON s.id = wi.supply_id AND s.deleted_at IS NULL
  LEFT JOIN units u       ON u.id = s.unit_id AND u.deleted_at IS NULL
  WHERE w.deleted_at IS NULL
    AND w.effective_area IS NOT NULL
    AND w.effective_area > 0
  GROUP BY w.id, w.project_id, w.field_id
),

-- =======================
-- MÉTRICAS POR PROYECTO/CAMPO (agregación final)
-- =======================
field_metrics AS (
  SELECT
    wb.project_id,
    wb.field_id,
    -- Superficie total (hectáreas)
    SUM(wb.effective_area) AS total_surface_ha,
    -- Cantidades de insumos
    SUM(COALESCE(sa.total_liters, 0))     AS total_liters,
    SUM(COALESCE(sa.total_kilograms, 0))  AS total_kilograms,
    SUM(COALESCE(sa.total_hectares, 0))   AS total_hectares_units,
    -- Costos
    SUM(wb.labor_cost_per_wo) AS total_labor_cost,
    SUM(COALESCE(sa.total_supplies_cost, 0)) AS total_supplies_cost,
    -- Costo directo total
    SUM(wb.labor_cost_per_wo + COALESCE(sa.total_supplies_cost, 0)) AS total_direct_cost,
    -- Métricas calculadas
    COUNT(DISTINCT wb.workorder_id) AS total_workorders,
    -- Costo promedio por hectárea
    CASE 
      WHEN SUM(wb.effective_area) > 0 
      THEN SUM(wb.labor_cost_per_wo + COALESCE(sa.total_supplies_cost, 0)) / SUM(wb.effective_area)
      ELSE 0 
    END AS cost_per_hectare,
    -- Porcentaje de costos por tipo
    CASE 
      WHEN SUM(wb.labor_cost_per_wo + COALESCE(sa.total_supplies_cost, 0)) > 0
      THEN (SUM(wb.labor_cost_per_wo) / SUM(wb.labor_cost_per_wo + COALESCE(sa.total_supplies_cost, 0))) * 100
      ELSE 0 
    END AS labor_cost_percentage,
    CASE 
      WHEN SUM(wb.labor_cost_per_wo + COALESCE(sa.total_supplies_cost, 0)) > 0
      THEN (SUM(COALESCE(sa.total_supplies_cost, 0)) / SUM(wb.labor_cost_per_wo + COALESCE(sa.total_supplies_cost, 0))) * 100
      ELSE 0 
    END AS supplies_cost_percentage
  FROM workorder_base wb
  LEFT JOIN supply_aggregation sa ON sa.workorder_id = wb.workorder_id
  GROUP BY wb.project_id, wb.field_id
)

SELECT
  fm.project_id,
  fm.field_id,
  -- COMPATIBILIDAD: Mantener nombres originales para el código Go
  fm.total_surface_ha AS surface_ha,           -- ← NOMBRE ORIGINAL
  fm.total_liters AS liters,                   -- ← NOMBRE ORIGINAL  
  fm.total_kilograms AS kilograms,             -- ← NOMBRE ORIGINAL
  fm.total_direct_cost AS direct_cost,         -- ← NOMBRE ORIGINAL
  -- Campos adicionales optimizados
  fm.total_hectares_units,
  fm.total_labor_cost,
  fm.total_supplies_cost,
  fm.total_workorders,
  fm.cost_per_hectare,
  fm.labor_cost_percentage,
  fm.supplies_cost_percentage,
  -- Indicadores de eficiencia
  CASE 
    WHEN fm.total_surface_ha > 0 AND fm.total_workorders > 0
    THEN fm.total_surface_ha / fm.total_workorders
    ELSE 0 
  END AS avg_surface_per_workorder,
  -- Densidad de insumos por hectárea
  CASE 
    WHEN fm.total_surface_ha > 0
    THEN fm.total_liters / fm.total_surface_ha
    ELSE 0 
  END AS liters_per_hectare,
  CASE 
    WHEN fm.total_surface_ha > 0
    THEN fm.total_kilograms / fm.total_surface_ha
    ELSE 0 
  END AS kilograms_per_hectare
FROM field_metrics fm
WHERE fm.total_surface_ha > 0; -- Solo campos con superficie válida

-- =======================
-- ÍNDICES OPTIMIZADOS PARA CLOUD SQL (GCP)
-- =======================

-- Índices parciales para soft-delete (estándar en GCP)
CREATE INDEX IF NOT EXISTS idx_workorders_metrics_notdel 
  ON workorders(project_id, field_id, effective_area) 
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_workorder_items_wo_notdel 
  ON workorder_items(workorder_id, supply_id) 
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_supplies_metrics_notdel 
  ON supplies(id, price, unit_id) 
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_labors_metrics_notdel 
  ON labors(id, price) 
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_units_metrics_notdel 
  ON units(id, name) 
  WHERE deleted_at IS NULL;

-- Índice compuesto para JOINs frecuentes
CREATE INDEX IF NOT EXISTS idx_workorders_metrics_composite 
  ON workorders(project_id, field_id, labor_id, effective_area) 
  WHERE deleted_at IS NULL;

-- Índice para agregaciones por proyecto/campo
CREATE INDEX IF NOT EXISTS idx_workorders_grouping 
  ON workorders(project_id, field_id, effective_area) 
  WHERE deleted_at IS NULL AND effective_area > 0;

-- Índice para optimizar agregaciones de supplies
CREATE INDEX IF NOT EXISTS idx_workorder_items_supply_metrics 
  ON workorder_items(workorder_id, supply_id, total_used, final_dose) 
  WHERE deleted_at IS NULL;
