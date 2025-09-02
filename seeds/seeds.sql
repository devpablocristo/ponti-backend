ROLLBACK;

-- ========================================
-- CLEAN
-- ========================================
TRUNCATE workorder_items RESTART IDENTITY CASCADE;
TRUNCATE invoices RESTART IDENTITY CASCADE;
TRUNCATE workorders RESTART IDENTITY CASCADE;
TRUNCATE labors RESTART IDENTITY CASCADE;
TRUNCATE supplies RESTART IDENTITY CASCADE;
TRUNCATE lots RESTART IDENTITY CASCADE;
TRUNCATE fields RESTART IDENTITY CASCADE;
TRUNCATE projects RESTART IDENTITY CASCADE;
TRUNCATE customers RESTART IDENTITY CASCADE;
TRUNCATE investors RESTART IDENTITY CASCADE;
TRUNCATE project_investors RESTART IDENTITY CASCADE;

-- ========================================
-- DATOS BASE SIMPLES
-- ========================================
INSERT INTO customers (id, name) VALUES 
  (1, 'Cliente Demo');

INSERT INTO investors (id, name) VALUES 
  (1, 'Inversor 1'),
  (2, 'Inversor 2'),
  (3, 'Inversor 3'),
  (4, 'Inversor 4'),
  (5, 'Inversor 5');

-- ========================================
-- TRES PROYECTOS CON DIFERENTES ESCENARIOS
-- ========================================
INSERT INTO projects (id, customer_id, campaign_id, name, admin_cost) VALUES
  (1, 1, 1, 'Proyecto A - Parcial', 1000),      -- 50% sembrado, 100% cosechado
  (2, 1, 1, 'Proyecto B - Completo', 500),      -- 100% sembrado y cosechado
  (3, 1, 1, 'Proyecto C - Sin Siembra', 750);   -- 0% sembrado, 0% cosechado

-- ========================================
-- RELACIÓN PROYECTO-INVERSOR (CORREGIDO)
-- ========================================
INSERT INTO project_investors (project_id, investor_id, percentage) VALUES
  -- Proyecto 1: 3 inversores (40% + 35% + 25% = 100%)
  (1, 1, 40.00),
  (1, 2, 35.00),
  (1, 3, 25.00),
  
  -- Proyecto 2: 2 inversores (60% + 40% = 100%)
  (2, 4, 60.00),
  (2, 5, 40.00),
  
  -- Proyecto 3: 1 inversor (100%)
  (3, 1, 100.00);

-- ========================================
-- TRES CAMPOS CON DIFERENTES ESCENARIOS
-- ========================================
INSERT INTO fields (id, project_id, name, lease_type_id) VALUES
  (101, 1, 'Campo A - Parcial', 1),    -- ID 1 = % INGRESO NETO
  (102, 2, 'Campo B - Completo', 2),   -- ID 2 = % UTILIDAD
  (103, 3, 'Campo C - Vacío', 3);      -- ID 3 = ARRIENDO FIJO

-- ========================================
-- SEIS LOTES CON DIFERENTES ESTADOS
-- ========================================
INSERT INTO lots (id, field_id, name, hectares, previous_crop_id, current_crop_id, season, tons, sowing_date) VALUES
  -- Proyecto 1: Campo A - Parcial (100 ha sembradas de 200 ha totales)
  (1001, 101, 'Lote A1', 100, 1, 2, '2024-2025', 200, '2024-10-15'),   -- 100 ha sembradas
  (1002, 101, 'Lote A2', 100, 2, 1, '2024-2025', 200, NULL),           -- 100 ha NO sembradas
  
  -- Proyecto 2: Campo B - Completo (150 ha sembradas de 150 ha totales)
  (1003, 102, 'Lote B1', 75, 1, 3, '2024-2025', 150, '2024-06-01'),    -- 75 ha sembradas
  (1004, 102, 'Lote B2', 75, 3, 1, '2024-2025', 150, '2024-06-05'),    -- 75 ha sembradas
  
  -- Proyecto 3: Campo C - Vacío (0 ha sembradas de 100 ha totales)
  (1005, 103, 'Lote C1', 50, 1, 2, '2024-2025', 100, NULL),            -- 50 ha NO sembradas
  (1006, 103, 'Lote C2', 50, 2, 1, '2024-2025', 100, NULL);            -- 50 ha NO sembradas

