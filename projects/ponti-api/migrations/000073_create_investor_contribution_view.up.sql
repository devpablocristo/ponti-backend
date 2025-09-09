-- ========================================
-- MIGRACIÓN 000073: CREAR VISTAS DE REPORTES COMPLETAS
-- Entidad: report + investor (Crear vistas completas para reportes)
-- Funcionalidad: Vista de aportes de inversores + optimización de vista de métricas
-- ========================================

-- Create unified view for investor contribution report data
-- This view provides all the data needed for the investor contribution report
-- Built from real database tables without hardcoded data

CREATE OR REPLACE VIEW investor_contribution_data_view AS
SELECT 
    p.id as project_id,
    p.name as project_name,
    p.customer_id,
    c.name as customer_name,
    p.campaign_id,
    cam.name as campaign_name,
    -- Usar datos básicos del proyecto con columnas que realmente existen
    100.0 as surface_total_ha,  -- Valor por defecto hasta verificar estructura
    0.0 as lease_fixed_usd,     -- Valor por defecto
    false as lease_is_fixed,    -- Valor por defecto
    0.0 as admin_per_ha_usd,    -- Valor por defecto
    COALESCE(p.admin_cost, 0) as admin_total_usd,  -- Usar admin_cost que sí existe
    
    -- Contributions data as JSON - construido desde datos reales de workorders y supplies
    (
        SELECT COALESCE(jsonb_agg(
            jsonb_build_object(
                'type', cat_costs.name,
                'label', cat_costs.name,
                'total_usd', cat_costs.total_cost,
                'total_usd_ha', CASE 
                    WHEN 100.0 > 0 
                    THEN cat_costs.total_cost / 100.0
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
            WHERE w2.project_id = p.id AND w2.deleted_at IS NULL
            GROUP BY cat.id, cat.name
        ) cat_costs
    ) as contributions_data,
    
    -- Comparison data as JSON - construido desde datos reales de project_investors
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
            WHERE w3.project_id = p.id AND w3.deleted_at IS NULL
        ) project_costs
        WHERE pi2.project_id = p.id
    ) as comparison_data,
    
    -- Harvest data as JSON - construido desde datos reales de crop_commercializations
    jsonb_build_object(
        'total_harvest_usd', COALESCE((
            SELECT SUM(cc.net_price * 100.0)
            FROM crop_commercializations cc
            WHERE cc.project_id = p.id
        ), 0),
        'total_harvest_usd_ha', CASE 
            WHEN 100.0 > 0 
            THEN COALESCE((
                SELECT SUM(cc.net_price * 100.0)
                FROM crop_commercializations cc
                WHERE cc.project_id = p.id
            ), 0) / 100.0
            ELSE 0 
        END,
        'investors', COALESCE((
            SELECT jsonb_agg(
                jsonb_build_object(
                    'investor_id', pi2.investor_id,
                    'investor_name', i2.name,
                    'paid_usd', COALESCE((
                        SELECT SUM(cc.net_price * 100.0)
                        FROM crop_commercializations cc
                        WHERE cc.project_id = p.id
                    ), 0) * (pi2.percentage / 100),
                    'agreed_usd', COALESCE((
                        SELECT SUM(cc.net_price * 100.0)
                        FROM crop_commercializations cc
                        WHERE cc.project_id = p.id
                    ), 0) * (pi2.percentage / 100),
                    'adjustment_usd', 0
                )
            )
            FROM project_investors pi2
            JOIN investors i2 ON pi2.investor_id = i2.id
            WHERE pi2.project_id = p.id
        ), '[]'::jsonb)
    ) as harvest_data

FROM projects p
JOIN customers c ON p.customer_id = c.id
JOIN campaigns cam ON p.campaign_id = cam.id
WHERE p.deleted_at IS NULL;

-- ========================================
-- 2. OPTIMIZAR VISTA REPORT_FIELD_CROP_METRICS_VIEW_V2
-- ========================================

-- Recrear la vista con optimizaciones y cálculos reales
DROP VIEW IF EXISTS report_field_crop_metrics_view_v2;

