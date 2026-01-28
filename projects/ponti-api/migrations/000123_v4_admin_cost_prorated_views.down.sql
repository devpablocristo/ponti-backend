-- =============================================================
-- MIGRATION 000123 V4 ADMIN COST PRORATED VIEWS (DOWN)
-- =============================================================
-- Nota: SQL en inglés, comentarios en español

BEGIN;

-- Volver a admin_cost prorrateado como función principal
CREATE OR REPLACE FUNCTION v4_ssot.admin_cost_per_ha_for_lot(p_lot_id bigint)
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

CREATE OR REPLACE FUNCTION v4_ssot.admin_cost_total_for_project(p_project_id bigint)
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(p.admin_cost, 0)::double precision
  FROM public.projects p
  WHERE p.id = p_project_id AND p.deleted_at IS NULL
$$;

CREATE OR REPLACE FUNCTION v4_ssot.operating_result_total_for_project(p_project_id bigint)
RETURNS double precision
LANGUAGE sql STABLE AS $$
  WITH project_totals AS (
    SELECT
      p.id,
      p.admin_cost,
      COALESCE(SUM(l.hectares), 0)::double precision AS total_hectares
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
      )::double precision AS total_lease
    FROM public.lots l
    JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
    WHERE f.project_id = p_project_id AND l.deleted_at IS NULL
  )
  SELECT COALESCE(
    (SELECT COALESCE(SUM(v4_ssot.income_net_total_for_lot(l.id)), 0)::double precision
     FROM public.lots l
     JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
     WHERE f.project_id = p_project_id AND l.deleted_at IS NULL)
    -
    v4_ssot.direct_costs_total_for_project(p_project_id)
    -
    (SELECT total_lease FROM lease_cost)
    -
    (SELECT COALESCE(admin_cost, 0)::double precision FROM project_totals)
  , 0)::double precision
$$;

CREATE OR REPLACE FUNCTION v4_ssot.total_invested_cost_for_project(p_project_id bigint)
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    (SELECT COALESCE(SUM(v4_ssot.direct_cost_for_lot(l.id)), 0)::double precision
     FROM public.lots l
     JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
     WHERE f.project_id = p_project_id AND l.deleted_at IS NULL)
    +
    (SELECT COALESCE(SUM(v4_ssot.rent_per_ha_for_lot(l.id) * l.hectares), 0)::double precision
     FROM public.lots l
     JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
     WHERE f.project_id = p_project_id AND l.deleted_at IS NULL)
    +
    v4_ssot.admin_cost_total_for_project(p_project_id)
  , 0)::double precision
$$;

CREATE OR REPLACE FUNCTION v4_ssot.total_budget_cost_for_project(p_project_id bigint)
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(p.admin_cost, 0)::numeric
  FROM public.projects p
  WHERE p.id = p_project_id AND p.deleted_at IS NULL
$$;

DROP FUNCTION IF EXISTS v4_ssot.admin_cost_prorated_per_ha_for_lot(bigint);

-- Restaurar views con admin_cost_per_ha_for_lot
CREATE OR REPLACE VIEW v4_calc.field_crop_aggregated AS
SELECT
  project_id,
  field_id,
  crop_id,
  MIN(lot_id) AS sample_lot_id,
  SUM(tons)::numeric AS production_tn,
  SUM(sowed_area_ha)::numeric AS sown_area_ha,
  SUM(surface_ha)::numeric AS surface_ha,
  SUM(v4_ssot.labor_cost_for_lot(lot_id))::numeric AS labor_costs_usd,
  SUM(total_insumos_usd)::numeric AS supply_costs_usd,
  SUM(v4_ssot.labor_cost_for_lot(lot_id) + total_insumos_usd)::numeric AS direct_cost_usd,
  SUM(v4_ssot.rent_fixed_only_for_lot(lot_id) * surface_ha)::numeric AS rent_fixed_usd,
  SUM(v4_ssot.rent_per_ha_for_lot(lot_id) * surface_ha)::numeric AS rent_total_usd,
  SUM(v4_ssot.admin_cost_per_ha_for_lot(lot_id) * surface_ha)::numeric AS administration_usd
FROM v4_calc.field_crop_supply_costs_by_lot
GROUP BY project_id, field_id, crop_id;

