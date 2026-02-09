-- ========================================
-- MIGRATION 000110 V4 SSOT FUNCTIONS (UP)
-- ========================================
-- Nota: SQL en inglés, comentarios en español

BEGIN;

CREATE OR REPLACE FUNCTION v4_ssot.lot_hectares(p_lot_id bigint) 
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(l.hectares, 0)
  FROM public.lots l
  WHERE l.id = p_lot_id AND l.deleted_at IS NULL
$$;

CREATE OR REPLACE FUNCTION v4_ssot.lot_tons(p_lot_id bigint) 
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(l.tons, 0)
  FROM public.lots l
  WHERE l.id = p_lot_id AND l.deleted_at IS NULL
$$;

CREATE OR REPLACE FUNCTION v4_ssot.labor_cost_for_lot(p_lot_id bigint) 
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

CREATE OR REPLACE FUNCTION v4_ssot.supply_cost_for_lot_base(p_lot_id bigint)
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    SUM(wi.total_used * s.price), 0
  )::numeric
  FROM public.workorders w
  JOIN public.workorder_items wi ON wi.workorder_id = w.id AND wi.deleted_at IS NULL
  JOIN public.supplies s ON s.id = wi.supply_id AND s.deleted_at IS NULL
  WHERE w.lot_id = p_lot_id
    AND w.deleted_at IS NULL
    AND w.effective_area IS NOT NULL
    AND w.effective_area > 0
$$;

CREATE OR REPLACE FUNCTION v4_ssot.supply_cost_for_lot_by_category(
  p_lot_id bigint,
  p_category_name text
)
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    SUM(wi.total_used * s.price), 0
  )::numeric
  FROM public.workorders w
  JOIN public.workorder_items wi ON wi.workorder_id = w.id AND wi.deleted_at IS NULL
  JOIN public.supplies s ON s.id = wi.supply_id AND s.deleted_at IS NULL
  JOIN public.categories c ON c.id = s.category_id
  WHERE w.lot_id = p_lot_id
    AND c.name = p_category_name
    AND w.deleted_at IS NULL
    AND w.effective_area > 0
    AND wi.total_used > 0
    AND s.price IS NOT NULL
$$;

CREATE OR REPLACE FUNCTION v4_ssot.surface_for_lot(p_lot_id bigint)
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(SUM(w.effective_area), 0)::numeric
  FROM public.workorders w
  WHERE w.lot_id = p_lot_id
    AND w.deleted_at IS NULL
    AND w.effective_area > 0
$$;

CREATE OR REPLACE FUNCTION v4_ssot.liters_for_lot(p_lot_id bigint)
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

CREATE OR REPLACE FUNCTION v4_ssot.kilograms_for_lot(p_lot_id bigint)
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

CREATE OR REPLACE FUNCTION v4_ssot.supply_cost_for_lot(p_lot_id bigint)
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    (SELECT COALESCE(SUM(wi.total_used * s.price), 0)::numeric
     FROM public.workorders w
     JOIN public.workorder_items wi ON wi.workorder_id = w.id AND wi.deleted_at IS NULL
     JOIN public.supplies s ON s.id = wi.supply_id AND s.deleted_at IS NULL
     WHERE w.deleted_at IS NULL
       AND w.effective_area > 0
       AND wi.total_used > 0
       AND s.price IS NOT NULL
       AND w.lot_id = p_lot_id)
    +
    (SELECT COALESCE(SUM(sm.quantity * s.price), 0)::numeric
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
  , 0)::numeric
$$;

CREATE OR REPLACE FUNCTION v4_ssot.net_price_usd_for_lot(p_lot_id bigint) 
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(cc.net_price, 0)
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  LEFT JOIN public.crop_commercializations cc
    ON cc.project_id = f.project_id
   AND cc.crop_id = l.current_crop_id
   AND cc.deleted_at IS NULL
  WHERE l.id = p_lot_id
    AND l.deleted_at IS NULL
  LIMIT 1
$$;

CREATE OR REPLACE FUNCTION v4_ssot.income_net_total_for_lot(p_lot_id bigint) 
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(l.tons, 0)::numeric * COALESCE(v4_ssot.net_price_usd_for_lot(l.id), 0)::numeric
  FROM public.lots l
  WHERE l.id = p_lot_id AND l.deleted_at IS NULL
$$;

