-- =============================================================================
-- Migration: 000320_recreate_all_field_crop_views
-- Description: Recrea TODAS las vistas v4_report.field_crop_* en orden correcto
-- Garantiza consistencia después de migraciones parciales
-- =============================================================================

-- 1. DROP en orden inverso (dependencias primero)
DROP VIEW IF EXISTS v4_report.field_crop_metrics CASCADE;
DROP VIEW IF EXISTS v4_report.field_crop_rentabilidad CASCADE;
DROP VIEW IF EXISTS v4_report.field_crop_economicos CASCADE;
DROP VIEW IF EXISTS v4_report.field_crop_insumos CASCADE;
DROP VIEW IF EXISTS v4_report.field_crop_labores CASCADE;
DROP VIEW IF EXISTS v4_report.field_crop_cultivos CASCADE;

-- =============================================================================
-- 2. FIELD_CROP_CULTIVOS (base)
-- =============================================================================
CREATE OR REPLACE VIEW v4_report.field_crop_cultivos AS
SELECT
  f.project_id,
  f.id AS field_id,
  f.name AS field_name,
  c.id AS current_crop_id,
  c.name AS crop_name,
  SUM(l.hectares)::numeric AS superficie_total,
  SUM(v3_lot_ssot.seeded_area_for_lot(l.id))::numeric AS superficie_sembrada_ha,
  SUM(l.harvested_area)::numeric AS area_cosechada_ha,
  SUM(l.tons)::numeric AS produccion_tn,
  v3_lot_ssot.gross_price_usd_for_lot(MIN(l.id)) AS precio_bruto_usd_tn,
  v3_lot_ssot.freight_cost_for_lot(MIN(l.id)) AS gasto_flete_usd_tn,
  v3_lot_ssot.commercial_cost_for_lot(MIN(l.id)) AS gasto_comercial_usd_tn,
  v3_lot_ssot.net_price_usd_for_lot(MIN(l.id)) AS precio_neto_usd_tn,
  CASE 
    WHEN SUM(v3_lot_ssot.seeded_area_for_lot(l.id)) > 0 
    THEN SUM(l.tons) / SUM(v3_lot_ssot.seeded_area_for_lot(l.id))
    ELSE 0
  END AS rendimiento_tn_ha,
  CASE 
    WHEN SUM(v3_lot_ssot.seeded_area_for_lot(l.id)) > 0 
    THEN (SUM(l.tons) / SUM(v3_lot_ssot.seeded_area_for_lot(l.id))) * v3_lot_ssot.net_price_usd_for_lot(MIN(l.id))
    ELSE 0
  END AS ingreso_neto_por_ha
FROM public.lots l
JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
JOIN public.crops c ON c.id = l.current_crop_id
WHERE l.deleted_at IS NULL AND l.current_crop_id IS NOT NULL
GROUP BY f.project_id, f.id, f.name, c.id, c.name;

-- =============================================================================
-- 3. FIELD_CROP_LABORES
-- =============================================================================
CREATE OR REPLACE VIEW v4_report.field_crop_labores AS
WITH lot_labores AS (
  SELECT
    f.project_id,
    f.id AS field_id,
    l.current_crop_id,
    l.id AS lot_id,
    v3_lot_ssot.seeded_area_for_lot(l.id) AS sowed_area_ha,
    v3_lot_ssot.labor_cost_by_category_for_lot(l.id, 'Siembra') AS siembra_usd,
    v3_lot_ssot.labor_cost_by_category_for_lot(l.id, 'Pulverización') AS pulverizacion_usd,
    v3_lot_ssot.labor_cost_by_category_for_lot(l.id, 'Riego') AS riego_usd,
    v3_lot_ssot.labor_cost_by_category_for_lot(l.id, 'Cosecha') AS cosecha_usd,
    v3_lot_ssot.labor_cost_by_category_for_lot(l.id, 'Otras Labores') AS otras_labores_usd,
    v3_lot_ssot.labor_cost_for_lot(l.id) AS total_labores_usd
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  WHERE l.deleted_at IS NULL AND l.current_crop_id IS NOT NULL
)
SELECT
  project_id,
  field_id,
  current_crop_id,
  SUM(sowed_area_ha)::numeric AS sown_area_ha,
  SUM(siembra_usd)::numeric AS siembra_usd,
  v3_core_ssot.safe_div(SUM(siembra_usd), SUM(sowed_area_ha)) AS siembra_usd_ha,
  SUM(pulverizacion_usd)::numeric AS pulverizacion_usd,
  v3_core_ssot.safe_div(SUM(pulverizacion_usd), SUM(sowed_area_ha)) AS pulverizacion_usd_ha,
  SUM(riego_usd)::numeric AS riego_usd,
  v3_core_ssot.safe_div(SUM(riego_usd), SUM(sowed_area_ha)) AS riego_usd_ha,
  SUM(cosecha_usd)::numeric AS cosecha_usd,
  v3_core_ssot.safe_div(SUM(cosecha_usd), SUM(sowed_area_ha)) AS cosecha_usd_ha,
  SUM(otras_labores_usd)::numeric AS otras_labores_usd,
  v3_core_ssot.safe_div(SUM(otras_labores_usd), SUM(sowed_area_ha)) AS otras_labores_usd_ha,
  SUM(total_labores_usd)::numeric AS total_labores_usd,
  v3_core_ssot.safe_div(SUM(total_labores_usd), SUM(sowed_area_ha)) AS total_labores_usd_ha