CREATE OR REPLACE VIEW v4_calc.field_crop_metrics_lot_base AS
SELECT
  f.project_id,
  f.id AS field_id,
  f.name AS field_name,
  l.current_crop_id,
  c.name AS crop_name,
  l.id AS lot_id,
  l.hectares,
  l.tons,
  COALESCE(v4_ssot.seeded_area_for_lot(l.id), 0)::numeric AS sowed_area_ha,
  COALESCE(v4_ssot.harvested_area_for_lot(l.id), 0)::numeric AS harvested_area_ha,
  COALESCE(v4_ssot.yield_tn_per_ha_for_lot(l.id), 0) AS yield_tn_per_ha,
  COALESCE(v4_ssot.labor_cost_for_lot(l.id), 0)::numeric AS labor_cost_usd,
  COALESCE(v4_ssot.supply_cost_for_lot_base(l.id), 0)::numeric AS supply_cost_usd,
  COALESCE(v4_ssot.net_price_usd_for_lot(l.id), 0)::numeric AS net_price_usd,
  COALESCE(v4_ssot.rent_per_ha_for_lot(l.id), 0)::numeric AS rent_per_ha,
  COALESCE(v4_ssot.admin_cost_per_ha_for_lot(l.id), 0)::numeric AS admin_per_ha,
  COALESCE(v4_ssot.board_price_for_lot(l.id), 0)::numeric AS board_price,
  COALESCE(v4_ssot.freight_cost_for_lot(l.id), 0)::numeric AS freight_cost,
  COALESCE(v4_ssot.commercial_cost_for_lot(l.id), 0)::numeric AS commercial_cost
FROM public.lots l
JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
JOIN public.projects p ON p.id = f.project_id AND p.deleted_at IS NULL
LEFT JOIN public.crops c ON c.id = l.current_crop_id AND c.deleted_at IS NULL
WHERE l.deleted_at IS NULL
  AND l.current_crop_id IS NOT NULL
  AND l.hectares > 0;

CREATE OR REPLACE VIEW v4_calc.investor_contribution_categories AS
WITH lot_base AS (
  SELECT
    f.project_id,
    l.id AS lot_id,
    l.hectares,
    COALESCE((
      SELECT SUM(w.effective_area)
      FROM public.workorders w
      JOIN public.labors lab ON w.labor_id = lab.id
      JOIN public.categories cat ON lab.category_id = cat.id
      WHERE w.lot_id = l.id
        AND w.deleted_at IS NULL
        AND cat.name = 'Siembra'
        AND cat.type_id = 4
    ), 0)::numeric AS seeded_area_ha
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  JOIN public.projects p ON p.id = f.project_id AND p.deleted_at IS NULL
  WHERE l.deleted_at IS NULL
),
seed_area AS (
  SELECT
    lb.project_id,
    COALESCE(SUM(lb.seeded_area_ha), 0)::numeric AS total_seeded_area_ha,
    COALESCE(SUM(v4_ssot.rent_fixed_only_for_lot(lb.lot_id) * lb.hectares), 0)::numeric AS rent_capitalizable_total_usd,
    COALESCE(SUM(v4_ssot.admin_cost_per_ha_for_lot(lb.lot_id) * lb.hectares), 0)::numeric AS administration_total_usd
  FROM lot_base lb
  GROUP BY lb.project_id
),
labor_totals AS (
  SELECT
    lb.project_id,
    COALESCE(SUM(CASE WHEN cat.name IN ('Pulverización', 'Otras Labores') THEN lab.price * w.effective_area ELSE 0 END), 0)::numeric AS general_labors_total_usd,
    COALESCE(SUM(CASE WHEN cat.name = 'Siembra' THEN lab.price * w.effective_area ELSE 0 END), 0)::numeric AS sowing_total_usd,
    COALESCE(SUM(CASE WHEN cat.name = 'Riego' THEN lab.price * w.effective_area ELSE 0 END), 0)::numeric AS irrigation_total_usd
  FROM lot_base lb
  JOIN public.workorders w ON w.lot_id = lb.lot_id AND w.deleted_at IS NULL
  JOIN public.labors lab ON lab.id = w.labor_id AND lab.deleted_at IS NULL
  JOIN public.categories cat ON cat.id = lab.category_id
  WHERE cat.type_id = 4
  GROUP BY lb.project_id
),
invested_totals AS (
  SELECT
    p.project_id,
    v4_ssot.seeds_invested_for_project_mb(p.project_id)::numeric AS seeds_total_usd,
    v4_ssot.agrochemicals_invested_for_project_mb(p.project_id)::numeric AS agrochemicals_total_usd,
    (
      SELECT COALESCE(SUM(sm.quantity * s.price), 0)::numeric
      FROM public.supply_movements sm
      JOIN public.supplies s ON s.id = sm.supply_id AND s.deleted_at IS NULL
      JOIN public.categories cat ON cat.id = s.category_id
      WHERE sm.project_id = p.project_id
        AND sm.deleted_at IS NULL
        AND sm.is_entry = TRUE
        AND sm.movement_type IN ('Stock', 'Remito oficial', 'Movimiento interno', 'Movimiento interno entrada')
        AND cat.type_id = 3
    ) AS fertilizers_total_usd
  FROM (SELECT DISTINCT project_id FROM lot_base) p
)
SELECT
  it.project_id,
  it.agrochemicals_total_usd,
  COALESCE(it.fertilizers_total_usd, 0)::numeric AS fertilizers_total_usd,
  it.seeds_total_usd,
  COALESCE(lt.general_labors_total_usd, 0)::numeric AS general_labors_total_usd,
  COALESCE(lt.sowing_total_usd, 0)::numeric AS sowing_total_usd,
  COALESCE(lt.irrigation_total_usd, 0)::numeric AS irrigation_total_usd,
  COALESCE(sa.rent_capitalizable_total_usd, 0)::numeric AS rent_capitalizable_total_usd,
  COALESCE(sa.administration_total_usd, 0)::numeric AS administration_total_usd,
  COALESCE(sa.total_seeded_area_ha, 0)::numeric AS total_seeded_area_ha
