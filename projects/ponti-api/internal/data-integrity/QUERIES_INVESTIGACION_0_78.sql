-- ========================================
-- QUERIES DE INVESTIGACIÓN: Diferencia $0.78
-- ========================================
-- Proyecto: Ponti Backend
-- Fecha: 2025-10-21
-- Propósito: Queries para investigar diferencias entre total_used y final_dose × effective_area

-- ========================================
-- 1. VERIFICACIÓN BÁSICA DE DIFERENCIAS
-- ========================================
-- Compara los dos métodos de cálculo (RAW vs SSOT)

SELECT 
  SUM((wo.effective_area * l.price) + COALESCE(wi.total_used * s.price, 0)) as raw_total,
  SUM((wo.effective_area * l.price) + COALESCE(wi.final_dose * wo.effective_area * s.price, 0)) as calculated_total,
  SUM((wo.effective_area * l.price) + COALESCE(wi.total_used * s.price, 0)) - 
    SUM((wo.effective_area * l.price) + COALESCE(wi.final_dose * wo.effective_area * s.price, 0)) as difference
FROM workorders wo
JOIN labors l ON l.id = wo.labor_id
LEFT JOIN workorder_items wi ON wi.workorder_id = wo.id AND wi.deleted_at IS NULL
LEFT JOIN supplies s ON s.id = wi.supply_id AND s.deleted_at IS NULL
WHERE wo.project_id = 11 
  AND wo.deleted_at IS NULL;

-- ========================================
-- 2. ITEMS CON DIFERENCIAS (DETALLADO)
-- ========================================
-- Lista todos los items donde total_used ≠ final_dose × effective_area

SELECT 
  wo.id as workorder_id,
  wo.number as workorder_number,
  wo.date as workorder_date,
  wi.id as item_id,
  s.name as supply_name,
  s.price as unit_price,
  wi.total_used,
  wi.final_dose,
  wo.effective_area,
  -- Cálculo esperado
  wi.final_dose * wo.effective_area as calculated_total_used,
  -- Diferencia en unidades
  wi.total_used - (wi.final_dose * wo.effective_area) as units_difference,
  -- Diferencia en costo
  (wi.total_used - (wi.final_dose * wo.effective_area)) * s.price as cost_difference,
  -- Porcentaje de diferencia
  CASE 
    WHEN wi.final_dose * wo.effective_area > 0 
    THEN ((wi.total_used - (wi.final_dose * wo.effective_area)) / (wi.final_dose * wo.effective_area) * 100)
    ELSE 0 
  END as pct_difference
FROM workorders wo
JOIN workorder_items wi ON wi.workorder_id = wo.id AND wi.deleted_at IS NULL
JOIN supplies s ON s.id = wi.supply_id AND s.deleted_at IS NULL
WHERE wo.project_id = 11 
  AND wo.deleted_at IS NULL 
  AND ABS(wi.total_used - (wi.final_dose * wo.effective_area)) > 0.0001
ORDER BY ABS((wi.total_used - (wi.final_dose * wo.effective_area)) * s.price) DESC;

-- ========================================
-- 3. RESUMEN POR TIPO DE INSUMO
-- ========================================
-- Agrupa las diferencias por categoría de insumo

SELECT 
  c.name as category,
  COUNT(*) as items_count,
  SUM((wi.total_used - (wi.final_dose * wo.effective_area)) * s.price) as total_cost_diff,
  AVG((wi.total_used - (wi.final_dose * wo.effective_area)) * s.price) as avg_cost_diff,
  MIN((wi.total_used - (wi.final_dose * wo.effective_area)) * s.price) as min_cost_diff,
  MAX((wi.total_used - (wi.final_dose * wo.effective_area)) * s.price) as max_cost_diff
FROM workorders wo
JOIN workorder_items wi ON wi.workorder_id = wo.id AND wi.deleted_at IS NULL
JOIN supplies s ON s.id = wi.supply_id AND s.deleted_at IS NULL
JOIN categories c ON c.id = s.category_id
WHERE wo.project_id = 11 
  AND wo.deleted_at IS NULL 
  AND ABS(wi.total_used - (wi.final_dose * wo.effective_area)) > 0.0001
GROUP BY c.name
ORDER BY ABS(SUM((wi.total_used - (wi.final_dose * wo.effective_area)) * s.price)) DESC;

-- ========================================
-- 4. RESUMEN POR WORKORDER
-- ========================================
-- Agrupa las diferencias por orden de trabajo

