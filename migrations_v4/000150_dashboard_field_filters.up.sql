-- ========================================
-- MIGRATION 000150 DASHBOARD FIELD FILTERS (UP)
-- ========================================
-- Nota: SQL en inglés, comentarios en español

BEGIN;

-- Vista de métricas por campo
CREATE OR REPLACE VIEW v4_report.dashboard_metrics_field AS
WITH
lot_data AS (
  SELECT
    lm.project_id,
    lm.field_id,
    lm.lot_id,
    lm.hectares,
    lm.seeded_area_ha,
    lm.harvested_area_ha,
    lm.direct_cost_per_ha_usd
  FROM v4_report.lot_metrics lm
),
field_hectares AS (
  SELECT project_id, field_id, SUM(hectares)::numeric AS total_hectares
  FROM v4_report.lot_metrics
  GROUP BY project_id, field_id
),
project_hectares AS (
  SELECT project_id, SUM(hectares)::numeric AS total_hectares
  FROM v4_report.lot_metrics
  GROUP BY project_id
),
rent_fixed_ssot AS (
  SELECT
    f.project_id,
    f.id AS field_id,
    SUM(v4_ssot.rent_fixed_only_for_lot(l.id) * l.hectares)::numeric AS rent_fixed_total_usd
  FROM public.fields f
  JOIN public.lots l ON l.field_id = f.id AND l.deleted_at IS NULL
  WHERE f.deleted_at IS NULL
  GROUP BY f.project_id, f.id
),
direct_costs AS (
  SELECT
    project_id,
    field_id,
    SUM(v4_ssot.direct_cost_for_lot(lot_id))::numeric AS direct_costs_total_usd
  FROM v4_report.lot_metrics
  GROUP BY project_id, field_id
),
admin_costs AS (
  SELECT
    project_id,
    field_id,
    SUM(v4_ssot.admin_cost_per_ha_for_lot(lot_id) * hectares)::numeric AS admin_total_usd
  FROM v4_report.lot_metrics
  GROUP BY project_id, field_id
),
rent_totals AS (
  SELECT
    project_id,
    field_id,
    SUM(v4_ssot.rent_per_ha_for_lot(lot_id) * hectares)::numeric AS rent_total_usd
  FROM v4_report.lot_metrics
  GROUP BY project_id, field_id
),
income_totals AS (
  SELECT
    project_id,
    field_id,
    SUM(v4_ssot.income_net_total_for_lot(lot_id))::numeric AS income_net_total_usd
  FROM v4_report.lot_metrics
  GROUP BY project_id, field_id
)
SELECT
  p.customer_id,
  p.id AS project_id,
  p.campaign_id,
  ld.field_id,

  COALESCE(SUM(ld.seeded_area_ha), 0)::numeric AS sowing_hectares,
  COALESCE(SUM(ld.hectares), 0)::numeric AS sowing_total_hectares,
  v4_core.percentage(
    COALESCE(SUM(ld.seeded_area_ha), 0)::numeric,
    COALESCE(SUM(ld.hectares), 0)::numeric
  ) AS sowing_progress_pct,

  COALESCE(SUM(ld.harvested_area_ha), 0)::numeric AS harvest_hectares,
  COALESCE(SUM(ld.hectares), 0)::numeric AS harvest_total_hectares,
  v4_core.percentage(
    COALESCE(SUM(ld.harvested_area_ha), 0)::numeric,
    COALESCE(SUM(ld.hectares), 0)::numeric
  ) AS harvest_progress_pct,

  COALESCE(
    SUM(ld.direct_cost_per_ha_usd * ld.seeded_area_ha) / NULLIF(SUM(ld.seeded_area_ha), 0),
    0
  )::numeric AS executed_costs_usd,
  (
    COALESCE(p.planned_cost, 0)::numeric *
    v4_core.safe_div(
      COALESCE(fh.total_hectares, 0)::numeric,
      COALESCE(ph.total_hectares, 0)::numeric
    )
  )::numeric AS budget_cost_usd,
  v4_core.percentage(
    COALESCE(
      SUM(ld.direct_cost_per_ha_usd * ld.seeded_area_ha) / NULLIF(SUM(ld.seeded_area_ha), 0),
      0
    )::numeric,
    (
      COALESCE(p.planned_cost, 0)::numeric *
      v4_core.safe_div(
        COALESCE(fh.total_hectares, 0)::numeric,
        COALESCE(ph.total_hectares, 0)::numeric
      )
    )::numeric
  ) AS costs_progress_pct,

  COALESCE(it.income_net_total_usd, 0)::numeric AS operating_result_income_usd,
  (
    COALESCE(it.income_net_total_usd, 0) -
    COALESCE(dc.direct_costs_total_usd, 0) -
    COALESCE(rt.rent_total_usd, 0) -
    COALESCE(ac.admin_total_usd, 0)
  )::numeric AS operating_result_usd,
  (
    COALESCE(dc.direct_costs_total_usd, 0) +
    COALESCE(ac.admin_total_usd, 0) +
    COALESCE(rfs.rent_fixed_total_usd, 0)
  )::numeric AS operating_result_total_costs_usd,
  v4_ssot.renta_pct(
    (
      COALESCE(it.income_net_total_usd, 0) -
      COALESCE(dc.direct_costs_total_usd, 0) -
      COALESCE(rt.rent_total_usd, 0) -
      COALESCE(ac.admin_total_usd, 0)
    )::numeric,
    (
      COALESCE(dc.direct_costs_total_usd, 0) +
      COALESCE(ac.admin_total_usd, 0) +
      COALESCE(rfs.rent_fixed_total_usd, 0)
    )::numeric
  ) AS operating_result_pct,

  COALESCE(ph.total_hectares, 0)::numeric AS project_total_hectares

