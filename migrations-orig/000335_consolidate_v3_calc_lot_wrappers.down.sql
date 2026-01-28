-- ========================================
-- MIGRACIÓN 000335: Consolidar v3_calc lot -> v3_lot_ssot (DOWN)
-- ========================================
--
-- Propósito: Revertir wrappers y restaurar definiciones locales en v3_calc.
-- Nota: Comentarios en español, código en inglés.
--
BEGIN;

-- Accessors básicos de lote
CREATE OR REPLACE FUNCTION v3_calc.lot_hectares(p_lot_id bigint)
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(l.hectares, 0)
  FROM public.lots l
  WHERE l.id = p_lot_id AND l.deleted_at IS NULL
$$;

CREATE OR REPLACE FUNCTION v3_calc.lot_tons(p_lot_id bigint)
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(l.tons, 0)
  FROM public.lots l
  WHERE l.id = p_lot_id AND l.deleted_at IS NULL
$$;

-- Costos por lote
CREATE OR REPLACE FUNCTION v3_calc.labor_cost_for_lot(p_lot_id bigint)
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(SUM(lb.price * w.effective_area), 0)::numeric
  FROM public.workorders w
  JOIN public.labors lb ON lb.id = w.labor_id
  WHERE w.deleted_at IS NULL
    AND w.effective_area > 0
    AND lb.price IS NOT NULL
    AND w.lot_id = p_lot_id
$$;

CREATE OR REPLACE FUNCTION v3_calc.supply_cost_for_lot(p_lot_id bigint)
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    -- Costos por workorder_items (uso directo en workorders)
    (SELECT COALESCE(SUM((wi.final_dose)::double precision * s.price * (w.effective_area)::double precision), 0)::double precision
     FROM public.workorders w
     JOIN public.workorder_items wi ON wi.workorder_id = w.id
     JOIN public.supplies s ON s.id = wi.supply_id
     WHERE w.deleted_at IS NULL
       AND w.effective_area > 0
       AND wi.final_dose > 0
       AND s.price IS NOT NULL
       AND w.lot_id = p_lot_id)
    +
    -- Costos por movimientos internos de salida (insumos transferidos a otros proyectos)
    (SELECT COALESCE(SUM(sm.quantity * s.price), 0)::double precision
     FROM public.supply_movements sm
     JOIN public.supplies s ON s.id = sm.supply_id
     JOIN public.workorders w ON w.lot_id = p_lot_id
     WHERE sm.deleted_at IS NULL
       AND s.deleted_at IS NULL
       AND w.deleted_at IS NULL
       AND sm.movement_type = 'Movimiento interno'
       AND sm.is_entry = false
       AND sm.project_id = w.project_id
       AND s.price IS NOT NULL
       AND sm.quantity > 0)
  , 0)::double precision
$$;

CREATE OR REPLACE FUNCTION v3_calc.direct_cost_for_lot(p_lot_id bigint)
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(v3_calc.labor_cost_for_lot(p_lot_id), 0)::double precision
       + COALESCE(v3_calc.supply_cost_for_lot(p_lot_id), 0)
$$;

-- Precios e ingresos por lote
CREATE OR REPLACE FUNCTION v3_calc.net_price_usd_for_lot(p_lot_id bigint)
RETURNS numeric
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

CREATE OR REPLACE FUNCTION v3_calc.income_net_total_for_lot(p_lot_id bigint)
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(l.tons, 0)::numeric * COALESCE(v3_calc.net_price_usd_for_lot(l.id), 0)::numeric
  FROM public.lots l
  WHERE l.id = p_lot_id AND l.deleted_at IS NULL
$$;

CREATE OR REPLACE FUNCTION v3_calc.income_net_per_ha_for_lot(p_lot_id bigint)
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT v3_calc.safe_div_dp(
           COALESCE(v3_calc.income_net_total_for_lot(p_lot_id), 0)::double precision,
           v3_calc.lot_hectares(p_lot_id)
         )
$$;

-- Costos por ha y rentas
CREATE OR REPLACE FUNCTION v3_calc.admin_cost_per_ha_for_lot(p_lot_id bigint)
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT CASE WHEN t.total_hectares > 0
              THEN COALESCE(p.admin_cost, 0)::double precision / t.total_hectares
              ELSE 0 END
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  JOIN public.projects p ON p.id = f.project_id AND p.deleted_at IS NULL
  CROSS JOIN LATERAL (
    SELECT v3_calc.total_hectares_for_project(f.project_id) AS total_hectares
  ) t
  WHERE l.id = p_lot_id AND l.deleted_at IS NULL
$$;

CREATE OR REPLACE FUNCTION v3_calc.cost_per_ha_for_lot(p_lot_id bigint)
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT v3_calc.safe_div_dp(
           COALESCE(v3_calc.direct_cost_for_lot(p_lot_id), 0)::double precision,
           v3_calc.lot_hectares(p_lot_id)
         )
$$;

CREATE OR REPLACE FUNCTION v3_calc.rent_per_ha_for_lot(p_lot_id bigint)
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT v3_calc.rent_per_ha(
           f.lease_type_id,
           f.lease_type_percent,
           f.lease_type_value,
           v3_calc.income_net_per_ha_for_lot(p_lot_id),
           v3_calc.cost_per_ha_for_lot(p_lot_id),
           v3_calc.admin_cost_per_ha_for_lot(p_lot_id)
         )
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  WHERE l.id = p_lot_id AND l.deleted_at IS NULL
$$;

CREATE OR REPLACE FUNCTION v3_calc.active_total_per_ha_for_lot(p_lot_id bigint)
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT v3_calc.active_total_per_ha(
           v3_calc.cost_per_ha_for_lot(p_lot_id),
           v3_calc.rent_per_ha_for_lot(p_lot_id),
           v3_calc.admin_cost_per_ha_for_lot(p_lot_id)
         )
$$;

CREATE OR REPLACE FUNCTION v3_calc.operating_result_per_ha_for_lot(p_lot_id bigint)
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT v3_calc.operating_result_per_ha(
           v3_calc.income_net_per_ha_for_lot(p_lot_id),
           v3_calc.active_total_per_ha_for_lot(p_lot_id)
         )
$$;

CREATE OR REPLACE FUNCTION v3_calc.yield_tn_per_ha_for_lot(p_lot_id bigint)
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT v3_calc.per_ha_dp(
           v3_calc.lot_tons(p_lot_id),
           v3_calc.lot_hectares(p_lot_id)
         )
$$;

-- Áreas por lote
CREATE OR REPLACE FUNCTION v3_calc.seeded_area_for_lot(p_lot_id bigint)
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT v3_calc.seeded_area(l.sowing_date, l.hectares::numeric)
  FROM public.lots l
  WHERE l.id = p_lot_id AND l.deleted_at IS NULL
$$;

CREATE OR REPLACE FUNCTION v3_calc.harvested_area_for_lot(p_lot_id bigint)
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT v3_calc.harvested_area(
           v3_calc.lot_tons(p_lot_id)::numeric,
           v3_calc.lot_hectares(p_lot_id)::numeric
         )
$$;

-- Rentabilidad
CREATE OR REPLACE FUNCTION v3_calc.renta_pct(
  operating_result_total_usd double precision,
  total_costs_usd double precision
) RETURNS double precision
LANGUAGE sql IMMUTABLE AS $$
  SELECT CASE WHEN COALESCE(total_costs_usd,0) > 0
              THEN (COALESCE(operating_result_total_usd,0) / total_costs_usd) * 100
              ELSE 0 END
$$;

COMMIT;
