-- ========================================
-- MIGRACIÓN 000082: CREAR VISTA VIEWS_FIXES
-- ========================================
-- Propósito: Vista centralizada para todos los fixes de vistas
-- Incluye: fix_labors_list_duplication, y otros fixes futuros
-- ========================================

-- ========================================
-- 1. CREAR FUNCIÓN PARA OBTENER VALOR DE DÓLAR POR MES
-- ========================================
CREATE OR REPLACE FUNCTION get_project_dollar_value(p_project_id BIGINT, p_month VARCHAR)
RETURNS DECIMAL AS $$
BEGIN
  RETURN (
    SELECT average_value 
    FROM project_dollar_values 
    WHERE project_id = p_project_id 
      AND month = p_month 
      AND deleted_at IS NULL
    LIMIT 1
  );
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- ========================================
-- 2. CREAR VISTA VIEWS_FIXES CON TODOS LOS FIXES
-- ========================================
CREATE OR REPLACE VIEW views_fixes AS
-- ========================================
-- FIX 1: LABORS LIST SIN DUPLICACIÓN
-- ========================================
-- Problema: fix_labors_list duplicaba labores cuando había múltiples meses
-- de dólar promedio para el mismo proyecto
-- Solución: Usar función que obtiene el valor específico del mes
-- ========================================
SELECT
    'fix_labors_list' AS fix_name,
    'Corrige duplicación de labores por múltiples meses de dólar promedio' AS description,
    'workorders' AS affected_table,
    'fix_labors_list_duplication' AS fix_type
UNION ALL
-- ========================================
-- FIX 2: PLACEHOLDER PARA FUTUROS FIXES
-- ========================================
SELECT
    'placeholder_fix' AS fix_name,
    'Placeholder para futuros fixes de vistas' AS description,
    'various' AS affected_table,
    'placeholder' AS fix_type;

-- ========================================
-- 3. RECREAR VISTA FIX_LABORS_LIST CORREGIDA
-- ========================================
DROP VIEW IF EXISTS fix_labors_list;

CREATE OR REPLACE VIEW fix_labors_list AS
SELECT
    w.id AS workorder_id,
    w.number AS workorder_number,
    w.date,
    p.id AS project_id,
    f.id AS field_id,
    p.name AS project_name,
    f.name AS field_name,
    c.name AS crop_name,
    lb.name AS labor_name,
    lc.name AS category_name,
    w.contractor,
    w.effective_area AS surface_ha,
    lb.price AS cost_ha,
    lb.contractor_name,
    inv.name AS investor_name,
    
    -- FIX: Usar función que obtiene el valor específico del mes (evita duplicación)
    -- NOTA: Esta vista ahora usa un mes por defecto, pero el código Go debe filtrar por mes
    COALESCE(
        get_project_dollar_value(w.project_id, '01'), -- Mes por defecto (formato correcto)
        get_default_fx_rate()
    ) AS usd_avg_value,
    
    -- Cálculos corregidos según especificaciones:
    
    -- Total neto: Costo labor * superficie (en USD)
    (lb.price * w.effective_area) AS net_total,
    
    -- Total IVA: Usar porcentaje configurable desde app_parameters
    (lb.price * w.effective_area * get_iva_percentage()) AS total_iva,
    
    -- Costo U$/Ha en pesos: Costo Ha * dólar promedio (mostrar en pesos)
    (lb.price * COALESCE(
        get_project_dollar_value(w.project_id, '01'), -- Mes por defecto (formato correcto)
        get_default_fx_rate()
    )) AS usd_cost_ha,
    
    -- Total U Neto en pesos: Total costo en pesos * Has
    (lb.price * COALESCE(
        get_project_dollar_value(w.project_id, '01'), -- Mes por defecto (formato correcto)
        get_default_fx_rate()
    ) * w.effective_area) AS usd_net_total,
    
    -- Campos de factura
    i.id AS invoice_id,
    i.number AS invoice_number,
    i.company AS invoice_company,
    i.date AS invoice_date,
    i.status AS invoice_status
    
FROM workorders w
INNER JOIN projects p ON w.project_id = p.id
INNER JOIN fields f ON w.field_id = f.id
INNER JOIN crops c ON w.crop_id = c.id
INNER JOIN labors lb ON w.labor_id = lb.id
INNER JOIN categories lc ON lb.category_id = lc.id
INNER JOIN investors inv ON w.investor_id = inv.id
LEFT JOIN invoices i ON i.work_order_id = w.id
WHERE w.deleted_at IS NULL
    AND p.deleted_at IS NULL;

-- ========================================
-- 4. COMENTARIOS EXPLICATIVOS
-- ========================================
-- La vista fix_labors_list ahora usa get_project_dollar_value() que evita duplicación
-- El código Go debe modificar la consulta para usar el mes específico:
-- Ejemplo: WHERE usd_avg_value = get_project_dollar_value(project_id, '2024-01')
-- O mejor aún, hacer la consulta directamente con el filtro por mes
