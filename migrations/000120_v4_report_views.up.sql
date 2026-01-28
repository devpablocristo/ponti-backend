-- ========================================
-- MIGRATION 000120 V4 REPORT VIEWS.UP (UP)
-- ========================================
-- Nota: SQL en inglés, comentarios en español

BEGIN;

CREATE OR REPLACE VIEW v4_report.lot_metrics AS
WITH 
project_totals AS (
  SELECT project_id, SUM(hectares)::numeric AS total_hectares
  FROM v4_calc.lot_base_costs GROUP BY project_id
),
field_totals AS (
  SELECT field_id, SUM(hectares)::numeric AS total_hectares
  FROM v4_calc.lot_base_costs GROUP BY field_id
)
SELECT
  c.project_id,
  c.field_id,
  c.lot_id,
  c.lot_name,
  c.hectares,
  c.sowed_area_ha,
  c.harvested_area_ha,
  c.yield_tn_per_ha,
  c.tons,
  c.sowing_date,
  c.labor_cost_usd,
  c.supplies_cost_usd AS supply_cost_usd,
  c.direct_cost_usd,
  c.income_net_total_usd,
  c.income_net_per_ha_usd,
  c.rent_per_ha_usd,
  c.admin_cost_per_ha_usd,
  c.direct_cost_per_ha_usd,
  
  (c.direct_cost_per_ha_usd + c.rent_per_ha_usd + c.admin_cost_per_ha_usd)::numeric AS active_total_per_ha_usd,
  (c.income_net_per_ha_usd - (c.direct_cost_per_ha_usd + c.rent_per_ha_usd + c.admin_cost_per_ha_usd))::numeric AS operating_result_per_ha_usd,
  (c.rent_per_ha_usd * c.hectares)::numeric AS rent_total_usd,
  (c.admin_cost_per_ha_usd * c.hectares)::numeric AS admin_total_usd,
  ((c.direct_cost_per_ha_usd + c.rent_per_ha_usd + c.admin_cost_per_ha_usd) * c.hectares)::numeric AS active_total_usd,
  ((c.income_net_per_ha_usd - (c.direct_cost_per_ha_usd + c.rent_per_ha_usd + c.admin_cost_per_ha_usd)) * c.hectares)::numeric AS operating_result_total_usd,
  c.direct_cost_usd AS direct_cost_total_usd,
  COALESCE(pt.total_hectares, 0) AS project_total_hectares,
  COALESCE(ft.total_hectares, 0) AS field_total_hectares
FROM v4_calc.lot_base_costs c
LEFT JOIN project_totals pt ON pt.project_id = c.project_id
LEFT JOIN field_totals ft ON ft.field_id = c.field_id;

CREATE OR REPLACE VIEW v4_report.lot_list AS
SELECT
  f.project_id,
  p.name AS project_name,
  f.id AS field_id,
  f.name AS field_name,
  l.id AS id,
  l.name AS lot_name,
  l.variety,
  l.season,
  l.previous_crop_id,
  prev_crop.name AS previous_crop,
  l.current_crop_id,
  curr_crop.name AS current_crop,
  l.hectares,
  l.updated_at,
  COALESCE(lm.sowed_area_ha, 0)::numeric AS sowed_area_ha,
  COALESCE(lm.harvested_area_ha, 0)::numeric AS harvested_area_ha,
  COALESCE(lm.yield_tn_per_ha, 0) AS yield_tn_per_ha,
  COALESCE(lm.direct_cost_per_ha_usd, 0)::numeric AS cost_usd_per_ha,
  COALESCE(lm.income_net_per_ha_usd, 0)::numeric AS income_net_per_ha_usd,
  COALESCE(lm.rent_per_ha_usd, 0)::numeric AS rent_per_ha_usd,
  COALESCE(lm.admin_cost_per_ha_usd, 0)::numeric AS admin_cost_per_ha_usd,
  COALESCE(lm.active_total_per_ha_usd, 0)::numeric AS active_total_per_ha_usd,
  COALESCE(lm.operating_result_per_ha_usd, 0)::numeric AS operating_result_per_ha_usd,
  COALESCE(lm.income_net_total_usd, 0)::numeric AS income_net_total_usd,
  COALESCE(lm.direct_cost_total_usd, 0)::numeric AS direct_cost_total_usd,
  COALESCE(lm.rent_total_usd, 0)::numeric AS rent_total_usd,
  COALESCE(lm.admin_total_usd, 0)::numeric AS admin_total_usd,
  COALESCE(lm.active_total_usd, 0)::numeric AS active_total_usd,
  COALESCE(lm.operating_result_total_usd, 0)::numeric AS operating_result_total_usd,
  l.sowing_date AS lot_sowing_date,
  NULL::date AS lot_harvest_date,
  l.tons,
  (
    SELECT MIN(w.date)
    FROM public.workorders w
    JOIN public.labors lb ON lb.id = w.labor_id AND lb.deleted_at IS NULL
    WHERE w.lot_id = l.id AND w.deleted_at IS NULL
  ) AS raw_sowing_date
FROM public.lots l
JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
JOIN public.projects p ON p.id = f.project_id AND p.deleted_at IS NULL
LEFT JOIN public.crops prev_crop ON prev_crop.id = l.previous_crop_id AND prev_crop.deleted_at IS NULL
LEFT JOIN public.crops curr_crop ON curr_crop.id = l.current_crop_id AND curr_crop.deleted_at IS NULL
LEFT JOIN v4_report.lot_metrics lm ON lm.lot_id = l.id
WHERE l.deleted_at IS NULL;

CREATE OR REPLACE VIEW v4_report.labor_list AS
SELECT
  w.id AS workorder_id,
  w.number AS workorder_number,
  w.date,
  w.project_id,
  p.name AS project_name,
  w.field_id,
  f.name AS field_name,
  w.lot_id,
  l.name AS lot_name,
  w.crop_id,
  c.name AS crop_name,
  w.labor_id,
  lb.name AS labor_name,
  lb.category_id AS labor_category_id,
  cat.name AS labor_category_name,
  w.contractor,
  lb.contractor_name,
  w.effective_area AS surface_ha,
  lb.price AS cost_per_ha,
  (lb.price * w.effective_area)::numeric AS total_labor_cost,
  v4_core.dollar_average_for_month(w.project_id, w.date) AS dollar_average_month,
  
  
  lb.price::numeric AS usd_cost_ha,
  
  (lb.price * w.effective_area)::numeric AS usd_net_total,
  w.investor_id,
  i.name AS investor_name