FROM public.projects p
LEFT JOIN lot_data ld ON ld.project_id = p.id
LEFT JOIN field_hectares fh ON fh.project_id = p.id AND fh.field_id = ld.field_id
LEFT JOIN project_hectares ph ON ph.project_id = p.id
LEFT JOIN rent_fixed_ssot rfs ON rfs.project_id = p.id AND rfs.field_id = ld.field_id
LEFT JOIN direct_costs dc ON dc.project_id = p.id AND dc.field_id = ld.field_id
LEFT JOIN admin_costs ac ON ac.project_id = p.id AND ac.field_id = ld.field_id
LEFT JOIN rent_totals rt ON rt.project_id = p.id AND rt.field_id = ld.field_id
LEFT JOIN income_totals it ON it.project_id = p.id AND it.field_id = ld.field_id
WHERE p.deleted_at IS NULL
GROUP BY
  p.customer_id,
  p.id,
  p.campaign_id,
  p.admin_cost,
  p.planned_cost,
  ph.total_hectares,
  fh.total_hectares,
  rfs.rent_fixed_total_usd,
  dc.direct_costs_total_usd,
  ac.admin_total_usd,
  rt.rent_total_usd,
  it.income_net_total_usd,
  ld.field_id;

-- Vista de balance de gestión por campo
CREATE OR REPLACE VIEW v4_report.dashboard_management_balance_field AS
WITH
lots_base AS (
  SELECT
    p.id AS project_id,
    p.customer_id,
    p.campaign_id,
    f.id AS field_id,
    l.id AS lot_id,
    l.hectares
  FROM public.projects p
  JOIN public.fields f ON f.project_id = p.id AND f.deleted_at IS NULL
  JOIN public.lots l ON l.field_id = f.id AND l.deleted_at IS NULL
  WHERE p.deleted_at IS NULL
),
income_totals AS (
  SELECT
    project_id,
    field_id,
    SUM(v4_ssot.income_net_total_for_lot(lot_id))::numeric AS income_usd
  FROM lots_base
  GROUP BY project_id, field_id
),
direct_costs AS (
  SELECT
    project_id,
    field_id,
    SUM(v4_ssot.direct_cost_for_lot(lot_id))::numeric AS direct_costs_usd
  FROM lots_base
  GROUP BY project_id, field_id
),
rent_totals AS (
  SELECT
    project_id,
    field_id,
    SUM(v4_ssot.rent_per_ha_for_lot(lot_id) * hectares)::numeric AS rent_total_usd
  FROM lots_base
  GROUP BY project_id, field_id
),
admin_totals AS (
  SELECT
    project_id,
    field_id,
    SUM(v4_ssot.admin_cost_per_ha_for_lot(lot_id) * hectares)::numeric AS admin_total_usd
  FROM lots_base
  GROUP BY project_id, field_id
),
supply_costs AS (
  SELECT
    project_id,
    field_id,
    SUM(semillas_usd)::numeric AS semillas_usd,
    SUM(herbicidas_usd + insecticidas_usd + fungicidas_usd + coadyuvantes_usd)::numeric AS agroquimicos_usd,
    SUM(fertilizantes_usd)::numeric AS fertilizantes_usd
  FROM v4_calc.field_crop_supply_costs_by_lot
  GROUP BY project_id, field_id
),
labor_costs AS (
  SELECT
    project_id,
    field_id,
    SUM(total_labores_usd)::numeric AS labor_total_usd
  FROM v4_calc.field_crop_labor_costs_by_lot
  GROUP BY project_id, field_id
)
SELECT
  lb.project_id,
  lb.customer_id,
  lb.campaign_id,
  lb.field_id,
  COALESCE(it.income_usd, 0) AS income_usd,
  (
    COALESCE(it.income_usd, 0) -
    COALESCE(dc.direct_costs_usd, 0) -
    COALESCE(rt.rent_total_usd, 0) -
    COALESCE(ad.admin_total_usd, 0)
  ) AS operating_result_usd,
  v4_ssot.renta_pct(
    (
      COALESCE(it.income_usd, 0) -
      COALESCE(dc.direct_costs_usd, 0) -
      COALESCE(rt.rent_total_usd, 0) -
      COALESCE(ad.admin_total_usd, 0)
    ),
    (
      COALESCE(dc.direct_costs_usd, 0) +
      COALESCE(rt.rent_total_usd, 0) +
      COALESCE(ad.admin_total_usd, 0)
    )
  ) AS operating_result_pct,
  COALESCE(dc.direct_costs_usd, 0) AS costos_directos_ejecutados_usd,
  COALESCE(dc.direct_costs_usd, 0) AS costos_directos_invertidos_usd,
  0::numeric AS costos_directos_stock_usd,
  COALESCE(sc.semillas_usd, 0) AS semillas_ejecutados_usd,
  COALESCE(sc.semillas_usd, 0) AS semillas_invertidos_usd,
  0::numeric AS semillas_stock_usd,
  COALESCE(sc.agroquimicos_usd, 0) AS agroquimicos_ejecutados_usd,
  COALESCE(sc.agroquimicos_usd, 0) AS agroquimicos_invertidos_usd,
  0::numeric AS agroquimicos_stock_usd,
  COALESCE(sc.fertilizantes_usd, 0) AS fertilizantes_ejecutados_usd,
  COALESCE(sc.fertilizantes_usd, 0) AS fertilizantes_invertidos_usd,
  0::numeric AS fertilizantes_stock_usd,
  COALESCE(lc.labor_total_usd, 0) AS labores_ejecutados_usd,
  COALESCE(lc.labor_total_usd, 0) AS labores_invertidos_usd,
  COALESCE(rt.rent_total_usd, 0) AS arriendo_ejecutados_usd,
  COALESCE(rt.rent_total_usd, 0) AS arriendo_invertidos_usd,
  COALESCE(ad.admin_total_usd, 0) AS estructura_ejecutados_usd,
  COALESCE(ad.admin_total_usd, 0) AS estructura_invertidos_usd,
  COALESCE(sc.semillas_usd, 0) AS semilla_cost,
  COALESCE(sc.agroquimicos_usd, 0) AS insumos_cost,
  COALESCE(lc.labor_total_usd, 0) AS labores_cost,
  COALESCE(sc.fertilizantes_usd, 0) AS fertilizantes_cost