FROM lot_labores
GROUP BY project_id, field_id, current_crop_id;

-- =============================================================================
-- 4. FIELD_CROP_INSUMOS
-- =============================================================================
CREATE OR REPLACE VIEW v4_report.field_crop_insumos AS
WITH lot_insumos AS (
  SELECT
    f.project_id,
    f.id AS field_id,
    l.current_crop_id,
    l.id AS lot_id,
    v3_lot_ssot.seeded_area_for_lot(l.id) AS sowed_area_ha,
    v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Semilla') AS semillas_usd,
    v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Curasemillas') AS curasemillas_usd,
    v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Herbicidas') AS herbicidas_usd,
    v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Insecticidas') AS insecticidas_usd,
    v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Fungicidas') AS fungicidas_usd,
    v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Coadyuvantes') AS coadyuvantes_usd,
    v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Fertilizantes') AS fertilizantes_usd,
    v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Otros Insumos') AS otros_insumos_usd
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  WHERE l.deleted_at IS NULL AND l.current_crop_id IS NOT NULL
)
SELECT
  project_id,
  field_id,
  current_crop_id,
  SUM(sowed_area_ha)::numeric AS sown_area_ha,
  SUM(semillas_usd)::numeric AS semillas_usd,
  v3_core_ssot.safe_div(SUM(semillas_usd), SUM(sowed_area_ha)) AS semillas_usd_ha,
  SUM(curasemillas_usd)::numeric AS curasemillas_usd,
  v3_core_ssot.safe_div(SUM(curasemillas_usd), SUM(sowed_area_ha)) AS curasemillas_usd_ha,
  SUM(herbicidas_usd)::numeric AS herbicidas_usd,
  v3_core_ssot.safe_div(SUM(herbicidas_usd), SUM(sowed_area_ha)) AS herbicidas_usd_ha,
  SUM(insecticidas_usd)::numeric AS insecticidas_usd,
  v3_core_ssot.safe_div(SUM(insecticidas_usd), SUM(sowed_area_ha)) AS insecticidas_usd_ha,
  SUM(fungicidas_usd)::numeric AS fungicidas_usd,
  v3_core_ssot.safe_div(SUM(fungicidas_usd), SUM(sowed_area_ha)) AS fungicidas_usd_ha,
  SUM(coadyuvantes_usd)::numeric AS coadyuvantes_usd,
  v3_core_ssot.safe_div(SUM(coadyuvantes_usd), SUM(sowed_area_ha)) AS coadyuvantes_usd_ha,
  SUM(fertilizantes_usd)::numeric AS fertilizantes_usd,
  v3_core_ssot.safe_div(SUM(fertilizantes_usd), SUM(sowed_area_ha)) AS fertilizantes_usd_ha,
  SUM(otros_insumos_usd)::numeric AS otros_insumos_usd,
  v3_core_ssot.safe_div(SUM(otros_insumos_usd), SUM(sowed_area_ha)) AS otros_insumos_usd_ha,
  (SUM(semillas_usd) + SUM(curasemillas_usd) + SUM(herbicidas_usd) + SUM(insecticidas_usd) + 
   SUM(fungicidas_usd) + SUM(coadyuvantes_usd) + SUM(fertilizantes_usd) + SUM(otros_insumos_usd))::numeric AS total_insumos_usd,
  v3_core_ssot.safe_div(
    SUM(semillas_usd) + SUM(curasemillas_usd) + SUM(herbicidas_usd) + SUM(insecticidas_usd) + 
    SUM(fungicidas_usd) + SUM(coadyuvantes_usd) + SUM(fertilizantes_usd) + SUM(otros_insumos_usd),
    SUM(sowed_area_ha)
  ) AS total_insumos_usd_ha