FROM public.workorders w
JOIN public.projects p ON p.id = w.project_id AND p.deleted_at IS NULL
JOIN public.fields f ON f.id = w.field_id AND f.deleted_at IS NULL
LEFT JOIN public.lots l ON l.id = w.lot_id AND l.deleted_at IS NULL
LEFT JOIN public.crops c ON c.id = w.crop_id AND c.deleted_at IS NULL
JOIN public.labors lb ON lb.id = w.labor_id AND lb.deleted_at IS NULL
LEFT JOIN public.categories cat ON cat.id = lb.category_id AND cat.deleted_at IS NULL
LEFT JOIN public.investors i ON i.id = w.investor_id AND i.deleted_at IS NULL
WHERE w.deleted_at IS NULL
  AND w.effective_area IS NOT NULL
  AND w.effective_area > 0
  AND lb.price IS NOT NULL;

CREATE OR REPLACE VIEW v4_report.labor_metrics AS
WITH wo AS (
  SELECT
    w.id AS workorder_id,
    w.project_id,
    w.field_id,
    w.date,
    w.effective_area::numeric AS effective_area,
    lb.price::numeric AS labor_price_per_ha
  FROM public.workorders w
  JOIN public.labors lb ON lb.id = w.labor_id
  WHERE w.deleted_at IS NULL
    AND lb.deleted_at IS NULL
    AND w.effective_area IS NOT NULL
    AND w.effective_area > 0
    AND lb.price IS NOT NULL
),
agg AS (
  SELECT
    project_id,
    field_id,
    COUNT(DISTINCT workorder_id) AS total_workorders,
    SUM(effective_area) AS surface_ha,
    SUM(v4_core.labor_cost(labor_price_per_ha, effective_area)) AS total_labor_cost,
    MIN(date) AS first_workorder_date,
    MAX(date) AS last_workorder_date
  FROM wo
  GROUP BY project_id, field_id
)
SELECT
  a.project_id,
  a.field_id,
  a.surface_ha,
  a.total_labor_cost,
  v4_core.cost_per_ha(a.total_labor_cost, a.surface_ha) AS avg_labor_cost_per_ha,
  a.total_workorders,
  a.first_workorder_date,
  a.last_workorder_date
FROM agg a;

CREATE OR REPLACE VIEW v4_report.field_crop_cultivos AS
WITH lot_base AS (
  SELECT
    f.project_id,
    f.id AS field_id,
    f.name AS field_name,
    l.current_crop_id AS crop_id,
    c.name AS crop_name,
    l.id AS lot_id,
    l.hectares,
    l.tons,
    v4_ssot.seeded_area_for_lot(l.id)::numeric AS sowed_area_ha,
    v4_ssot.harvested_area_for_lot(l.id)::numeric AS harvested_area_ha
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  JOIN public.crops c ON c.id = l.current_crop_id AND c.deleted_at IS NULL
  WHERE l.deleted_at IS NULL AND l.current_crop_id IS NOT NULL
)
SELECT
  project_id,
  field_id,
  field_name,
  crop_id AS current_crop_id,
  crop_name,
  SUM(hectares)::numeric AS superficie_total,
  SUM(sowed_area_ha)::numeric AS superficie_sembrada_ha,
  SUM(harvested_area_ha)::numeric AS area_cosechada_ha,
  SUM(tons)::numeric AS produccion_tn,
  v4_ssot.board_price_for_lot(MIN(lot_id)) AS precio_bruto_usd_tn,
  v4_ssot.freight_cost_for_lot(MIN(lot_id)) AS gasto_flete_usd_tn,
  v4_ssot.commercial_cost_for_lot(MIN(lot_id)) AS gasto_comercial_usd_tn,
  v4_ssot.net_price_usd_for_lot(MIN(lot_id)) AS precio_neto_usd_tn,
  v4_core.safe_div(SUM(tons), SUM(hectares)::numeric) AS rendimiento_tn_ha,
  (v4_core.safe_div(SUM(tons), SUM(hectares)::numeric) * 
   v4_ssot.net_price_usd_for_lot(MIN(lot_id))) AS ingreso_neto_por_ha
FROM lot_base
GROUP BY project_id, field_id, field_name, crop_id, crop_name;

CREATE OR REPLACE VIEW v4_report.field_crop_economicos AS
SELECT
  project_id,
  field_id,
  crop_id AS current_crop_id,
  (labor_costs_usd + supply_costs_usd) AS gastos_directos_usd,
  v4_core.safe_div(
    (labor_costs_usd + supply_costs_usd),
    surface_ha
  ) AS gastos_directos_usd_ha,
  (
    (production_tn * v4_ssot.net_price_usd_for_lot(sample_lot_id)) -
    (labor_costs_usd + supply_costs_usd)
  ) AS margen_bruto_usd,
  (
    (v4_core.safe_div(production_tn, surface_ha) * v4_ssot.net_price_usd_for_lot(sample_lot_id)) -
    v4_core.safe_div((labor_costs_usd + supply_costs_usd), surface_ha)
  ) AS margen_bruto_usd_ha,
  
  rent_fixed_usd AS arriendo_usd,
  v4_core.safe_div(rent_fixed_usd, surface_ha) AS arriendo_usd_ha,
  administration_usd AS adm_estructura_usd,
  v4_core.safe_div(administration_usd, surface_ha) AS adm_estructura_usd_ha,
  
  (
    (production_tn * v4_ssot.net_price_usd_for_lot(sample_lot_id)) -
    (labor_costs_usd + supply_costs_usd) - rent_total_usd - administration_usd
  ) AS resultado_operativo_usd,
  (
    (v4_core.safe_div(production_tn, surface_ha) * v4_ssot.net_price_usd_for_lot(sample_lot_id)) -
    v4_core.safe_div((labor_costs_usd + supply_costs_usd), surface_ha) -
    v4_core.safe_div(rent_total_usd, surface_ha) -
    v4_core.safe_div(administration_usd, surface_ha)
  ) AS resultado_operativo_usd_ha
FROM v4_calc.field_crop_aggregated;