-- ========================================
-- SEIS LABORES CON PRECIOS REDONDOS
-- ========================================
INSERT INTO labors (id, project_id, name, category_id, price, contractor_name) VALUES
  -- Proyecto 1
  (1, 1, 'Siembra', 9, 50, 'Contratista 1'),      -- ID 9 = Siembra
  (2, 1, 'Cosecha', 13, 100, 'Contratista 2'),    -- ID 13 = Cosecha
  (3, 1, 'Fertilización', 8, 75, 'Contratista 3'), -- ID 8 = Fertilizantes
  
  -- Proyecto 2
  (4, 2, 'Siembra', 9, 50, 'Contratista 1'),
  (5, 2, 'Cosecha', 13, 100, 'Contratista 2'),
  (6, 2, 'Fertilización', 8, 75, 'Contratista 3');

-- ========================================
-- SEIS INSUMOS CON PRECIOS REDONDOS
-- ========================================
INSERT INTO supplies (id, project_id, name, type_id, category_id, price) VALUES
  -- Proyecto 1
  (1, 1, 'Fertilizante', 3, 8, 2),           -- ID 3 = Fertilizantes, ID 8 = Fertilizantes
  (2, 1, 'Semilla', 1, 1, 10),               -- ID 1 = Semillas, ID 1 = Semilla
  
  -- Proyecto 2
  (3, 2, 'Fertilizante', 3, 8, 2),
  (4, 2, 'Semilla', 1, 1, 10),
  
  -- Proyecto 3
  (5, 3, 'Fertilizante', 3, 8, 2),
  (6, 3, 'Semilla', 1, 1, 10);

-- ========================================
-- WORKORDERS CON DIFERENTES ESCENARIOS
-- ========================================
INSERT INTO workorders (id, project_id, field_id, lot_id, crop_id, labor_id, investor_id, date, effective_area) VALUES
  -- Proyecto 1: Campo A - Parcial (100 ha de 200 ha)
  (1, 1, 101, 1001, 2, 1, 1, '2024-10-15', 100),   -- Siembra Lote A1 - 100 ha × $50 = $5,000
  (2, 1, 101, 1001, 2, 3, 1, '2024-11-10', 100),   -- Fertilización Lote A1 - 100 ha × $75 = $7,500
  (3, 1, 101, 1001, 2, 2, 1, '2025-03-20', 100),   -- Cosecha Lote A1 - 100 ha × $100 = $10,000
  
  -- Proyecto 2: Campo B - Completo (150 ha de 150 ha)
  (4, 2, 102, 1003, 3, 4, 1, '2024-06-01', 75),    -- Siembra Lote B1 - 75 ha × $50 = $3,750
  (5, 2, 102, 1003, 3, 6, 1, '2024-07-01', 75),    -- Fertilización Lote B1 - 75 ha × $75 = $5,625
  (6, 2, 102, 1003, 3, 5, 1, '2024-12-15', 75),    -- Cosecha Lote B1 - 75 ha × $100 = $7,500
  
  (7, 2, 102, 1004, 1, 4, 1, '2024-06-05', 75),    -- Siembra Lote B2 - 75 ha × $50 = $3,750
  (8, 2, 102, 1004, 1, 6, 1, '2024-07-05', 75),    -- Fertilización Lote B2 - 75 ha × $75 = $5,625
  (9, 2, 102, 1004, 1, 5, 1, '2024-12-20', 75);    -- Cosecha Lote B2 - 75 ha × $100 = $7,500

-- ========================================
-- WORKORDER_ITEMS CON NÚMEROS REDONDOS
-- ========================================
INSERT INTO workorder_items (workorder_id, supply_id, final_dose, total_used) VALUES
  -- Proyecto 1: Fertilización
  (2, 1, 100, 10000),    -- Workorder 2: 100 kg/ha × 100 ha = 10,000 kg × $2 = $20,000
  
  -- Proyecto 1: Semillas
  (1, 2, 50, 5000),      -- Workorder 1: 50 kg/ha × 100 ha = 5,000 kg × $10 = $50,000
  
  -- Proyecto 2: Fertilización
  (5, 3, 100, 7500),     -- Workorder 5: 100 kg/ha × 75 ha = 7,500 kg × $2 = $15,000
  (8, 3, 100, 7500),     -- Workorder 8: 100 kg/ha × 75 ha = 7,500 kg × $2 = $15,000
  
  -- Proyecto 2: Semillas
  (4, 4, 50, 3750),      -- Workorder 4: 50 kg/ha × 75 ha = 3,750 kg × $10 = $37,500
  (7, 4, 50, 3750);      -- Workorder 7: 50 kg/ha × 75 ha = 3,750 kg × $10 = $37,500