FROM lot_insumos
GROUP BY project_id, field_id, current_crop_id;

-- =============================================================================
-- 5. FIELD_CROP_ECONOMICOS (con arriendo TOTAL)
-- =============================================================================
CREATE OR REPLACE VIEW v4_report.field_crop_economicos AS
WITH lot_base AS (
  SELECT
    f.project_id,
    f.id AS field_id,
    l.current_crop_id AS crop_id,
    l.id AS lot_id,
    l.hectares,
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
    SUM(hectares)::numeric AS surface_ha,
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
    -- Arriendo TOTAL (fijo + % ingresos) para mostrar en UI y Resultado Operativo
    SUM(v3_lot_ssot.rent_per_ha_for_lot(lot_id) * sowed_area_ha)::numeric AS rent_total_usd,
    -- Arriendo FIJO para Total Invertido
    SUM(v3_lot_ssot.rent_fixed_only_for_lot(lot_id) * sowed_area_ha)::numeric AS rent_fixed_usd,
    SUM(v3_lot_ssot.admin_cost_per_ha_for_lot(lot_id) * sowed_area_ha)::numeric AS administration_usd
  FROM lot_base
  GROUP BY project_id, field_id, crop_id
)
SELECT
  project_id,
  field_id,
  crop_id AS current_crop_id,
  (labor_costs_usd + supply_costs_usd) AS gastos_directos_usd,
  v3_core_ssot.safe_div((labor_costs_usd + supply_costs_usd), sown_area_ha) AS gastos_directos_usd_ha,
  (
    (production_tn * v3_lot_ssot.net_price_usd_for_lot(sample_lot_id)) -
    (labor_costs_usd + supply_costs_usd)
  ) AS margen_bruto_usd,
  v3_core_ssot.safe_div(
    (production_tn * v3_lot_ssot.net_price_usd_for_lot(sample_lot_id)) - (labor_costs_usd + supply_costs_usd),
    sown_area_ha
  ) AS margen_bruto_usd_ha,
  -- Arriendo TOTAL para mostrar en UI
  rent_total_usd AS arriendo_usd,
  v3_core_ssot.safe_div(rent_total_usd, sown_area_ha) AS arriendo_usd_ha,
  administration_usd AS adm_estructura_usd,
  v3_core_ssot.safe_div(administration_usd, sown_area_ha) AS adm_estructura_usd_ha,
  -- Resultado Operativo usa arriendo TOTAL
  (
    (production_tn * v3_lot_ssot.net_price_usd_for_lot(sample_lot_id)) -
    (labor_costs_usd + supply_costs_usd) - rent_total_usd - administration_usd
  ) AS resultado_operativo_usd,
  v3_core_ssot.safe_div(
    (production_tn * v3_lot_ssot.net_price_usd_for_lot(sample_lot_id)) -
    (labor_costs_usd + supply_costs_usd) - rent_total_usd - administration_usd,
    sown_area_ha
  ) AS resultado_operativo_usd_ha
FROM aggregated;

-- =============================================================================
-- 6. FIELD_CROP_RENTABILIDAD (usa arriendo FIJO para Total Invertido)
-- =============================================================================
CREATE OR REPLACE VIEW v4_report.field_crop_rentabilidad AS
WITH lot_base AS (
  SELECT
    f.project_id,
    f.id AS field_id,
    l.current_crop_id AS crop_id,
    l.id AS lot_id,
    l.hectares,
    l.tons,
    v3_lot_ssot.seeded_area_for_lot(l.id) AS sowed_area_ha
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
    SUM(hectares)::numeric AS surface_ha,
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
    ) AS direct_cost_usd,
    -- Arriendo FIJO para Total Invertido (capitalizable)
    SUM(v3_lot_ssot.rent_fixed_only_for_lot(lot_id) * sowed_area_ha)::numeric AS rent_usd,
    SUM(v3_lot_ssot.admin_cost_per_ha_for_lot(lot_id) * sowed_area_ha)::numeric AS administration_usd
  FROM lot_base
  GROUP BY project_id, field_id, crop_id
)
SELECT
  project_id,
  field_id,
  crop_id AS current_crop_id,
  (direct_cost_usd + rent_usd + administration_usd) AS total_invertido_usd,
  v3_core_ssot.safe_div(direct_cost_usd + rent_usd + administration_usd, sown_area_ha) AS total_invertido_usd_ha,
  v3_lot_ssot.renta_pct(
    (production_tn * v3_lot_ssot.net_price_usd_for_lot(sample_lot_id) - direct_cost_usd - rent_usd - administration_usd)::double precision,
    (direct_cost_usd + rent_usd + administration_usd)::double precision
  ) AS renta_pct,
  v3_core_ssot.safe_div(
    v3_core_ssot.safe_div(direct_cost_usd + rent_usd + administration_usd, sown_area_ha),
    v3_lot_ssot.net_price_usd_for_lot(sample_lot_id)
  ) AS rinde_indiferencia_total_usd_tn
