ROLLBACK;

-- ========================================
-- CLEAN - Limpiar todas las tablas relacionadas
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
  (1, 'Inversor Demo');

-- ========================================
-- DOS PROYECTOS SIMPLES
-- ========================================
INSERT INTO projects (id, customer_id, campaign_id, name, admin_cost, status) VALUES
  (1, 1, 1, 'Proyecto A', 1000, 'active'),
  (2, 1, 1, 'Proyecto B', 500, 'active');

-- ========================================
-- RELACIÓN PROYECTO-INVERSOR
-- ========================================
INSERT INTO project_investors (project_id, investor_id, investment_percentage) VALUES
  (1, 1, 100.00),
  (2, 1, 100.00);

-- ========================================
-- DOS CAMPOS SIMPLES
-- ========================================
INSERT INTO fields (id, project_id, name, lease_type_id, hectares) VALUES
  (101, 1, 'Campo 1', 1, 200),    -- ID 1 = % INGRESO NETO
  (102, 2, 'Campo 2', 2, 150);    -- ID 2 = % UTILIDAD

-- ========================================
-- CUATRO LOTES CON NÚMEROS REDONDOS
-- ========================================
INSERT INTO lots (id, field_id, name, hectares, previous_crop_id, current_crop_id, season, tons, sowing_date, harvest_date) VALUES
  -- Proyecto 1: Campo 1
  (1001, 101, 'Lote A', 100, 1, 2, '2024-2025', 200, '2024-10-15', '2025-03-20'),   -- 100 ha sembradas y cosechadas
  (1002, 101, 'Lote B', 100, 2, 1, '2024-2025', 200, NULL, NULL),                   -- 100 ha NO sembradas aún
  
  -- Proyecto 2: Campo 2
  (1003, 102, 'Lote C', 75, 1, 3, '2024-2025', 150, '2024-06-01', '2024-12-15'),   -- 75 ha sembradas y cosechadas
  (1004, 102, 'Lote D', 75, 3, 1, '2024-2025', 150, '2024-06-05', '2024-12-20');   -- 75 ha sembradas y cosechadas

-- ========================================
-- SEIS LABORES CON PRECIOS REDONDOS
-- ========================================
INSERT INTO labors (id, project_id, name, category_id, price, contractor_name, status) VALUES
  -- Proyecto 1
  (1, 1, 'Siembra', 9, 50, 'Contratista 1', 'completed'),      -- ID 9 = Siembra
  (2, 1, 'Cosecha', 13, 100, 'Contratista 2', 'completed'),    -- ID 13 = Cosecha
  (3, 1, 'Fertilización', 8, 75, 'Contratista 3', 'completed'), -- ID 8 = Fertilizantes
  
  -- Proyecto 2
  (4, 2, 'Siembra', 9, 50, 'Contratista 1', 'completed'),
  (5, 2, 'Cosecha', 13, 100, 'Contratista 2', 'completed'),
  (6, 2, 'Fertilización', 8, 75, 'Contratista 3', 'completed');

-- ========================================
-- SEIS INSUMOS CON PRECIOS REDONDOS
-- ========================================
INSERT INTO supplies (id, project_id, name, type_id, category_id, price, unit) VALUES
  -- Proyecto 1
  (1, 1, 'Fertilizante', 3, 8, 2, 'kg'),           -- ID 3 = Fertilizantes, ID 8 = Fertilizantes
  (2, 1, 'Semilla', 1, 1, 10, 'kg'),               -- ID 1 = Semillas, ID 1 = Semilla
  (3, 1, 'Herbicida', 2, 4, 5, 'L'),               -- ID 2 = Agroquímicos, ID 4 = Herbicidas
  
  -- Proyecto 2
  (4, 2, 'Fertilizante', 3, 8, 2, 'kg'),
  (5, 2, 'Semilla', 1, 1, 10, 'kg'),
  (6, 2, 'Herbicida', 2, 4, 5, 'L');

-- ========================================
-- DOCE WORKORDERS CON NÚMEROS REDONDOS
-- ========================================
INSERT INTO workorders (id, project_id, field_id, lot_id, crop_id, labor_id, investor_id, date, effective_area, status, total_cost) VALUES
  -- Proyecto 1: Campo 1
  (1, 1, 101, 1001, 2, 1, 1, '2024-10-15', 100, 'completed', 5000),   -- Siembra Lote A - 100 ha × $50 = $5,000
  (2, 1, 101, 1001, 2, 3, 1, '2024-11-10', 100, 'completed', 7500),   -- Fertilización Lote A - 100 ha × $75 = $7,500
  (3, 1, 101, 1001, 2, 2, 1, '2025-03-20', 100, 'completed', 10000),  -- Cosecha Lote A - 100 ha × $100 = $10,000
  
  -- Proyecto 2: Campo 2
  (4, 2, 102, 1003, 3, 4, 1, '2024-06-01', 75, 'completed', 3750),    -- Siembra Lote C - 75 ha × $50 = $3,750
  (5, 2, 102, 1003, 3, 6, 1, '2024-07-01', 75, 'completed', 5625),    -- Fertilización Lote C - 75 ha × $75 = $5,625
  (6, 2, 102, 1003, 3, 5, 1, '2024-12-15', 75, 'completed', 7500),    -- Cosecha Lote C - 75 ha × $100 = $7,500
  
  (7, 2, 102, 1004, 1, 4, 1, '2024-06-05', 75, 'completed', 3750),    -- Siembra Lote D - 75 ha × $50 = $3,750
  (8, 2, 102, 1004, 1, 6, 1, '2024-07-05', 75, 'completed', 5625),    -- Fertilización Lote D - 75 ha × $75 = $5,625
  (9, 2, 102, 1004, 1, 5, 1, '2024-12-20', 75, 'completed', 7500);    -- Cosecha Lote D - 75 ha × $100 = $7,500

