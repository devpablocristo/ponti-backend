-- =============================================================================
-- Migration: 000326_fix_summary_results_rent_bug
-- Descripción: Copia v3 pero corrige el bug del arriendo
-- Bug: v3 usa rent_fixed_only_for_lot (solo fijo: 119,770)
-- Fix: Cambiar a rent_per_ha_for_lot (total: 161,773)
-- =============================================================================

DROP VIEW IF EXISTS v4_report.summary_results;

CREATE VIEW v4_report.summary_results AS
WITH lot_base AS (
  SELECT 
    l.id AS lot_id,
    f.project_id,
    l.current_crop_id,
    c.name AS crop_name,
    l.hectares,
    COALESCE(l.tons, 0) AS tons,
    COALESCE((
      SELECT sum(w.effective_area)
      FROM workorders w
      JOIN labors lab ON w.labor_id = lab.id
      JOIN categories cat ON lab.category_id = cat.id
      WHERE w.lot_id = l.id 
        AND w.deleted_at IS NULL 
        AND cat.name = 'Siembra' 
        AND cat.type_id = 4
    ), 0) AS seeded_area_ha
  FROM lots l
  JOIN fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  JOIN projects p ON p.id = f.project_id AND p.deleted_at IS NULL
  LEFT JOIN crops c ON c.id = l.current_crop_id AND c.deleted_at IS NULL
  WHERE l.deleted_at IS NULL AND l.hectares > 0
),
by_crop AS (
  SELECT
    lb.project_id,
    lb.current_crop_id,
    lb.crop_name::text AS crop_name,
    COALESCE(sum(lb.seeded_area_ha), 0)::numeric AS surface_ha,
    COALESCE(sum(v3_calc.income_net_total_for_lot(lb.lot_id)), 0)::numeric AS net_income_usd,
    COALESCE(sum(v3_lot_ssot.labor_cost_for_lot(lb.lot_id)::double precision + v3_lot_ssot.supply_cost_for_lot_base(lb.lot_id)::double precision), 0)::numeric AS direct_costs_usd,
    -- FIX: Cambiado de rent_fixed_only_for_lot a rent_per_ha_for_lot
    COALESCE(sum(lb.seeded_area_ha::double precision * v3_lot_ssot.rent_per_ha_for_lot(lb.lot_id)), 0)::numeric AS rent_usd,
    COALESCE(sum(lb.seeded_area_ha::double precision * v3_calc.admin_cost_per_ha_for_lot(lb.lot_id)), 0)::numeric AS structure_usd
  FROM lot_base lb
  WHERE lb.current_crop_id IS NOT NULL
  GROUP BY lb.project_id, lb.current_crop_id, lb.crop_name
),
project_totals AS (
  SELECT
    project_id,
    sum(surface_ha) AS total_surface_ha,
    sum(net_income_usd) AS total_net_income_usd,
    sum(direct_costs_usd) AS total_direct_costs_usd,
    sum(rent_usd) AS total_rent_usd,
    sum(structure_usd) AS total_structure_usd,
    sum(direct_costs_usd + rent_usd + structure_usd) AS total_invested_usd,
    sum(net_income_usd - (direct_costs_usd + rent_usd + structure_usd)) AS total_operating_result_usd
  FROM by_crop
  GROUP BY project_id
)
SELECT
  bc.project_id,
  bc.current_crop_id,
  bc.crop_name,
  bc.surface_ha,
  bc.net_income_usd,
  bc.direct_costs_usd,
  bc.rent_usd,
  bc.structure_usd,
  (bc.direct_costs_usd + bc.rent_usd + bc.structure_usd)::numeric AS total_invested_usd,
  (bc.net_income_usd - (bc.direct_costs_usd + bc.rent_usd + bc.structure_usd))::numeric AS operating_result_usd,
  v3_calc.renta_pct(
    (bc.net_income_usd - (bc.direct_costs_usd + bc.rent_usd + bc.structure_usd))::double precision,
    (bc.direct_costs_usd + bc.rent_usd + bc.structure_usd)::double precision
  )::numeric AS crop_return_pct,
  pt.total_surface_ha,
  pt.total_net_income_usd,
  pt.total_direct_costs_usd,
  pt.total_rent_usd,
  pt.total_structure_usd,
  pt.total_invested_usd AS total_invested_project_usd,
  pt.total_operating_result_usd,
  v3_calc.renta_pct(
    pt.total_operating_result_usd::double precision,
    pt.total_invested_usd::double precision
  )::numeric AS project_return_pct
FROM by_crop bc
JOIN project_totals pt ON pt.project_id = bc.project_id;

COMMENT ON VIEW v4_report.summary_results IS 
'FIX 000326: Igual que v3 pero con arriendo corregido (rent_per_ha_for_lot en vez de rent_fixed_only).';
