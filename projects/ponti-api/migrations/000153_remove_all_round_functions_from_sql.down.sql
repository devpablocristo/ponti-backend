-- ========================================
-- MIGRATION 000153: REMOVE ALL ROUND FUNCTIONS FROM SQL (DOWN)
-- ========================================
-- 
-- Purpose: Revertir eliminación de ROUND() en las vistas y funciones SQL
-- Date: 2025-01-27
-- Author: System
-- 
-- Note: Code in English, comments in Spanish.

-- ========================================
-- REVERTIR FUNCIÓN DE PORCENTAJES CON REDONDEO
-- ========================================
-- Nota: Volver a usar redondeo a 3 decimales
CREATE OR REPLACE FUNCTION v3_calc.percentage_rounded(numeric, numeric) RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT ROUND(v3_calc.safe_div($1, $2) * 100, 3)
$$;

-- ========================================
-- REVERTIR VISTA v3_dashboard_crop_incidence CON REDONDEO
-- ========================================
-- Nota: Volver a usar redondeo a 3 decimales
CREATE OR REPLACE VIEW public.v3_dashboard_crop_incidence AS
WITH lot_base AS (
  SELECT
    l.id               AS lot_id,
    f.project_id       AS project_id,
    l.current_crop_id  AS current_crop_id,
    c.name             AS crop_name,
    l.hectares         AS hectares
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  LEFT JOIN public.crops c ON c.id = l.current_crop_id AND c.deleted_at IS NULL
  WHERE l.deleted_at IS NULL AND l.hectares IS NOT NULL AND l.hectares > 0
),
-- Calcular superficie total del proyecto PRIMERO (todos los lotes)
project_totals AS (
  SELECT 
    project_id,
    SUM(hectares)::numeric AS total_project_hectares
  FROM lot_base
  GROUP BY project_id
),
-- Agrupar por cultivo solo los lotes que tienen cultivo asignado
by_crop AS (
  SELECT 
    project_id, 
    current_crop_id, 
    crop_name, 
    SUM(hectares)::numeric AS crop_hectares
  FROM lot_base
  WHERE current_crop_id IS NOT NULL
  GROUP BY project_id, current_crop_id, crop_name
),
-- Calcular porcentajes base
crop_percentages AS (
  SELECT 
    bc.project_id,
    bc.current_crop_id,
    bc.crop_name,
    bc.crop_hectares,
    -- CON REDONDEO: volver a redondear a 3 decimales
    ROUND(v3_calc.safe_div(bc.crop_hectares, pt.total_project_hectares) * 100, 3) AS crop_incidence_pct
  FROM by_crop bc
  JOIN project_totals pt ON pt.project_id = bc.project_id
),
-- Calcular total de porcentajes
total_percentages AS (
  SELECT 
    project_id,
    SUM(crop_incidence_pct) AS total_pct
  FROM crop_percentages
  GROUP BY project_id
),
-- Aplicar lógica de ajuste solo si suma > 99%
adjusted_percentages AS (
  SELECT 
    cp.project_id,
    cp.current_crop_id,
    cp.crop_name,
    cp.crop_hectares,
    CASE 
      WHEN tp.total_pct > 99 THEN
        -- Si es el último cultivo (mayor crop_id), ajustar para que sume 100%
        CASE 
          WHEN cp.current_crop_id = (
            SELECT MAX(current_crop_id) 
            FROM crop_percentages cp2 
            WHERE cp2.project_id = cp.project_id
          ) THEN ROUND(100 - (tp.total_pct - cp.crop_incidence_pct), 3)
          ELSE cp.crop_incidence_pct
        END
      ELSE cp.crop_incidence_pct
    END AS crop_incidence_pct
  FROM crop_percentages cp
  JOIN total_percentages tp ON tp.project_id = cp.project_id
)
SELECT
  ap.project_id,
  ap.current_crop_id,
  ap.crop_name,
  ap.crop_hectares,
  ap.crop_incidence_pct,
  -- CON REDONDEO: usar función SSOT para costo directo por hectárea del cultivo
  ROUND(v3_calc.cost_per_ha_for_crop_ssot(ap.project_id, ap.current_crop_id)::numeric, 3) AS cost_per_ha_usd
FROM adjusted_percentages ap;