-- ========================================
-- VER TODAS LAS COLUMNAS DISPONIBLES
-- ========================================
SELECT '=== VER TODAS LAS COLUMNAS DISPONIBLES ===' as info;
SELECT * FROM dashboard_sowing_progress_view WHERE project_id = 1 LIMIT 1;

-- ========================================
-- MÓDULO 1: AVANCE DE SIEMBRA
-- ========================================
-- RESULTADOS REALES:
-- Proyecto 1: 100 ha sembradas / 200 ha totales = 50.00%
-- Proyecto 2: 150 ha sembradas / 150 ha totales = 100.00%
-- Proyecto 3: 0 ha sembradas / 100 ha totales = 0.00%
SELECT '=== MÓDULO 1: AVANCE DE SIEMBRA ===' as info;
SELECT 
  project_id, 
  sowing_hectares, 
  sowing_total_hectares, 
  sowing_progress_pct
FROM dashboard_sowing_progress_view
WHERE project_id IN (1)
ORDER BY project_id;

-- ========================================
-- MÓDULO 2: AVANCE DE COSTOS
-- ========================================
-- RESULTADOS REALES:
-- Proyecto 1: $92,500 ejecutados (labores $22,500 + insumos $70,000)
-- Proyecto 2: $138,750 ejecutados (labores $33,750 + insumos $105,000)
-- Proyecto 3: $0 ejecutados (sin workorders)
SELECT '=== MÓDULO 2: AVANCE DE COSTOS ===' as info;
SELECT 
  project_id, 
  executed_supplies_usd,
  executed_costs_usd,
  costs_progress_pct
FROM dashboard_costs_progress_view
WHERE project_id IN (1)
ORDER BY project_id;

-- ========================================
-- MÓDULO 3: AVANCE DE COSECHA
-- ========================================
-- RESULTADOS REALES:
-- Proyecto 1: 100 ha cosechadas / 200 ha totales = 50.00%
-- Proyecto 2: 150 ha cosechadas / 150 ha totales = 100.00%
-- Proyecto 3: 0 ha cosechadas / 100 ha totales = 0.00%
SELECT '=== MÓDULO 3: AVANCE DE COSECHA ===' as info;
SELECT 
  project_id, 
  harvest_hectares, 
  harvest_total_hectares, 
  harvest_progress_pct
FROM dashboard_harvest_progress_view
WHERE project_id IN (1,2,3)
ORDER BY project_id;

-- ========================================
-- MÓDULO 4: RESULTADO OPERATIVO
-- ========================================
-- RESULTADOS REALES:
-- Proyecto 1: $0 ingresos - $92,500 costos = -$92,500 (-100%)
-- Proyecto 2: $0 ingresos - $138,750 costos = -$138,750 (-100%)
-- Proyecto 3: $0 ingresos - $0 costos = $0 (0%)
SELECT '=== MÓDULO 4: RESULTADO OPERATIVO ===' as info;
SELECT 
  project_id, 
  income_usd,
  operating_result_usd,
  operating_result_total_costs_usd,
  operating_result_pct
FROM dashboard_operating_result_view
WHERE project_id IN (1,2,3)
ORDER BY project_id;

-- ========================================
-- MÓDULO 5: AVANCE DE APORTES
-- ========================================
-- RESULTADOS REALES:
-- Proyecto 1: 3 inversores (40% + 35% + 25% = 100%)
-- Proyecto 2: 2 inversores (60% + 40% = 100%)
-- Proyecto 3: 1 inversor (100%)
SELECT '=== MÓDULO 5: AVANCE DE APORTES ===' as info;
SELECT 
  project_id, 
  investor_id,
  investor_name,
  investor_percentage_pct,
  contributions_progress_pct
FROM dashboard_contributions_progress_view
ORDER BY project_id;

