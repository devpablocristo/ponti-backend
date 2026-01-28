-- =============================================================================
-- MIGRACIÓN 000331: Per-ha usa superficie total (sin siembra)
-- =============================================================================
--
-- Propósito:
-- - Lotes: costo usd/ha divide por hectáreas del lote
-- - Reporte por cultivo: divide por superficie total
-- - Resumen: arriendo fijo aparece aunque no haya siembra
-- Nota: Comentarios en español, código en inglés
--

BEGIN;

-- 1) lot_base_costs: costo por ha usa superficie total
CREATE OR REPLACE VIEW v4_calc.lot_base_costs AS
WITH
raw AS (
  SELECT
    f.project_id,
    f.id AS field_id,
    l.current_crop_id,
    l.id AS lot_id,
    l.name AS lot_name,
    COALESCE(l.hectares, 0) AS hectares,
    COALESCE(l.tons, 0) AS tons,
    l.sowing_date
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  WHERE l.deleted_at IS NULL
),
areas AS (
  SELECT
    r.lot_id,
    v4_ssot.seeded_area_for_lot(r.lot_id) AS sowed_area_ha,
    v4_ssot.harvested_area_for_lot(r.lot_id) AS harvested_area_ha
  FROM raw r
),
costs AS (
  SELECT
    lot_id,
    MAX(COALESCE(labor_cost_usd, 0))::numeric AS labor_cost_usd,
    MAX(COALESCE(supplies_cost_usd, 0))::numeric AS supplies_cost_usd,
    MAX(COALESCE(direct_cost_usd, 0))::numeric AS direct_cost_usd
  FROM public.v3_workorder_metrics
  GROUP BY lot_id
),
ssot_values AS (
  SELECT
    r.lot_id,
    v4_ssot.yield_tn_per_ha_for_lot(r.lot_id) AS yield_tn_per_ha,
    v4_ssot.income_net_total_for_lot(r.lot_id) AS income_net_total_usd,
    v4_ssot.rent_per_ha_for_lot(r.lot_id) AS rent_per_ha_usd,
    v4_ssot.rent_fixed_only_for_lot(r.lot_id) AS rent_fixed_per_ha_usd,
    v4_ssot.admin_cost_per_ha_for_lot(r.lot_id) AS admin_cost_per_ha_usd
  FROM raw r
),
derived AS (
  SELECT
    r.project_id, r.field_id, r.current_crop_id, r.lot_id, r.lot_name,
    r.hectares, r.tons, r.sowing_date,
    COALESCE(a.sowed_area_ha, 0)::numeric AS sowed_area_ha,
    COALESCE(a.harvested_area_ha, 0)::numeric AS harvested_area_ha,
    COALESCE(c.labor_cost_usd, 0)::numeric AS labor_cost_usd,
    COALESCE(c.supplies_cost_usd, 0)::numeric AS supplies_cost_usd,
    COALESCE(c.direct_cost_usd, 0)::numeric AS direct_cost_usd,
    COALESCE(s.yield_tn_per_ha, 0) AS yield_tn_per_ha,
    COALESCE(s.income_net_total_usd, 0)::numeric AS income_net_total_usd,
    COALESCE(s.rent_per_ha_usd, 0)::numeric AS rent_per_ha_usd,
    COALESCE(s.rent_fixed_per_ha_usd, 0)::numeric AS rent_fixed_per_ha_usd,
    COALESCE(s.admin_cost_per_ha_usd, 0)::numeric AS admin_cost_per_ha_usd
  FROM raw r
  LEFT JOIN areas a ON a.lot_id = r.lot_id
  LEFT JOIN costs c ON c.lot_id = r.lot_id
  LEFT JOIN ssot_values s ON s.lot_id = r.lot_id
)
SELECT
  project_id,
  field_id,
  current_crop_id,
  lot_id,
  lot_name,
  hectares,
  tons,
  sowing_date,
  sowed_area_ha,
  harvested_area_ha,
  yield_tn_per_ha,
  labor_cost_usd,
  supplies_cost_usd,
  direct_cost_usd,
  income_net_total_usd,
  v4_core.per_ha(income_net_total_usd, hectares::numeric) AS income_net_per_ha_usd,
  -- FIX: costo por ha usa superficie total
  v4_core.per_ha(direct_cost_usd, hectares::numeric) AS direct_cost_per_ha_usd,
  rent_per_ha_usd,
  rent_fixed_per_ha_usd,
  admin_cost_per_ha_usd