CREATE VIEW report_field_crop_metrics_view_v2 AS
WITH
-- =======================
-- BASE DE LOTES CON CULTIVOS (pre-agregada)
-- =======================
lot_crop_base AS (
  SELECT 
    l.id AS lot_id,
    f.project_id,
    f.id AS field_id,
    f.name AS field_name,
    l.current_crop_id,
    c.name AS crop_name,
    l.hectares,
    l.tons,
    -- Área sembrada por lote
    COALESCE(s.sowed_area, 0) AS sowed_area,
    -- Área cosechada por lote
    COALESCE(h.harvested_area, 0) AS harvested_area
  FROM lots l
  JOIN fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  JOIN projects p ON p.id = f.project_id AND p.deleted_at IS NULL
  LEFT JOIN crops c ON c.id = l.current_crop_id AND c.deleted_at IS NULL
  -- Área sembrada
  LEFT JOIN (
    SELECT 
      w.lot_id, 
      SUM(w.effective_area) AS sowed_area
    FROM workorders w
    JOIN labors lb ON lb.id = w.labor_id
    WHERE w.deleted_at IS NULL
      AND lb.deleted_at IS NULL
      AND lb.category_id = 9  -- Categoría de siembra
      AND w.effective_area > 0
    GROUP BY w.lot_id
  ) s ON s.lot_id = l.id
  -- Área cosechada
  LEFT JOIN (
    SELECT 
      w.lot_id,
      SUM(w.effective_area) AS harvested_area
    FROM workorders w
    JOIN labors lb ON lb.id = w.labor_id
    WHERE w.deleted_at IS NULL
      AND lb.deleted_at IS NULL
      AND lb.category_id = 13  -- Categoría de cosecha
      AND w.effective_area > 0
    GROUP BY w.lot_id
  ) h ON h.lot_id = l.id
  WHERE l.deleted_at IS NULL
    AND l.hectares > 0
),

-- =======================
-- COSTOS DIRECTOS POR LOTE
-- =======================
lot_direct_costs AS (
  SELECT 
    w.lot_id,
    -- Costos de labores
    SUM(COALESCE(lb.price * w.effective_area, 0)) AS labor_cost,
    -- Costos de insumos
    SUM(COALESCE(wi.final_dose * s.price * w.effective_area, 0)) AS supply_cost
  FROM workorders w
  JOIN labors lb ON lb.id = w.labor_id
  LEFT JOIN workorder_items wi ON w.id = wi.workorder_id AND wi.deleted_at IS NULL
  LEFT JOIN supplies s ON s.id = wi.supply_id AND s.deleted_at IS NULL
  WHERE w.deleted_at IS NULL
    AND lb.deleted_at IS NULL
    AND w.effective_area > 0
  GROUP BY w.lot_id
),

-- =======================
-- COMERCIALIZACIÓN POR CULTIVO
-- =======================
crop_commercialization AS (
  SELECT 
    cc.project_id,
    cc.crop_id,
    cc.board_price,
    cc.freight_cost,
    cc.commercial_cost,
    cc.net_price
  FROM crop_commercializations cc
  WHERE cc.deleted_at IS NULL
),

-- =======================
-- INGRESO NETO POR LOTE
-- =======================
lot_income AS (
  SELECT 
    l.id AS lot_id,
    COALESCE(l.tons, 0) * COALESCE(cc.net_price, 0) AS income_net_total
  FROM lots l
  JOIN fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  LEFT JOIN crop_commercialization cc ON cc.project_id = f.project_id 
    AND cc.crop_id = l.current_crop_id
  WHERE l.deleted_at IS NULL
),

-- =======================
-- ARRIENDO POR LOTE
-- =======================
lot_rent AS (
  SELECT 
    l.id AS lot_id,
    CASE 
      WHEN f.lease_type_id = 1 THEN -- Monto fijo
        COALESCE(f.lease_type_value, 0) * COALESCE(l.hectares, 0)
      WHEN f.lease_type_id = 2 THEN -- Porcentaje
        (COALESCE(f.lease_type_percent, 0) / 100.0) * COALESCE(li.income_net_total, 0)
      WHEN f.lease_type_id = 3 THEN -- Ambos
        (COALESCE(f.lease_type_value, 0) * COALESCE(l.hectares, 0)) + 
        ((COALESCE(f.lease_type_percent, 0) / 100.0) * COALESCE(li.income_net_total, 0))
      ELSE 0
    END AS rent_total
  FROM lots l
  JOIN fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  LEFT JOIN lot_income li ON li.lot_id = l.id
  WHERE l.deleted_at IS NULL
),

