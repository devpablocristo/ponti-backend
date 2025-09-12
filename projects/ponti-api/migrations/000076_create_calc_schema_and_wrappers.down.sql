-- ========================================
-- MIGRACIÓN 000084: ELIMINAR ESQUEMA calc Y WRAPPERS (DOWN)
-- ========================================
-- 
-- Objetivo: Revertir creación de schema calc y wrappers en public
-- Fecha: 2025-09-12
-- Autor: Sistema
-- 
-- Nota: Código en inglés, comentarios en español.

BEGIN;

-- Eliminar wrappers en public (mantener orden para evitar dependencias)
DROP FUNCTION IF EXISTS public.calculate_yield(numeric, numeric);
DROP FUNCTION IF EXISTS public.calculate_supply_cost(double precision, numeric, numeric);
DROP FUNCTION IF EXISTS public.calculate_sowed_area(date, numeric);
DROP FUNCTION IF EXISTS public.calculate_labor_cost(numeric, numeric);
DROP FUNCTION IF EXISTS public.calculate_harvested_area(numeric, numeric);
DROP FUNCTION IF EXISTS public.calculate_cost_per_ha(numeric, numeric);

-- Eliminar funciones en schema calc
DROP FUNCTION IF EXISTS calc.norm_dose(numeric, numeric);
DROP FUNCTION IF EXISTS calc.units_per_ha(numeric, numeric);
DROP FUNCTION IF EXISTS calc.indifference_price_usd_tn(double precision, double precision);
DROP FUNCTION IF EXISTS calc.renta_pct(double precision, double precision);
DROP FUNCTION IF EXISTS calc.operating_result_per_ha(double precision, double precision);
DROP FUNCTION IF EXISTS calc.active_total_per_ha(double precision, double precision, double precision);
DROP FUNCTION IF EXISTS calc.rent_per_ha(integer, double precision, double precision, double precision, double precision, double precision);
DROP FUNCTION IF EXISTS calc.income_net_per_ha(numeric, numeric);
DROP FUNCTION IF EXISTS calc.income_net_total(numeric, numeric);
DROP FUNCTION IF EXISTS calc.cost_per_ha(numeric, numeric);
DROP FUNCTION IF EXISTS calc.supply_cost(double precision, numeric, numeric);
DROP FUNCTION IF EXISTS calc.labor_cost(numeric, numeric);
DROP FUNCTION IF EXISTS calc.yield_tn_per_ha_over_harvested(numeric, numeric);
DROP FUNCTION IF EXISTS calc.yield_tn_per_ha_over_hectares(numeric, numeric);
DROP FUNCTION IF EXISTS calc.harvested_area(numeric, numeric);
DROP FUNCTION IF EXISTS calc.seeded_area(date, numeric);
DROP FUNCTION IF EXISTS calc.dose_per_ha(numeric, numeric);
DROP FUNCTION IF EXISTS calc.per_ha_dp(double precision, double precision);
DROP FUNCTION IF EXISTS calc.per_ha(numeric, numeric);
DROP FUNCTION IF EXISTS calc.percentage_capped(numeric, numeric);
DROP FUNCTION IF EXISTS calc.percentage(numeric, numeric);
DROP FUNCTION IF EXISTS calc.safe_div_dp(double precision, double precision);
DROP FUNCTION IF EXISTS calc.safe_div(numeric, numeric);
DROP FUNCTION IF EXISTS calc.coalesce0(double precision);
DROP FUNCTION IF EXISTS calc.coalesce0(numeric);

-- Eliminar schema calc
DROP SCHEMA IF EXISTS calc;

COMMIT;


