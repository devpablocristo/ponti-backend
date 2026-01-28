-- ========================================
-- MIGRACIÓN 000335: Consolidar v3_calc lot -> v3_lot_ssot (UP)
-- ========================================
--
-- Propósito: Unificar funciones de lote en v3_calc para que deleguen en v3_lot_ssot.
-- Enfoque: v3_calc queda como wrapper de v3_lot_ssot (SSOT único).
-- Nota: Comentarios en español, código en inglés.
--
BEGIN;

-- Accessors básicos de lote
CREATE OR REPLACE FUNCTION v3_calc.lot_hectares(p_lot_id bigint)
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT v3_lot_ssot.lot_hectares(p_lot_id)
$$;

CREATE OR REPLACE FUNCTION v3_calc.lot_tons(p_lot_id bigint)
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT v3_lot_ssot.lot_tons(p_lot_id)
$$;

-- Costos por lote
CREATE OR REPLACE FUNCTION v3_calc.labor_cost_for_lot(p_lot_id bigint)
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT v3_lot_ssot.labor_cost_for_lot(p_lot_id)
$$;

CREATE OR REPLACE FUNCTION v3_calc.supply_cost_for_lot(p_lot_id bigint)
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT v3_lot_ssot.supply_cost_for_lot(p_lot_id)
$$;

CREATE OR REPLACE FUNCTION v3_calc.direct_cost_for_lot(p_lot_id bigint)
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT v3_lot_ssot.direct_cost_for_lot(p_lot_id)
$$;

-- Precios e ingresos por lote
CREATE OR REPLACE FUNCTION v3_calc.net_price_usd_for_lot(p_lot_id bigint)
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT v3_lot_ssot.net_price_usd_for_lot(p_lot_id)
$$;

CREATE OR REPLACE FUNCTION v3_calc.income_net_total_for_lot(p_lot_id bigint)
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT v3_lot_ssot.income_net_total_for_lot(p_lot_id)
$$;

CREATE OR REPLACE FUNCTION v3_calc.income_net_per_ha_for_lot(p_lot_id bigint)
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT v3_lot_ssot.income_net_per_ha_for_lot(p_lot_id)
$$;

-- Costos por ha y rentas
CREATE OR REPLACE FUNCTION v3_calc.admin_cost_per_ha_for_lot(p_lot_id bigint)
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT v3_lot_ssot.admin_cost_per_ha_for_lot(p_lot_id)
$$;

CREATE OR REPLACE FUNCTION v3_calc.cost_per_ha_for_lot(p_lot_id bigint)
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT v3_lot_ssot.cost_per_ha_for_lot(p_lot_id)
$$;

CREATE OR REPLACE FUNCTION v3_calc.rent_per_ha_for_lot(p_lot_id bigint)
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT v3_lot_ssot.rent_per_ha_for_lot(p_lot_id)
$$;

CREATE OR REPLACE FUNCTION v3_calc.active_total_per_ha_for_lot(p_lot_id bigint)
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT v3_lot_ssot.active_total_per_ha_for_lot(p_lot_id)
$$;

CREATE OR REPLACE FUNCTION v3_calc.operating_result_per_ha_for_lot(p_lot_id bigint)
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT v3_lot_ssot.operating_result_per_ha_for_lot(p_lot_id)
$$;

CREATE OR REPLACE FUNCTION v3_calc.yield_tn_per_ha_for_lot(p_lot_id bigint)
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT v3_lot_ssot.yield_tn_per_ha_for_lot(p_lot_id)
$$;

-- Áreas por lote
CREATE OR REPLACE FUNCTION v3_calc.seeded_area_for_lot(p_lot_id bigint)
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT v3_lot_ssot.seeded_area_for_lot(p_lot_id)
$$;

CREATE OR REPLACE FUNCTION v3_calc.harvested_area_for_lot(p_lot_id bigint)
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT v3_lot_ssot.harvested_area_for_lot(p_lot_id)
$$;

-- Rentabilidad
CREATE OR REPLACE FUNCTION v3_calc.renta_pct(
  operating_result_total_usd double precision,
  total_costs_usd double precision
) RETURNS double precision
LANGUAGE sql IMMUTABLE AS $$
  SELECT v3_lot_ssot.renta_pct(operating_result_total_usd, total_costs_usd)
$$;

COMMIT;