CREATE OR REPLACE FUNCTION v4_ssot.income_net_per_ha_for_lot(p_lot_id bigint) 
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT v4_core.safe_div(
           COALESCE(v4_ssot.income_net_total_for_lot(p_lot_id), 0),
           v4_ssot.lot_hectares(p_lot_id)::numeric
         )
$$;

CREATE OR REPLACE FUNCTION v4_ssot.direct_cost_for_lot(p_lot_id bigint) 
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(v4_ssot.labor_cost_for_lot(p_lot_id), 0)::numeric
       + COALESCE(v4_ssot.supply_cost_for_lot(p_lot_id), 0)
$$;

CREATE OR REPLACE FUNCTION v4_ssot.cost_per_ha_for_lot(p_lot_id bigint) 
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT v4_core.safe_div(
           COALESCE(v4_ssot.direct_cost_for_lot(p_lot_id), 0)::numeric,
           v4_ssot.lot_hectares(p_lot_id)
         )
$$;

CREATE OR REPLACE FUNCTION v4_ssot.total_hectares_for_project(p_project_id bigint) 
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(SUM(l.hectares), 0)::numeric
  FROM public.fields f
  LEFT JOIN public.lots l ON l.field_id = f.id AND l.deleted_at IS NULL
  WHERE f.project_id = p_project_id AND f.deleted_at IS NULL
$$;

CREATE OR REPLACE FUNCTION v4_ssot.admin_cost_prorated_per_ha_for_lot(p_lot_id bigint)
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT CASE
           WHEN t.total_hectares > 0 THEN COALESCE(p.admin_cost, 0)::numeric / t.total_hectares
           ELSE 0::numeric
         END
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  JOIN public.projects p ON p.id = f.project_id AND p.deleted_at IS NULL
  CROSS JOIN LATERAL (
    SELECT v4_ssot.total_hectares_for_project(f.project_id)::numeric AS total_hectares
  ) t
  WHERE l.id = p_lot_id AND l.deleted_at IS NULL
$$;

CREATE OR REPLACE FUNCTION v4_ssot.admin_cost_per_ha_for_lot(p_lot_id bigint) 
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(p.admin_cost, 0)::numeric
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  JOIN public.projects p ON p.id = f.project_id AND p.deleted_at IS NULL
  WHERE l.id = p_lot_id AND l.deleted_at IS NULL
$$;

CREATE OR REPLACE FUNCTION v4_ssot.rent_per_ha_for_lot(p_lot_id bigint) 
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT v4_core.rent_per_ha(
           f.lease_type_id,
           f.lease_type_percent,
           f.lease_type_value,
           v4_ssot.income_net_per_ha_for_lot(p_lot_id),
           v4_ssot.cost_per_ha_for_lot(p_lot_id),
           v4_ssot.admin_cost_per_ha_for_lot(p_lot_id)
         )::numeric
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  WHERE l.id = p_lot_id AND l.deleted_at IS NULL
$$;

CREATE OR REPLACE FUNCTION v4_ssot.rent_per_ha_for_lot_fixed(p_lot_id BIGINT)
RETURNS numeric AS $$
DECLARE
  calculated_rent numeric;
BEGIN
  calculated_rent := v4_ssot.rent_per_ha_for_lot(p_lot_id);
  
  IF calculated_rent < 0 THEN
    RETURN 0;
  ELSE
    RETURN calculated_rent;
  END IF;
END;
$$ LANGUAGE plpgsql STABLE;

CREATE OR REPLACE FUNCTION v4_ssot.active_total_per_ha_for_lot(p_lot_id bigint) 
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT v4_core.active_total_per_ha(
           v4_ssot.cost_per_ha_for_lot(p_lot_id),
           v4_ssot.rent_per_ha_for_lot(p_lot_id),
           v4_ssot.admin_cost_per_ha_for_lot(p_lot_id)
         )
$$;

CREATE OR REPLACE FUNCTION v4_ssot.operating_result_per_ha_for_lot(p_lot_id bigint) 
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT v4_core.operating_result_per_ha(
           v4_ssot.income_net_per_ha_for_lot(p_lot_id),
           v4_ssot.active_total_per_ha_for_lot(p_lot_id)
         )
$$;

CREATE OR REPLACE FUNCTION v4_ssot.yield_tn_per_ha_for_lot(p_lot_id bigint) 
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT v4_core.per_ha(
           v4_ssot.lot_tons(p_lot_id),
           v4_ssot.lot_hectares(p_lot_id)
         )