FROM (SELECT DISTINCT project_id, customer_id, campaign_id, field_id FROM lots_base) lb
LEFT JOIN income_totals it ON it.project_id = lb.project_id AND it.field_id = lb.field_id
LEFT JOIN direct_costs dc ON dc.project_id = lb.project_id AND dc.field_id = lb.field_id
LEFT JOIN rent_totals rt ON rt.project_id = lb.project_id AND rt.field_id = lb.field_id
LEFT JOIN admin_totals ad ON ad.project_id = lb.project_id AND ad.field_id = lb.field_id
LEFT JOIN supply_costs sc ON sc.project_id = lb.project_id AND sc.field_id = lb.field_id
LEFT JOIN labor_costs lc ON lc.project_id = lb.project_id AND lc.field_id = lb.field_id;

-- Vista de incidencia de cultivos por campo
CREATE OR REPLACE VIEW v4_report.dashboard_crop_incidence_field AS
WITH
field_totals AS (
  SELECT
    project_id,
    field_id,
    SUM(surface_ha)::numeric AS total_hectares
  FROM v4_calc.field_crop_aggregated
  GROUP BY project_id, field_id
),
by_crop AS (
  SELECT
    project_id,
    field_id,
    crop_id AS current_crop_id,
    SUM(surface_ha)::numeric AS crop_hectares
  FROM v4_calc.field_crop_aggregated
  GROUP BY project_id, field_id, crop_id
)
SELECT
  p.id AS project_id,
  p.customer_id,
  p.campaign_id,
  bc.field_id,
  bc.current_crop_id,
  c.name AS crop_name,
  bc.crop_hectares,
  v4_core.percentage(bc.crop_hectares, COALESCE(ft.total_hectares, 0)::numeric) AS crop_incidence_pct,
  v4_ssot.cost_per_ha_for_crop_ssot(p.id, bc.current_crop_id)::numeric AS cost_per_ha_usd
FROM by_crop bc
JOIN public.projects p ON p.id = bc.project_id AND p.deleted_at IS NULL
LEFT JOIN field_totals ft ON ft.project_id = bc.project_id AND ft.field_id = bc.field_id
LEFT JOIN public.crops c ON c.id = bc.current_crop_id AND c.deleted_at IS NULL
ORDER BY bc.project_id, bc.field_id, c.name;

-- Vista de indicadores operativos por campo
CREATE OR REPLACE VIEW v4_report.dashboard_operational_indicators_field AS
SELECT
  p.id AS project_id,
  p.customer_id,
  p.campaign_id,
  f.id AS field_id,
  MIN(w.date) AS start_date,
  MAX(w.date) AS end_date,
  v4_core.calculate_campaign_closing_date(MAX(w.date)) AS campaign_closing_date,
  MIN(w.number) AS first_workorder_id,
  MAX(w.number) AS last_workorder_id,
  v4_ssot.last_stock_count_date_for_project(p.id) AS last_stock_count_date
FROM public.projects p
JOIN public.fields f ON f.project_id = p.id AND f.deleted_at IS NULL
LEFT JOIN public.workorders w ON w.field_id = f.id AND w.deleted_at IS NULL
WHERE p.deleted_at IS NULL
GROUP BY p.id, p.customer_id, p.campaign_id, f.id;

COMMIT;
