-- ========================================
-- MIGRACIÓN 000114: CREATE v3_workorder_ssot SCHEMA (UP)
-- ========================================
-- 
-- Propósito: Crear esquema v3_workorder_ssot con funciones específicas de workorders
-- Dependencias: Requiere v3_core_ssot (000113)
-- Alcance: Funciones para cálculos de workorders que se usan en múltiples módulos
-- Fecha: 2025-10-04
-- Autor: Sistema
-- 
-- Nota: Código en inglés, comentarios en español
-- Usa v3_core_ssot para operaciones básicas

BEGIN;

-- ========================================
-- CREAR ESQUEMA v3_workorder_ssot
-- ========================================
CREATE SCHEMA IF NOT EXISTS v3_workorder_ssot;

COMMENT ON SCHEMA v3_workorder_ssot IS 'Funciones SSOT de workorders: cálculos por workorder, lote y proyecto';

-- ========================================
-- GRUPO 1: COSTOS POR WORKORDER (2 funciones)
-- ========================================
-- Propósito: Calcular costos básicos por workorder

CREATE OR REPLACE FUNCTION v3_workorder_ssot.labor_cost_for_workorder(p_workorder_id bigint)
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT v3_core_ssot.labor_cost(lb.price::numeric, w.effective_area::numeric)
  FROM public.workorders w
  JOIN public.labors lb ON lb.id = w.labor_id AND lb.deleted_at IS NULL
  WHERE w.id = p_workorder_id
    AND w.deleted_at IS NULL
    AND w.effective_area > 0
$$;

CREATE OR REPLACE FUNCTION v3_workorder_ssot.supply_cost_for_workorder(p_workorder_id bigint)
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    SUM(v3_core_ssot.supply_cost(
      wi.final_dose::double precision,
      s.price::numeric,
      w.effective_area::numeric
    )), 0
  )::numeric
  FROM public.workorders w
  LEFT JOIN public.workorder_items wi ON wi.workorder_id = w.id AND wi.deleted_at IS NULL
  LEFT JOIN public.supplies s ON s.id = wi.supply_id AND s.deleted_at IS NULL
  WHERE w.id = p_workorder_id
    AND w.deleted_at IS NULL
$$;

-- ========================================
-- GRUPO 2: AGREGACIONES POR LOTE (3 funciones)
-- ========================================
-- Propósito: Agregar costos y superficies por lote

CREATE OR REPLACE FUNCTION v3_workorder_ssot.surface_for_lot(p_lot_id bigint)
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(SUM(w.effective_area), 0)::numeric
  FROM public.workorders w
  WHERE w.lot_id = p_lot_id
    AND w.deleted_at IS NULL
    AND w.effective_area > 0
$$;

CREATE OR REPLACE FUNCTION v3_workorder_ssot.labor_cost_for_lot_wo(p_lot_id bigint)
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    SUM(v3_core_ssot.labor_cost(lb.price::numeric, w.effective_area::numeric)), 0
  )::numeric
  FROM public.workorders w
  JOIN public.labors lb ON lb.id = w.labor_id AND lb.deleted_at IS NULL
  WHERE w.lot_id = p_lot_id
    AND w.deleted_at IS NULL
    AND w.effective_area > 0
$$;

CREATE OR REPLACE FUNCTION v3_workorder_ssot.supply_cost_for_lot_wo(p_lot_id bigint)
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    SUM(v3_core_ssot.supply_cost(
      wi.final_dose::double precision,
      s.price::numeric,
      w.effective_area::numeric
    )), 0
  )::numeric
  FROM public.workorders w
  LEFT JOIN public.workorder_items wi ON wi.workorder_id = w.id AND wi.deleted_at IS NULL
  LEFT JOIN public.supplies s ON s.id = wi.supply_id AND s.deleted_at IS NULL
  WHERE w.lot_id = p_lot_id
    AND w.deleted_at IS NULL
$$;

-- ========================================
-- GRUPO 3: INSUMOS POR LOTE (2 funciones)
-- ========================================
-- Propósito: Calcular consumos de insumos por lote

CREATE OR REPLACE FUNCTION v3_workorder_ssot.liters_for_lot(p_lot_id bigint)
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    SUM(CASE WHEN s.unit_id = 1 THEN (wi.final_dose * w.effective_area) ELSE 0 END), 0
  )::numeric
  FROM public.workorders w
  LEFT JOIN public.workorder_items wi ON wi.workorder_id = w.id AND wi.deleted_at IS NULL
  LEFT JOIN public.supplies s ON s.id = wi.supply_id AND s.deleted_at IS NULL
  WHERE w.lot_id = p_lot_id
    AND w.deleted_at IS NULL
$$;

CREATE OR REPLACE FUNCTION v3_workorder_ssot.kilograms_for_lot(p_lot_id bigint)
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    SUM(CASE WHEN s.unit_id = 2 THEN (wi.final_dose * w.effective_area) ELSE 0 END), 0
  )::numeric
  FROM public.workorders w
  LEFT JOIN public.workorder_items wi ON wi.workorder_id = w.id AND wi.deleted_at IS NULL
  LEFT JOIN public.supplies s ON s.id = wi.supply_id AND s.deleted_at IS NULL
  WHERE w.lot_id = p_lot_id
    AND w.deleted_at IS NULL
$$;

COMMIT;

-- Comentarios sobre funciones clave
COMMENT ON FUNCTION v3_workorder_ssot.labor_cost_for_workorder(bigint) IS 'Calcula costo de labor por workorder';
COMMENT ON FUNCTION v3_workorder_ssot.supply_cost_for_workorder(bigint) IS 'Calcula costo de insumos por workorder';
COMMENT ON FUNCTION v3_workorder_ssot.surface_for_lot(bigint) IS 'Calcula superficie total trabajada en un lote';
