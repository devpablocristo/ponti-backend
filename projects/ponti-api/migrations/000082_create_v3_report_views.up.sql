-- ========================================
-- MIGRATION 000082: CREATE v3_report_views (UP)
-- ========================================
-- 
-- Purpose: Create report field crop metrics view
-- Date: 2025-09-12
-- Author: System
-- 
-- Note: Code in English, comments in Spanish.

-- -------------------------------------------------------------------
-- v3_report_field_crop_metrics_view: métricas por field y cultivo
-- -------------------------------------------------------------------
CREATE OR REPLACE VIEW public.v3_report_field_crop_metrics_view AS
WITH lot_base AS (
  SELECT
    l.id AS lot_id,
    f.project_id,
    f.id AS field_id,
    f.name AS field_name,
    l.current_crop_id,
    c.name AS crop_name,
    l.hectares,
    COALESCE(l.tons, 0)::numeric AS tons,
    v3_calc.seeded_area(l.sowing_date, l.hectares::numeric)::double precision AS seeded_area_ha,
    v3_calc.harvested_area(l.tons::numeric, l.hectares::numeric)::double precision AS harvested_area_ha
  FROM public.lots l
  JOIN public.fields   f ON f.id = l.field_id AND f.deleted_at IS NULL
  JOIN public.projects p ON p.id = f.project_id AND p.deleted_at IS NULL
  LEFT JOIN public.crops c ON c.id = l.current_crop_id AND c.deleted_at IS NULL
  WHERE l.deleted_at IS NULL AND l.hectares > 0
)
SELECT
  lb.project_id,
  lb.field_id,
  lb.field_name::text AS field_name,
  lb.current_crop_id,
  lb.crop_name::text  AS crop_name,

  COALESCE(SUM(v3_calc.income_net_total_for_lot(lb.lot_id)), 0)             AS income_usd,

  -- Costos directos ejecutados (labores + insumos)
  COALESCE(SUM(v3_calc.direct_cost_for_lot(lb.lot_id)), 0)                  AS direct_costs_executed_usd,

  -- Costos directos invertidos (labores+insumos+arriendo+estructura)
  (
    COALESCE(SUM(v3_calc.direct_cost_for_lot(lb.lot_id)), 0)::double precision
    + COALESCE(SUM(v3_calc.rent_per_ha_for_lot(lb.lot_id) * lb.hectares), 0)::double precision
    + COALESCE(SUM(v3_calc.admin_cost_per_ha_for_lot(lb.lot_id) * lb.hectares), 0)::double precision
  )                                                                        AS direct_costs_invested_usd,

  -- Componentes de invertidos
  COALESCE(SUM(v3_calc.rent_per_ha_for_lot(lb.lot_id) * lb.hectares), 0)     AS rent_invested_usd,
  COALESCE(SUM(v3_calc.admin_cost_per_ha_for_lot(lb.lot_id) * lb.hectares), 0) AS structure_invested_usd,

  -- Resultado operativo y ratio
  COALESCE(SUM(v3_calc.operating_result_per_ha_for_lot(lb.lot_id) * lb.hectares), 0) AS operating_result_usd,
  v3_calc.renta_pct(
    COALESCE(SUM(v3_calc.operating_result_per_ha_for_lot(lb.lot_id) * lb.hectares), 0)::double precision,
    COALESCE(SUM(v3_calc.direct_cost_for_lot(lb.lot_id)), 0)::double precision
  )                                                                          AS operating_result_pct
FROM lot_base lb
LEFT JOIN public.crop_commercializations cc
  ON cc.project_id = lb.project_id
 AND cc.crop_id   = lb.current_crop_id
 AND cc.deleted_at IS NULL
WHERE lb.current_crop_id IS NOT NULL
GROUP BY lb.project_id, lb.field_id, lb.field_name, lb.current_crop_id, lb.crop_name;

