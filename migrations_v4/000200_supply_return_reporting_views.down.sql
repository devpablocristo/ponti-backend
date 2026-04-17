BEGIN;

CREATE OR REPLACE FUNCTION v4_ssot.seeds_invested_for_project_mb(p_project_id bigint)
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    (SELECT SUM(sm.quantity * s.price)
     FROM public.supply_movements sm
     JOIN public.supplies s ON s.id = sm.supply_id
     JOIN public.categories c ON c.id = s.category_id AND c.deleted_at IS NULL
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
     JOIN public.categories c ON c.id = s.category_id AND c.deleted_at IS NULL
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

CREATE OR REPLACE VIEW v4_calc.dashboard_fertilizers_invested_by_project AS
SELECT
  sm.project_id,
  COALESCE(SUM(sm.quantity * s.price), 0) AS fertilizantes_invertidos_usd
FROM public.supply_movements sm
JOIN public.supplies s ON s.id = sm.supply_id
JOIN public.categories c ON s.category_id = c.id AND c.deleted_at IS NULL
WHERE sm.deleted_at IS NULL
  AND s.deleted_at IS NULL
  AND sm.is_entry = TRUE
  AND sm.movement_type IN ('Stock', 'Remito oficial', 'Movimiento interno', 'Movimiento interno entrada')
  AND c.type_id = 3
GROUP BY sm.project_id;

CREATE OR REPLACE VIEW v4_calc.investor_contribution_categories AS
WITH lot_base AS (
  SELECT
    f.project_id,
    l.id AS lot_id,
    l.hectares,
    COALESCE((
      SELECT SUM(w.effective_area)
      FROM public.workorders w
      JOIN public.labors lab ON w.labor_id = lab.id AND lab.deleted_at IS NULL
      JOIN public.categories cat ON lab.category_id = cat.id AND cat.deleted_at IS NULL
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
  JOIN public.labors lab ON lab.id = w.labor_id AND lab.deleted_at IS NULL AND lab.deleted_at IS NULL
  JOIN public.categories cat ON cat.id = lab.category_id AND cat.deleted_at IS NULL
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
      JOIN public.categories cat ON cat.id = s.category_id AND cat.deleted_at IS NULL
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

CREATE OR REPLACE VIEW v4_calc.investor_real_contributions AS
WITH investor_base AS (
  SELECT
    pi.project_id,
    pi.investor_id,
    i.name AS investor_name,
    pi.percentage AS share_pct_agreed
  FROM public.project_investors pi
  JOIN public.investors i ON i.id = pi.investor_id AND i.deleted_at IS NULL
  WHERE pi.deleted_at IS NULL
),
admin_config AS (
  SELECT
    project_id,
    COUNT(*) FILTER (WHERE deleted_at IS NULL) > 0 AS has_custom_admin
  FROM public.admin_cost_investors
  GROUP BY project_id
),
category_totals AS (
  SELECT
    project_id,
    agrochemicals_total_usd,
    fertilizers_total_usd,
    seeds_total_usd,
    general_labors_total_usd,
    sowing_total_usd,
    irrigation_total_usd,
    rent_capitalizable_total_usd,
    administration_total_usd,
    total_seeded_area_ha,
    (agrochemicals_total_usd + fertilizers_total_usd + seeds_total_usd +
     general_labors_total_usd + sowing_total_usd + irrigation_total_usd +
     rent_capitalizable_total_usd + administration_total_usd) AS total_contributions_usd
  FROM v4_calc.investor_contribution_categories
),
investor_agrochemicals_real AS (
  SELECT
    sm.project_id,
    sm.investor_id,
    COALESCE(SUM(sm.quantity * s.price), 0)::numeric AS agrochemicals_real_usd
  FROM public.supply_movements sm
  JOIN public.supplies s ON s.id = sm.supply_id AND s.deleted_at IS NULL
  JOIN public.categories cat ON cat.id = s.category_id AND cat.deleted_at IS NULL
  WHERE sm.deleted_at IS NULL
    AND sm.is_entry = TRUE
    AND sm.movement_type IN ('Stock', 'Remito oficial', 'Movimiento interno', 'Movimiento interno entrada')
    AND cat.type_id = 2
  GROUP BY sm.project_id, sm.investor_id
),
investor_fertilizers_real AS (
  SELECT
    sm.project_id,
    sm.investor_id,
    COALESCE(SUM(sm.quantity * s.price), 0)::numeric AS fertilizers_real_usd
  FROM public.supply_movements sm
  JOIN public.supplies s ON s.id = sm.supply_id AND s.deleted_at IS NULL
  JOIN public.categories cat ON cat.id = s.category_id AND cat.deleted_at IS NULL
  WHERE sm.deleted_at IS NULL
    AND sm.is_entry = TRUE
    AND sm.movement_type IN ('Stock', 'Remito oficial', 'Movimiento interno', 'Movimiento interno entrada')
    AND cat.type_id = 3
  GROUP BY sm.project_id, sm.investor_id
),
investor_seeds_real AS (
  SELECT
    sm.project_id,
    sm.investor_id,
    COALESCE(SUM(sm.quantity * s.price), 0)::numeric AS seeds_real_usd
  FROM public.supply_movements sm
  JOIN public.supplies s ON s.id = sm.supply_id AND s.deleted_at IS NULL
  JOIN public.categories cat ON cat.id = s.category_id AND cat.deleted_at IS NULL
  WHERE sm.deleted_at IS NULL
    AND sm.is_entry = TRUE
    AND sm.movement_type IN ('Stock', 'Remito oficial', 'Movimiento interno', 'Movimiento interno entrada')
    AND cat.type_id = 1
  GROUP BY sm.project_id, sm.investor_id
),
workorder_alloc AS (
  SELECT
    w.id AS workorder_id,
    w.project_id,
    w.investor_id,
    1::numeric AS factor
  FROM public.workorders w
  WHERE w.deleted_at IS NULL
    AND NOT EXISTS (
      SELECT 1
      FROM public.workorder_investor_splits wis
      WHERE wis.workorder_id = w.id
        AND wis.deleted_at IS NULL
    )
  UNION ALL
  SELECT
    w.id AS workorder_id,
    w.project_id,
    wis.investor_id,
    (wis.percentage::numeric / 100)::numeric AS factor
  FROM public.workorders w
  JOIN public.workorder_investor_splits wis
    ON wis.workorder_id = w.id
   AND wis.deleted_at IS NULL
  WHERE w.deleted_at IS NULL
),
investor_general_labors_real AS (
  SELECT
    w.project_id,
    a.investor_id,
    COALESCE(SUM(lab.price * (w.effective_area * a.factor)), 0)::numeric AS general_labors_real_usd
  FROM public.workorders w
  JOIN workorder_alloc a ON a.workorder_id = w.id
  JOIN public.labors lab ON w.labor_id = lab.id AND lab.deleted_at IS NULL AND lab.deleted_at IS NULL
  JOIN public.categories cat ON cat.id = lab.category_id AND cat.deleted_at IS NULL
  WHERE w.deleted_at IS NULL
    AND cat.type_id = 4
    AND cat.name IN ('Pulverización', 'Otras Labores')
  GROUP BY w.project_id, a.investor_id
),
investor_sowing_real AS (
  SELECT
    w.project_id,
    a.investor_id,
    COALESCE(SUM(lab.price * (w.effective_area * a.factor)), 0)::numeric AS sowing_real_usd
  FROM public.workorders w
  JOIN workorder_alloc a ON a.workorder_id = w.id
  JOIN public.labors lab ON w.labor_id = lab.id AND lab.deleted_at IS NULL AND lab.deleted_at IS NULL
  JOIN public.categories cat ON cat.id = lab.category_id AND cat.deleted_at IS NULL
  WHERE w.deleted_at IS NULL
    AND cat.type_id = 4
    AND cat.name = 'Siembra'
  GROUP BY w.project_id, a.investor_id
),
investor_irrigation_real AS (
  SELECT
    w.project_id,
    a.investor_id,
    COALESCE(SUM(lab.price * (w.effective_area * a.factor)), 0)::numeric AS irrigation_real_usd
  FROM public.workorders w
  JOIN workorder_alloc a ON a.workorder_id = w.id
  JOIN public.labors lab ON w.labor_id = lab.id AND lab.deleted_at IS NULL AND lab.deleted_at IS NULL
  JOIN public.categories cat ON cat.id = lab.category_id AND cat.deleted_at IS NULL
  WHERE w.deleted_at IS NULL
    AND cat.type_id = 4
    AND cat.name = 'Riego'
  GROUP BY w.project_id, a.investor_id
),
investor_rent_real AS (
  SELECT
    f.project_id,
    fi.investor_id,
    SUM(
      v4_ssot.rent_fixed_only_for_lot(l.id) * l.hectares *
      COALESCE(fi.percentage, 0)::numeric / 100
    )::numeric AS rent_real_usd
  FROM public.fields f
  JOIN public.lots l ON l.field_id = f.id AND l.deleted_at IS NULL
  LEFT JOIN public.field_investors fi ON fi.field_id = f.id AND fi.deleted_at IS NULL
  WHERE f.deleted_at IS NULL
  GROUP BY f.project_id, fi.investor_id
),
investor_admin_real AS (
  SELECT
    ib.project_id,
    ib.investor_id,
    (
      CASE
        WHEN COALESCE(ac.has_custom_admin, FALSE)
          THEN (ct.administration_total_usd * COALESCE(aci.percentage, 0)::numeric / 100)
        ELSE (ct.administration_total_usd * ib.share_pct_agreed / 100)
      END
    )::numeric AS admin_real_usd
  FROM investor_base ib
  JOIN category_totals ct ON ct.project_id = ib.project_id
  LEFT JOIN admin_config ac ON ac.project_id = ib.project_id
  LEFT JOIN public.admin_cost_investors aci
    ON aci.project_id = ib.project_id
   AND aci.investor_id = ib.investor_id
   AND aci.deleted_at IS NULL
)
SELECT
  ib.project_id,
  ib.investor_id,
  ib.investor_name,
  ib.share_pct_agreed,
  COALESCE(agro.agrochemicals_real_usd, 0)::numeric AS agrochemicals_real_usd,
  COALESCE(fert.fertilizers_real_usd, 0)::numeric AS fertilizers_real_usd,
  COALESCE(seed.seeds_real_usd, 0)::numeric AS seeds_real_usd,
  COALESCE(glabor.general_labors_real_usd, 0)::numeric AS general_labors_real_usd,
  COALESCE(sow.sowing_real_usd, 0)::numeric AS sowing_real_usd,
  COALESCE(irrig.irrigation_real_usd, 0)::numeric AS irrigation_real_usd,
  COALESCE(rri.rent_real_usd, 0)::numeric AS rent_real_usd,
  COALESCE(ia.admin_real_usd, 0)::numeric AS administration_real_usd,
  (
    COALESCE(agro.agrochemicals_real_usd, 0) +
    COALESCE(fert.fertilizers_real_usd, 0) +
    COALESCE(seed.seeds_real_usd, 0) +
    COALESCE(glabor.general_labors_real_usd, 0) +
    COALESCE(sow.sowing_real_usd, 0) +
    COALESCE(irrig.irrigation_real_usd, 0) +
    COALESCE(rri.rent_real_usd, 0) +
    COALESCE(ia.admin_real_usd, 0)
  )::numeric AS total_real_contribution_usd,
  (
    COALESCE(agro.agrochemicals_real_usd, 0) - (ct.agrochemicals_total_usd * ib.share_pct_agreed / 100)
  )::numeric AS agrochemicals_diff_usd,
  (
    COALESCE(fert.fertilizers_real_usd, 0) - (ct.fertilizers_total_usd * ib.share_pct_agreed / 100)
  )::numeric AS fertilizers_diff_usd,
  (
    COALESCE(seed.seeds_real_usd, 0) - (ct.seeds_total_usd * ib.share_pct_agreed / 100)
  )::numeric AS seeds_diff_usd,
  (
    COALESCE(glabor.general_labors_real_usd, 0) - (ct.general_labors_total_usd * ib.share_pct_agreed / 100)
  )::numeric AS general_labors_diff_usd,
  (
    COALESCE(sow.sowing_real_usd, 0) - (ct.sowing_total_usd * ib.share_pct_agreed / 100)
  )::numeric AS sowing_diff_usd,
  (
    COALESCE(irrig.irrigation_real_usd, 0) - (ct.irrigation_total_usd * ib.share_pct_agreed / 100)
  )::numeric AS irrigation_diff_usd,
  (
    COALESCE(rri.rent_real_usd, 0) - (ct.rent_capitalizable_total_usd * ib.share_pct_agreed / 100)
  )::numeric AS rent_diff_usd,
  (
    COALESCE(ia.admin_real_usd, 0) - (ct.administration_total_usd * ib.share_pct_agreed / 100)
  )::numeric AS administration_diff_usd
FROM investor_base ib
JOIN category_totals ct ON ct.project_id = ib.project_id
LEFT JOIN investor_agrochemicals_real agro
  ON agro.project_id = ib.project_id
 AND agro.investor_id = ib.investor_id
LEFT JOIN investor_fertilizers_real fert
  ON fert.project_id = ib.project_id
 AND fert.investor_id = ib.investor_id
LEFT JOIN investor_seeds_real seed
  ON seed.project_id = ib.project_id
 AND seed.investor_id = ib.investor_id
LEFT JOIN investor_general_labors_real glabor
  ON glabor.project_id = ib.project_id
 AND glabor.investor_id = ib.investor_id
LEFT JOIN investor_sowing_real sow
  ON sow.project_id = ib.project_id
 AND sow.investor_id = ib.investor_id
LEFT JOIN investor_irrigation_real irrig
  ON irrig.project_id = ib.project_id
 AND irrig.investor_id = ib.investor_id
LEFT JOIN investor_rent_real rri
  ON rri.project_id = ib.project_id
 AND rri.investor_id = ib.investor_id
LEFT JOIN investor_admin_real ia
  ON ia.project_id = ib.project_id
 AND ia.investor_id = ib.investor_id;

COMMIT;