$$;

CREATE OR REPLACE FUNCTION v4_ssot.seeded_area_for_lot(p_lot_id bigint) 
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(SUM(
    CASE WHEN lb.category_id = 9 THEN w.effective_area ELSE 0 END
  ), 0)::numeric
  FROM public.workorders w
  JOIN public.labors lb ON lb.id = w.labor_id AND lb.deleted_at IS NULL
  WHERE w.lot_id = p_lot_id
    AND w.deleted_at IS NULL
    AND w.effective_area IS NOT NULL
    AND w.effective_area > 0
$$;

CREATE OR REPLACE FUNCTION v4_ssot.harvested_area_for_lot(p_lot_id bigint) 
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT v4_core.harvested_area(
           v4_ssot.lot_tons(p_lot_id)::numeric,
           v4_ssot.lot_hectares(p_lot_id)::numeric
         )
$$;

CREATE OR REPLACE FUNCTION v4_ssot.renta_pct(
  operating_result_total_usd numeric, 
  total_costs_usd numeric
) RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT CASE WHEN COALESCE(total_costs_usd,0) > 0
              THEN (COALESCE(operating_result_total_usd,0) / total_costs_usd) * 100
              ELSE 0 END
$$;

CREATE OR REPLACE FUNCTION v4_ssot.direct_cost_usd(
  p_labor_cost_usd numeric,
  p_supply_cost_usd numeric
) RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT COALESCE(p_labor_cost_usd, 0) + COALESCE(p_supply_cost_usd, 0)
$$;

CREATE OR REPLACE FUNCTION v4_ssot.direct_costs_total_for_project(p_project_id bigint) 
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    (SELECT COALESCE(SUM(v4_ssot.direct_cost_for_lot(l.id)), 0)::numeric
     FROM public.lots l
     JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
     WHERE f.project_id = p_project_id
       AND l.deleted_at IS NULL)
  , 0)::numeric
$$;

CREATE OR REPLACE FUNCTION v4_ssot.operating_result_total_for_project(p_project_id bigint) 
RETURNS numeric
LANGUAGE sql STABLE AS $$
  WITH project_totals AS (
    SELECT
      p.id,
      p.admin_cost,
      COALESCE(SUM(l.hectares), 0)::numeric AS total_hectares
    FROM public.projects p
    LEFT JOIN public.fields f ON f.project_id = p.id AND f.deleted_at IS NULL
    LEFT JOIN public.lots l ON l.field_id = f.id AND l.deleted_at IS NULL
    WHERE p.id = p_project_id AND p.deleted_at IS NULL
    GROUP BY p.id, p.admin_cost
  ),
  lease_cost AS (
    SELECT
      COALESCE(
        SUM(v4_ssot.rent_per_ha_for_lot(l.id) * l.hectares),
        0
      )::numeric AS total_lease
    FROM public.lots l
    JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
    WHERE f.project_id = p_project_id AND l.deleted_at IS NULL
  )
  SELECT COALESCE(
    (SELECT COALESCE(SUM(v4_ssot.income_net_total_for_lot(l.id)), 0)::numeric
     FROM public.lots l
     JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
     WHERE f.project_id = p_project_id AND l.deleted_at IS NULL)
    -
    v4_ssot.direct_costs_total_for_project(p_project_id)
    -
    (SELECT total_lease FROM lease_cost)
    -
    (SELECT COALESCE(admin_cost * total_hectares, 0)::numeric FROM project_totals)
  , 0)::numeric
$$;

CREATE OR REPLACE FUNCTION v4_ssot.total_invested_cost_for_project(p_project_id bigint) 
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    (SELECT COALESCE(SUM(v4_ssot.direct_cost_for_lot(l.id)), 0)::numeric
     FROM public.lots l
     JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
     WHERE f.project_id = p_project_id AND l.deleted_at IS NULL)
    +
    (SELECT COALESCE(SUM(v4_ssot.rent_per_ha_for_lot(l.id) * l.hectares), 0)::numeric
     FROM public.lots l
     JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
     WHERE f.project_id = p_project_id AND l.deleted_at IS NULL)
    +
    (SELECT COALESCE(SUM(v4_ssot.admin_cost_per_ha_for_lot(l.id) * l.hectares), 0)::numeric
     FROM public.lots l
     JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
     WHERE f.project_id = p_project_id AND l.deleted_at IS NULL)
  , 0)::numeric