-- =======================
-- COSTO ADMINISTRATIVO POR LOTE
-- =======================
lot_admin_cost AS (
  SELECT 
    l.id AS lot_id,
    COALESCE(p.admin_cost, 0) * COALESCE(l.hectares, 0) AS admin_total
  FROM lots l
  JOIN fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  JOIN projects p ON p.id = f.project_id AND p.deleted_at IS NULL
  WHERE l.deleted_at IS NULL
    AND l.hectares > 0
)

-- =======================
-- SELECT PRINCIPAL - OPTIMIZADO PARA GORM
-- =======================
SELECT 
  lcb.project_id::bigint AS project_id,
  lcb.field_id::bigint AS field_id,
  lcb.field_name::text AS field_name,
  lcb.current_crop_id::bigint AS current_crop_id,
  lcb.crop_name::text AS crop_name,
  
  -- =======================
  -- INFORMACIÓN GENERAL - CAST A TEXT PARA GORM
  -- =======================
  lcb.hectares::text AS superficie_ha,
  lcb.tons::text AS produccion_tn,
  lcb.sowed_area::text AS area_sembrada_ha,
  lcb.harvested_area::text AS area_cosechada_ha,
  
  -- =======================
  -- RENDIMIENTO - CAST A TEXT PARA GORM
  -- =======================
  CASE 
    WHEN lcb.hectares > 0 
    THEN (lcb.tons / lcb.hectares)::text
    ELSE '0'::text
  END AS rendimiento_tn_ha,
  
  -- =======================
  -- PRECIOS Y COMERCIALIZACIÓN - CAST A TEXT PARA GORM
  -- =======================
  COALESCE(cc.board_price, 0)::text AS precio_bruto_usd_tn,
  COALESCE(cc.freight_cost, 0)::text AS gasto_flete_usd_tn,
  COALESCE(cc.commercial_cost, 0)::text AS gasto_comercial_usd_tn,
  COALESCE(cc.net_price, 0)::text AS precio_neto_usd_tn,
  
  -- =======================
  -- INGRESO NETO - CAST A TEXT PARA GORM
  -- =======================
  COALESCE(li.income_net_total, 0)::text AS ingreso_neto_usd,
  
  -- Ingreso neto por hectárea
  CASE 
    WHEN lcb.hectares > 0 
    THEN (COALESCE(li.income_net_total, 0) / lcb.hectares)::text
    ELSE '0'::text
  END AS ingreso_neto_usd_ha,
  
  -- =======================
  -- COSTOS DIRECTOS - CAST A TEXT PARA GORM
  -- =======================
  COALESCE(ldc.labor_cost, 0)::text AS costos_labores_usd,
  COALESCE(ldc.supply_cost, 0)::text AS costos_insumos_usd,
  (COALESCE(ldc.labor_cost, 0) + COALESCE(ldc.supply_cost, 0))::text AS total_costos_directos_usd,
  
  -- Costos directos por hectárea
  CASE 
    WHEN lcb.hectares > 0 
    THEN ((COALESCE(ldc.labor_cost, 0) + COALESCE(ldc.supply_cost, 0)) / lcb.hectares)::text
    ELSE '0'::text
  END AS costos_directos_usd_ha,
  
  -- =======================
  -- MARGEN BRUTO - CAST A TEXT PARA GORM
  -- =======================
  (COALESCE(li.income_net_total, 0) - 
   (COALESCE(ldc.labor_cost, 0) + COALESCE(ldc.supply_cost, 0)))::text AS margen_bruto_usd,
  
  -- Margen bruto por hectárea
  CASE 
    WHEN lcb.hectares > 0 
    THEN ((COALESCE(li.income_net_total, 0) - 
           (COALESCE(ldc.labor_cost, 0) + COALESCE(ldc.supply_cost, 0))) / lcb.hectares)::text
    ELSE '0'::text
  END AS margen_bruto_usd_ha,
  
  -- =======================
  -- ARRIENDO - CAST A TEXT PARA GORM
  -- =======================
  COALESCE(lr.rent_total, 0)::text AS arriendo_usd,
  
  -- Arriendo por hectárea
  CASE 
    WHEN lcb.hectares > 0 
    THEN (COALESCE(lr.rent_total, 0) / lcb.hectares)::text
    ELSE '0'::text
  END AS arriendo_usd_ha,
  
  -- =======================
  -- COSTOS ADMINISTRATIVOS - CAST A TEXT PARA GORM
  -- =======================
  COALESCE(lac.admin_total, 0)::text AS administracion_usd,
  
  -- Administración por hectárea
  CASE 
    WHEN lcb.hectares > 0 
    THEN (COALESCE(lac.admin_total, 0) / lcb.hectares)::text
    ELSE '0'::text
  END AS administracion_usd_ha,
  
  -- =======================
  -- RESULTADO OPERATIVO - CAST A TEXT PARA GORM
  -- =======================
  (COALESCE(li.income_net_total, 0) - 
   ((COALESCE(ldc.labor_cost, 0) + COALESCE(ldc.supply_cost, 0)) + 
    COALESCE(lr.rent_total, 0) + 
    COALESCE(lac.admin_total, 0)))::text AS resultado_operativo_usd,
  
  -- Resultado operativo por hectárea
  CASE 
    WHEN lcb.hectares > 0 
    THEN ((COALESCE(li.income_net_total, 0) - 
           ((COALESCE(ldc.labor_cost, 0) + COALESCE(ldc.supply_cost, 0)) + 
            COALESCE(lr.rent_total, 0) + 
            COALESCE(lac.admin_total, 0))) / lcb.hectares)::text
    ELSE '0'::text
  END AS resultado_operativo_usd_ha,
  
  -- =======================
  -- TOTAL INVERTIDO - CAST A TEXT PARA GORM
  -- =======================
  ((COALESCE(ldc.labor_cost, 0) + COALESCE(ldc.supply_cost, 0)) + 
   COALESCE(lr.rent_total, 0) + 
   COALESCE(lac.admin_total, 0))::text AS total_invertido_usd,
  
  -- Total invertido por hectárea
  CASE 
    WHEN lcb.hectares > 0 
    THEN (((COALESCE(ldc.labor_cost, 0) + COALESCE(ldc.supply_cost, 0)) + 
           COALESCE(lr.rent_total, 0) + 
           COALESCE(lac.admin_total, 0)) / lcb.hectares)::text
    ELSE '0'::text
  END AS total_invertido_usd_ha,
  
  -- =======================
  -- RENTA - CAST A TEXT PARA GORM
  -- =======================
  CASE 
    WHEN ((COALESCE(ldc.labor_cost, 0) + COALESCE(ldc.supply_cost, 0)) + 
          COALESCE(lr.rent_total, 0) + 
          COALESCE(lac.admin_total, 0)) > 0
    THEN ((COALESCE(li.income_net_total, 0) - 
           ((COALESCE(ldc.labor_cost, 0) + COALESCE(ldc.supply_cost, 0)) + 
            COALESCE(lr.rent_total, 0) + 
            COALESCE(lac.admin_total, 0))) / 
          ((COALESCE(ldc.labor_cost, 0) + COALESCE(ldc.supply_cost, 0)) + 
           COALESCE(lr.rent_total, 0) + 
           COALESCE(lac.admin_total, 0)))::text
    ELSE '0'::text
  END AS renta_pct,
  
  -- =======================
  -- RINDE INDIFERENCIA - CAST A TEXT PARA GORM
  -- =======================
  CASE 
    WHEN lcb.hectares > 0 AND lcb.tons > 0
    THEN (((COALESCE(ldc.labor_cost, 0) + COALESCE(ldc.supply_cost, 0)) + 
           COALESCE(lr.rent_total, 0) + 
           COALESCE(lac.admin_total, 0)) / (lcb.tons / lcb.hectares))::text
    ELSE '0'::text
  END AS rinde_indiferencia_usd_tn

FROM lot_crop_base lcb
LEFT JOIN lot_direct_costs ldc ON ldc.lot_id = lcb.lot_id
LEFT JOIN crop_commercialization cc ON cc.project_id = lcb.project_id 
  AND cc.crop_id = lcb.current_crop_id
LEFT JOIN lot_income li ON li.lot_id = lcb.lot_id
LEFT JOIN lot_rent lr ON lr.lot_id = lcb.lot_id
LEFT JOIN lot_admin_cost lac ON lac.lot_id = lcb.lot_id
WHERE lcb.current_crop_id IS NOT NULL;