-- -------------------------------------------------------------------
-- v3_report_summary_results_view: resumen de resultados generales por cultivo
-- -------------------------------------------------------------------
CREATE OR REPLACE VIEW public.v3_report_summary_results_view AS
WITH lot_base AS (
  SELECT
    l.id AS lot_id,
    f.project_id,
    l.current_crop_id,
    c.name AS crop_name,
    l.hectares,
    COALESCE(l.tons, 0)::numeric AS tons,
    v3_calc.seeded_area(l.sowing_date, l.hectares::numeric)::double precision AS seeded_area_ha
  FROM public.lots l
  JOIN public.fields   f ON f.id = l.field_id AND f.deleted_at IS NULL
  JOIN public.projects p ON p.id = f.project_id AND p.deleted_at IS NULL
  LEFT JOIN public.crops c ON c.id = l.current_crop_id AND c.deleted_at IS NULL
  WHERE l.deleted_at IS NULL AND l.hectares > 0
),
by_crop AS (
  SELECT
    lb.project_id,
    lb.current_crop_id,
    lb.crop_name::text AS crop_name,
    
    -- Superficie total por cultivo
    COALESCE(SUM(lb.seeded_area_ha), 0)::numeric AS surface_ha,
    
    -- Ingreso neto total por cultivo
    COALESCE(SUM(v3_calc.income_net_total_for_lot(lb.lot_id)), 0)::numeric AS net_income_usd,
    
    -- Costos directos totales por cultivo (labores + insumos)
    COALESCE(SUM(v3_calc.direct_cost_for_lot(lb.lot_id)), 0)::numeric AS direct_costs_usd,
    
    -- Arriendo total por cultivo
    COALESCE(SUM(v3_calc.rent_per_ha_for_lot(lb.lot_id) * lb.hectares), 0)::numeric AS rent_usd,
    
    -- Estructura total por cultivo
    COALESCE(SUM(v3_calc.admin_cost_per_ha_for_lot(lb.lot_id) * lb.hectares), 0)::numeric AS structure_usd,
    
    -- Total invertido por cultivo
    (
      COALESCE(SUM(v3_calc.direct_cost_for_lot(lb.lot_id)), 0)::numeric
      + COALESCE(SUM(v3_calc.rent_per_ha_for_lot(lb.lot_id) * lb.hectares), 0)::numeric
      + COALESCE(SUM(v3_calc.admin_cost_per_ha_for_lot(lb.lot_id) * lb.hectares), 0)::numeric
    ) AS total_invested_usd,
    
    -- Resultado operativo total por cultivo
    COALESCE(SUM(v3_calc.operating_result_per_ha_for_lot(lb.lot_id) * lb.hectares), 0)::numeric AS operating_result_usd
    
  FROM lot_base lb
  WHERE lb.current_crop_id IS NOT NULL
  GROUP BY lb.project_id, lb.current_crop_id, lb.crop_name
),
project_totals AS (
  SELECT
    project_id,
    SUM(surface_ha)::numeric AS total_surface_ha,
    SUM(net_income_usd)::numeric AS total_net_income_usd,
    SUM(direct_costs_usd)::numeric AS total_direct_costs_usd,
    SUM(rent_usd)::numeric AS total_rent_usd,
    SUM(structure_usd)::numeric AS total_structure_usd,
    SUM(total_invested_usd)::numeric AS total_invested_usd,
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
  
  -- Renta del cultivo (%)
  v3_calc.renta_pct(
    bc.operating_result_usd::double precision,
    bc.total_invested_usd::double precision
  )::numeric AS crop_return_pct,
  
  -- Totales del proyecto (para comparación)
  pt.total_surface_ha,
  pt.total_net_income_usd,
  pt.total_direct_costs_usd,
  pt.total_rent_usd,
  pt.total_structure_usd,
  pt.total_invested_usd AS total_invested_project_usd,
  pt.total_operating_result_usd,
  
  -- Renta total del proyecto (%)
  v3_calc.renta_pct(
    pt.total_operating_result_usd::double precision,
    pt.total_invested_usd::double precision
  )::numeric AS project_return_pct

FROM by_crop bc
JOIN project_totals pt ON pt.project_id = bc.project_id
ORDER BY bc.project_id, bc.current_crop_id;

-- -------------------------------------------------------------------
-- v3_investor_contribution_data_view: datos de contribución de inversores v3
-- -------------------------------------------------------------------
CREATE OR REPLACE VIEW public.v3_investor_contribution_data_view AS
WITH project_base AS (
  SELECT
    p.id AS project_id,
    p.name AS project_name,
    p.customer_id,
    c.name AS customer_name,
    p.campaign_id,
    cam.name AS campaign_name,
    -- Calcular superficie total usando funciones SSOT
    COALESCE(SUM(l.hectares), 0)::numeric AS surface_total_ha,
    -- Datos de arriendo usando funciones SSOT
    COALESCE(SUM(v3_calc.rent_per_ha_for_lot(l.id) * l.hectares), 0)::numeric AS lease_fixed_usd,
    false AS lease_is_fixed, -- Por defecto, se puede calcular dinámicamente
    -- Datos de administración usando funciones SSOT
    COALESCE(SUM(v3_calc.admin_cost_per_ha_for_lot(l.id) * l.hectares), 0)::numeric AS admin_per_ha_usd,
    COALESCE(SUM(v3_calc.admin_cost_per_ha_for_lot(l.id) * l.hectares), 0)::numeric AS admin_total_usd
  FROM public.projects p
  JOIN public.customers c ON p.customer_id = c.id AND c.deleted_at IS NULL
  JOIN public.campaigns cam ON p.campaign_id = cam.id AND cam.deleted_at IS NULL
  LEFT JOIN public.fields f ON f.project_id = p.id AND f.deleted_at IS NULL
  LEFT JOIN public.lots l ON l.field_id = f.id AND l.deleted_at IS NULL
  WHERE p.deleted_at IS NULL
  GROUP BY p.id, p.name, p.customer_id, c.name, p.campaign_id, cam.name
),
contributions_data AS (
  SELECT
    pb.project_id,
    -- Contributions data as JSON - construido desde datos reales usando funciones SSOT
    (
      SELECT COALESCE(jsonb_agg(
        jsonb_build_object(
          'type', cat_costs.name,
          'label', cat_costs.name,
          'total_usd', cat_costs.total_cost,
          'total_usd_ha', CASE 
            WHEN pb.surface_total_ha > 0 
            THEN cat_costs.total_cost / pb.surface_total_ha
            ELSE 0 
          END,
          'investors', '[]'::jsonb,
          'requires_manual_attribution', false
        )
      ), '[]'::jsonb)
      FROM (
        SELECT cat.name, SUM(wi.total_used * s.price) as total_cost
        FROM workorders w2
        JOIN workorder_items wi ON w2.id = wi.workorder_id
        JOIN supplies s ON wi.supply_id = s.id AND s.deleted_at IS NULL
        JOIN categories cat ON s.category_id = cat.id
        WHERE w2.project_id = pb.project_id AND w2.deleted_at IS NULL
        GROUP BY cat.id, cat.name
      ) cat_costs
    ) as contributions_data
  FROM project_base pb
),
comparison_data AS (
  SELECT
    pb.project_id,
    -- Comparison data as JSON - construido desde datos reales
    (
      SELECT COALESCE(jsonb_agg(
        jsonb_build_object(
          'investor_id', pi2.investor_id,
          'investor_name', i2.name,
          'agreed_share_pct', pi2.percentage,
          'agreed_usd', total_project_cost * (pi2.percentage / 100),
          'actual_usd', total_project_cost * (pi2.percentage / 100),
          'adjustment_usd', 0
        )
      ), '[]'::jsonb)
      FROM project_investors pi2
      JOIN investors i2 ON pi2.investor_id = i2.id
      CROSS JOIN (
        SELECT COALESCE(SUM(wi.total_used * s.price), 0) as total_project_cost
        FROM workorders w3
        JOIN workorder_items wi ON w3.id = wi.workorder_id
        JOIN supplies s ON wi.supply_id = s.id AND s.deleted_at IS NULL
        WHERE w3.project_id = pb.project_id AND w3.deleted_at IS NULL
      ) project_costs
      WHERE pi2.project_id = pb.project_id
    ) as comparison_data
  FROM project_base pb
),
harvest_data AS (
  SELECT
    pb.project_id,
    -- Harvest data as JSON - construido desde datos reales usando funciones SSOT
    jsonb_build_object(
      'total_harvest_usd', COALESCE((
        SELECT SUM(v3_calc.income_net_total_for_lot(l.id))
        FROM lots l
        JOIN fields f ON l.field_id = f.id
        WHERE f.project_id = pb.project_id AND l.deleted_at IS NULL
      ), 0),
      'total_harvest_usd_ha', CASE 
        WHEN pb.surface_total_ha > 0 
        THEN COALESCE((
          SELECT SUM(v3_calc.income_net_total_for_lot(l.id))
          FROM lots l
          JOIN fields f ON l.field_id = f.id
          WHERE f.project_id = pb.project_id AND l.deleted_at IS NULL
        ), 0) / pb.surface_total_ha
        ELSE 0 
      END,
      'investors', COALESCE((
        SELECT jsonb_agg(
          jsonb_build_object(
            'investor_id', pi2.investor_id,
            'investor_name', i2.name,
            'paid_usd', COALESCE((
              SELECT SUM(v3_calc.income_net_total_for_lot(l.id))
              FROM lots l
              JOIN fields f ON l.field_id = f.id
              WHERE f.project_id = pb.project_id AND l.deleted_at IS NULL
            ), 0) * (pi2.percentage / 100),
            'agreed_usd', COALESCE((
              SELECT SUM(v3_calc.income_net_total_for_lot(l.id))
              FROM lots l
              JOIN fields f ON l.field_id = f.id
              WHERE f.project_id = pb.project_id AND l.deleted_at IS NULL
            ), 0) * (pi2.percentage / 100),
            'adjustment_usd', 0
          )
        )
        FROM project_investors pi2
        JOIN investors i2 ON pi2.investor_id = i2.id
        WHERE pi2.project_id = pb.project_id
      ), '[]'::jsonb)
    ) as harvest_data
  FROM project_base pb
)
SELECT 
  pb.project_id,
  pb.project_name,
  pb.customer_id,
  pb.customer_name,
  pb.campaign_id,
  pb.campaign_name,
  pb.surface_total_ha,
  pb.lease_fixed_usd,
  pb.lease_is_fixed,
  pb.admin_per_ha_usd,
  pb.admin_total_usd,
  cd.contributions_data,
  compd.comparison_data,
  hd.harvest_data
FROM project_base pb
LEFT JOIN contributions_data cd ON cd.project_id = pb.project_id
LEFT JOIN comparison_data compd ON compd.project_id = pb.project_id
LEFT JOIN harvest_data hd ON hd.project_id = pb.project_id;