$$;

CREATE OR REPLACE FUNCTION v4_ssot.supply_cost_for_project(p_project_id bigint)
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    (SELECT COALESCE(SUM(wi.total_used * s.price), 0)::numeric
     FROM public.workorders w
     JOIN public.workorder_items wi ON wi.workorder_id = w.id AND wi.deleted_at IS NULL
     JOIN public.supplies s ON s.id = wi.supply_id AND s.deleted_at IS NULL
     WHERE w.deleted_at IS NULL
       AND w.effective_area > 0
       AND wi.total_used > 0
       AND s.price IS NOT NULL
       AND w.project_id = p_project_id)
    +
    (SELECT COALESCE(SUM(sm.quantity * s.price), 0)::numeric
     FROM public.supply_movements sm
     JOIN public.supplies s ON s.id = sm.supply_id
     WHERE sm.deleted_at IS NULL
       AND s.deleted_at IS NULL
       AND sm.movement_type = 'Movimiento interno'
       AND sm.is_entry = false
       AND sm.project_id = p_project_id
       AND s.price IS NOT NULL
       AND sm.quantity > 0)
  , 0)::numeric
$$;

CREATE OR REPLACE FUNCTION v4_ssot.supply_cost_received_for_project(p_project_id bigint) 
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    (SELECT COALESCE(SUM(sm.quantity * s.price), 0)::numeric
     FROM public.supply_movements sm
     JOIN public.supplies s ON s.id = sm.supply_id
     WHERE sm.deleted_at IS NULL
       AND s.deleted_at IS NULL
       AND sm.movement_type = 'Movimiento interno entrada'
       AND sm.is_entry = true
       AND sm.project_id = p_project_id
       AND s.price IS NOT NULL
       AND sm.quantity > 0)
  , 0)::numeric
$$;

CREATE OR REPLACE FUNCTION v4_ssot.total_costs_for_crop(p_project_id bigint, p_crop_id bigint) 
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    (SELECT COALESCE(SUM(v4_ssot.direct_cost_for_lot(l.id)), 0)::numeric
     FROM public.lots l
     JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
     WHERE f.project_id = p_project_id
       AND l.current_crop_id = p_crop_id
       AND l.deleted_at IS NULL)
  , 0)::numeric
$$;

CREATE OR REPLACE FUNCTION v4_ssot.cost_per_ha_for_crop(p_project_id bigint, p_crop_id bigint) 
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT v4_core.per_ha(
    v4_ssot.total_costs_for_crop(p_project_id, p_crop_id),
    (SELECT COALESCE(SUM(l.hectares), 0)::numeric
     FROM public.lots l
     JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
     WHERE f.project_id = p_project_id 
       AND l.current_crop_id = p_crop_id 
       AND l.deleted_at IS NULL)
  )
$$;

CREATE OR REPLACE FUNCTION v4_ssot.cost_per_ha_for_crop_ssot(p_project_id bigint, p_crop_id bigint) 
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT v4_ssot.cost_per_ha_for_crop(p_project_id, p_crop_id)
$$;

CREATE OR REPLACE FUNCTION v4_ssot.seeds_invested_for_project_mb(p_project_id bigint) 
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    (SELECT SUM(sm.quantity * s.price)
     FROM public.supply_movements sm
     JOIN public.supplies s ON s.id = sm.supply_id
     JOIN public.categories c ON c.id = s.category_id
     WHERE sm.project_id = p_project_id 
       AND sm.deleted_at IS NULL
       AND s.deleted_at IS NULL
       AND c.type_id = 1
       AND sm.is_entry = TRUE
       AND sm.movement_type IN ('Stock', 'Remito oficial', 'Movimiento interno', 'Movimiento interno entrada'))
  , 0)::numeric
$$;

CREATE OR REPLACE FUNCTION v4_ssot.agrochemicals_invested_for_project_mb(p_project_id bigint) 
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    (SELECT SUM(sm.quantity * s.price)
     FROM public.supply_movements sm
     JOIN public.supplies s ON s.id = sm.supply_id
     JOIN public.categories c ON c.id = s.category_id
     WHERE sm.project_id = p_project_id 
       AND sm.deleted_at IS NULL
       AND s.deleted_at IS NULL
       AND c.type_id = 2
       AND sm.is_entry = TRUE
       AND sm.movement_type IN ('Stock', 'Remito oficial', 'Movimiento interno', 'Movimiento interno entrada'))
  , 0)::numeric