FROM derived d;

COMMENT ON VIEW v4_calc.lot_base_costs IS 
'Backbone cálculos por lote. FIX 000331: per_ha usa superficie total.';

-- 2) field_crop_labores: per_ha usa superficie total
CREATE OR REPLACE VIEW v4_report.field_crop_labores AS
WITH lot_base AS (
  SELECT
    f.project_id,
    f.id AS field_id,
    l.current_crop_id AS crop_id,
    l.id AS lot_id,
    l.hectares AS surface_ha,
    v3_lot_ssot.seeded_area_for_lot(l.id)::numeric AS sowed_area_ha
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  WHERE l.deleted_at IS NULL AND l.current_crop_id IS NOT NULL
),
labor_costs AS (
  SELECT
    lb.project_id,
    lb.field_id,
    lb.crop_id,
    lb.lot_id,
    lb.sowed_area_ha,
    lb.surface_ha,
    COALESCE(SUM(CASE WHEN cat.name = 'Siembra' THEN lab.price * w.effective_area ELSE 0 END), 0)::numeric AS siembra_usd,
    COALESCE(SUM(CASE WHEN cat.name = 'Pulverización' THEN lab.price * w.effective_area ELSE 0 END), 0)::numeric AS pulverizacion_usd,
    COALESCE(SUM(CASE WHEN cat.name = 'Riego' THEN lab.price * w.effective_area ELSE 0 END), 0)::numeric AS riego_usd,
    COALESCE(SUM(CASE WHEN cat.name = 'Cosecha' THEN lab.price * w.effective_area ELSE 0 END), 0)::numeric AS cosecha_usd,
    COALESCE(SUM(CASE WHEN cat.name = 'Otras Labores' THEN lab.price * w.effective_area ELSE 0 END), 0)::numeric AS otras_labores_usd
  FROM lot_base lb
  LEFT JOIN public.workorders w ON w.lot_id = lb.lot_id AND w.deleted_at IS NULL
  LEFT JOIN public.labors lab ON lab.id = w.labor_id AND lab.deleted_at IS NULL
  LEFT JOIN public.categories cat ON cat.id = lab.category_id
  GROUP BY lb.project_id, lb.field_id, lb.crop_id, lb.lot_id, lb.sowed_area_ha, lb.surface_ha
)
SELECT
  project_id,
  field_id,
  crop_id AS current_crop_id,
  COALESCE(SUM(siembra_usd), 0) AS siembra_total_usd,
  COALESCE(SUM(pulverizacion_usd), 0) AS pulverizacion_total_usd,
  COALESCE(SUM(riego_usd), 0) AS riego_total_usd,
  COALESCE(SUM(cosecha_usd), 0) AS cosecha_total_usd,
  COALESCE(SUM(otras_labores_usd), 0) AS otras_labores_total_usd,
  v3_core_ssot.safe_div(COALESCE(SUM(siembra_usd), 0), COALESCE(SUM(surface_ha), 1)::numeric) AS siembra_usd_ha,
  v3_core_ssot.safe_div(COALESCE(SUM(pulverizacion_usd), 0), COALESCE(SUM(surface_ha), 1)::numeric) AS pulverizacion_usd_ha,
  v3_core_ssot.safe_div(COALESCE(SUM(riego_usd), 0), COALESCE(SUM(surface_ha), 1)::numeric) AS riego_usd_ha,
  v3_core_ssot.safe_div(COALESCE(SUM(cosecha_usd), 0), COALESCE(SUM(surface_ha), 1)::numeric) AS cosecha_usd_ha,
  v3_core_ssot.safe_div(COALESCE(SUM(otras_labores_usd), 0), COALESCE(SUM(surface_ha), 1)::numeric) AS otras_labores_usd_ha,
  COALESCE(SUM(siembra_usd) + SUM(pulverizacion_usd) + SUM(riego_usd) + 
   SUM(cosecha_usd) + SUM(otras_labores_usd), 0) AS total_labores_usd,
  v3_core_ssot.safe_div(
    COALESCE(SUM(siembra_usd) + SUM(pulverizacion_usd) + SUM(riego_usd) + 
     SUM(cosecha_usd) + SUM(otras_labores_usd), 0),
    COALESCE(SUM(surface_ha), 1)::numeric
  ) AS total_labores_usd_ha
