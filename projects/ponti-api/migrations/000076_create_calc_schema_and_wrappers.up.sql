-- ========================================
-- MIGRACIÓN 000084: CREAR ESQUEMA calc Y WRAPPERS EN public (UP)
-- ========================================
-- 
-- Objetivo: Centralizar cálculos (DRY/SSOT) en schema calc y exponer wrappers en public
-- Fecha: 2025-09-12
-- Autor: Sistema
-- 
-- Nota: Código en inglés, comentarios en español.

BEGIN;

-- ============================================================================
-- SSOT de cálculos (DRY) — esquema calc
--  - Helpers genéricos (safe_div, %)
--  - Cálculos de negocio por-ha / por-lote / costos / rentas
--  - Wrappers de compatibilidad en schema public (mantienen firmas existentes)
-- ============================================================================

CREATE SCHEMA IF NOT EXISTS calc;

-- -----------------------------
-- Helpers genéricos / seguros
-- -----------------------------
CREATE OR REPLACE FUNCTION calc.coalesce0(numeric) RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT COALESCE($1, 0)
$$;

CREATE OR REPLACE FUNCTION calc.coalesce0(double precision) RETURNS double precision
LANGUAGE sql IMMUTABLE AS $$
  SELECT COALESCE($1, 0)
$$;

CREATE OR REPLACE FUNCTION calc.safe_div(numeric, numeric) RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT CASE WHEN $2 IS NULL OR $2 = 0 THEN 0 ELSE $1 / $2 END
$$;

CREATE OR REPLACE FUNCTION calc.safe_div_dp(double precision, double precision) RETURNS double precision
LANGUAGE sql IMMUTABLE AS $$
  SELECT CASE WHEN $2 IS NULL OR $2 = 0 THEN 0 ELSE $1 / $2 END
$$;

CREATE OR REPLACE FUNCTION calc.percentage(numeric, numeric) RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT calc.safe_div($1, $2) * 100
$$;

CREATE OR REPLACE FUNCTION calc.percentage_capped(numeric, numeric) RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT LEAST(calc.safe_div($1, $2) * 100, 100)
$$;

-- Conversión genérica a "por ha"
CREATE OR REPLACE FUNCTION calc.per_ha(numeric, numeric) RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT calc.safe_div($1, $2)
$$;

CREATE OR REPLACE FUNCTION calc.per_ha_dp(double precision, double precision) RETURNS double precision
LANGUAGE sql IMMUTABLE AS $$
  SELECT calc.safe_div_dp($1, $2)
$$;

-- -----------------------------------
-- Cálculos de dominio (agro / gestión)
-- -----------------------------------

-- Dosis normalizada (suma de dosis sobre superficie)
CREATE OR REPLACE FUNCTION calc.dose_per_ha(total_dose numeric, surface_ha numeric) RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT calc.safe_div(total_dose, surface_ha)
$$;

-- Área sembrada: si hay fecha de siembra, cuenta la superficie
CREATE OR REPLACE FUNCTION calc.seeded_area(sowing_date date, hectares numeric) RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT CASE WHEN sowing_date IS NOT NULL THEN COALESCE(hectares,0) ELSE 0 END
$$;

-- Área cosechada: si hay toneladas > 0, cuenta la superficie
CREATE OR REPLACE FUNCTION calc.harvested_area(tons numeric, hectares numeric) RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT CASE WHEN tons IS NOT NULL AND tons > 0 THEN COALESCE(hectares,0) ELSE 0 END
$$;

-- Rendimiento tn/ha (base hectáreas declaradas)
CREATE OR REPLACE FUNCTION calc.yield_tn_per_ha_over_hectares(tons numeric, hectares numeric) RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT calc.safe_div( COALESCE(tons,0), COALESCE(hectares,0) )
$$;

-- Rendimiento tn/ha (base área cosechada)
CREATE OR REPLACE FUNCTION calc.yield_tn_per_ha_over_harvested(tons numeric, harvested_area numeric) RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT calc.safe_div( COALESCE(tons,0), COALESCE(harvested_area,0) )
$$;

-- Costos
CREATE OR REPLACE FUNCTION calc.labor_cost(labor_price numeric, effective_area numeric) RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT COALESCE(labor_price,0) * COALESCE(effective_area,0)
$$;

CREATE OR REPLACE FUNCTION calc.supply_cost(final_dose double precision, supply_price numeric, effective_area numeric) RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT COALESCE(final_dose,0)::numeric * COALESCE(supply_price,0) * COALESCE(effective_area,0)
$$;