FROM invested_totals it
LEFT JOIN labor_totals lt ON lt.project_id = it.project_id
LEFT JOIN seed_area sa ON sa.project_id = it.project_id;

CREATE OR REPLACE VIEW v4_report.investor_project_base AS
SELECT
  p.id AS project_id,
  p.name AS project_name,
  p.customer_id,
  c.name AS customer_name,
  p.campaign_id,
  cam.name AS campaign_name,
  COALESCE(SUM(l.hectares), 0::double precision)::numeric AS surface_total_ha,
  COALESCE(SUM(v4_ssot.rent_fixed_only_for_lot(l.id) * l.hectares), 0::double precision)::numeric AS lease_fixed_total_usd,
  CASE
    WHEN COALESCE(SUM(v4_ssot.rent_fixed_only_for_lot(l.id) * l.hectares), 0::double precision) > 0::double precision THEN TRUE
    ELSE FALSE
  END AS lease_is_fixed,
  CASE
    WHEN COALESCE(SUM(l.hectares), 0::double precision) > 0::double precision THEN
      COALESCE(SUM(v4_ssot.rent_fixed_only_for_lot(l.id) * l.hectares), 0::double precision) / SUM(l.hectares)
    ELSE 0::double precision
  END::numeric AS lease_per_ha_usd,
  COALESCE(SUM(v4_ssot.admin_cost_per_ha_for_lot(l.id) * l.hectares), 0::double precision)::numeric AS admin_total_usd,
  CASE
    WHEN COALESCE(SUM(l.hectares), 0::double precision) > 0::double precision THEN
      COALESCE(SUM(v4_ssot.admin_cost_per_ha_for_lot(l.id) * l.hectares), 0::double precision) / SUM(l.hectares)
    ELSE 0::double precision
  END::numeric AS admin_per_ha_usd
FROM public.projects p
JOIN public.customers c ON p.customer_id = c.id AND c.deleted_at IS NULL
JOIN public.campaigns cam ON p.campaign_id = cam.id AND cam.deleted_at IS NULL
LEFT JOIN public.fields f ON f.project_id = p.id AND f.deleted_at IS NULL
LEFT JOIN public.lots l ON l.field_id = f.id AND l.deleted_at IS NULL
WHERE p.deleted_at IS NULL
GROUP BY
  p.id,
  p.name,
  p.customer_id,
  c.name,
  p.campaign_id,
  cam.name;

COMMIT;