FROM labor_costs
GROUP BY project_id, field_id, crop_id;

-- 3) field_crop_insumos: per_ha usa superficie total
CREATE OR REPLACE VIEW v4_report.field_crop_insumos AS
WITH lot_base AS (
  SELECT
    f.project_id,
    f.id AS field_id,
    l.current_crop_id AS crop_id,
    l.id AS lot_id,
    l.hectares AS surface_ha,
    v3_lot_ssot.seeded_area_for_lot(l.id)::numeric AS sowed_area_ha
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  WHERE l.deleted_at IS NULL AND l.current_crop_id IS NOT NULL
),
supply_costs AS (
  SELECT
    lb.project_id,
    lb.field_id,
    lb.crop_id,
    lb.lot_id,
    lb.sowed_area_ha,
    lb.surface_ha,
    v3_lot_ssot.supply_cost_for_lot_by_category(lb.lot_id, 'Semilla') AS semillas_usd,
    v3_lot_ssot.supply_cost_for_lot_by_category(lb.lot_id, 'Curasemillas') AS curasemillas_usd,
    v3_lot_ssot.supply_cost_for_lot_by_category(lb.lot_id, 'Herbicidas') AS herbicidas_usd,
    v3_lot_ssot.supply_cost_for_lot_by_category(lb.lot_id, 'Insecticidas') AS insecticidas_usd,
    v3_lot_ssot.supply_cost_for_lot_by_category(lb.lot_id, 'Fungicidas') AS fungicidas_usd,
    v3_lot_ssot.supply_cost_for_lot_by_category(lb.lot_id, 'Coadyuvantes') AS coadyuvantes_usd,
    v3_lot_ssot.supply_cost_for_lot_by_category(lb.lot_id, 'Fertilizantes') AS fertilizantes_usd,
    v3_lot_ssot.supply_cost_for_lot_by_category(lb.lot_id, 'Otros Insumos') AS otros_insumos_usd
  FROM lot_base lb
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
  v3_core_ssot.safe_div(COALESCE(SUM(semillas_usd), 0), COALESCE(SUM(surface_ha), 1)::numeric) AS semillas_usd_ha,
  v3_core_ssot.safe_div(COALESCE(SUM(curasemillas_usd), 0), COALESCE(SUM(surface_ha), 1)::numeric) AS curasemillas_usd_ha,
  v3_core_ssot.safe_div(COALESCE(SUM(herbicidas_usd), 0), COALESCE(SUM(surface_ha), 1)::numeric) AS herbicidas_usd_ha,
  v3_core_ssot.safe_div(COALESCE(SUM(insecticidas_usd), 0), COALESCE(SUM(surface_ha), 1)::numeric) AS insecticidas_usd_ha,
  v3_core_ssot.safe_div(COALESCE(SUM(fungicidas_usd), 0), COALESCE(SUM(surface_ha), 1)::numeric) AS fungicidas_usd_ha,
  v3_core_ssot.safe_div(COALESCE(SUM(coadyuvantes_usd), 0), COALESCE(SUM(surface_ha), 1)::numeric) AS coadyuvantes_usd_ha,
  v3_core_ssot.safe_div(COALESCE(SUM(fertilizantes_usd), 0), COALESCE(SUM(surface_ha), 1)::numeric) AS fertilizantes_usd_ha,
  v3_core_ssot.safe_div(COALESCE(SUM(otros_insumos_usd), 0), COALESCE(SUM(surface_ha), 1)::numeric) AS otros_insumos_usd_ha,
  COALESCE(SUM(semillas_usd) + SUM(curasemillas_usd) + SUM(herbicidas_usd) + 
   SUM(insecticidas_usd) + SUM(fungicidas_usd) + SUM(coadyuvantes_usd) +
   SUM(fertilizantes_usd) + SUM(otros_insumos_usd), 0) AS total_insumos_usd,
  v3_core_ssot.safe_div(
    COALESCE(SUM(semillas_usd) + SUM(curasemillas_usd) + SUM(herbicidas_usd) + 
     SUM(insecticidas_usd) + SUM(fungicidas_usd) + SUM(coadyuvantes_usd) +
     SUM(fertilizantes_usd) + SUM(otros_insumos_usd), 0),
    COALESCE(SUM(surface_ha), 1)::numeric
  ) AS total_insumos_usd_ha
FROM supply_costs
GROUP BY project_id, field_id, crop_id;

-- 4) field_crop_economicos: per_ha usa superficie total y arriendo fijo visible
CREATE OR REPLACE VIEW v4_report.field_crop_economicos AS
WITH lot_base AS (
  SELECT
    f.project_id,
    f.id AS field_id,
    l.current_crop_id AS crop_id,
    l.id AS lot_id,
    l.hectares AS surface_ha,
    v3_lot_ssot.seeded_area_for_lot(l.id)::numeric AS sowed_area_ha,
    l.tons
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  WHERE l.deleted_at IS NULL AND l.current_crop_id IS NOT NULL
),
aggregated AS (
  SELECT
    project_id,
    field_id,
    crop_id,
    MIN(lot_id) AS sample_lot_id,
    SUM(tons)::numeric AS production_tn,
    SUM(sowed_area_ha)::numeric AS sown_area_ha,
    SUM(surface_ha)::numeric AS surface_ha,
    SUM(v3_lot_ssot.labor_cost_for_lot(lot_id))::numeric AS labor_costs_usd,
    SUM(
      v3_lot_ssot.supply_cost_for_lot_by_category(lot_id, 'Semilla') +
      v3_lot_ssot.supply_cost_for_lot_by_category(lot_id, 'Curasemillas') +
      v3_lot_ssot.supply_cost_for_lot_by_category(lot_id, 'Herbicidas') +
      v3_lot_ssot.supply_cost_for_lot_by_category(lot_id, 'Insecticidas') +
      v3_lot_ssot.supply_cost_for_lot_by_category(lot_id, 'Fungicidas') +
      v3_lot_ssot.supply_cost_for_lot_by_category(lot_id, 'Coadyuvantes') +
      v3_lot_ssot.supply_cost_for_lot_by_category(lot_id, 'Fertilizantes') +
      v3_lot_ssot.supply_cost_for_lot_by_category(lot_id, 'Otros Insumos')
    )::numeric AS supply_costs_usd,
    -- Arriendo FIJO para mostrar en UI y total invertido
    SUM(v3_lot_ssot.rent_fixed_only_for_lot(lot_id) * surface_ha)::numeric AS rent_fixed_usd,
    -- Arriendo TOTAL (fijo + % ingresos) para resultado operativo
    SUM(v3_lot_ssot.rent_per_ha_for_lot(lot_id) * surface_ha)::numeric AS rent_total_usd,
    SUM(v3_lot_ssot.admin_cost_per_ha_for_lot(lot_id) * surface_ha)::numeric AS administration_usd
  FROM lot_base
  GROUP BY project_id, field_id, crop_id
)
SELECT
  project_id,
  field_id,
  crop_id AS current_crop_id,
  (labor_costs_usd + supply_costs_usd) AS gastos_directos_usd,
  v3_core_ssot.safe_div(
    (labor_costs_usd + supply_costs_usd),
    surface_ha
  ) AS gastos_directos_usd_ha,
  (
    (production_tn * v3_lot_ssot.net_price_usd_for_lot(sample_lot_id)) -
    (labor_costs_usd + supply_costs_usd)
  ) AS margen_bruto_usd,
  (
    (v3_core_ssot.safe_div(production_tn, surface_ha) * v3_lot_ssot.net_price_usd_for_lot(sample_lot_id)) -
    v3_core_ssot.safe_div((labor_costs_usd + supply_costs_usd), surface_ha)
  ) AS margen_bruto_usd_ha,
  -- Mostrar arriendo FIJO (capitalizable)
  rent_fixed_usd AS arriendo_usd,
  v3_core_ssot.safe_div(rent_fixed_usd, surface_ha) AS arriendo_usd_ha,
  administration_usd AS adm_estructura_usd,
  v3_core_ssot.safe_div(administration_usd, surface_ha) AS adm_estructura_usd_ha,
  -- Resultado operativo usa arriendo TOTAL
  (
    (production_tn * v3_lot_ssot.net_price_usd_for_lot(sample_lot_id)) -
    (labor_costs_usd + supply_costs_usd) - rent_total_usd - administration_usd
  ) AS resultado_operativo_usd,
  (
    (v3_core_ssot.safe_div(production_tn, surface_ha) * v3_lot_ssot.net_price_usd_for_lot(sample_lot_id)) -
    v3_core_ssot.safe_div((labor_costs_usd + supply_costs_usd), surface_ha) -
    v3_core_ssot.safe_div(rent_total_usd, surface_ha) -
    v3_core_ssot.safe_div(administration_usd, surface_ha)
  ) AS resultado_operativo_usd_ha