FROM aggregated;

-- =============================================================================
-- 7. FIELD_CROP_METRICS (vista principal con nombres en español)
-- =============================================================================
CREATE OR REPLACE VIEW v4_report.field_crop_metrics AS
SELECT
  c.project_id,
  c.field_id,
  c.field_name,
  c.current_crop_id,
  c.crop_name,
  -- Nombres en español (paridad con modelo Go)
  c.superficie_total AS superficie_ha,
  c.produccion_tn,
  c.superficie_sembrada_ha AS area_sembrada_ha,
  c.area_cosechada_ha,
  c.rendimiento_tn_ha,
  c.precio_bruto_usd_tn,
  c.gasto_flete_usd_tn,
  c.gasto_comercial_usd_tn,
  c.precio_neto_usd_tn,
  c.ingreso_neto_por_ha AS ingreso_neto_usd_ha,
  (c.ingreso_neto_por_ha * c.superficie_sembrada_ha) AS ingreso_neto_usd,
  -- Labores
  COALESCE(lb.total_labores_usd, 0) AS costos_labores_usd,
  COALESCE(lb.total_labores_usd_ha, 0) AS costos_labores_usd_ha,
  -- Insumos
  COALESCE(i.total_insumos_usd, 0) AS costos_insumos_usd,
  COALESCE(i.total_insumos_usd_ha, 0) AS costos_insumos_usd_ha,
  -- Costos directos
  COALESCE(e.gastos_directos_usd, 0) AS total_costos_directos_usd,
  COALESCE(e.gastos_directos_usd_ha, 0) AS costos_directos_usd_ha,
  -- Margen bruto
  COALESCE(e.margen_bruto_usd, 0) AS margen_bruto_usd,
  COALESCE(e.margen_bruto_usd_ha, 0) AS margen_bruto_usd_ha,
  -- Arriendo (TOTAL)
  COALESCE(e.arriendo_usd, 0) AS arriendo_usd,
  COALESCE(e.arriendo_usd_ha, 0) AS arriendo_usd_ha,
  -- Administración
  COALESCE(e.adm_estructura_usd, 0) AS administracion_usd,
  COALESCE(e.adm_estructura_usd_ha, 0) AS administracion_usd_ha,
  -- Resultado operativo
  COALESCE(e.resultado_operativo_usd, 0) AS resultado_operativo_usd,
  COALESCE(e.resultado_operativo_usd_ha, 0) AS resultado_operativo_usd_ha,
  -- Rentabilidad
  COALESCE(r.total_invertido_usd, 0) AS total_invertido_usd,
  COALESCE(r.total_invertido_usd_ha, 0) AS total_invertido_usd_ha,
  COALESCE(r.renta_pct, 0) AS renta_pct,
  COALESCE(r.rinde_indiferencia_total_usd_tn, 0) AS rinde_indiferencia_usd_tn
FROM v4_report.field_crop_cultivos c
LEFT JOIN v4_report.field_crop_labores lb
  ON lb.project_id = c.project_id AND lb.field_id = c.field_id AND lb.current_crop_id = c.current_crop_id
LEFT JOIN v4_report.field_crop_insumos i
  ON i.project_id = c.project_id AND i.field_id = c.field_id AND i.current_crop_id = c.current_crop_id
LEFT JOIN v4_report.field_crop_economicos e
  ON e.project_id = c.project_id AND e.field_id = c.field_id AND e.current_crop_id = c.current_crop_id
LEFT JOIN v4_report.field_crop_rentabilidad r
  ON r.project_id = c.project_id AND r.field_id = c.field_id AND r.current_crop_id = c.current_crop_id;

COMMENT ON VIEW v4_report.field_crop_metrics IS 
'Migration 000320: Recreación completa. Nombres español. arriendo = TOTAL.';