CREATE OR REPLACE FUNCTION calc.cost_per_ha(total_cost numeric, hectares numeric) RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT calc.per_ha(total_cost, hectares)
$$;

-- Ingresos
CREATE OR REPLACE FUNCTION calc.income_net_total(tons numeric, net_price_usd numeric) RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT COALESCE(tons,0) * COALESCE(net_price_usd,0)
$$;

CREATE OR REPLACE FUNCTION calc.income_net_per_ha(income_net_total numeric, hectares numeric) RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT calc.per_ha(income_net_total, hectares)
$$;

-- Renta por ha (según tipo de arriendo)
-- 1: % sobre ingreso neto por ha
-- 2: % sobre (ingreso neto - costo directo/ha - admin/ha)
-- 3: valor fijo por ha
-- 4: valor fijo por ha + % sobre ingreso neto por ha
CREATE OR REPLACE FUNCTION calc.rent_per_ha(
  lease_type_id integer,
  lease_type_percent double precision,
  lease_type_value double precision,
  income_net_per_ha double precision,
  cost_per_ha double precision,
  admin_cost_per_ha double precision
) RETURNS double precision
LANGUAGE sql IMMUTABLE AS $$
  SELECT
    CASE
      WHEN lease_type_id = 1 THEN COALESCE(lease_type_percent,0)/100.0 * COALESCE(income_net_per_ha,0)
      WHEN lease_type_id = 2 THEN COALESCE(lease_type_percent,0)/100.0 *
                               (COALESCE(income_net_per_ha,0) - COALESCE(cost_per_ha,0) - COALESCE(admin_cost_per_ha,0))
      WHEN lease_type_id = 3 THEN COALESCE(lease_type_value,0)
      WHEN lease_type_id = 4 THEN COALESCE(lease_type_value,0) +
                               (COALESCE(lease_type_percent,0)/100.0 * COALESCE(income_net_per_ha,0))
      ELSE 0
    END
$$;

-- Overload para lease_type_id bigint (compatibilidad con fields.lease_type_id bigserial)
CREATE OR REPLACE FUNCTION calc.rent_per_ha(
  lease_type_id bigint,
  lease_type_percent double precision,
  lease_type_value double precision,
  income_net_per_ha double precision,
  cost_per_ha double precision,
  admin_cost_per_ha double precision
) RETURNS double precision
LANGUAGE sql IMMUTABLE AS $$
  SELECT calc.rent_per_ha(lease_type_id::integer, lease_type_percent, lease_type_value,
                          income_net_per_ha, cost_per_ha, admin_cost_per_ha)
$$;

-- Activo total por ha = costo directo/ha + renta/ha + admin/ha
CREATE OR REPLACE FUNCTION calc.active_total_per_ha(
  direct_cost_per_ha double precision,
  rent_per_ha double precision,
  admin_cost_per_ha double precision
) RETURNS double precision
LANGUAGE sql IMMUTABLE AS $$
  SELECT COALESCE(direct_cost_per_ha,0) + COALESCE(rent_per_ha,0) + COALESCE(admin_cost_per_ha,0)
$$;

-- Resultado operativo por ha = ingreso neto/ha - activo total/ha
CREATE OR REPLACE FUNCTION calc.operating_result_per_ha(
  income_net_per_ha double precision,
  active_total_per_ha double precision
) RETURNS double precision
LANGUAGE sql IMMUTABLE AS $$
  SELECT COALESCE(income_net_per_ha,0) - COALESCE(active_total_per_ha,0)
$$;

-- % de renta = resultado operativo / costos totales
CREATE OR REPLACE FUNCTION calc.renta_pct(operating_result_total_usd double precision, total_costs_usd double precision) RETURNS double precision
LANGUAGE sql IMMUTABLE AS $$
  SELECT CASE WHEN COALESCE(total_costs_usd,0) > 0
              THEN (COALESCE(operating_result_total_usd,0) / total_costs_usd) * 100
              ELSE 0 END
$$;

-- Precio de indiferencia (USD / tn) = invertido_por_ha / (tn/ha)
CREATE OR REPLACE FUNCTION calc.indifference_price_usd_tn(total_invested_per_ha double precision, yield_tn_per_ha double precision) RETURNS double precision
LANGUAGE sql IMMUTABLE AS $$
  SELECT calc.per_ha_dp(total_invested_per_ha, yield_tn_per_ha)
$$;

-- Unidades por ha (litros/ha, kg/ha, etc.)
CREATE OR REPLACE FUNCTION calc.units_per_ha(units numeric, area numeric) RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT calc.per_ha(units, area)
$$;