FROM aggregated;

-- 5) field_crop_rentabilidad: per_ha usa superficie total
CREATE OR REPLACE VIEW v4_report.field_crop_rentabilidad AS
WITH lot_base AS (
  SELECT
    f.project_id,
    f.id AS field_id,
    l.current_crop_id AS crop_id,
    l.id AS lot_id,
    l.hectares AS surface_ha,
    l.tons,
    v3_lot_ssot.seeded_area_for_lot(l.id)::numeric AS sowed_area_ha
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  WHERE l.deleted_at IS NULL AND l.current_crop_id IS NOT NULL
),
aggregated AS (
  SELECT
    project_id,
    field_id,
    crop_id,
    MIN(lot_id) AS sample_lot_id,
    SUM(tons)::numeric AS production_tn,
    SUM(sowed_area_ha)::numeric AS sown_area_ha,
    SUM(surface_ha)::numeric AS surface_ha,
    SUM(
      v3_lot_ssot.labor_cost_for_lot(lot_id) +
      v3_lot_ssot.supply_cost_for_lot_by_category(lot_id, 'Semilla') +
      v3_lot_ssot.supply_cost_for_lot_by_category(lot_id, 'Curasemillas') +
      v3_lot_ssot.supply_cost_for_lot_by_category(lot_id, 'Herbicidas') +
      v3_lot_ssot.supply_cost_for_lot_by_category(lot_id, 'Insecticidas') +
      v3_lot_ssot.supply_cost_for_lot_by_category(lot_id, 'Fungicidas') +
      v3_lot_ssot.supply_cost_for_lot_by_category(lot_id, 'Coadyuvantes') +
      v3_lot_ssot.supply_cost_for_lot_by_category(lot_id, 'Fertilizantes') +
      v3_lot_ssot.supply_cost_for_lot_by_category(lot_id, 'Otros Insumos')
    )::numeric AS direct_cost_usd,
    -- Total Invertido usa arriendo FIJO (capitalizable)
    SUM(v3_lot_ssot.rent_fixed_only_for_lot(lot_id) * surface_ha)::numeric AS rent_usd,
    SUM(v3_lot_ssot.admin_cost_per_ha_for_lot(lot_id) * surface_ha)::numeric AS administration_usd
  FROM lot_base
  GROUP BY project_id, field_id, crop_id
)
SELECT
  project_id,
  field_id,
  crop_id AS current_crop_id,
  (direct_cost_usd + rent_usd + administration_usd) AS total_invertido_usd,
  v3_core_ssot.safe_div(
    (direct_cost_usd + rent_usd + administration_usd),
    surface_ha
  ) AS total_invertido_usd_ha,
  v3_lot_ssot.renta_pct(
    (
      (production_tn * v3_lot_ssot.net_price_usd_for_lot(sample_lot_id)) -
      direct_cost_usd - rent_usd - administration_usd
    ),
    (direct_cost_usd + rent_usd + administration_usd)
  ) AS renta_pct,
  v3_core_ssot.safe_div(
    v3_core_ssot.safe_div((direct_cost_usd + rent_usd + administration_usd), surface_ha),
    v3_lot_ssot.net_price_usd_for_lot(sample_lot_id)
  ) AS rinde_indiferencia_total_usd_tn