CREATE OR REPLACE VIEW v4_report.field_crop_insumos AS
WITH supply_costs AS (
  SELECT
    project_id,
    field_id,
    crop_id,
    lot_id,
    sowed_area_ha,
    surface_ha,
    semillas_usd,
    curasemillas_usd,
    herbicidas_usd,
    insecticidas_usd,
    fungicidas_usd,
    coadyuvantes_usd,
    fertilizantes_usd,
    otros_insumos_usd,
    total_insumos_usd
  FROM v4_calc.field_crop_supply_costs_by_lot
)
SELECT
  project_id,
  field_id,
  crop_id AS current_crop_id,
  COALESCE(SUM(semillas_usd), 0) AS semillas_total_usd,
  COALESCE(SUM(curasemillas_usd), 0) AS curasemillas_total_usd,
  COALESCE(SUM(herbicidas_usd), 0) AS herbicidas_total_usd,
  COALESCE(SUM(insecticidas_usd), 0) AS insecticidas_total_usd,
  COALESCE(SUM(fungicidas_usd), 0) AS fungicidas_total_usd,
  COALESCE(SUM(coadyuvantes_usd), 0) AS coadyuvantes_total_usd,
  COALESCE(SUM(fertilizantes_usd), 0) AS fertilizantes_total_usd,
  COALESCE(SUM(otros_insumos_usd), 0) AS otros_insumos_total_usd,
  v4_core.safe_div(COALESCE(SUM(semillas_usd), 0), COALESCE(SUM(surface_ha), 1)::numeric) AS semillas_usd_ha,
  v4_core.safe_div(COALESCE(SUM(curasemillas_usd), 0), COALESCE(SUM(surface_ha), 1)::numeric) AS curasemillas_usd_ha,
  v4_core.safe_div(COALESCE(SUM(herbicidas_usd), 0), COALESCE(SUM(surface_ha), 1)::numeric) AS herbicidas_usd_ha,
  v4_core.safe_div(COALESCE(SUM(insecticidas_usd), 0), COALESCE(SUM(surface_ha), 1)::numeric) AS insecticidas_usd_ha,
  v4_core.safe_div(COALESCE(SUM(fungicidas_usd), 0), COALESCE(SUM(surface_ha), 1)::numeric) AS fungicidas_usd_ha,
  v4_core.safe_div(COALESCE(SUM(coadyuvantes_usd), 0), COALESCE(SUM(surface_ha), 1)::numeric) AS coadyuvantes_usd_ha,
  v4_core.safe_div(COALESCE(SUM(fertilizantes_usd), 0), COALESCE(SUM(surface_ha), 1)::numeric) AS fertilizantes_usd_ha,
  v4_core.safe_div(COALESCE(SUM(otros_insumos_usd), 0), COALESCE(SUM(surface_ha), 1)::numeric) AS otros_insumos_usd_ha,
  COALESCE(SUM(total_insumos_usd), 0) AS total_insumos_usd,
  v4_core.safe_div(
    COALESCE(SUM(total_insumos_usd), 0),
    COALESCE(SUM(surface_ha), 1)::numeric
  ) AS total_insumos_usd_ha
FROM supply_costs
GROUP BY project_id, field_id, crop_id;

CREATE OR REPLACE VIEW v4_report.field_crop_labores AS
SELECT
  project_id,
  field_id,
  crop_id AS current_crop_id,
  COALESCE(SUM(siembra_usd), 0) AS siembra_total_usd,
  COALESCE(SUM(pulverizacion_usd), 0) AS pulverizacion_total_usd,
  COALESCE(SUM(riego_usd), 0) AS riego_total_usd,
  COALESCE(SUM(cosecha_usd), 0) AS cosecha_total_usd,
  COALESCE(SUM(otras_labores_usd), 0) AS otras_labores_total_usd,
  v4_core.safe_div(COALESCE(SUM(siembra_usd), 0), COALESCE(SUM(surface_ha), 1)::numeric) AS siembra_usd_ha,
  v4_core.safe_div(COALESCE(SUM(pulverizacion_usd), 0), COALESCE(SUM(surface_ha), 1)::numeric) AS pulverizacion_usd_ha,
  v4_core.safe_div(COALESCE(SUM(riego_usd), 0), COALESCE(SUM(surface_ha), 1)::numeric) AS riego_usd_ha,
  v4_core.safe_div(COALESCE(SUM(cosecha_usd), 0), COALESCE(SUM(surface_ha), 1)::numeric) AS cosecha_usd_ha,
  v4_core.safe_div(COALESCE(SUM(otras_labores_usd), 0), COALESCE(SUM(surface_ha), 1)::numeric) AS otras_labores_usd_ha,
  COALESCE(SUM(total_labores_usd), 0) AS total_labores_usd,
  v4_core.safe_div(
    COALESCE(SUM(total_labores_usd), 0),
    COALESCE(SUM(surface_ha), 1)::numeric
  ) AS total_labores_usd_ha
FROM v4_calc.field_crop_labor_costs_by_lot
GROUP BY project_id, field_id, crop_id;

CREATE OR REPLACE VIEW v4_report.field_crop_metrics AS
WITH
aggregated_base AS (
  SELECT
    project_id,
    field_id,
    field_name,
    current_crop_id,
    crop_name,
    superficie_total,
    superficie_sembrada_ha,
    area_cosechada_ha,
    produccion_tn,
    rendimiento_tn_ha,
    precio_bruto_usd_tn,
    gasto_flete_usd_tn,
    gasto_comercial_usd_tn,
    precio_neto_usd_tn,
    ingreso_neto_total
  FROM v4_calc.field_crop_metrics_aggregated
),
costs AS (
  SELECT
    project_id,
    field_id,
    crop_id,
    labor_costs_usd,
    supply_costs_usd,
    rent_total_usd AS arriendo_total_usd,
    administration_usd AS admin_total_usd
  FROM v4_calc.field_crop_aggregated
)
SELECT
  a.project_id,
  a.field_id,
  a.field_name,
  a.current_crop_id,
  a.crop_name,
  a.superficie_total AS superficie_ha,
  a.produccion_tn,
  a.superficie_sembrada_ha AS area_sembrada_ha,
  a.area_cosechada_ha,
  a.rendimiento_tn_ha,
  a.precio_bruto_usd_tn,
  a.gasto_flete_usd_tn,
  a.gasto_comercial_usd_tn,
  a.precio_neto_usd_tn,
  a.ingreso_neto_total AS ingreso_neto_usd,
  v4_core.safe_div(a.ingreso_neto_total, a.superficie_total) AS ingreso_neto_usd_ha,
  c.labor_costs_usd AS costos_labores_usd,
  v4_core.safe_div(c.labor_costs_usd, a.superficie_total) AS costos_labores_usd_ha,
  c.supply_costs_usd AS costos_insumos_usd,
  v4_core.safe_div(c.supply_costs_usd, a.superficie_total) AS costos_insumos_usd_ha,
  (c.labor_costs_usd + c.supply_costs_usd)::numeric AS total_costos_directos_usd,
  v4_core.safe_div(c.labor_costs_usd + c.supply_costs_usd, a.superficie_total) AS costos_directos_usd_ha,
  (a.ingreso_neto_total - c.labor_costs_usd - c.supply_costs_usd)::numeric AS margen_bruto_usd,
  v4_core.safe_div(a.ingreso_neto_total - c.labor_costs_usd - c.supply_costs_usd, a.superficie_total) AS margen_bruto_usd_ha,
  c.arriendo_total_usd AS arriendo_usd,
  v4_core.safe_div(c.arriendo_total_usd, a.superficie_total) AS arriendo_usd_ha,
  c.admin_total_usd AS administracion_usd,
  v4_core.safe_div(c.admin_total_usd, a.superficie_total) AS administracion_usd_ha,
  (a.ingreso_neto_total - c.labor_costs_usd - c.supply_costs_usd - c.arriendo_total_usd - c.admin_total_usd)::numeric AS resultado_operativo_usd,
  v4_core.safe_div(
    a.ingreso_neto_total - c.labor_costs_usd - c.supply_costs_usd - c.arriendo_total_usd - c.admin_total_usd,
    a.superficie_total
  ) AS resultado_operativo_usd_ha,
  (c.labor_costs_usd + c.supply_costs_usd + c.arriendo_total_usd + c.admin_total_usd)::numeric AS total_invertido_usd,
  v4_core.safe_div(
    c.labor_costs_usd + c.supply_costs_usd + c.arriendo_total_usd + c.admin_total_usd,
    a.superficie_total
  ) AS total_invertido_usd_ha,
  CASE WHEN (c.labor_costs_usd + c.supply_costs_usd + c.arriendo_total_usd + c.admin_total_usd) > 0
    THEN ((a.ingreso_neto_total - c.labor_costs_usd - c.supply_costs_usd - c.arriendo_total_usd - c.admin_total_usd) /
          (c.labor_costs_usd + c.supply_costs_usd + c.arriendo_total_usd + c.admin_total_usd) * 100)::double precision
    ELSE 0
  END AS renta_pct,
  CASE WHEN a.precio_neto_usd_tn > 0 AND a.superficie_total > 0
    THEN ((c.labor_costs_usd + c.supply_costs_usd + c.arriendo_total_usd + c.admin_total_usd) / a.superficie_total / a.precio_neto_usd_tn)::numeric
    ELSE 0
  END AS rinde_indiferencia_usd_tn
