-- ========================================
-- MIGRATION 000191 INVESTOR REAL CONTRIBUTIONS (SPLITS) (UP)
-- ========================================
-- Nota: SQL en inglés, comentarios en español

BEGIN;

-- Reemplaza la vista para que los costos de labores puedan repartirse por inversor
-- cuando existe un split en la OT (sin duplicar workorders).
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

-- Allocation: if a workorder has splits, use them; otherwise attribute 100% to workorders.investor_id.
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

  ct.total_contributions_usd AS project_total_contributions_usd,

  CASE
    WHEN ct.total_contributions_usd > 0
    THEN (
      (
        COALESCE(agro.agrochemicals_real_usd, 0) +
        COALESCE(fert.fertilizers_real_usd, 0) +
        COALESCE(seed.seeds_real_usd, 0) +
        COALESCE(glabor.general_labors_real_usd, 0) +
        COALESCE(sow.sowing_real_usd, 0) +
        COALESCE(irrig.irrigation_real_usd, 0) +
        COALESCE(rri.rent_real_usd, 0) +
        COALESCE(ia.admin_real_usd, 0)
      ) / ct.total_contributions_usd * 100
    )::numeric
    ELSE 0::numeric
  END AS contributions_progress_pct
FROM investor_base ib
JOIN category_totals ct ON ct.project_id = ib.project_id
LEFT JOIN investor_agrochemicals_real agro
  ON agro.project_id = ib.project_id AND agro.investor_id = ib.investor_id
LEFT JOIN investor_fertilizers_real fert
  ON fert.project_id = ib.project_id AND fert.investor_id = ib.investor_id
LEFT JOIN investor_seeds_real seed
  ON seed.project_id = ib.project_id AND seed.investor_id = ib.investor_id
LEFT JOIN investor_general_labors_real glabor
  ON glabor.project_id = ib.project_id AND glabor.investor_id = ib.investor_id
LEFT JOIN investor_sowing_real sow
  ON sow.project_id = ib.project_id AND sow.investor_id = ib.investor_id
LEFT JOIN investor_irrigation_real irrig
  ON irrig.project_id = ib.project_id AND irrig.investor_id = ib.investor_id
LEFT JOIN investor_rent_real rri
  ON rri.project_id = ib.project_id AND rri.investor_id = ib.investor_id
LEFT JOIN investor_admin_real ia
  ON ia.project_id = ib.project_id AND ia.investor_id = ib.investor_id;

COMMIT;