SELECT 
  wo.id,
  wo.number,
  wo.date,
  COUNT(wi.id) as items_with_diff,
  SUM((wi.total_used - (wi.final_dose * wo.effective_area)) * s.price) as total_cost_diff
FROM workorders wo
JOIN workorder_items wi ON wi.workorder_id = wo.id AND wi.deleted_at IS NULL
JOIN supplies s ON s.id = wi.supply_id AND s.deleted_at IS NULL
WHERE wo.project_id = 11 
  AND wo.deleted_at IS NULL 
  AND ABS(wi.total_used - (wi.final_dose * wo.effective_area)) > 0.0001
GROUP BY wo.id, wo.number, wo.date
ORDER BY ABS(SUM((wi.total_used - (wi.final_dose * wo.effective_area)) * s.price)) DESC;

-- ========================================
-- 5. COMPARACIÓN RAW vs SSOT (COMO EN CONTROLES)
-- ========================================
-- Replica exactamente los cálculos de los controles de integridad

WITH raw_calc AS (
  SELECT COALESCE(SUM(labor_cost + supply_cost), 0) AS total_cost
  FROM (
    SELECT 
      wo.id,
      (wo.effective_area * l.price) AS labor_cost,
      COALESCE((
        SELECT SUM(wi.total_used * s.price)
        FROM workorder_items wi
        JOIN supplies s ON s.id = wi.supply_id
        WHERE wi.workorder_id = wo.id 
          AND wi.deleted_at IS NULL
      ), 0) AS supply_cost
    FROM workorders wo
    JOIN labors l ON l.id = wo.labor_id
    WHERE wo.deleted_at IS NULL
      AND wo.project_id = 11
  ) AS costs
),
ssot_calc AS (
  SELECT COALESCE(SUM(wm.direct_cost_usd), 0) AS total_cost
  FROM v3_workorder_metrics wm
  WHERE wm.project_id = 11
)
SELECT 
  raw_calc.total_cost as raw_method_left,
  ssot_calc.total_cost as ssot_method_right,
  raw_calc.total_cost - ssot_calc.total_cost as difference
FROM raw_calc, ssot_calc;

-- ========================================
-- 6. ESTADÍSTICAS GENERALES
-- ========================================
-- Resumen estadístico de las diferencias

SELECT 
  COUNT(*) as total_items_with_diff,
  SUM((wi.total_used - (wi.final_dose * wo.effective_area)) * s.price) as total_cost_diff,
  AVG((wi.total_used - (wi.final_dose * wo.effective_area)) * s.price) as avg_cost_diff,
  MIN((wi.total_used - (wi.final_dose * wo.effective_area)) * s.price) as min_cost_diff,
  MAX((wi.total_used - (wi.final_dose * wo.effective_area)) * s.price) as max_cost_diff,
  STDDEV((wi.total_used - (wi.final_dose * wo.effective_area)) * s.price) as stddev_cost_diff
FROM workorders wo
JOIN workorder_items wi ON wi.workorder_id = wo.id AND wi.deleted_at IS NULL
JOIN supplies s ON s.id = wi.supply_id AND s.deleted_at IS NULL
WHERE wo.project_id = 11 
  AND wo.deleted_at IS NULL 
  AND ABS(wi.total_used - (wi.final_dose * wo.effective_area)) > 0.0001;

-- ========================================
-- 7. ITEMS PERFECTAMENTE SINCRONIZADOS
-- ========================================
-- Cuántos items SÍ están correctos (para comparar)

SELECT 
  COUNT(*) as items_in_sync
FROM workorders wo
JOIN workorder_items wi ON wi.workorder_id = wo.id AND wi.deleted_at IS NULL
WHERE wo.project_id = 11 
  AND wo.deleted_at IS NULL 
  AND ABS(wi.total_used - (wi.final_dose * wo.effective_area)) <= 0.0001;

-- ========================================
-- 8. ANÁLISIS DE REDONDEO
-- ========================================
-- Detecta patrones de redondeo (ej: redondeado a 1 decimal, 2 decimales, etc.)