-- ========================================
-- WORKORDER_ITEMS CON NÚMEROS REDONDOS
-- ========================================
INSERT INTO workorder_items (workorder_id, supply_id, final_dose, total_used, unit_cost, total_cost) VALUES
  -- Proyecto 1: Fertilización
  (2, 1, 100, 10000, 2, 20000),    -- Workorder 2: 100 kg/ha × 100 ha = 10,000 kg × $2 = $20,000
  
  -- Proyecto 1: Semillas
  (1, 2, 50, 5000, 10, 50000),     -- Workorder 1: 50 kg/ha × 100 ha = 5,000 kg × $10 = $50,000
  
  -- Proyecto 2: Fertilización
  (5, 4, 100, 7500, 2, 15000),     -- Workorder 5: 100 kg/ha × 75 ha = 7,500 kg × $2 = $15,000
  (8, 4, 100, 7500, 2, 15000),     -- Workorder 8: 100 kg/ha × 75 ha = 7,500 kg × $2 = $15,000
  
  -- Proyecto 2: Semillas
  (4, 5, 50, 3750, 10, 37500),     -- Workorder 4: 50 kg/ha × 75 ha = 3,750 kg × $10 = $37,500
  (7, 5, 50, 3750, 10, 37500);     -- Workorder 7: 50 kg/ha × 75 ha = 3,750 kg × $10 = $37,500

-- ========================================
-- VERIFICACIÓN Y VALIDACIÓN
-- ========================================
SELECT '=== RESUMEN DE DATOS INSERTADOS ===' as info;

SELECT 'Customers' as tabla, COUNT(*) as total FROM customers
UNION ALL
SELECT 'Projects', COUNT(*) FROM projects
UNION ALL
SELECT 'Fields', COUNT(*) FROM fields
UNION ALL
SELECT 'Lots', COUNT(*) FROM lots
UNION ALL
SELECT 'Labors', COUNT(*) FROM labors
UNION ALL
SELECT 'Supplies', COUNT(*) FROM supplies
UNION ALL
SELECT 'Workorders', COUNT(*) FROM workorders
UNION ALL
SELECT 'Workorder Items', COUNT(*) FROM workorder_items
UNION ALL
SELECT 'Project Investors', COUNT(*) FROM project_investors;

-- ========================================
-- VERIFICACIÓN DEL DASHBOARD
-- ========================================
SELECT '=== VERIFICACIÓN DASHBOARD ===' as info;

-- 1. Avance de siembra por proyecto
SELECT 
  'SOWING PROGRESS' as metric,
  customer_id, project_id, campaign_id, field_id,
  sowing_hectares,           -- Hectáreas sembradas
  sowing_total_hectares,     -- Total de hectáreas del proyecto
  sowing_progress_percent    -- Porcentaje de avance de siembra
FROM dashboard_view 
WHERE project_id IS NOT NULL 
ORDER BY customer_id, project_id, campaign_id, field_id;

-- 2. Avance de costos por proyecto
SELECT 
  'COSTS PROGRESS' as metric,
  customer_id, project_id, campaign_id, field_id,
  executed_costs_usd,        -- Costos directos ejecutados
  executed_labors_usd,       -- Labores ejecutadas
  executed_supplies_usd,     -- Insumos utilizados
  budget_cost_usd,           -- Costos administrativos
  budget_total_usd,          -- Presupuesto total del proyecto
  costs_progress_pct         -- Porcentaje de avance de costos
FROM dashboard_view 
WHERE project_id IS NOT NULL 
ORDER BY customer_id, project_id, campaign_id, field_id;

-- 3. Avance de cosecha por proyecto
SELECT 
  'HARVEST PROGRESS' as metric,
  customer_id, project_id, campaign_id, field_id,
  harvest_hectares,           -- Hectáreas cosechadas
  harvest_total_hectares,     -- Total de hectáreas cosechables
  harvest_progress_percent    -- Porcentaje de avance de cosecha
FROM dashboard_view 
WHERE project_id IS NOT NULL 
ORDER BY customer_id, project_id, campaign_id, field_id;

-- ========================================
-- CÁLCULOS ESPERADOS PARA VERIFICACIÓN
-- ========================================
SELECT '=== CÁLCULOS ESPERADOS ===' as info;

-- Proyecto 1: Campo 1 (200 ha totales)
-- Sowing: 100 ha sembradas / 200 ha totales = 50%
-- Costs: Labores + Insumos + Admin = (22,500 + 70,000 + 1,000) = 93,500
-- Harvest: 100 ha cosechadas / 100 ha cosechables = 100%

-- Proyecto 2: Campo 2 (150 ha totales)
-- Sowing: 150 ha sembradas / 150 ha totales = 100%
-- Costs: Labores + Insumos + Admin = (22,500 + 67,500 + 500) = 90,500
-- Harvest: 150 ha cosechadas / 150 ha cosechables = 100%