-- Alias para la función existente de dosis normalizada
CREATE OR REPLACE FUNCTION calc.norm_dose(dose numeric, area numeric) RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT CASE WHEN area > 0 THEN dose / area ELSE NULL END
$$;

-- -------------------------------------------------------
-- Wrappers de compatibilidad (mantienen firmas en schema public)
-- -------------------------------------------------------
CREATE OR REPLACE FUNCTION public.calculate_cost_per_ha(p_total_cost numeric, p_hectares numeric) RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT calc.cost_per_ha(p_total_cost, p_hectares)
$$;

-- Overload para aceptar hectares como double precision
CREATE OR REPLACE FUNCTION public.calculate_cost_per_ha(p_total_cost numeric, p_hectares double precision) RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT calc.cost_per_ha(p_total_cost, p_hectares::numeric)
$$;

CREATE OR REPLACE FUNCTION public.calculate_harvested_area(p_tons numeric, p_hectares numeric) RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT calc.harvested_area(p_tons, p_hectares)
$$;

CREATE OR REPLACE FUNCTION public.calculate_labor_cost(p_labor_price numeric, p_effective_area numeric) RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT calc.labor_cost(p_labor_price, p_effective_area)
$$;

CREATE OR REPLACE FUNCTION public.calculate_sowed_area(p_sowing_date date, p_hectares numeric) RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT calc.seeded_area(p_sowing_date, p_hectares)
$$;

CREATE OR REPLACE FUNCTION public.calculate_supply_cost(p_final_dose double precision, p_supply_price numeric, p_effective_area numeric) RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT calc.supply_cost(p_final_dose, p_supply_price, p_effective_area)
$$;

CREATE OR REPLACE FUNCTION public.calculate_yield(p_tons numeric, p_hectares numeric) RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT calc.yield_tn_per_ha_over_hectares(p_tons, p_hectares)
$$;

-- Nota: public.calculate_campaign_closing_date() queda igual, ya usa get_campaign_closure_days().

-- -------------------------------------------------------
-- Helpers de consulta de negocio (leen tablas transaccionales)
--  Estos reemplazan el uso de vistas base_* en vistas v3
-- -------------------------------------------------------

-- Helpers básicos de lot (evitan repetir SELECT contra public.lots)
CREATE OR REPLACE FUNCTION calc.lot_hectares(p_lot_id bigint) RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(l.hectares, 0)
  FROM public.lots l
  WHERE l.id = p_lot_id AND l.deleted_at IS NULL
$$;

CREATE OR REPLACE FUNCTION calc.lot_tons(p_lot_id bigint) RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(l.tons, 0)
  FROM public.lots l
  WHERE l.id = p_lot_id AND l.deleted_at IS NULL
$$;

-- Costo de labores por lote
CREATE OR REPLACE FUNCTION calc.labor_cost_for_lot(p_lot_id bigint) RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(SUM(lb.price * w.effective_area), 0)::numeric
  FROM public.workorders w
  JOIN public.labors lb ON lb.id = w.labor_id
  WHERE w.deleted_at IS NULL
    AND w.effective_area > 0
    AND lb.price IS NOT NULL
    AND w.lot_id = p_lot_id
$$;

-- Costo de insumos por lote
CREATE OR REPLACE FUNCTION calc.supply_cost_for_lot(p_lot_id bigint) RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(SUM((wi.final_dose)::double precision * s.price * (w.effective_area)::double precision), 0)::double precision
  FROM public.workorders w
  JOIN public.workorder_items wi ON wi.workorder_id = w.id
  JOIN public.supplies s ON s.id = wi.supply_id
  WHERE w.deleted_at IS NULL
    AND w.effective_area > 0
    AND wi.final_dose > 0
    AND s.price IS NOT NULL
    AND w.lot_id = p_lot_id
$$;

-- Costo directo por lote
CREATE OR REPLACE FUNCTION calc.direct_cost_for_lot(p_lot_id bigint) RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(calc.labor_cost_for_lot(p_lot_id), 0)::double precision
       + COALESCE(calc.supply_cost_for_lot(p_lot_id), 0)
$$;

-- Precio neto (USD/tn) vigente para el lote (según project/crop)
CREATE OR REPLACE FUNCTION calc.net_price_usd_for_lot(p_lot_id bigint) RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(cc.net_price, 0)
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  LEFT JOIN public.crop_commercializations cc
    ON cc.project_id = f.project_id
   AND cc.crop_id    = l.current_crop_id
   AND cc.deleted_at IS NULL
  WHERE l.id = p_lot_id
    AND l.deleted_at IS NULL
  LIMIT 1