SELECT 
  wo.id,
  wo.number,
  wi.id as item_id,
  s.name,
  wi.total_used,
  wi.final_dose * wo.effective_area as calculated,
  wi.total_used - (wi.final_dose * wo.effective_area) as diff,
  -- Compara con redondeos comunes
  ROUND((wi.final_dose * wo.effective_area)::numeric, 1) as rounded_1_decimal,
  ROUND((wi.final_dose * wo.effective_area)::numeric, 2) as rounded_2_decimals,
  ROUND((wi.final_dose * wo.effective_area)::numeric, 3) as rounded_3_decimals,
  -- Verifica si coincide con algún redondeo
  CASE 
    WHEN ABS(wi.total_used - ROUND((wi.final_dose * wo.effective_area)::numeric, 1)) < 0.0001 
      THEN 'Redondeado a 1 decimal'
    WHEN ABS(wi.total_used - ROUND((wi.final_dose * wo.effective_area)::numeric, 2)) < 0.0001 
      THEN 'Redondeado a 2 decimales'
    WHEN ABS(wi.total_used - ROUND((wi.final_dose * wo.effective_area)::numeric, 3)) < 0.0001 
      THEN 'Redondeado a 3 decimales'
    ELSE 'Otro patrón'
  END as rounding_pattern
FROM workorders wo
JOIN workorder_items wi ON wi.workorder_id = wo.id AND wi.deleted_at IS NULL
JOIN supplies s ON s.id = wi.supply_id AND s.deleted_at IS NULL
WHERE wo.project_id = 11 
  AND wo.deleted_at IS NULL 
  AND ABS(wi.total_used - (wi.final_dose * wo.effective_area)) > 0.0001
ORDER BY ABS(wi.total_used - (wi.final_dose * wo.effective_area)) DESC;

-- ========================================
-- 9. AUDITORÍA: ¿CUÁNDO SE CREARON/MODIFICARON?
-- ========================================
-- Verifica si hay patrones temporales en las diferencias

SELECT 
  DATE(wo.created_at) as creation_date,
  COUNT(*) as items_with_diff,
  SUM((wi.total_used - (wi.final_dose * wo.effective_area)) * s.price) as total_diff
FROM workorders wo
JOIN workorder_items wi ON wi.workorder_id = wo.id AND wi.deleted_at IS NULL
JOIN supplies s ON s.id = wi.supply_id AND s.deleted_at IS NULL
WHERE wo.project_id = 11 
  AND wo.deleted_at IS NULL 
  AND ABS(wi.total_used - (wi.final_dose * wo.effective_area)) > 0.0001
GROUP BY DATE(wo.created_at)
ORDER BY DATE(wo.created_at);

-- ========================================
-- 10. CORRECCIÓN PROPUESTA (NO EJECUTAR SIN AUTORIZACIÓN)
-- ========================================
-- Query para corregir total_used si se decide que debe ser exacto
-- ⚠️  SOLO PARA REFERENCIA - NO EJECUTAR SIN APROBACIÓN

/*
UPDATE workorder_items wi
SET total_used = wi.final_dose * (
  SELECT wo.effective_area 
  FROM workorders wo 
  WHERE wo.id = wi.workorder_id
)
WHERE workorder_id IN (
  SELECT wo.id 
  FROM workorders wo 
  WHERE wo.project_id = 11 
    AND wo.deleted_at IS NULL
)
AND ABS(wi.total_used - (wi.final_dose * (
  SELECT wo.effective_area 
  FROM workorders wo 
  WHERE wo.id = wi.workorder_id
))) > 0.0001;
*/

-- ========================================
-- 11. EXPORTAR PARA EXCEL
-- ========================================
-- Query simplificado para exportar y analizar en Excel

SELECT 
  wo.number as "Orden",
  wo.date as "Fecha",
  s.name as "Insumo",
  wi.total_used as "Total Usado (Almacenado)",
  wi.final_dose as "Dosis Final",
  wo.effective_area as "Área Efectiva",
  wi.final_dose * wo.effective_area as "Total Calculado",
  wi.total_used - (wi.final_dose * wo.effective_area) as "Diferencia Unidades",
  s.price as "Precio Unitario",
  (wi.total_used - (wi.final_dose * wo.effective_area)) * s.price as "Diferencia USD"
FROM workorders wo
JOIN workorder_items wi ON wi.workorder_id = wo.id AND wi.deleted_at IS NULL
JOIN supplies s ON s.id = wi.supply_id AND s.deleted_at IS NULL
WHERE wo.project_id = 11 
  AND wo.deleted_at IS NULL 
  AND ABS(wi.total_used - (wi.final_dose * wo.effective_area)) > 0.0001
ORDER BY ABS((wi.total_used - (wi.final_dose * wo.effective_area)) * s.price) DESC;