$$;

CREATE OR REPLACE FUNCTION v4_ssot.direct_costs_invested_for_project_mb(p_project_id bigint) 
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    (SELECT SUM(sm.quantity * s.price)
     FROM public.supply_movements sm
     JOIN public.supplies s ON s.id = sm.supply_id
     WHERE sm.project_id = p_project_id 
       AND sm.deleted_at IS NULL
       AND s.deleted_at IS NULL
       AND sm.movement_type IN ('Stock', 'Remito oficial'))
  , 0)::numeric
$$;

CREATE OR REPLACE FUNCTION v4_ssot.stock_value_for_project_mb(p_project_id bigint) 
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    (SELECT SUM(sm.quantity * s.price)
     FROM public.supply_movements sm
     JOIN public.supplies s ON s.id = sm.supply_id
     WHERE sm.project_id = p_project_id 
       AND sm.deleted_at IS NULL
       AND s.deleted_at IS NULL
       AND sm.movement_type IN ('Stock', 'Remito oficial'))
  , 0)::numeric - v4_ssot.direct_costs_total_for_project(p_project_id)
$$;

CREATE OR REPLACE FUNCTION v4_ssot.first_workorder_date_for_project(p_project_id bigint)
RETURNS date
LANGUAGE sql STABLE AS $$
  SELECT MIN(date)
  FROM public.workorders
  WHERE project_id = p_project_id 
    AND deleted_at IS NULL
$$;

CREATE OR REPLACE FUNCTION v4_ssot.last_workorder_date_for_project(p_project_id bigint)
RETURNS date
LANGUAGE sql STABLE AS $$
  SELECT MAX(date)
  FROM public.workorders
  WHERE project_id = p_project_id 
    AND deleted_at IS NULL
$$;

CREATE OR REPLACE FUNCTION v4_ssot.first_workorder_id_for_project(p_project_id bigint)
RETURNS bigint
LANGUAGE sql STABLE AS $$
  SELECT id
  FROM public.workorders
  WHERE project_id = p_project_id 
    AND deleted_at IS NULL
  ORDER BY date ASC, id ASC
  LIMIT 1
$$;

CREATE OR REPLACE FUNCTION v4_ssot.last_workorder_id_for_project(p_project_id bigint)
RETURNS bigint
LANGUAGE sql STABLE AS $$
  SELECT id
  FROM public.workorders
  WHERE project_id = p_project_id 
    AND deleted_at IS NULL
  ORDER BY date DESC, id DESC
  LIMIT 1
$$;

CREATE OR REPLACE FUNCTION v4_ssot.last_stock_count_date_for_project(p_project_id bigint)
RETURNS date
LANGUAGE sql STABLE AS $$
  SELECT MAX(close_date)
  FROM public.stocks
  WHERE project_id = p_project_id 
    AND deleted_at IS NULL
$$;

CREATE OR REPLACE FUNCTION v4_ssot.direct_cost_per_ha_usd(
  p_labor_cost_usd numeric,
  p_supply_cost_usd numeric,
  p_sowed_area_ha numeric
) RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT v4_core.safe_div(
    COALESCE(p_labor_cost_usd, 0) + COALESCE(p_supply_cost_usd, 0),
    COALESCE(p_sowed_area_ha, 0)
  )
$$;

CREATE OR REPLACE FUNCTION v4_ssot.supply_movements_invested_total_for_project(p_project_id bigint)
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    (SELECT SUM(sm.quantity * s.price)
     FROM public.supply_movements sm
     JOIN public.supplies s ON s.id = sm.supply_id
     WHERE sm.project_id = p_project_id 
       AND sm.deleted_at IS NULL
       AND s.deleted_at IS NULL
       AND sm.is_entry = TRUE
       AND sm.movement_type IN ('Stock', 'Remito oficial', 'Movimiento interno', 'Movimiento interno entrada'))
  , 0)::numeric
$$;

CREATE OR REPLACE FUNCTION v4_ssot.lease_executed_for_project(p_project_id bigint)
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(SUM(
    CASE
      WHEN f.lease_type_id IN (3, 4) THEN f.lease_type_value * l.hectares
      ELSE 0
    END
  ), 0)::numeric
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  WHERE f.project_id = p_project_id AND l.deleted_at IS NULL
$$;