$$;

-- Ingreso neto total por lote (USD)
CREATE OR REPLACE FUNCTION calc.income_net_total_for_lot(p_lot_id bigint) RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(l.tons, 0)::numeric * COALESCE(calc.net_price_usd_for_lot(l.id), 0)::numeric
  FROM public.lots l
  WHERE l.id = p_lot_id AND l.deleted_at IS NULL
$$;

-- Ingreso neto por ha (USD/ha)
CREATE OR REPLACE FUNCTION calc.income_net_per_ha_for_lot(p_lot_id bigint) RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT calc.safe_div_dp(
           COALESCE(calc.income_net_total_for_lot(p_lot_id), 0)::double precision,
           calc.lot_hectares(p_lot_id)
         )
$$;

-- Hectáreas totales por proyecto (para prorrateo de admin)
CREATE OR REPLACE FUNCTION calc.total_hectares_for_project(p_project_id bigint) RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(SUM(l.hectares), 0)::double precision
  FROM public.fields f
  LEFT JOIN public.lots l ON l.field_id = f.id AND l.deleted_at IS NULL
  WHERE f.project_id = p_project_id AND f.deleted_at IS NULL
$$;

-- Costo de administración por ha para un lote
CREATE OR REPLACE FUNCTION calc.admin_cost_per_ha_for_lot(p_lot_id bigint) RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT CASE WHEN t.total_hectares > 0
              THEN COALESCE(p.admin_cost, 0)::double precision / t.total_hectares
              ELSE 0 END
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  JOIN public.projects p ON p.id = f.project_id AND p.deleted_at IS NULL
  CROSS JOIN LATERAL (
    SELECT calc.total_hectares_for_project(f.project_id) AS total_hectares
  ) t
  WHERE l.id = p_lot_id AND l.deleted_at IS NULL
$$;

-- Costo directo por ha para un lote
CREATE OR REPLACE FUNCTION calc.cost_per_ha_for_lot(p_lot_id bigint) RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT calc.safe_div_dp(
           COALESCE(calc.direct_cost_for_lot(p_lot_id), 0)::double precision,
           calc.lot_hectares(p_lot_id)
         )
$$;

-- Renta por ha para un lote (según reglas del field)
CREATE OR REPLACE FUNCTION calc.rent_per_ha_for_lot(p_lot_id bigint) RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT calc.rent_per_ha(
           f.lease_type_id,
           f.lease_type_percent,
           f.lease_type_value,
           calc.income_net_per_ha_for_lot(p_lot_id),
           calc.cost_per_ha_for_lot(p_lot_id),
           calc.admin_cost_per_ha_for_lot(p_lot_id)
         )
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  WHERE l.id = p_lot_id AND l.deleted_at IS NULL
$$;

-- Activo total por ha para un lote
CREATE OR REPLACE FUNCTION calc.active_total_per_ha_for_lot(p_lot_id bigint) RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT calc.active_total_per_ha(
           calc.cost_per_ha_for_lot(p_lot_id),
           calc.rent_per_ha_for_lot(p_lot_id),
           calc.admin_cost_per_ha_for_lot(p_lot_id)
         )
$$;

-- Resultado operativo por ha para un lote
CREATE OR REPLACE FUNCTION calc.operating_result_per_ha_for_lot(p_lot_id bigint) RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT calc.operating_result_per_ha(
           calc.income_net_per_ha_for_lot(p_lot_id),
           calc.active_total_per_ha_for_lot(p_lot_id)
         )
$$;

-- Rendimiento tn/ha para un lote
CREATE OR REPLACE FUNCTION calc.yield_tn_per_ha_for_lot(p_lot_id bigint) RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT calc.per_ha_dp(
           calc.lot_tons(p_lot_id),
           calc.lot_hectares(p_lot_id)
         )
$$;

-- Área sembrada (ha) para un lote
CREATE OR REPLACE FUNCTION calc.seeded_area_for_lot(p_lot_id bigint) RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT calc.seeded_area(l.sowing_date, l.hectares::numeric)
  FROM public.lots l
  WHERE l.id = p_lot_id AND l.deleted_at IS NULL
$$;

-- Área cosechada (ha) para un lote
CREATE OR REPLACE FUNCTION calc.harvested_area_for_lot(p_lot_id bigint) RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT calc.harvested_area(
           calc.lot_tons(p_lot_id)::numeric,
           calc.lot_hectares(p_lot_id)::numeric
         )
$$;

COMMIT;