-- ========================================
-- MÓDULO 6: BALANCE DE GESTIÓN
-- ========================================
-- RESULTADOS REALES:
-- Proyecto 1: Semillas $50,000, Insumos $20,000, Labores $22,500 = $92,500 total + $1,000 estructura = $93,500
-- Proyecto 2: Semillas $75,000, Insumos $30,000, Labores $33,750 = $138,750 total + $500 estructura = $139,250  
-- Proyecto 3: Todo en $0 + $750 estructura = $750
SELECT '=== MÓDULO 6: BALANCE DE GESTIÓN ===' as info;
SELECT 
  project_id,
  -- SEMILLA
  seeds_executed_usd,             -- Semilla Ejecutados
  seeds_invested_usd,             -- Semilla Invertidos  
  seeds_stock_usd,                -- Semilla Stock (Invertidos - Ejecutados)
  -- INSUMOS
  supplies_executed_usd,          -- Insumos Ejecutados (no semilla)
  supplies_invested_usd,          -- Insumos Invertidos (no semilla)
  supplies_stock_usd,             -- Insumos Stock (Invertidos - Ejecutados)
  -- LABORES
  labors_executed_usd,            -- Labores Ejecutados
  labors_invested_usd,            -- Labores Invertidos
  labors_stock_usd,               -- Labores Stock (Invertidos - Ejecutados)
  -- COSTOS DIRECTOS TOTALES
  direct_costs_executed_usd,      -- Costos Directos Ejecutados (Semilla + Insumos + Labores)
  direct_costs_invested_usd,      -- Costos Directos Invertidos (Semilla + Insumos + Labores)
  direct_costs_stock_usd,         -- Costos Directos Stock (Invertidos - Ejecutados)
  -- OTROS COSTOS
  lease_invested_usd,             -- Arriendo Invertidos (30% de comercializaciones)
  structure_invested_usd,         -- Estructura Invertidos (admin_cost del proyecto)
  total_invested_usd              -- Total Invertido (Directos + Arriendo + Estructura)
FROM dashboard_balance_management_view 
WHERE project_id IN (1,2,3)
ORDER BY project_id;

-- =============================================
-- MÓDULO 7: INCIDENCIA DE COSTOS POR CULTIVO
-- =============================================
-- RESULTADOS REALES:
-- Proyecto 1: 200 ha totales - Cultivo 1 (Soja): 100 ha (50%) $0 costos = $0/ha, Cultivo 2 (Maíz): 100 ha (50%) $92,500 costos = $925/ha
-- Proyecto 2: 150 ha totales - Cultivo 1 (Soja): 75 ha (50%) $0 costos = $0/ha, Cultivo 2 (Maíz): 75 ha (50%) $138,750 costos = $1,850/ha
-- Proyecto 3: 100 ha totales - Sin cultivos específicos, $0 costos = $0/ha
SELECT '=== MÓDULO 7: INCIDENCIA DE COSTOS POR CULTIVO ===' as info;
SELECT 
  project_id, 
  crop_id,
  crop_name,
  crop_hectares,
  project_total_hectares,
  incidence_pct,
  crop_direct_costs_usd,
  cost_per_ha_usd
FROM dashboard_crop_cost_incidence_view 
ORDER BY project_id;

-- ========================================
-- MÓDULO 8: INDICADORES OPERATIVOS
-- ========================================
-- RESULTADOS REALES:
-- Proyecto 1: Primera orden 2024-10-15, Última orden 2025-03-20, Arqueo stock 2024-11-10
-- Proyecto 2: Primera orden 2024-06-01, Última orden 2024-12-20, Arqueo stock 2024-07-01
-- Proyecto 3: Sin órdenes de trabajo, sin arqueo de stock
SELECT '=== MÓDULO 8: INDICADORES OPERATIVOS ===' as info;
SELECT 
  project_id,
  first_workorder_date,        -- Fecha de la primera orden de trabajo
  first_workorder_number,      -- Número de la primera orden de trabajo
  last_workorder_date,         -- Fecha de la última orden de trabajo
  last_workorder_number,       -- Número de la última orden de trabajo
  last_stock_count_date,       -- Fecha del último arqueo de stock
  campaign_closing_date        -- Fecha de cierre de campaña (placeholder)
FROM dashboard_operational_indicators_view 
WHERE project_id IN (1,2,3)
ORDER BY project_id;