CREATE OR REPLACE FUNCTION v4_ssot.lease_invested_for_project(p_project_id bigint)
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(SUM(v4_ssot.rent_per_ha_for_lot(l.id) * l.hectares), 0)::numeric
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  WHERE f.project_id = p_project_id AND l.deleted_at IS NULL
$$;

CREATE OR REPLACE FUNCTION v4_ssot.admin_cost_total_for_project(p_project_id bigint)
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    (SELECT p.admin_cost * v4_ssot.total_hectares_for_project(p_project_id)
     FROM public.projects p
     WHERE p.id = p_project_id AND p.deleted_at IS NULL)
  , 0)::numeric
$$;

CREATE OR REPLACE FUNCTION v4_ssot.total_costs_for_project(p_project_id bigint)
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    v4_ssot.direct_costs_total_for_project(p_project_id) + 
    v4_ssot.lease_invested_for_project(p_project_id) + 
    v4_ssot.admin_cost_total_for_project(p_project_id)
  , 0)::numeric
$$;

CREATE OR REPLACE FUNCTION v4_ssot.crop_incidence_for_project(p_project_id bigint)
RETURNS TABLE(
  current_crop_id bigint,
  crop_name text,
  crop_hectares numeric,
  crop_incidence_pct numeric
)
LANGUAGE sql STABLE AS $$
  WITH lot_base AS (
    SELECT
      l.current_crop_id,
      c.name AS crop_name,
      l.hectares
    FROM public.lots l
    JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
    LEFT JOIN public.crops c ON c.id = l.current_crop_id AND c.deleted_at IS NULL
    WHERE f.project_id = p_project_id 
      AND l.deleted_at IS NULL 
      AND l.hectares IS NOT NULL 
      AND l.hectares > 0
  ),
  project_totals AS (
    SELECT SUM(hectares)::numeric AS total_project_hectares
    FROM lot_base
  ),
  by_crop AS (
    SELECT 
      current_crop_id, 
      crop_name, 
      SUM(hectares)::numeric AS crop_hectares
    FROM lot_base
    WHERE current_crop_id IS NOT NULL
    GROUP BY current_crop_id, crop_name
  ),
  crop_percentages AS (
    SELECT
      bc.current_crop_id,
      bc.crop_name,
      bc.crop_hectares,
      pt.total_project_hectares,
      v4_core.percentage_rounded(bc.crop_hectares, pt.total_project_hectares) AS base_percentage,
      ROW_NUMBER() OVER (ORDER BY bc.crop_name) AS crop_order,
      COUNT(*) OVER () AS total_crops
    FROM by_crop bc
    CROSS JOIN project_totals pt
  ),
  project_sums AS (
    SELECT SUM(base_percentage) AS total_percentage
    FROM crop_percentages
  )
  SELECT
    cp.current_crop_id,
    cp.crop_name,
    cp.crop_hectares,
    CASE 
      WHEN ps.total_percentage > 99.000 AND cp.crop_order = cp.total_crops THEN
        100.000 - COALESCE((
          SELECT SUM(base_percentage) 
          FROM crop_percentages cp2 
          WHERE cp2.crop_order < cp.crop_order
        ), 0)
      ELSE
        cp.base_percentage
    END AS crop_incidence_pct
  FROM crop_percentages cp
  CROSS JOIN project_sums ps
  ORDER BY cp.crop_name
$$;

CREATE OR REPLACE FUNCTION v4_ssot.first_workorder_number_for_project(p_project_id bigint)
RETURNS text
LANGUAGE sql STABLE AS $$
  SELECT number::text
  FROM public.workorders
  WHERE project_id = p_project_id
    AND deleted_at IS NULL
  ORDER BY date ASC, id ASC
  LIMIT 1
$$;

CREATE OR REPLACE FUNCTION v4_ssot.last_workorder_number_for_project(p_project_id bigint)
RETURNS text
LANGUAGE sql STABLE AS $$
  SELECT number::text
  FROM public.workorders
  WHERE project_id = p_project_id
    AND deleted_at IS NULL
  ORDER BY date DESC, id DESC
  LIMIT 1
$$;

CREATE OR REPLACE FUNCTION v4_ssot.board_price_for_lot(p_lot_id bigint)
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(cc.board_price, 0)
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  LEFT JOIN public.crop_commercializations cc 
    ON cc.project_id = f.project_id 
   AND cc.crop_id = l.current_crop_id
   AND cc.deleted_at IS NULL
  WHERE l.id = p_lot_id 
    AND l.deleted_at IS NULL
  LIMIT 1