FROM aggregated;

-- 6) field_crop_metrics: per_ha usa superficie total
CREATE OR REPLACE VIEW v4_report.field_crop_metrics AS
WITH 
lot_base AS (
  SELECT
    f.project_id,
    f.id AS field_id,
    f.name AS field_name,
    l.current_crop_id,
    c.name AS crop_name,
    l.id AS lot_id,
    l.hectares,
    l.tons,
    COALESCE(v3_lot_ssot.seeded_area_for_lot(l.id), 0)::numeric AS sowed_area_ha,
    COALESCE(v3_lot_ssot.harvested_area_for_lot(l.id), 0)::numeric AS harvested_area_ha,
    COALESCE(v3_lot_ssot.yield_tn_per_ha_for_lot(l.id), 0) AS yield_tn_per_ha,
    COALESCE(v3_lot_ssot.labor_cost_for_lot(l.id), 0)::numeric AS labor_cost_usd,
    COALESCE(v3_lot_ssot.supply_cost_for_lot_base(l.id), 0)::numeric AS supply_cost_usd,
    COALESCE(v3_lot_ssot.net_price_usd_for_lot(l.id), 0)::numeric AS net_price_usd,
    COALESCE(v3_lot_ssot.rent_per_ha_for_lot(l.id), 0)::numeric AS rent_per_ha,
    COALESCE(v3_calc.admin_cost_per_ha_for_lot(l.id), 0)::numeric AS admin_per_ha,
    COALESCE(v3_report_ssot.board_price_for_lot(l.id), 0)::numeric AS board_price,
    COALESCE(v3_report_ssot.freight_cost_for_lot(l.id), 0)::numeric AS freight_cost,
    COALESCE(v3_report_ssot.commercial_cost_for_lot(l.id), 0)::numeric AS commercial_cost
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  JOIN public.projects p ON p.id = f.project_id AND p.deleted_at IS NULL
  LEFT JOIN public.crops c ON c.id = l.current_crop_id AND c.deleted_at IS NULL
  WHERE l.deleted_at IS NULL 
    AND l.current_crop_id IS NOT NULL
    AND l.hectares > 0
),
aggregated AS (
  SELECT
    lb.project_id,
    lb.field_id,
    lb.field_name,
    lb.current_crop_id,
    lb.crop_name,
    SUM(lb.hectares)::numeric AS superficie_total,
    SUM(lb.sowed_area_ha)::numeric AS superficie_sembrada_ha,
    SUM(lb.harvested_area_ha)::numeric AS area_cosechada_ha,
    SUM(lb.tons)::numeric AS produccion_tn,
    CASE WHEN SUM(lb.sowed_area_ha) > 0 
      THEN SUM(lb.yield_tn_per_ha * lb.sowed_area_ha) / SUM(lb.sowed_area_ha)
      ELSE 0 
    END AS rendimiento_tn_ha,
    SUM(lb.labor_cost_usd)::numeric AS costos_labores_usd,
    SUM(lb.supply_cost_usd)::numeric AS costos_insumos_usd,
    CASE WHEN SUM(lb.tons) > 0 
      THEN SUM(lb.board_price * lb.tons) / SUM(lb.tons)
      ELSE 0 
    END AS precio_bruto_usd_tn,
    CASE WHEN SUM(lb.tons) > 0 
      THEN SUM(lb.freight_cost * lb.tons) / SUM(lb.tons)
      ELSE 0 
    END AS gasto_flete_usd_tn,
    CASE WHEN SUM(lb.tons) > 0 
      THEN SUM(lb.commercial_cost * lb.tons) / SUM(lb.tons)
      ELSE 0 
    END AS gasto_comercial_usd_tn,
    CASE WHEN SUM(lb.tons) > 0 
      THEN SUM(lb.net_price_usd * lb.tons) / SUM(lb.tons)
      ELSE 0 
    END AS precio_neto_usd_tn,
    -- Arriendo y admin por superficie total
    SUM(lb.rent_per_ha * lb.hectares)::numeric AS arriendo_total_usd,
    SUM(lb.admin_per_ha * lb.hectares)::numeric AS admin_total_usd,
    SUM(lb.tons * lb.net_price_usd)::numeric AS ingreso_neto_total
  FROM lot_base lb
  GROUP BY lb.project_id, lb.field_id, lb.field_name, lb.current_crop_id, lb.crop_name
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
  v3_core_ssot.safe_div(a.ingreso_neto_total, a.superficie_total) AS ingreso_neto_usd_ha,
  a.costos_labores_usd,
  v3_core_ssot.safe_div(a.costos_labores_usd, a.superficie_total) AS costos_labores_usd_ha,
  a.costos_insumos_usd,
  v3_core_ssot.safe_div(a.costos_insumos_usd, a.superficie_total) AS costos_insumos_usd_ha,
  (a.costos_labores_usd + a.costos_insumos_usd)::numeric AS total_costos_directos_usd,
  v3_core_ssot.safe_div(a.costos_labores_usd + a.costos_insumos_usd, a.superficie_total) AS costos_directos_usd_ha,
  (a.ingreso_neto_total - a.costos_labores_usd - a.costos_insumos_usd)::numeric AS margen_bruto_usd,
  v3_core_ssot.safe_div(a.ingreso_neto_total - a.costos_labores_usd - a.costos_insumos_usd, a.superficie_total) AS margen_bruto_usd_ha,
  a.arriendo_total_usd AS arriendo_usd,
  v3_core_ssot.safe_div(a.arriendo_total_usd, a.superficie_total) AS arriendo_usd_ha,
  a.admin_total_usd AS administracion_usd,
  v3_core_ssot.safe_div(a.admin_total_usd, a.superficie_total) AS administracion_usd_ha,
  (a.ingreso_neto_total - a.costos_labores_usd - a.costos_insumos_usd - a.arriendo_total_usd - a.admin_total_usd)::numeric AS resultado_operativo_usd,
  v3_core_ssot.safe_div(
    a.ingreso_neto_total - a.costos_labores_usd - a.costos_insumos_usd - a.arriendo_total_usd - a.admin_total_usd, 
    a.superficie_total
  ) AS resultado_operativo_usd_ha,
  (a.costos_labores_usd + a.costos_insumos_usd + a.arriendo_total_usd + a.admin_total_usd)::numeric AS total_invertido_usd,
  v3_core_ssot.safe_div(
    a.costos_labores_usd + a.costos_insumos_usd + a.arriendo_total_usd + a.admin_total_usd, 
    a.superficie_total
  ) AS total_invertido_usd_ha,
  CASE WHEN (a.costos_labores_usd + a.costos_insumos_usd + a.arriendo_total_usd + a.admin_total_usd) > 0
    THEN ((a.ingreso_neto_total - a.costos_labores_usd - a.costos_insumos_usd - a.arriendo_total_usd - a.admin_total_usd) / 
          (a.costos_labores_usd + a.costos_insumos_usd + a.arriendo_total_usd + a.admin_total_usd) * 100)::double precision
    ELSE 0
  END AS renta_pct,
  CASE WHEN a.precio_neto_usd_tn > 0 AND a.superficie_total > 0
    THEN ((a.costos_labores_usd + a.costos_insumos_usd + a.arriendo_total_usd + a.admin_total_usd) / a.superficie_total / a.precio_neto_usd_tn)::numeric
    ELSE 0
  END AS rinde_indiferencia_usd_tn
FROM aggregated a;

COMMENT ON VIEW v4_report.field_crop_metrics IS 
'OPTIMIZADO 000324: per_ha usa superficie total (000331).';

-- 7) summary_results: usa superficie_total
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

COMMENT ON VIEW v4_report.summary_results IS 
'SSOT: Agrega desde field_crop_metrics. per_ha usa superficie total (000331).';

COMMIT;