FROM aggregated_base a
LEFT JOIN costs c
  ON c.project_id = a.project_id
  AND c.field_id = a.field_id
  AND c.crop_id = a.current_crop_id;

CREATE OR REPLACE VIEW v4_report.field_crop_rentabilidad AS
SELECT
  project_id,
  field_id,
  crop_id AS current_crop_id,
  (direct_cost_usd + rent_fixed_usd + administration_usd) AS total_invertido_usd,
  v4_core.safe_div(
    (direct_cost_usd + rent_fixed_usd + administration_usd),
    surface_ha
  ) AS total_invertido_usd_ha,
  v4_ssot.renta_pct(
    (
      (production_tn * v4_ssot.net_price_usd_for_lot(sample_lot_id)) -
      direct_cost_usd - rent_fixed_usd - administration_usd
    ),
    (direct_cost_usd + rent_fixed_usd + administration_usd)
  ) AS renta_pct,
  v4_core.safe_div(
    v4_core.safe_div((direct_cost_usd + rent_fixed_usd + administration_usd), surface_ha),
    v4_ssot.net_price_usd_for_lot(sample_lot_id)
  ) AS rinde_indiferencia_total_usd_tn
FROM v4_calc.field_crop_aggregated;

CREATE OR REPLACE VIEW v4_report.summary_results AS
WITH 
by_crop AS (
  SELECT
    project_id,
    current_crop_id,
    crop_name,
    SUM(superficie_ha)::numeric AS surface_ha,
    SUM(ingreso_neto_usd)::numeric AS net_income_usd,
    SUM(total_costos_directos_usd)::numeric AS direct_costs_usd,
    SUM(arriendo_usd)::numeric AS rent_usd,
    SUM(administracion_usd)::numeric AS structure_usd,
    SUM(total_invertido_usd)::numeric AS total_invested_usd,
    SUM(resultado_operativo_usd)::numeric AS operating_result_usd
  FROM v4_report.field_crop_metrics
  WHERE current_crop_id IS NOT NULL
  GROUP BY project_id, current_crop_id, crop_name
),
project_totals AS (
  SELECT
    project_id,
    SUM(surface_ha)::numeric AS total_surface_ha,
    SUM(net_income_usd)::numeric AS total_net_income_usd,
    SUM(direct_costs_usd)::numeric AS total_direct_costs_usd,
    SUM(rent_usd)::numeric AS total_rent_usd,
    SUM(structure_usd)::numeric AS total_structure_usd,
    SUM(total_invested_usd)::numeric AS total_invested_project_usd,
    SUM(operating_result_usd)::numeric AS total_operating_result_usd
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
  bc.total_invested_usd,
  bc.operating_result_usd,
  CASE WHEN bc.total_invested_usd > 0 
    THEN (bc.operating_result_usd / bc.total_invested_usd * 100)::numeric
    ELSE 0::numeric 
  END AS crop_return_pct,
  pt.total_surface_ha,
  pt.total_net_income_usd,
  pt.total_direct_costs_usd,
  pt.total_rent_usd,
  pt.total_structure_usd,
  pt.total_invested_project_usd,
  pt.total_operating_result_usd,
  CASE WHEN pt.total_invested_project_usd > 0 
    THEN (pt.total_operating_result_usd / pt.total_invested_project_usd * 100)::numeric
    ELSE 0::numeric 
  END AS project_return_pct
FROM by_crop bc
JOIN project_totals pt ON pt.project_id = bc.project_id;

CREATE OR REPLACE VIEW v4_report.dashboard_management_balance AS
SELECT
  p.id AS project_id,
  COALESCE(SUM(v4_ssot.income_net_total_for_lot(l.id)), 0) AS income_usd,
  v4_ssot.operating_result_total_for_project(p.id) AS operating_result_usd,
  v4_ssot.renta_pct(
    v4_ssot.operating_result_total_for_project(p.id),
    v4_ssot.total_costs_for_project(p.id)
  ) AS operating_result_pct,
  v4_ssot.direct_costs_total_for_project(p.id) AS costos_directos_ejecutados_usd,
  (v4_ssot.supply_movements_invested_total_for_project(p.id) +
   COALESCE(SUM(v4_ssot.labor_cost_pre_harvest_for_lot(l.id)), 0)) AS costos_directos_invertidos_usd,
  ((v4_ssot.supply_movements_invested_total_for_project(p.id) +
    COALESCE(SUM(v4_ssot.labor_cost_pre_harvest_for_lot(l.id)), 0)) -
   v4_ssot.direct_costs_total_for_project(p.id)) AS costos_directos_stock_usd,
  COALESCE(sc.semillas_ejecutados_usd, 0) AS semillas_ejecutados_usd,
  v4_ssot.seeds_invested_for_project_mb(p.id) AS semillas_invertidos_usd,
  (v4_ssot.seeds_invested_for_project_mb(p.id) -
   COALESCE(sc.semillas_ejecutados_usd, 0)) AS semillas_stock_usd,
  COALESCE(sc.agroquimicos_ejecutados_usd, 0) AS agroquimicos_ejecutados_usd,
  v4_ssot.agrochemicals_invested_for_project_mb(p.id) AS agroquimicos_invertidos_usd,
  (v4_ssot.agrochemicals_invested_for_project_mb(p.id) -
   COALESCE(sc.agroquimicos_ejecutados_usd, 0)) AS agroquimicos_stock_usd,
  COALESCE(sc.fertilizantes_ejecutados_usd, 0) AS fertilizantes_ejecutados_usd,
  COALESCE(fi.fertilizantes_invertidos_usd, 0) AS fertilizantes_invertidos_usd,
  (COALESCE(fi.fertilizantes_invertidos_usd, 0) -
   COALESCE(sc.fertilizantes_ejecutados_usd, 0)) AS fertilizantes_stock_usd,
  COALESCE(SUM(v4_ssot.labor_cost_for_lot(l.id)), 0) AS labores_ejecutados_usd,
  COALESCE(SUM(v4_ssot.labor_cost_pre_harvest_for_lot(l.id)), 0) AS labores_invertidos_usd,
  
  v4_ssot.lease_invested_for_project(p.id) AS arriendo_ejecutados_usd,
  v4_ssot.lease_executed_for_project(p.id) AS arriendo_invertidos_usd,
  v4_ssot.admin_cost_total_for_project(p.id) AS estructura_ejecutados_usd,
  v4_ssot.admin_cost_total_for_project(p.id) AS estructura_invertidos_usd,
  COALESCE(sc.semillas_ejecutados_usd, 0) AS semilla_cost,
  COALESCE(sc.agroquimicos_ejecutados_usd, 0) AS insumos_cost,
  COALESCE(SUM(v4_ssot.labor_cost_for_lot(l.id)), 0) AS labores_cost,
  COALESCE(sc.fertilizantes_ejecutados_usd, 0) AS fertilizantes_cost
FROM public.projects p
LEFT JOIN public.fields f ON f.project_id = p.id AND f.deleted_at IS NULL
LEFT JOIN public.lots l ON l.field_id = f.id AND l.deleted_at IS NULL
LEFT JOIN v4_calc.dashboard_supply_costs_by_project sc ON sc.project_id = p.id
LEFT JOIN v4_calc.dashboard_fertilizers_invested_by_project fi ON fi.project_id = p.id
WHERE p.deleted_at IS NULL
GROUP BY p.id, sc.semillas_ejecutados_usd, sc.agroquimicos_ejecutados_usd, sc.fertilizantes_ejecutados_usd, fi.fertilizantes_invertidos_usd;

CREATE OR REPLACE VIEW v4_report.dashboard_metrics AS
WITH lot_data AS (
  SELECT
    lm.project_id,
    lm.lot_id,
    lm.hectares,
    lm.sowed_area_ha,
    lm.harvested_area_ha,
    lm.direct_cost_per_ha_usd
  FROM v4_report.lot_metrics lm
),
project_hectares AS (
  SELECT
    project_id,
    SUM(hectares) AS total_hectares
  FROM v4_report.lot_metrics
  GROUP BY project_id
),
rent_fixed_ssot AS (
  SELECT
    f.project_id,
    SUM(v4_ssot.rent_fixed_only_for_lot(l.id) * l.hectares)::double precision AS rent_fixed_total_usd
  FROM public.fields f
  JOIN public.lots l ON l.field_id = f.id AND l.deleted_at IS NULL
  WHERE f.deleted_at IS NULL
  GROUP BY f.project_id
)
SELECT
  p.customer_id,
  p.id AS project_id,
  p.campaign_id,
  
  
  COALESCE(SUM(ld.sowed_area_ha), 0)::double precision AS sowing_hectares,
  COALESCE(SUM(ld.hectares), 0)::double precision AS sowing_total_hectares,
  v4_core.percentage(
    COALESCE(SUM(ld.sowed_area_ha), 0)::numeric,
    COALESCE(SUM(ld.hectares), 0)::numeric
  ) AS sowing_progress_pct,
  
  
  COALESCE(SUM(ld.harvested_area_ha), 0)::double precision AS harvest_hectares,
  COALESCE(SUM(ld.hectares), 0)::double precision AS harvest_total_hectares,
  v4_core.percentage(
    COALESCE(SUM(ld.harvested_area_ha), 0)::numeric,
    COALESCE(SUM(ld.hectares), 0)::numeric
  ) AS harvest_progress_pct,
  
  
  COALESCE(
    SUM(ld.direct_cost_per_ha_usd * ld.sowed_area_ha) / NULLIF(SUM(ld.sowed_area_ha), 0),
    0
  )::double precision AS executed_costs_usd,
  COALESCE(p.planned_cost, 0)::double precision AS budget_cost_usd,
  v4_core.percentage(
    COALESCE(
      SUM(ld.direct_cost_per_ha_usd * ld.sowed_area_ha) / NULLIF(SUM(ld.sowed_area_ha), 0),
      0
    )::numeric,
    COALESCE(p.planned_cost, 0)::numeric
  ) AS costs_progress_pct,
  
  
  COALESCE(SUM(v4_ssot.income_net_total_for_lot(ld.lot_id)), 0) AS operating_result_income_usd,
  v4_ssot.operating_result_total_for_project(p.id) AS operating_result_usd,
  (
    COALESCE(v4_ssot.direct_costs_total_for_project(p.id), 0) +
    COALESCE(p.admin_cost * ph.total_hectares, 0) +
    COALESCE(rfs.rent_fixed_total_usd, 0)
  )::double precision AS operating_result_total_costs_usd,
  v4_ssot.renta_pct(
    v4_ssot.operating_result_total_for_project(p.id),
    (
      COALESCE(v4_ssot.direct_costs_total_for_project(p.id), 0) +
      COALESCE(p.admin_cost * ph.total_hectares, 0) +
      COALESCE(rfs.rent_fixed_total_usd, 0)
    )::double precision
  ) AS operating_result_pct,
  
  COALESCE(ph.total_hectares, 0)::numeric AS project_total_hectares
  
FROM public.projects p
LEFT JOIN lot_data ld ON ld.project_id = p.id
LEFT JOIN project_hectares ph ON ph.project_id = p.id
LEFT JOIN rent_fixed_ssot rfs ON rfs.project_id = p.id
WHERE p.deleted_at IS NULL
GROUP BY p.customer_id, p.id, p.campaign_id, p.admin_cost, p.planned_cost, ph.total_hectares, rfs.rent_fixed_total_usd;

CREATE OR REPLACE VIEW v4_report.dashboard_crop_incidence AS
SELECT
  p.id AS project_id,
  ci.current_crop_id,
  ci.crop_name,
  ci.crop_hectares,
  ci.crop_incidence_pct,
  v4_ssot.cost_per_ha_for_crop_ssot(p.id, ci.current_crop_id)::numeric AS cost_per_ha_usd
FROM public.projects p
CROSS JOIN LATERAL v4_ssot.crop_incidence_for_project(p.id) ci
WHERE p.deleted_at IS NULL
ORDER BY p.id, ci.crop_name;

CREATE OR REPLACE VIEW v4_report.dashboard_operational_indicators AS
SELECT
  p.id AS project_id,
  v4_ssot.first_workorder_date_for_project(p.id) AS start_date,
  v4_ssot.last_workorder_date_for_project(p.id) AS end_date,
  v4_core.calculate_campaign_closing_date(
    v4_ssot.last_workorder_date_for_project(p.id)
  ) AS campaign_closing_date,
  v4_ssot.first_workorder_number_for_project(p.id) AS first_workorder_id,
  v4_ssot.last_workorder_number_for_project(p.id) AS last_workorder_id,
  v4_ssot.last_stock_count_date_for_project(p.id) AS last_stock_count_date
FROM public.projects p
WHERE p.deleted_at IS NULL;

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

CREATE OR REPLACE VIEW v4_report.investor_contribution_categories AS
SELECT *
FROM v4_calc.investor_contribution_categories;

CREATE OR REPLACE VIEW v4_report.investor_distributions AS
SELECT
  irc.project_id,
  irc.investor_id,
  irc.investor_name,
  irc.share_pct_agreed,
  
  (irc.project_total_contributions_usd * irc.share_pct_agreed / 100)::numeric AS agreed_contribution_usd,
  irc.total_real_contribution_usd AS real_contribution_usd,
  (
    irc.total_real_contribution_usd -
    (irc.project_total_contributions_usd * irc.share_pct_agreed / 100)
  )::numeric AS adjustment_usd,
  
  irc.agrochemicals_real_usd,
  irc.fertilizers_real_usd,
  irc.seeds_real_usd,
  irc.general_labors_real_usd,
  irc.sowing_real_usd,
  irc.irrigation_real_usd,
  irc.rent_real_usd,
  irc.administration_real_usd,
  irc.project_total_contributions_usd
FROM v4_calc.investor_real_contributions irc
ORDER BY irc.project_id, irc.investor_id;

CREATE OR REPLACE VIEW v4_report.investor_contribution_data AS
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
project_surface_data AS (
  SELECT
    project_id,
    MAX(surface_total_ha)::numeric AS surface_total_ha
  FROM v4_report.investor_project_base
  GROUP BY project_id
),
investor_harvest_real AS (
  SELECT
    w.project_id,
    w.investor_id,
    COALESCE(SUM(lab.price * w.effective_area), 0)::numeric AS harvest_real_usd
  FROM public.workorders w
  JOIN public.labors lab ON w.labor_id = lab.id AND lab.deleted_at IS NULL
  JOIN public.categories cat ON cat.id = lab.category_id
  WHERE w.deleted_at IS NULL
    AND cat.type_id = 4
    AND cat.name = 'Cosecha'
  GROUP BY w.project_id, w.investor_id
),
harvest_totals AS (
  SELECT
    psd.project_id,
    COALESCE(SUM(hr.harvest_real_usd), 0)::numeric AS total_harvest_usd,
    CASE
      WHEN COALESCE(psd.surface_total_ha, 0) > 0
      THEN COALESCE(SUM(hr.harvest_real_usd), 0) / psd.surface_total_ha
      ELSE 0
    END::numeric AS total_harvest_usd_ha
  FROM project_surface_data psd
  LEFT JOIN investor_harvest_real hr ON hr.project_id = psd.project_id
  GROUP BY psd.project_id, psd.surface_total_ha
)
SELECT
  pb.project_id,
  pb.project_name,
  pb.customer_id,
  pb.customer_name,
  pb.campaign_id,
  pb.campaign_name,
  
  (
    SELECT jsonb_agg(
      jsonb_build_object(
        'investor_id', irc.investor_id,
        'investor_name', irc.investor_name,
        'share_pct', irc.share_pct_agreed::numeric
      )
      ORDER BY irc.investor_id
    )
    FROM v4_calc.investor_real_contributions irc
    WHERE irc.project_id = pb.project_id
  ) AS investor_headers,
  
  jsonb_build_object(
    'surface_total_ha', pb.surface_total_ha::numeric,
    'lease_fixed_total_usd', pb.lease_fixed_total_usd::numeric,
    'lease_is_fixed', pb.lease_is_fixed,
    'lease_per_ha_usd', pb.lease_per_ha_usd::numeric,
    'admin_total_usd', pb.admin_total_usd::numeric,
    'admin_per_ha_usd', pb.admin_per_ha_usd::numeric
  ) AS general_project_data,
  
  (
    SELECT jsonb_agg(
      jsonb_build_object(
        'key', cat.key,
        'sort_index', cat.sort_index,
        'type', cat.type,
        'label', cat.label,
        'total_usd', cat.total_usd::numeric,
        'total_usd_ha', cat.total_usd_ha::numeric,
        'investors', cat.investors,
        'requires_manual_attribution', cat.requires_manual_attribution,
        'attribution_note', cat.attribution_note
      )
      ORDER BY cat.sort_index
    )
    FROM (
      
      SELECT
        'agrochemicals'::text AS key,
        1 AS sort_index,
        'pre_harvest'::text AS type,
        'Agroquímicos'::text AS label,
        cc.agrochemicals_total_usd AS total_usd,
        CASE
          WHEN pb.surface_total_ha > 0
          THEN cc.agrochemicals_total_usd / pb.surface_total_ha
          ELSE 0
        END AS total_usd_ha,
        (
          SELECT jsonb_agg(
            jsonb_build_object(
              'investor_id', irc.investor_id,
              'investor_name', irc.investor_name,
              'amount_usd', irc.agrochemicals_real_usd::numeric,
              'share_pct', 
                CASE
                  WHEN cc.agrochemicals_total_usd > 0
                  THEN (irc.agrochemicals_real_usd / cc.agrochemicals_total_usd * 100)
                  ELSE 0
                END::numeric
            )
            ORDER BY irc.investor_id
          )
          FROM v4_calc.investor_real_contributions irc
          WHERE irc.project_id = pb.project_id
        ) AS investors,
        false AS requires_manual_attribution,
        NULL AS attribution_note
      FROM v4_report.investor_contribution_categories cc
      WHERE cc.project_id = pb.project_id

      UNION ALL

      
      SELECT
        'fertilizers'::text,
        2,
        'pre_harvest'::text,
        'Fertilizantes'::text,
        cc.fertilizers_total_usd,
        CASE
          WHEN pb.surface_total_ha > 0
          THEN cc.fertilizers_total_usd / pb.surface_total_ha
          ELSE 0
        END,
        (
          SELECT jsonb_agg(
            jsonb_build_object(
              'investor_id', irc.investor_id,
              'investor_name', irc.investor_name,
              'amount_usd', irc.fertilizers_real_usd::numeric,
              'share_pct', 
                CASE
                  WHEN cc.fertilizers_total_usd > 0
                  THEN (irc.fertilizers_real_usd / cc.fertilizers_total_usd * 100)
                  ELSE 0
                END::numeric
            )
            ORDER BY irc.investor_id
          )
          FROM v4_calc.investor_real_contributions irc
          WHERE irc.project_id = pb.project_id
        ),
        false,
        NULL
      FROM v4_report.investor_contribution_categories cc
      WHERE cc.project_id = pb.project_id

      UNION ALL

      
      SELECT
        'seeds'::text,
        3,
        'pre_harvest'::text,
        'Semilla'::text,
        cc.seeds_total_usd,
        CASE
          WHEN pb.surface_total_ha > 0
          THEN cc.seeds_total_usd / pb.surface_total_ha
          ELSE 0
        END,
        (
          SELECT jsonb_agg(
            jsonb_build_object(
              'investor_id', irc.investor_id,
              'investor_name', irc.investor_name,
              'amount_usd', irc.seeds_real_usd::numeric,
              'share_pct', 
                CASE
                  WHEN cc.seeds_total_usd > 0
                  THEN (irc.seeds_real_usd / cc.seeds_total_usd * 100)
                  ELSE 0
                END::numeric
            )
            ORDER BY irc.investor_id
          )
          FROM v4_calc.investor_real_contributions irc
          WHERE irc.project_id = pb.project_id
        ),
        false,
        NULL
      FROM v4_report.investor_contribution_categories cc
      WHERE cc.project_id = pb.project_id

      UNION ALL

      
      SELECT
        'general_labors'::text,
        4,
        'pre_harvest'::text,
        'Labores Generales'::text,
        cc.general_labors_total_usd,
        CASE
          WHEN pb.surface_total_ha > 0
          THEN cc.general_labors_total_usd / pb.surface_total_ha
          ELSE 0
        END,
        (
          SELECT jsonb_agg(
            jsonb_build_object(
              'investor_id', irc.investor_id,
              'investor_name', irc.investor_name,
              'amount_usd', irc.general_labors_real_usd::numeric,
              'share_pct', 
                CASE
                  WHEN cc.general_labors_total_usd > 0
                  THEN (irc.general_labors_real_usd / cc.general_labors_total_usd * 100)
                  ELSE 0
                END::numeric
            )
            ORDER BY irc.investor_id
          )
          FROM v4_calc.investor_real_contributions irc
          WHERE irc.project_id = pb.project_id
        ),
        false,
        NULL
      FROM v4_report.investor_contribution_categories cc
      WHERE cc.project_id = pb.project_id

      UNION ALL

      
      SELECT
        'sowing'::text,
        5,
        'pre_harvest'::text,
        'Siembra'::text,
        cc.sowing_total_usd,
        CASE
          WHEN pb.surface_total_ha > 0
          THEN cc.sowing_total_usd / pb.surface_total_ha
          ELSE 0
        END,
        (
          SELECT jsonb_agg(
            jsonb_build_object(
              'investor_id', irc.investor_id,
              'investor_name', irc.investor_name,
              'amount_usd', irc.sowing_real_usd::numeric,
              'share_pct', 
                CASE
                  WHEN cc.sowing_total_usd > 0
                  THEN (irc.sowing_real_usd / cc.sowing_total_usd * 100)
                  ELSE 0
                END::numeric
            )
            ORDER BY irc.investor_id
          )
          FROM v4_calc.investor_real_contributions irc
          WHERE irc.project_id = pb.project_id
        ),
        false,
        NULL
      FROM v4_report.investor_contribution_categories cc
      WHERE cc.project_id = pb.project_id

      UNION ALL

      
      SELECT
        'irrigation'::text,
        6,
        'pre_harvest'::text,
        'Riego'::text,
        cc.irrigation_total_usd,
        CASE
          WHEN pb.surface_total_ha > 0
          THEN cc.irrigation_total_usd / pb.surface_total_ha
          ELSE 0
        END,
        (
          SELECT jsonb_agg(
            jsonb_build_object(
              'investor_id', irc.investor_id,
              'investor_name', irc.investor_name,
              'amount_usd', irc.irrigation_real_usd::numeric,
              'share_pct', 
                CASE
                  WHEN cc.irrigation_total_usd > 0
                  THEN (irc.irrigation_real_usd / cc.irrigation_total_usd * 100)
                  ELSE 0
                END::numeric
            )
            ORDER BY irc.investor_id
          )
          FROM v4_calc.investor_real_contributions irc
          WHERE irc.project_id = pb.project_id
        ),
        false,
        NULL
      FROM v4_report.investor_contribution_categories cc
      WHERE cc.project_id = pb.project_id

      UNION ALL

      
      SELECT
        'capitalizable_lease'::text,
        7,
        'pre_harvest'::text,
        'Arriendo Capitalizable'::text,
        cc.rent_capitalizable_total_usd,
        CASE
          WHEN pb.surface_total_ha > 0
          THEN cc.rent_capitalizable_total_usd / pb.surface_total_ha
          ELSE 0
        END,
        (
          SELECT jsonb_agg(
            jsonb_build_object(
              'investor_id', irc.investor_id,
              'investor_name', irc.investor_name,
              'amount_usd', irc.rent_real_usd::numeric,
              'share_pct', 
                CASE
                  WHEN cc.rent_capitalizable_total_usd > 0
                  THEN (irc.rent_real_usd / cc.rent_capitalizable_total_usd * 100)
                  ELSE 0
                END::numeric
            )
            ORDER BY irc.investor_id
          )
          FROM v4_calc.investor_real_contributions irc
          WHERE irc.project_id = pb.project_id
        ),
        true,
        'Requiere atribución manual por inversor'
      FROM v4_report.investor_contribution_categories cc
      WHERE cc.project_id = pb.project_id

      UNION ALL

      
      SELECT
        'administration_structure'::text,
        8,
        'pre_harvest'::text,
        'Administración y Estructura'::text,
        cc.administration_total_usd,
        CASE
          WHEN pb.surface_total_ha > 0
          THEN cc.administration_total_usd / pb.surface_total_ha
          ELSE 0
        END,
        (
          SELECT jsonb_agg(
            jsonb_build_object(
              'investor_id', irc.investor_id,
              'investor_name', irc.investor_name,
              'amount_usd', irc.administration_real_usd::numeric,
              'share_pct', 
                CASE
                  WHEN cc.administration_total_usd > 0
                  THEN (irc.administration_real_usd / cc.administration_total_usd * 100)
                  ELSE 0
                END::numeric
            )
            ORDER BY irc.investor_id
          )
          FROM v4_calc.investor_real_contributions irc
          WHERE irc.project_id = pb.project_id
        ),
        true,
        'Requiere atribución manual por inversor'
      FROM v4_report.investor_contribution_categories cc
      WHERE cc.project_id = pb.project_id
    ) AS cat
  ) AS contribution_categories,
  
  (
    SELECT jsonb_agg(
      jsonb_build_object(
        'investor_id', irc.investor_id,
        'investor_name', irc.investor_name,
        'agreed_share_pct', irc.share_pct_agreed,
        'agreed_usd', (irc.project_total_contributions_usd * irc.share_pct_agreed / 100)::numeric,
        'actual_usd', irc.total_real_contribution_usd::numeric,
        'share_pct', irc.contributions_progress_pct::numeric,
        'adjustment_usd', (irc.total_real_contribution_usd - (irc.project_total_contributions_usd * irc.share_pct_agreed / 100))::numeric
      )
      ORDER BY irc.investor_id
    )
    FROM v4_calc.investor_real_contributions irc
    WHERE irc.project_id = pb.project_id
  ) AS investor_contribution_comparison,
  
  jsonb_build_object(
    'rows', jsonb_build_array(
      jsonb_build_object(
        'key', 'harvest',
        'type', 'harvest',
        'total_usd', COALESCE(ht.total_harvest_usd, 0)::numeric,
        'total_us_ha', COALESCE(ht.total_harvest_usd_ha, 0)::numeric,
        'investors', COALESCE((
          SELECT jsonb_agg(
            jsonb_build_object(
              'investor_id', ib.investor_id,
              'investor_name', ib.investor_name,
              'amount_usd', COALESCE(hr.harvest_real_usd, 0)::numeric,
              'share_pct', 
                CASE 
                  WHEN COALESCE(ht.total_harvest_usd, 0)::numeric > 0 
                  THEN (COALESCE(hr.harvest_real_usd, 0)::numeric / COALESCE(ht.total_harvest_usd, 0)::numeric * 100)::numeric
                  ELSE 0::numeric
                END
            )
            ORDER BY ib.investor_id
          )
          FROM investor_base ib
          LEFT JOIN investor_harvest_real hr
            ON hr.project_id = ib.project_id AND hr.investor_id = ib.investor_id
          WHERE ib.project_id = pb.project_id
        ), '[]'::jsonb)
      ),
      jsonb_build_object(
        'key', 'totals',
        'type', 'totals',
        'total_usd', COALESCE(ht.total_harvest_usd, 0)::numeric,
        'total_us_ha', COALESCE(ht.total_harvest_usd_ha, 0)::numeric,
        'investors', COALESCE((
          SELECT jsonb_agg(
            jsonb_build_object(
              'investor_id', ib.investor_id,
              'investor_name', ib.investor_name,
              'amount_usd', COALESCE(hr.harvest_real_usd, 0)::numeric,
              'share_pct', 
                CASE 
                  WHEN COALESCE(ht.total_harvest_usd, 0)::numeric > 0 
                  THEN (COALESCE(hr.harvest_real_usd, 0)::numeric / COALESCE(ht.total_harvest_usd, 0)::numeric * 100)::numeric
                  ELSE 0::numeric
                END
            )
            ORDER BY ib.investor_id
          )
          FROM investor_base ib
          LEFT JOIN investor_harvest_real hr
            ON hr.project_id = ib.project_id AND hr.investor_id = ib.investor_id
          WHERE ib.project_id = pb.project_id
        ), '[]'::jsonb)
      )
    ),
    'footer_payment_agreed', COALESCE((
      SELECT jsonb_agg(
        jsonb_build_object(
          'investor_id', ib.investor_id,
          'investor_name', ib.investor_name,
          'amount_usd', (COALESCE(ht.total_harvest_usd, 0)::numeric * ib.share_pct_agreed::numeric / 100)::numeric,
          'share_pct', ib.share_pct_agreed::numeric
        )
        ORDER BY ib.investor_id
      )
      FROM investor_base ib
      WHERE ib.project_id = pb.project_id
    ), '[]'::jsonb),
    'footer_payment_adjustment', COALESCE((
      SELECT jsonb_agg(
        jsonb_build_object(
          'investor_id', ib.investor_id,
          'investor_name', ib.investor_name,
          'amount_usd', (
            COALESCE(hr.harvest_real_usd, 0)::numeric -
            (COALESCE(ht.total_harvest_usd, 0)::numeric * ib.share_pct_agreed::numeric / 100)
          )::numeric,
          'share_pct', ib.share_pct_agreed::numeric
        )
        ORDER BY ib.investor_id
      )
      FROM investor_base ib
      LEFT JOIN investor_harvest_real hr
        ON hr.project_id = ib.project_id AND hr.investor_id = ib.investor_id
      WHERE ib.project_id = pb.project_id
    ), '[]'::jsonb)
  ) AS harvest_settlement
FROM v4_report.investor_project_base pb
JOIN v4_report.investor_contribution_categories cc ON cc.project_id = pb.project_id
LEFT JOIN harvest_totals ht ON ht.project_id = pb.project_id
ORDER BY pb.project_id;

CREATE OR REPLACE VIEW v4_report.dashboard_contributions_progress AS
SELECT
  project_id,
  investor_id,
  investor_name,
  share_pct_agreed AS investor_percentage_pct,
  contributions_progress_pct
FROM v4_calc.investor_real_contributions;

COMMIT;