$$;

CREATE OR REPLACE FUNCTION v4_ssot.freight_cost_for_lot(p_lot_id bigint)
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(cc.freight_cost, 0)
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  LEFT JOIN public.crop_commercializations cc 
    ON cc.project_id = f.project_id 
   AND cc.crop_id = l.current_crop_id
   AND cc.deleted_at IS NULL
  WHERE l.id = p_lot_id 
    AND l.deleted_at IS NULL
  LIMIT 1
$$;

CREATE OR REPLACE FUNCTION v4_ssot.commercial_cost_for_lot(p_lot_id bigint)
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(cc.board_price * (cc.commercial_cost / 100.0), 0)
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  LEFT JOIN public.crop_commercializations cc 
    ON cc.project_id = f.project_id 
   AND cc.crop_id = l.current_crop_id
   AND cc.deleted_at IS NULL
  WHERE l.id = p_lot_id 
    AND l.deleted_at IS NULL
  LIMIT 1
$$;

CREATE OR REPLACE FUNCTION v4_ssot.rent_fixed_only_for_lot(p_lot_id bigint) 
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT
    CASE
      WHEN f.lease_type_id = 1 THEN 0
      
      WHEN f.lease_type_id = 2 THEN 0
      
      WHEN f.lease_type_id = 3 THEN COALESCE(f.lease_type_value, 0)
      
      WHEN f.lease_type_id = 4 THEN COALESCE(f.lease_type_value, 0)
      
      ELSE 0
    END::numeric
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  WHERE l.id = p_lot_id AND l.deleted_at IS NULL
$$;

CREATE OR REPLACE FUNCTION v4_ssot.labor_cost_pre_harvest_for_lot(p_lot_id bigint) 
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(SUM(lb.price * w.effective_area), 0)::numeric
  FROM public.workorders w
  JOIN public.labors lb ON lb.id = w.labor_id AND lb.deleted_at IS NULL
  JOIN public.categories cat ON cat.id = lb.category_id
  WHERE w.deleted_at IS NULL
    AND w.effective_area > 0
    AND lb.price IS NOT NULL
    AND w.lot_id = p_lot_id
    AND cat.type_id = 4
    AND cat.name != 'Cosecha'  -- EXCLUIR COSECHA
$$;

CREATE OR REPLACE FUNCTION v4_ssot.total_budget_cost_for_project(p_project_id bigint)
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(p.admin_cost * 10, 0)::numeric
  FROM public.projects p
  WHERE p.id = p_project_id AND p.deleted_at IS NULL
$$;

CREATE OR REPLACE FUNCTION v4_ssot.direct_costs_invested_for_project(p_project_id bigint)
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    (SELECT COALESCE(SUM(lb.price * l.hectares), 0)::numeric
     FROM public.lots l
     JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
     JOIN public.labors lb ON lb.project_id = f.project_id AND lb.deleted_at IS NULL
     WHERE f.project_id = p_project_id AND l.deleted_at IS NULL)
    +
    (SELECT COALESCE(SUM(s.price * st.initial_units), 0)::numeric
     FROM public.supplies s
     JOIN public.stocks st ON st.supply_id = s.id AND st.deleted_at IS NULL
     WHERE s.project_id = p_project_id AND s.deleted_at IS NULL
       AND st.initial_units IS NOT NULL)
    +
    v4_ssot.supply_cost_received_for_project(p_project_id)
  , 0)::numeric
$$;

CREATE OR REPLACE FUNCTION v4_ssot.stock_value_for_project(p_project_id bigint)
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    (SELECT COALESCE(SUM(s.price * st.initial_units), 0)::numeric
     FROM public.supplies s
     JOIN public.stocks st ON st.supply_id = s.id AND st.deleted_at IS NULL
     WHERE s.project_id = p_project_id 
       AND s.deleted_at IS NULL
       AND st.initial_units IS NOT NULL)
    -
    (SELECT COALESCE(SUM(wi.total_used * s.price), 0)::numeric
     FROM public.workorders w
     JOIN public.workorder_items wi ON wi.workorder_id = w.id AND wi.deleted_at IS NULL
     JOIN public.supplies s ON s.id = wi.supply_id AND s.deleted_at IS NULL
     WHERE w.project_id = p_project_id AND w.deleted_at IS NULL)
  , 0)::numeric
$$;

COMMIT;
