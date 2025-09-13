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
TRUNCATE app_parameters RESTART IDENTITY CASCADE;
TRUNCATE stocks RESTART IDENTITY CASCADE;

-- ========================================
-- CONFIGURACIÓN DE APLICACIÓN (app_parameters)
-- ========================================
-- Esta sección configura los parámetros unificados del sistema
-- Reemplaza las tablas separadas: units, calc_values, etc.
-- 
-- CATEGORÍAS:
-- - units: Unidades de medida (Lt, Kg, Ha)
-- - calculations: Valores de cálculo (IVA, días, tasas)
-- - business_rules: Reglas de negocio (porcentajes, etc.)
--
-- USO EN EL CÓDIGO:
-- - get_app_parameter('unit_liters') → 'Lt'
-- - get_iva_percentage() → 0.105
-- - get_campaign_closure_days() → 30
-- - get_default_fx_rate() → 1.0000
INSERT INTO app_parameters (key, value, type, category, description) VALUES
-- UNIDADES (antes tabla units separada)
('unit_liters', 'Lt', 'string', 'units', 'Unit of measurement: Liters'),
('unit_kilos', 'Kg', 'string', 'units', 'Unit of measurement: Kilograms'),
('unit_hectares', 'Ha', 'string', 'units', 'Unit of measurement: Hectares'),

-- CÁLCULOS (antes tabla calc_values separada)
('iva_percentage', '0.105', 'decimal', 'calculations', 'VAT percentage for labors (10.5%)'),
('campaign_closure_days', '30', 'integer', 'calculations', 'Days for campaign closure after end date'),
('default_fx_rate', '1.0000', 'decimal', 'calculations', 'Default exchange rate USD/USD');

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
  (3, 1, 1, 'Proyecto C - Sin Siembra', 750),   -- 0% sembrado, 0% cosechado
  (4, 1, 1, 'Proyecto D - Con Ingresos', 800),  -- 100% sembrado, cosechado y vendido
  (5, 1, 1, 'Proyecto prueba Agus', 0);         -- Ejemplo para verificar costos directos

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
  (3, 1, 100.00),
  
  -- Proyecto 4: 2 inversores (70% + 30% = 100%)
  (4, 2, 70.00),
  (4, 3, 30.00),
  
  -- Proyecto 5: 1 inversor (100%)
  (5, 1, 100.00);

-- ========================================
-- TRES CAMPOS CON DIFERENTES ESCENARIOS
-- ========================================
INSERT INTO fields (id, project_id, name, lease_type_id, lease_type_percent, lease_type_value) VALUES
  (101, 1, 'Campo A - Parcial', 1, 30.00, NULL),    -- ID 1 = % INGRESO NETO (30%)
  (102, 2, 'Campo B - Completo', 2, 25.00, NULL),   -- ID 2 = % UTILIDAD (25%)
  (103, 3, 'Campo C - Vacío', 3, NULL, 150.00),      -- ID 3 = ARRIENDO FIJO ($150/ha)
  (104, 4, 'Campo D - Con Ingresos', 1, 40.00, NULL), -- ID 1 = % INGRESO NETO (40%)
  (105, 5, 'Ejemplo', 1, NULL, NULL),               -- Campo para proyecto prueba Agus
  (106, 5, 'Ejemplo 2', 1, NULL, NULL);             -- Campo para proyecto prueba Agus

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
  (1006, 103, 'Lote C2', 50, 2, 1, '2024-2025', 100, NULL),            -- 50 ha NO sembradas
  
  -- Proyecto 4: Campo D - Con Ingresos (120 ha sembradas de 120 ha totales)
  (1007, 104, 'Lote D1', 60, 1, 2, '2024-2025', 120, '2024-05-01'),    -- 60 ha sembradas
  (1008, 104, 'Lote D2', 60, 2, 1, '2024-2025', 120, '2024-05-05'),    -- 60 ha sembradas
  
  -- Proyecto 5: Campos Ejemplo (35 ha totales)
  (1009, 105, 'Lote Ejemplo 1', 15, 1, 2, '2024-2025', NULL, NULL),     -- 15 ha para proyecto prueba Agus
  (1010, 106, 'Lote Ejemplo 2', 20, 2, 1, '2024-2025', NULL, NULL);     -- 20 ha para proyecto prueba Agus

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
  (6, 2, 'Fertilización', 8, 75, 'Contratista 3'),
  
  -- Proyecto 4
  (7, 4, 'Siembra', 9, 50, 'Contratista 1'),      -- ID 9 = Siembra
  (8, 4, 'Cosecha', 13, 100, 'Contratista 2'),    -- ID 13 = Cosecha
  (9, 4, 'Fertilización', 8, 75, 'Contratista 3'), -- ID 8 = Fertilizantes
  
  -- Proyecto 5
  (10, 5, 'Labor Ejemplo', 9, 5.00, 'Contratista Ejemplo'); -- Labor para proyecto prueba Agus

-- ========================================
-- INSUMOS CON DIFERENTES UNIDADES DE MEDIDA
-- ========================================
-- NOTA: unit_id ahora hace referencia a app_parameters
-- - unit_id 1 = 'unit_liters' (Lt) - Para herbicidas, insecticidas, fungicidas, adyuvantes
-- - unit_id 2 = 'unit_kilos' (Kg) - Para fertilizantes, semillas
-- - unit_id 3 = 'unit_hectares' (Ha) - Para servicios por hectárea
--
-- EJEMPLOS DE USO:
-- - Herbicidas: 2.5 Lt/ha × 100 ha = 250 Lt × $15/Lt = $3,750
-- - Insecticidas: 1.5 Lt/ha × 75 ha = 112.5 Lt × $25/Lt = $2,812.50
-- - Fungicidas: 2.0 Lt/ha × 75 ha = 150 Lt × $18/Lt = $2,700
-- - Adyuvantes: 0.5 Lt/ha × 50 ha = 25 Lt × $8/Lt = $200
--
-- TYPES DISPONIBLES:
-- - type_id 1 = 'Semilla'
-- - type_id 2 = 'Agroquímicos' (herbicidas, insecticidas, fungicidas, adyuvantes)
-- - type_id 3 = 'Fertilizantes'
-- - type_id 4 = 'Labores'
--
-- En el código Go, se usa getUnitName() que consulta app_parameters
INSERT INTO supplies (id, project_id, name, type_id, category_id, unit_id, price) VALUES
  -- Proyecto 1
  (1, 1, 'Fertilizante', 3, 8, 2, 2),           -- type_id 3 = Fertilizantes, category_id 8, unit_id 2 = Kg
  (2, 1, 'Semilla', 1, 1, 2, 10),               -- type_id 1 = Semilla, category_id 1, unit_id 2 = Kg
  (3, 1, 'Herbicida', 2, 9, 1, 15),             -- type_id 2 = Agroquímicos, category_id 9, unit_id 1 = Lt
  
  -- Proyecto 2
  (4, 2, 'Fertilizante', 3, 8, 2, 2),           -- type_id 3 = Fertilizantes, unit_id 2 = Kg
  (5, 2, 'Semilla', 1, 1, 2, 10),               -- type_id 1 = Semilla, unit_id 2 = Kg
  (6, 2, 'Insecticida', 2, 10, 1, 25),          -- type_id 2 = Agroquímicos, category_id 10, unit_id 1 = Lt
  (7, 2, 'Fungicida', 2, 11, 1, 18),            -- type_id 2 = Agroquímicos, category_id 11, unit_id 1 = Lt
  
  -- Proyecto 3
  (8, 3, 'Fertilizante', 3, 8, 2, 2),           -- type_id 3 = Fertilizantes, unit_id 2 = Kg
  (9, 3, 'Semilla', 1, 1, 2, 10),               -- type_id 1 = Semilla, unit_id 2 = Kg
  (10, 3, 'Adyuvante', 2, 12, 1, 8),            -- type_id 2 = Agroquímicos, category_id 12, unit_id 1 = Lt
  
  -- Proyecto 4
  (11, 4, 'Fertilizante', 3, 8, 2, 2),          -- type_id 3 = Fertilizantes, unit_id 2 = Kg
  (12, 4, 'Semilla', 1, 1, 2, 10),              -- type_id 1 = Semilla, unit_id 2 = Kg
  (13, 4, 'Herbicida', 2, 9, 1, 15),            -- type_id 2 = Agroquímicos, category_id 9, unit_id 1 = Lt
  (14, 4, 'Insecticida', 2, 10, 1, 25),         -- type_id 2 = Agroquímicos, category_id 10, unit_id 1 = Lt
  
  -- Proyecto 5 (reutiliza insumos del proyecto 1)
  (15, 5, 'Fertilizante Ejemplo', 3, 8, 2, 2),  -- type_id 3 = Fertilizantes, unit_id 2 = Kg
  (16, 5, 'Herbicida Ejemplo', 2, 9, 1, 15);    -- type_id 2 = Agroquímicos, category_id 9, unit_id 1 = Lt

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
  (9, 2, 102, 1004, 1, 5, 1, '2024-12-20', 75),    -- Cosecha Lote B2 - 75 ha × $100 = $7,500
  
  -- Proyecto 3: Campo C - Con Costos Mínimos (100 ha de 100 ha) - CORREGIDO PARA SER REALISTA
  (16, 3, 103, 1005, 2, 1, 1, '2024-08-01', 50),    -- Siembra Lote C1 - 50 ha × $50 = $2,500
  (17, 3, 103, 1005, 2, 3, 1, '2024-09-01', 50),    -- Fertilización Lote C1 - 50 ha × $75 = $3,750
  (18, 3, 103, 1005, 2, 2, 1, '2025-01-15', 50),    -- Cosecha Lote C1 - 50 ha × $100 = $5,000
  
  (19, 3, 103, 1006, 1, 1, 1, '2024-08-05', 50),    -- Siembra Lote C2 - 50 ha × $50 = $2,500
  (20, 3, 103, 1006, 1, 3, 1, '2024-09-05', 50),    -- Fertilización Lote C2 - 50 ha × $75 = $3,750
  (21, 3, 103, 1006, 1, 2, 1, '2025-01-20', 50),    -- Cosecha Lote C2 - 50 ha × $100 = $5,000
  
  -- Proyecto 4: Campo D - Con Ingresos (120 ha de 120 ha)
  (10, 4, 104, 1007, 2, 7, 1, '2024-05-01', 60),   -- Siembra Lote D1 - 60 ha × $50 = $3,000
  (11, 4, 104, 1007, 2, 9, 1, '2024-06-01', 60),   -- Fertilización Lote D1 - 60 ha × $75 = $4,500
  (12, 4, 104, 1007, 2, 8, 1, '2024-11-15', 60),   -- Cosecha Lote D1 - 60 ha × $100 = $6,000
  
  (13, 4, 104, 1008, 1, 7, 1, '2024-05-05', 60),   -- Siembra Lote D2 - 60 ha × $50 = $3,000
  (14, 4, 104, 1008, 1, 9, 1, '2024-06-05', 60),   -- Fertilización Lote D2 - 60 ha × $75 = $4,500
  (15, 4, 104, 1008, 1, 8, 1, '2024-11-20', 60),   -- Cosecha Lote D2 - 60 ha × $100 = $6,000
  
  -- Proyecto 5: Ejemplo para verificar costos directos
  (101, 5, 105, 1009, 1, 10, 1, '2024-10-01', 15.00), -- Labor Ejemplo Lote Ejemplo 1 - 15 ha × $5 = $75
  (102, 5, 106, 1010, 1, 10, 1, '2024-10-01', 15.00); -- Labor Ejemplo Lote Ejemplo 2 - 15 ha × $5 = $75

-- ========================================
-- WORKORDER_ITEMS CON NÚMEROS REDONDOS
-- ========================================
INSERT INTO workorder_items (workorder_id, supply_id, final_dose, total_used) VALUES
  -- Proyecto 1: Fertilización (Kg)
  (2, 1, 100, 10000),    -- Workorder 2: 100 kg/ha × 100 ha = 10,000 kg × $2 = $20,000
  
  -- Proyecto 1: Semillas (Kg)
  (1, 2, 50, 5000),      -- Workorder 1: 50 kg/ha × 100 ha = 5,000 kg × $10 = $50,000
  
  -- Proyecto 1: Herbicida (Lt)
  (1, 3, 2.5, 250),      -- Workorder 1: 2.5 Lt/ha × 100 ha = 250 Lt × $15 = $3,750
  
  -- Proyecto 2: Fertilización (Kg)
  (5, 4, 100, 7500),     -- Workorder 5: 100 kg/ha × 75 ha = 7,500 kg × $2 = $15,000
  (8, 4, 100, 7500),     -- Workorder 8: 100 kg/ha × 75 ha = 7,500 kg × $2 = $15,000
  
  -- Proyecto 2: Semillas (Kg)
  (4, 5, 50, 3750),      -- Workorder 4: 50 kg/ha × 75 ha = 3,750 kg × $10 = $37,500
  (7, 5, 50, 3750),      -- Workorder 7: 50 kg/ha × 75 ha = 3,750 kg × $10 = $37,500
  
  -- Proyecto 2: Insecticida (Lt)
  (5, 6, 1.5, 112.5),    -- Workorder 5: 1.5 Lt/ha × 75 ha = 112.5 Lt × $25 = $2,812.50
  
  -- Proyecto 2: Fungicida (Lt)
  (8, 7, 2.0, 150),      -- Workorder 8: 2.0 Lt/ha × 75 ha = 150 Lt × $18 = $2,700
  
  -- Proyecto 3: Fertilización (Kg)
  (17, 8, 100, 5000),    -- Workorder 17: 100 kg/ha × 50 ha = 5,000 kg × $2 = $10,000
  (20, 8, 100, 5000),    -- Workorder 20: 100 kg/ha × 50 ha = 5,000 kg × $2 = $10,000
  
  -- Proyecto 3: Semillas (Kg)
  (16, 9, 50, 2500),     -- Workorder 16: 50 kg/ha × 50 ha = 2,500 kg × $10 = $25,000
  (19, 9, 50, 2500),     -- Workorder 19: 50 kg/ha × 50 ha = 2,500 kg × $10 = $25,000
  
  -- Proyecto 3: Adyuvante (Lt)
  (17, 10, 0.5, 25),     -- Workorder 17: 0.5 Lt/ha × 50 ha = 25 Lt × $8 = $200
  
  -- Proyecto 4: Fertilización (Kg)
  (11, 11, 100, 6000),   -- Workorder 11: 100 kg/ha × 60 ha = 6,000 kg × $2 = $12,000
  (14, 11, 100, 6000),   -- Workorder 14: 100 kg/ha × 60 ha = 6,000 kg × $2 = $12,000
  
  -- Proyecto 4: Semillas (Kg)
  (10, 12, 50, 3000),    -- Workorder 10: 50 kg/ha × 60 ha = 3,000 kg × $10 = $30,000
  (13, 12, 50, 3000),    -- Workorder 13: 50 kg/ha × 60 ha = 3,000 kg × $10 = $30,000
  
  -- Proyecto 4: Herbicida (Lt)
  (10, 13, 2.0, 120),    -- Workorder 10: 2.0 Lt/ha × 60 ha = 120 Lt × $15 = $1,800
  
  -- Proyecto 4: Insecticida (Lt)
  (14, 14, 1.8, 108),    -- Workorder 14: 1.8 Lt/ha × 60 ha = 108 Lt × $25 = $2,700
  
  -- Proyecto 5: Fertilizante (Kg)
  (101, 15, 3.3327, 50), -- Workorder 101: 3.3327 kg/ha × 15 ha = 50 kg × $2 = $100
  
  -- Proyecto 5: Herbicida (Lt)
  (101, 16, 1.5, 22.5);  -- Workorder 101: 1.5 Lt/ha × 15 ha = 22.5 Lt × $15 = $337.50

-- ========================================
-- INVENTARIOS DE STOCK (stocks)
-- ========================================
-- Datos de inventarios físicos de insumos para control de stock
-- Representa el cierre de inventarios mensuales con existencia física vs. consumo
--
-- ESTRUCTURA:
-- - project_id: Proyecto al que pertenece el stock
-- - field_id: Campo específico del stock
-- - supply_id: Insumo inventariado
-- - investor_id: Inversor responsable del stock
-- - close_date: Fecha de cierre del inventario
-- - real_stock_units: Existencia física real encontrada
-- - initial_units: Existencia inicial del período
-- - year_period/month_period: Período del inventario
--
-- EJEMPLOS REALISTAS:
-- - Fertilizantes: Stock inicial vs. consumo en fertilizaciones
-- - Semillas: Stock inicial vs. consumo en siembras
-- - Herbicidas/Insecticidas: Stock inicial vs. consumo en aplicaciones
INSERT INTO stocks (id, project_id, supply_id, investor_id, close_date, real_stock_units, initial_units, year_period, month_period) VALUES
  -- PROYECTO 1: Inventarios mensuales (Octubre 2024 - Marzo 2025)
  -- Campo 101 (Campo A - Parcial)
  (1, 1, 1, 1, '2024-10-31', 8000, 10000, 2024, 10),    -- Fertilizante: 10,000 kg inicial - 2,000 kg usado = 8,000 kg stock
  (2, 1, 2, 1, '2024-10-31', 0, 5000, 2024, 10),       -- Semilla: 5,000 kg inicial - 5,000 kg usado = 0 kg stock
  (3, 1, 3, 1, '2024-10-31', 0, 250, 2024, 10),        -- Herbicida: 250 Lt inicial - 250 Lt usado = 0 Lt stock
  
  (4, 1, 1, 1, '2024-11-30', 6000, 8000, 2024, 11),    -- Fertilizante: 8,000 kg stock - 2,000 kg usado = 6,000 kg stock
  (5, 1, 2, 1, '2024-11-30', 0, 0, 2024, 11),          -- Semilla: Sin stock (ya usado en octubre)
  (6, 1, 3, 1, '2024-11-30', 0, 0, 2024, 11),          -- Herbicida: Sin stock (ya usado en octubre)
  
  (7, 1, 1, 1, '2024-12-31', 4000, 6000, 2024, 12),    -- Fertilizante: 6,000 kg stock - 2,000 kg usado = 4,000 kg stock
  
  (8, 1, 1, 1, '2025-01-31', 2000, 4000, 2025, 1),     -- Fertilizante: 4,000 kg stock - 2,000 kg usado = 2,000 kg stock
  
  (9, 1, 1, 1, '2025-02-28', 0, 2000, 2025, 2),        -- Fertilizante: 2,000 kg stock - 2,000 kg usado = 0 kg stock
  
  (10, 1, 1, 1, '2025-03-31', 0, 0, 2025, 3),          -- Fertilizante: Sin stock (agotado en febrero)

  -- PROYECTO 2: Inventarios mensuales (Junio 2024 - Diciembre 2024)
  -- Campo 102 (Campo B - Completo)
  (11, 2, 4, 1, '2024-06-30', 6000, 7500, 2024, 6),    -- Fertilizante: 7,500 kg inicial - 1,500 kg usado = 6,000 kg stock
  (12, 2, 5, 1, '2024-06-30', 0, 3750, 2024, 6),       -- Semilla: 3,750 kg inicial - 3,750 kg usado = 0 kg stock
  (13, 2, 6, 1, '2024-06-30', 0, 112.5, 2024, 6),      -- Insecticida: 112.5 Lt inicial - 112.5 Lt usado = 0 Lt stock
  
  (14, 2, 4, 1, '2024-07-31', 3000, 6000, 2024, 7),    -- Fertilizante: 6,000 kg stock - 3,000 kg usado = 3,000 kg stock
  (15, 2, 7, 1, '2024-07-31', 0, 150, 2024, 7),        -- Fungicida: 150 Lt inicial - 150 Lt usado = 0 Lt stock
  
  (16, 2, 4, 1, '2024-08-31', 0, 3000, 2024, 8),       -- Fertilizante: 3,000 kg stock - 3,000 kg usado = 0 kg stock
  
  (17, 2, 4, 1, '2024-09-30', 0, 0, 2024, 9),          -- Fertilizante: Sin stock (agotado en agosto)
  
  (18, 2, 4, 1, '2024-10-31', 0, 0, 2024, 10),         -- Fertilizante: Sin stock
  
  (19, 2, 4, 1, '2024-11-30', 0, 0, 2024, 11),         -- Fertilizante: Sin stock
  
  (20, 2, 4, 1, '2024-12-31', 0, 0, 2024, 12),         -- Fertilizante: Sin stock

  -- PROYECTO 3: Inventarios mensuales (Agosto 2024 - Enero 2025)
  -- Campo 103 (Campo C - Vacío) - Solo algunos inventarios
  (21, 3, 8, 1, '2024-08-31', 4000, 5000, 2024, 8),    -- Fertilizante: 5,000 kg inicial - 1,000 kg usado = 4,000 kg stock
  (22, 3, 9, 1, '2024-08-31', 0, 2500, 2024, 8),       -- Semilla: 2,500 kg inicial - 2,500 kg usado = 0 kg stock
  (23, 3, 10, 1, '2024-08-31', 0, 25, 2024, 8),        -- Adyuvante: 25 Lt inicial - 25 Lt usado = 0 Lt stock
  
  (24, 3, 8, 1, '2024-09-30', 2000, 4000, 2024, 9),    -- Fertilizante: 4,000 kg stock - 2,000 kg usado = 2,000 kg stock
  
  (25, 3, 8, 1, '2024-10-31', 0, 2000, 2024, 10),      -- Fertilizante: 2,000 kg stock - 2,000 kg usado = 0 kg stock
  
  (26, 3, 8, 1, '2024-11-30', 0, 0, 2024, 11),         -- Fertilizante: Sin stock (agotado en octubre)
  
  (27, 3, 8, 1, '2024-12-31', 0, 0, 2024, 12),         -- Fertilizante: Sin stock
  
  (28, 3, 8, 1, '2025-01-31', 0, 0, 2025, 1),          -- Fertilizante: Sin stock

  -- PROYECTO 4: Inventarios mensuales (Mayo 2024 - Noviembre 2024)
  -- Campo 104 (Campo D - Con Ingresos)
  (29, 4, 11, 1, '2024-05-31', 4000, 6000, 2024, 5),   -- Fertilizante: 6,000 kg inicial - 2,000 kg usado = 4,000 kg stock
  (30, 4, 12, 1, '2024-05-31', 0, 3000, 2024, 5),      -- Semilla: 3,000 kg inicial - 3,000 kg usado = 0 kg stock
  (31, 4, 13, 1, '2024-05-31', 0, 120, 2024, 5),       -- Herbicida: 120 Lt inicial - 120 Lt usado = 0 Lt stock
  
  (32, 4, 11, 1, '2024-06-30', 0, 4000, 2024, 6),      -- Fertilizante: 4,000 kg stock - 4,000 kg usado = 0 kg stock
  (33, 4, 14, 1, '2024-06-30', 0, 108, 2024, 6),       -- Insecticida: 108 Lt inicial - 108 Lt usado = 0 Lt stock
  
  (34, 4, 11, 1, '2024-07-31', 0, 0, 2024, 7),         -- Fertilizante: Sin stock (agotado en junio)
  
  (35, 4, 11, 1, '2024-08-31', 0, 0, 2024, 8),         -- Fertilizante: Sin stock
  
  (36, 4, 11, 1, '2024-09-30', 0, 0, 2024, 9),         -- Fertilizante: Sin stock
  
  (37, 4, 11, 1, '2024-10-31', 0, 0, 2024, 10),        -- Fertilizante: Sin stock
  
  (38, 4, 11, 1, '2024-11-30', 0, 0, 2024, 11),        -- Fertilizante: Sin stock

  -- PROYECTO 5: Inventarios mensuales (Octubre 2024)
  -- Campo 105 (Ejemplo) - Inventario simple
  (39, 5, 15, 1, '2024-10-31', 47, 50, 2024, 10),      -- Fertilizante Ejemplo: 50 kg inicial - 3 kg usado = 47 kg stock
  (40, 5, 16, 1, '2024-10-31', 0, 22.5, 2024, 10),     -- Herbicida Ejemplo: 22.5 Lt inicial - 22.5 Lt usado = 0 Lt stock
  
  -- Campo 106 (Ejemplo 2) - Inventario simple
  (41, 5, 15, 1, '2024-10-31', 47, 50, 2024, 10),      -- Fertilizante Ejemplo: 50 kg inicial - 3 kg usado = 47 kg stock
  (42, 5, 16, 1, '2024-10-31', 0, 22.5, 2024, 10),     -- Herbicida Ejemplo: 22.5 Lt inicial - 22.5 Lt usado = 0 Lt stock

  -- ========================================
  -- STOCKS ADICIONALES PARA COMPLETAR INVENTARIOS
  -- ========================================
  -- Añadir más inventarios para que todos los proyectos tengan cobertura completa
  
  -- PROYECTO 1: Inventarios adicionales (Abril 2025 - Junio 2025)
  (43, 1, 1, 1, '2025-04-30', 0, 0, 2025, 4),          -- Fertilizante: Sin stock (agotado)
  (44, 1, 2, 1, '2025-04-30', 0, 0, 2025, 4),          -- Semilla: Sin stock
  (45, 1, 3, 1, '2025-04-30', 0, 0, 2025, 4),          -- Herbicida: Sin stock
  
  (46, 1, 1, 1, '2025-05-31', 0, 0, 2025, 5),          -- Fertilizante: Sin stock
  (47, 1, 2, 1, '2025-05-31', 0, 0, 2025, 5),          -- Semilla: Sin stock
  (48, 1, 3, 1, '2025-05-31', 0, 0, 2025, 5),          -- Herbicida: Sin stock
  
  (49, 1, 1, 1, '2025-06-30', 0, 0, 2025, 6),          -- Fertilizante: Sin stock
  (50, 1, 2, 1, '2025-06-30', 0, 0, 2025, 6),          -- Semilla: Sin stock
  (51, 1, 3, 1, '2025-06-30', 0, 0, 2025, 6),          -- Herbicida: Sin stock

  -- PROYECTO 2: Inventarios adicionales (Enero 2025 - Marzo 2025)
  (52, 2, 4, 1, '2025-01-31', 0, 0, 2025, 1),          -- Fertilizante: Sin stock
  (53, 2, 5, 1, '2025-01-31', 0, 0, 2025, 1),          -- Semilla: Sin stock
  (54, 2, 6, 1, '2025-01-31', 0, 0, 2025, 1),          -- Insecticida: Sin stock
  (55, 2, 7, 1, '2025-01-31', 0, 0, 2025, 1),          -- Fungicida: Sin stock
  
  (56, 2, 4, 1, '2025-02-28', 0, 0, 2025, 2),          -- Fertilizante: Sin stock
  (57, 2, 5, 1, '2025-02-28', 0, 0, 2025, 2),          -- Semilla: Sin stock
  (58, 2, 6, 1, '2025-02-28', 0, 0, 2025, 2),          -- Insecticida: Sin stock
  (59, 2, 7, 1, '2025-02-28', 0, 0, 2025, 2),          -- Fungicida: Sin stock
  
  (60, 2, 4, 1, '2025-03-31', 0, 0, 2025, 3),          -- Fertilizante: Sin stock
  (61, 2, 5, 1, '2025-03-31', 0, 0, 2025, 3),          -- Semilla: Sin stock
  (62, 2, 6, 1, '2025-03-31', 0, 0, 2025, 3),          -- Insecticida: Sin stock
  (63, 2, 7, 1, '2025-03-31', 0, 0, 2025, 3),          -- Fungicida: Sin stock

  -- PROYECTO 3: Inventarios adicionales (Febrero 2025 - Abril 2025)
  (64, 3, 8, 1, '2025-02-28', 0, 0, 2025, 2),          -- Fertilizante: Sin stock
  (65, 3, 9, 1, '2025-02-28', 0, 0, 2025, 2),          -- Semilla: Sin stock
  (66, 3, 10, 1, '2025-02-28', 0, 0, 2025, 2),         -- Adyuvante: Sin stock
  
  (67, 3, 8, 1, '2025-03-31', 0, 0, 2025, 3),          -- Fertilizante: Sin stock
  (68, 3, 9, 1, '2025-03-31', 0, 0, 2025, 3),          -- Semilla: Sin stock
  (69, 3, 10, 1, '2025-03-31', 0, 0, 2025, 3),         -- Adyuvante: Sin stock
  
  (70, 3, 8, 1, '2025-04-30', 0, 0, 2025, 4),          -- Fertilizante: Sin stock
  (71, 3, 9, 1, '2025-04-30', 0, 0, 2025, 4),          -- Semilla: Sin stock
  (72, 3, 10, 1, '2025-04-30', 0, 0, 2025, 4),         -- Adyuvante: Sin stock

  -- PROYECTO 4: Inventarios adicionales (Diciembre 2024 - Febrero 2025)
  (73, 4, 11, 1, '2024-12-31', 0, 0, 2024, 12),        -- Fertilizante: Sin stock
  (74, 4, 12, 1, '2024-12-31', 0, 0, 2024, 12),        -- Semilla: Sin stock
  (75, 4, 13, 1, '2024-12-31', 0, 0, 2024, 12),        -- Herbicida: Sin stock
  (76, 4, 14, 1, '2024-12-31', 0, 0, 2024, 12),        -- Insecticida: Sin stock
  
  (77, 4, 11, 1, '2025-01-31', 0, 0, 2025, 1),         -- Fertilizante: Sin stock
  (78, 4, 12, 1, '2025-01-31', 0, 0, 2025, 1),         -- Semilla: Sin stock
  (79, 4, 13, 1, '2025-01-31', 0, 0, 2025, 1),         -- Herbicida: Sin stock
  (80, 4, 14, 1, '2025-01-31', 0, 0, 2025, 1),         -- Insecticida: Sin stock
  
  (81, 4, 11, 1, '2025-02-28', 0, 0, 2025, 2),         -- Fertilizante: Sin stock
  (82, 4, 12, 1, '2025-02-28', 0, 0, 2025, 2),         -- Semilla: Sin stock
  (83, 4, 13, 1, '2025-02-28', 0, 0, 2025, 2),         -- Herbicida: Sin stock
  (84, 4, 14, 1, '2025-02-28', 0, 0, 2025, 2),         -- Insecticida: Sin stock

  -- PROYECTO 5: Inventarios adicionales (Noviembre 2024 - Enero 2025)
  (85, 5, 15, 1, '2024-11-30', 44, 47, 2024, 11),      -- Fertilizante Ejemplo: 47 kg stock - 3 kg usado = 44 kg stock
  (86, 5, 16, 1, '2024-11-30', 0, 0, 2024, 11),        -- Herbicida Ejemplo: Sin stock
  
  (87, 5, 15, 1, '2024-12-31', 41, 44, 2024, 12),      -- Fertilizante Ejemplo: 44 kg stock - 3 kg usado = 41 kg stock
  (88, 5, 16, 1, '2024-12-31', 0, 0, 2024, 12),        -- Herbicida Ejemplo: Sin stock
  
  (89, 5, 15, 1, '2025-01-31', 38, 41, 2025, 1),       -- Fertilizante Ejemplo: 41 kg stock - 3 kg usado = 38 kg stock
  (90, 5, 16, 1, '2025-01-31', 0, 0, 2025, 1),         -- Herbicida Ejemplo: Sin stock
  
  (91, 5, 15, 1, '2025-02-28', 35, 38, 2025, 2),       -- Fertilizante Ejemplo: 38 kg stock - 3 kg usado = 35 kg stock
  (92, 5, 16, 1, '2025-02-28', 0, 0, 2025, 2),         -- Herbicida Ejemplo: Sin stock
  
  (93, 5, 15, 1, '2025-03-31', 32, 35, 2025, 3),       -- Fertilizante Ejemplo: 35 kg stock - 3 kg usado = 32 kg stock
  (94, 5, 16, 1, '2025-03-31', 0, 0, 2025, 3),         -- Herbicida Ejemplo: Sin stock

  -- ========================================
  -- STOCKS ADICIONALES PARA PROYECTOS 1-4
  -- ========================================
  -- Añadir inventarios adicionales para que todos los proyectos tengan stock restante
  
  -- PROYECTO 1: Inventarios adicionales (Julio 2025 - Septiembre 2025)
  (95, 1, 1, 1, '2025-07-31', 500, 1000, 2025, 7),     -- Fertilizante: 1,000 kg inicial - 500 kg usado = 500 kg stock
  (96, 1, 2, 1, '2025-07-31', 200, 500, 2025, 7),      -- Semilla: 500 kg inicial - 300 kg usado = 200 kg stock
  (97, 1, 3, 1, '2025-07-31', 50, 100, 2025, 7),       -- Herbicida: 100 Lt inicial - 50 Lt usado = 50 Lt stock
  
  (98, 1, 1, 1, '2025-08-31', 300, 500, 2025, 8),      -- Fertilizante: 500 kg stock - 200 kg usado = 300 kg stock
  (99, 1, 2, 1, '2025-08-31', 100, 200, 2025, 8),      -- Semilla: 200 kg stock - 100 kg usado = 100 kg stock
  (100, 1, 3, 1, '2025-08-31', 25, 50, 2025, 8),       -- Herbicida: 50 Lt stock - 25 Lt usado = 25 Lt stock
  
  (101, 1, 1, 1, '2025-09-30', 150, 300, 2025, 9),     -- Fertilizante: 300 kg stock - 150 kg usado = 150 kg stock
  (102, 1, 2, 1, '2025-09-30', 50, 100, 2025, 9),      -- Semilla: 100 kg stock - 50 kg usado = 50 kg stock
  (103, 1, 3, 1, '2025-09-30', 10, 25, 2025, 9),       -- Herbicida: 25 Lt stock - 15 Lt usado = 10 Lt stock

  -- PROYECTO 2: Inventarios adicionales (Abril 2025 - Junio 2025)
  (104, 2, 4, 1, '2025-04-30', 800, 1500, 2025, 4),    -- Fertilizante: 1,500 kg inicial - 700 kg usado = 800 kg stock
  (105, 2, 5, 1, '2025-04-30', 300, 600, 2025, 4),     -- Semilla: 600 kg inicial - 300 kg usado = 300 kg stock
  (106, 2, 6, 1, '2025-04-30', 75, 150, 2025, 4),      -- Insecticida: 150 Lt inicial - 75 Lt usado = 75 Lt stock
  (107, 2, 7, 1, '2025-04-30', 100, 200, 2025, 4),     -- Fungicida: 200 Lt inicial - 100 Lt usado = 100 Lt stock
  
  (108, 2, 4, 1, '2025-05-31', 500, 800, 2025, 5),     -- Fertilizante: 800 kg stock - 300 kg usado = 500 kg stock
  (109, 2, 5, 1, '2025-05-31', 150, 300, 2025, 5),     -- Semilla: 300 kg stock - 150 kg usado = 150 kg stock
  (110, 2, 6, 1, '2025-05-31', 40, 75, 2025, 5),       -- Insecticida: 75 Lt stock - 35 Lt usado = 40 Lt stock
  (111, 2, 7, 1, '2025-05-31', 60, 100, 2025, 5),      -- Fungicida: 100 Lt stock - 40 Lt usado = 60 Lt stock
  
  (112, 2, 4, 1, '2025-06-30', 250, 500, 2025, 6),     -- Fertilizante: 500 kg stock - 250 kg usado = 250 kg stock
  (113, 2, 5, 1, '2025-06-30', 75, 150, 2025, 6),      -- Semilla: 150 kg stock - 75 kg usado = 75 kg stock
  (114, 2, 6, 1, '2025-06-30', 20, 40, 2025, 6),       -- Insecticida: 40 Lt stock - 20 Lt usado = 20 Lt stock
  (115, 2, 7, 1, '2025-06-30', 30, 60, 2025, 6),       -- Fungicida: 60 Lt stock - 30 Lt usado = 30 Lt stock

  -- PROYECTO 3: Inventarios adicionales (Mayo 2025 - Julio 2025)
  (116, 3, 8, 1, '2025-05-31', 600, 1200, 2025, 5),    -- Fertilizante: 1,200 kg inicial - 600 kg usado = 600 kg stock
  (117, 3, 9, 1, '2025-05-31', 400, 800, 2025, 5),     -- Semilla: 800 kg inicial - 400 kg usado = 400 kg stock
  (118, 3, 10, 1, '2025-05-31', 80, 160, 2025, 5),     -- Adyuvante: 160 Lt inicial - 80 Lt usado = 80 Lt stock
  
  (119, 3, 8, 1, '2025-06-30', 350, 600, 2025, 6),     -- Fertilizante: 600 kg stock - 250 kg usado = 350 kg stock
  (120, 3, 9, 1, '2025-06-30', 200, 400, 2025, 6),     -- Semilla: 400 kg stock - 200 kg usado = 200 kg stock
  (121, 3, 10, 1, '2025-06-30', 50, 80, 2025, 6),      -- Adyuvante: 80 Lt stock - 30 Lt usado = 50 Lt stock
  
  (122, 3, 8, 1, '2025-07-31', 200, 350, 2025, 7),     -- Fertilizante: 350 kg stock - 150 kg usado = 200 kg stock
  (123, 3, 9, 1, '2025-07-31', 100, 200, 2025, 7),     -- Semilla: 200 kg stock - 100 kg usado = 100 kg stock
  (124, 3, 10, 1, '2025-07-31', 25, 50, 2025, 7),      -- Adyuvante: 50 Lt stock - 25 Lt usado = 25 Lt stock

  -- PROYECTO 4: Inventarios adicionales (Marzo 2025 - Mayo 2025)
  (125, 4, 11, 1, '2025-03-31', 700, 1400, 2025, 3),   -- Fertilizante: 1,400 kg inicial - 700 kg usado = 700 kg stock
  (126, 4, 12, 1, '2025-03-31', 500, 1000, 2025, 3),   -- Semilla: 1,000 kg inicial - 500 kg usado = 500 kg stock
  (127, 4, 13, 1, '2025-03-31', 120, 240, 2025, 3),    -- Herbicida: 240 Lt inicial - 120 Lt usado = 120 Lt stock
  (128, 4, 14, 1, '2025-03-31', 90, 180, 2025, 3),     -- Insecticida: 180 Lt inicial - 90 Lt usado = 90 Lt stock
  
  (129, 4, 11, 1, '2025-04-30', 400, 700, 2025, 4),    -- Fertilizante: 700 kg stock - 300 kg usado = 400 kg stock
  (130, 4, 12, 1, '2025-04-30', 250, 500, 2025, 4),    -- Semilla: 500 kg stock - 250 kg usado = 250 kg stock
  (131, 4, 13, 1, '2025-04-30', 60, 120, 2025, 4),     -- Herbicida: 120 Lt stock - 60 Lt usado = 60 Lt stock
  (132, 4, 14, 1, '2025-04-30', 45, 90, 2025, 4),      -- Insecticida: 90 Lt stock - 45 Lt usado = 45 Lt stock
  
  (133, 4, 11, 1, '2025-05-31', 200, 400, 2025, 5),    -- Fertilizante: 400 kg stock - 200 kg usado = 200 kg stock
  (134, 4, 12, 1, '2025-05-31', 125, 250, 2025, 5),    -- Semilla: 250 kg stock - 125 kg usado = 125 kg stock
  (135, 4, 13, 1, '2025-05-31', 30, 60, 2025, 5),      -- Herbicida: 60 Lt stock - 30 Lt usado = 30 Lt stock
  (136, 4, 14, 1, '2025-05-31', 20, 45, 2025, 5);      -- Insecticida: 45 Lt stock - 25 Lt usado = 20 Lt stock

-- ========================================
-- COMERCIALIZACIONES DE CULTIVOS
-- ========================================
-- Datos de comercialización para activar cálculos de ingresos netos y arriendos
-- Precios realistas del mercado argentino (USD/ton)
INSERT INTO crop_commercializations (project_id, crop_id, board_price, freight_cost, commercial_cost, net_price, created_at) VALUES
  -- Proyecto 1: Soja y Maíz
  (1, 1, 450.00, 25.00, 15.00, 410.00, '2024-10-01'),  -- Soja: $450 - $25 - $15 = $410/ton
  (1, 2, 280.00, 20.00, 12.00, 248.00, '2024-10-01'),  -- Maíz: $280 - $20 - $12 = $248/ton
  
  -- Proyecto 2: Soja y Maíz
  (2, 1, 450.00, 25.00, 15.00, 410.00, '2024-06-01'),  -- Soja: $450 - $25 - $15 = $410/ton
  (2, 2, 280.00, 20.00, 12.00, 248.00, '2024-06-01'),  -- Maíz: $280 - $20 - $12 = $248/ton
  (2, 3, 320.00, 22.00, 14.00, 284.00, '2024-06-01'),  -- Trigo: $320 - $22 - $14 = $284/ton
  
  -- Proyecto 3: Soja y Maíz (para futuras siembras)
  (3, 1, 450.00, 25.00, 15.00, 410.00, '2024-08-01'),  -- Soja: $450 - $25 - $15 = $410/ton
  (3, 2, 280.00, 20.00, 12.00, 248.00, '2024-08-01'),  -- Maíz: $280 - $20 - $12 = $248/ton
  
  -- Proyecto 4: Soja y Maíz
  (4, 1, 450.00, 25.00, 15.00, 410.00, '2024-05-01'),  -- Soja: $450 - $25 - $15 = $410/ton
  (4, 2, 280.00, 20.00, 12.00, 248.00, '2024-05-01');  -- Maíz: $280 - $20 - $12 = $248/ton

-- ========================================
-- VALORES DEL DÓLAR PARA PROYECTO 1
-- ========================================
-- Proyecto 1: Valores del dólar para diferentes meses
INSERT INTO project_dollar_values (project_id, year, month, start_value, end_value, average_value) VALUES
  (1, 2024, '01', 850.00, 900.00, 875.00),  -- Enero 2024
  (1, 2024, '02', 900.00, 950.00, 925.00),  -- Febrero 2024
  (1, 2024, '03', 950.00, 1000.00, 975.00), -- Marzo 2024
  (1, 2024, '04', 1000.00, 1050.00, 1025.00), -- Abril 2024
  (1, 2024, '05', 1050.00, 1100.00, 1075.00), -- Mayo 2024
  (1, 2024, '06', 1100.00, 1150.00, 1125.00), -- Junio 2024
  (1, 2024, '07', 1150.00, 1200.00, 1175.00), -- Julio 2024
  (1, 2024, '08', 1200.00, 1250.00, 1225.00), -- Agosto 2024
  (1, 2024, '09', 1250.00, 1300.00, 1275.00), -- Septiembre 2024
  (1, 2024, '10', 1300.00, 1350.00, 1325.00), -- Octubre 2024
  (1, 2024, '11', 1350.00, 1400.00, 1375.00), -- Noviembre 2024
  (1, 2024, '12', 1400.00, 1450.00, 1425.00), -- Diciembre 2024
  (1, 2025, '01', 1450.00, 1500.00, 1475.00), -- Enero 2025
  (1, 2025, '02', 1500.00, 1550.00, 1525.00), -- Febrero 2025
  (1, 2025, '03', 1550.00, 1600.00, 1575.00); -- Marzo 2025

-- ========================================
-- FACTURAS PARA PROYECTO 1
-- ========================================
-- Proyecto 1: Facturas de labores ejecutadas
-- Workorders del Proyecto 1: 1 (Siembra), 2 (Fertilización), 3 (Cosecha)
INSERT INTO invoices (id, work_order_id, number, company, date, status) VALUES
  (1, 1, 'INV-2024-001', 'Contratista 1', '2024-10-20', 'paid'),    -- Siembra Lote A1
  (2, 2, 'INV-2024-002', 'Contratista 3', '2024-11-15', 'paid'),    -- Fertilización Lote A1
  (3, 3, 'INV-2024-003', 'Contratista 2', '2025-03-25', 'pending'); -- Cosecha Lote A1

-- ========================================
-- FACTURA CON INGRESOS PARA PROYECTO 4
-- ========================================
-- Proyecto 4: Factura de venta de cosecha (120 ha × 2.5 ton/ha × $400/ton = $120,000)
-- Usar work_order_id de las cosechas del Proyecto 4 (workorders 12 y 15)
INSERT INTO invoices (id, work_order_id, number, company, date, status) VALUES
  (4, 12, 'INV-2024-004', 'Empresa Demo', '2024-12-01', 'paid'),  -- Cosecha Lote D1
  (5, 15, 'INV-2024-005', 'Empresa Demo', '2024-12-01', 'paid');  -- Cosecha Lote D2

-- ========================================
-- VERIFICACIÓN DE CÁLCULOS DE ARRIENDO CON COMERCIALIZACIONES
-- ========================================
-- Verificar que los cálculos de arriendo funcionen con los nuevos datos de comercialización
SELECT '=== VERIFICACIÓN DE CÁLCULOS DE ARRIENDO ===' as info;

-- Verificar datos de comercialización cargados
SELECT '=== DATOS DE COMERCIALIZACIÓN ===' as info;
SELECT 
  project_id,
  crop_id,
  board_price,
  freight_cost,
  commercial_cost,
  net_price,
  created_at
FROM crop_commercializations 
ORDER BY project_id, crop_id;

-- Verificar cálculos de lotes con ingresos netos
SELECT '=== CÁLCULOS DE LOTES CON INGRESOS NETOS ===' as info;
SELECT 
  l.id as lot_id,
  f.project_id,
  l.current_crop_id,
  l.hectares,
  l.tons,
  CAST(l.tons AS DECIMAL(10,2)) / NULLIF(CAST(l.hectares AS DECIMAL(10,2)), 0) as yield_tonha,
  cc.net_price as net_price_usd,
  (CAST(l.tons AS DECIMAL(10,2)) / NULLIF(CAST(l.hectares AS DECIMAL(10,2)), 0)) * CAST(cc.net_price AS DECIMAL(10,2)) as income_net_per_ha,
  lt.name as lease_type_name,
  f.lease_type_percent,
  f.lease_type_value
FROM lots l
LEFT JOIN fields f ON l.field_id = f.id
LEFT JOIN lease_types lt ON f.lease_type_id = lt.id
LEFT JOIN crop_commercializations cc ON cc.project_id = f.project_id AND cc.crop_id = l.current_crop_id
WHERE l.deleted_at IS NULL
ORDER BY f.project_id, l.id;

-- ========================================
-- VER TODAS LAS COLUMNAS DISPONIBLES
-- ========================================
SELECT '=== VER TODAS LAS COLUMNAS DISPONIBLES ===' as info;
SELECT * FROM dashboard_sowing_progress_view_v2 WHERE project_id = 1 LIMIT 1;

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
FROM dashboard_sowing_progress_view_v2
WHERE project_id IN (1)
ORDER BY project_id;

-- ========================================
-- MÓDULO 2: AVANCE DE COSTOS
-- ========================================
-- RESULTADOS REALES:
-- Proyecto 1: $92,500 ejecutados (labores $22,500 + insumos $70,000)
-- Proyecto 2: $138,750 ejecutados (labores $33,750 + insumos $105,000)
-- Proyecto 3: $70,000 ejecutados (labores $20,000 + insumos $50,000) - CORREGIDO
-- Proyecto 4: $111,000 ejecutados (labores $27,000 + insumos $84,000)
SELECT '=== MÓDULO 2: AVANCE DE COSTOS ===' as info;
SELECT 
  project_id, 
  executed_costs_usd,
  budget_cost_usd,
  costs_progress_pct
FROM dashboard_costs_progress_view_v2
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
FROM dashboard_harvest_progress_view_v2
WHERE project_id IN (1,2,3)
ORDER BY project_id;

-- ========================================
-- MÓDULO 4: RESULTADO OPERATIVO
-- ========================================
-- RESULTADOS REALES (después de aplicar migración 000061):
-- Proyecto 1: $48,000 ingresos - $92,500 costos directos - $1,000 admin = -$45,500 (-48.7%)
-- Proyecto 2: $5,000 ingresos - $138,750 costos directos - $500 admin = -$134,250 (-96.4%)
-- Proyecto 3: $15,000 ingresos - $70,000 costos directos - $750 admin = -$55,750 (-78.7%) - CORREGIDO
-- Proyecto 4: $38,400 ingresos - $111,000 costos directos - $800 admin = -$73,400 (-65.7%)
-- NOTA: Los ingresos se calculan en la vista dashboard_operating_result_view_v2 según tipo de arriendo
SELECT '=== MÓDULO 4: RESULTADO OPERATIVO ===' as info;
SELECT 
  project_id, 
  operating_result_usd,
  operating_result_total_costs_usd,
  operating_result_pct
FROM dashboard_operating_result_view_v2
ORDER BY project_id;

-- ========================================
-- MÓDULO 5: AVANCE DE APORTES
-- ========================================
-- RESULTADOS REALES:
-- Proyecto 1: 3 inversores (40% + 35% + 25% = 100%)
-- Proyecto 2: 2 inversores (60% + 40% = 100%)
-- Proyecto 3: 1 inversor (100%)
-- Proyecto 4: 2 inversores (70% + 30% = 100%)
SELECT '=== MÓDULO 5: AVANCE DE APORTES ===' as info;
SELECT 
  project_id, 
  investor_id,
  investor_name,
  investor_percentage_pct,
  contributions_progress_pct
FROM dashboard_contributions_progress_view_v2
WHERE project_id IN (1,2,3,4)
ORDER BY project_id;

-- ========================================
-- MÓDULO 6: BALANCE DE GESTIÓN
-- ========================================
-- RESULTADOS REALES:
-- Proyecto 1: Semillas $50,000, Insumos $20,000, Labores $22,500 = $92,500 total + $1,000 estructura = $93,500
-- Proyecto 2: Semillas $75,000, Insumos $30,000, Labores $33,750 = $138,750 total + $500 estructura = $139,250  
-- Proyecto 3: Semillas $50,000, Insumos $20,000, Labores $20,000 = $90,000 total + $750 estructura = $90,750 - CORREGIDO
-- Proyecto 4: Semillas $60,000, Insumos $24,000, Labores $27,000 = $111,000 total + $800 estructura = $111,800
SELECT '=== MÓDULO 6: BALANCE DE GESTIÓN ===' as info;
SELECT 
  project_id,
  income_usd,                     -- Ingresos
  costos_directos_ejecutados_usd, -- Costos Directos Ejecutados
  costos_directos_invertidos_usd, -- Costos Directos Invertidos
  arriendo_invertidos_usd,        -- Arriendo Invertidos
  estructura_invertidos_usd,      -- Estructura Invertidos
  operating_result_usd,           -- Resultado Operativo
  operating_result_pct            -- Resultado Operativo Porcentaje
FROM dashboard_management_balance_view_v2 
WHERE project_id IN (1,2,3,4)
ORDER BY project_id;

-- =============================================
-- MÓDULO 7: INCIDENCIA DE COSTOS POR CULTIVO
-- =============================================
-- RESULTADOS REALES:
-- Proyecto 1: 200 ha totales - Cultivo 1 (Soja): 100 ha (50%) $0 costos = $0/ha, Cultivo 2 (Maíz): 100 ha (50%) $92,500 costos = $925/ha
-- Proyecto 2: 150 ha totales - Cultivo 1 (Soja): 75 ha (50%) $0 costos = $0/ha, Cultivo 2 (Maíz): 75 ha (50%) $138,750 costos = $1,850/ha
-- Proyecto 3: 100 ha totales - Sin cultivos específicos, $0 costos = $0/ha
-- Proyecto 4: 120 ha totales - Cultivo 1 (Soja): 60 ha (50%) $0 costos = $0/ha, Cultivo 2 (Maíz): 60 ha (50%) $111,000 costos = $1,850/ha
SELECT '=== MÓDULO 7: INCIDENCIA DE COSTOS POR CULTIVO ===' as info;
SELECT 
  project_id, 
  current_crop_id,
  crop_name,
  crop_hectares,
  crop_incidence_pct
FROM dashboard_crop_incidence_view_v2 
ORDER BY project_id;

-- ========================================
-- MÓDULO 8: INDICADORES OPERATIVOS
-- ========================================
-- RESULTADOS REALES:
-- Proyecto 1: Primera orden 2024-10-15, Última orden 2025-03-20, Arqueo stock 2024-11-10
-- Proyecto 2: Primera orden 2024-06-01, Última orden 2024-12-20, Arqueo stock 2024-07-01
-- Proyecto 3: Sin órdenes de trabajo, sin arqueo de stock
-- Proyecto 4: Primera orden 2024-05-01, Última orden 2024-11-20, Arqueo stock 2024-06-01
SELECT '=== MÓDULO 8: INDICADORES OPERATIVOS ===' as info;
SELECT 
  project_id,
  start_date,                  -- Fecha de inicio
  end_date,                    -- Fecha de fin
  campaign_closing_date        -- Fecha de cierre de campaña
FROM dashboard_operational_indicators_view_v2 
WHERE project_id IN (1,2,3,4)
ORDER BY project_id;

-- ========================================
-- VERIFICACIÓN DEL EJEMPLO PROYECTO PRUEBA AGUS
-- ========================================
SELECT '=== VERIFICACIÓN PROYECTO PRUEBA AGUS ===' as info;

-- Verificar datos del proyecto
SELECT '=== DATOS DEL PROYECTO ===' as info;
SELECT 
  p.id as project_id,
  p.name as project_name,
  f.id as field_id,
  f.name as field_name,
  l.id as lot_id,
  l.name as lot_name,
  l.hectares
FROM projects p
LEFT JOIN fields f ON f.project_id = p.id
LEFT JOIN lots l ON l.field_id = f.id
WHERE p.id = 5
ORDER BY f.id, l.id;

-- Verificar órdenes de trabajo
SELECT '=== ÓRDENES DE TRABAJO ===' as info;
SELECT 
  w.id as workorder_id,
  w.project_id,
  w.field_id,
  w.lot_id,
  w.labor_id,
  l.name as labor_name,
  l.price as labor_price,
  w.effective_area,
  (l.price * w.effective_area) as labor_cost
FROM workorders w
LEFT JOIN labors l ON l.id = w.labor_id
WHERE w.project_id = 5
ORDER BY w.id;

-- Verificar items de órdenes de trabajo
SELECT '=== ITEMS DE ÓRDENES DE TRABAJO ===' as info;
SELECT 
  wi.workorder_id,
  wi.supply_id,
  s.name as supply_name,
  s.price as supply_price,
  wi.final_dose,
  wi.total_used,
  (wi.final_dose * s.price) as supply_cost_per_ha
FROM workorder_items wi
LEFT JOIN supplies s ON s.id = wi.supply_id
WHERE wi.workorder_id IN (101, 102)
ORDER BY wi.workorder_id;

-- Verificar métricas del proyecto
SELECT '=== MÉTRICAS DEL PROYECTO ===' as info;
SELECT 
  project_id,
  surface_ha,
  liters,
  kilograms,
  direct_cost
FROM workorder_metrics_view_v2
WHERE project_id = 5;

-- ========================================
-- VERIFICACIÓN DE APP_PARAMETERS
-- ========================================
SELECT '=== VERIFICACIÓN DE APP_PARAMETERS ===' as info;

-- Verificar parámetros cargados
SELECT '=== PARÁMETROS CARGADOS ===' as info;
SELECT 
  key,
  value,
  type,
  category,
  description
FROM app_parameters
ORDER BY category, key;

-- Verificar funciones SQL
SELECT '=== FUNCIONES SQL ===' as info;
SELECT 
  'get_iva_percentage()' as function_name,
  get_iva_percentage() as result;

SELECT 
  'get_campaign_closure_days()' as function_name,
  get_campaign_closure_days() as result;

SELECT 
  'get_default_fx_rate()' as function_name,
  get_default_fx_rate() as result;

-- Verificar uso en supplies
SELECT '=== USO EN SUPPLIES ===' as info;
SELECT 
  s.id,
  s.name,
  s.unit_id,
  CASE s.unit_id
    WHEN 1 THEN 'Lt (unit_liters)'
    WHEN 2 THEN 'Kg (unit_kilos)'
    WHEN 3 THEN 'Ha (unit_hectares)'
    ELSE 'Unknown'
  END as unit_description,
  s.price,
  s.price || ' USD/' || CASE s.unit_id
    WHEN 1 THEN 'Lt'
    WHEN 2 THEN 'Kg'
    WHEN 3 THEN 'Ha'
    ELSE 'Unknown'
  END as price_per_unit
FROM supplies s
ORDER BY s.unit_id, s.id;

-- Verificar uso en workorder_items con diferentes unidades
SELECT '=== USO EN WORKORDER_ITEMS ===' as info;
SELECT 
  wi.workorder_id,
  s.name as supply_name,
  CASE s.unit_id
    WHEN 1 THEN 'Lt (unit_liters)'
    WHEN 2 THEN 'Kg (unit_kilos)'
    WHEN 3 THEN 'Ha (unit_hectares)'
    ELSE 'Unknown'
  END as unit_description,
  wi.final_dose,
  wi.total_used,
  wi.final_dose || ' ' || CASE s.unit_id
    WHEN 1 THEN 'Lt/ha'
    WHEN 2 THEN 'Kg/ha'
    WHEN 3 THEN 'Ha/ha'
    ELSE 'Unknown/ha'
  END as dose_per_hectare,
  wi.total_used || ' ' || CASE s.unit_id
    WHEN 1 THEN 'Lt'
    WHEN 2 THEN 'Kg'
    WHEN 3 THEN 'Ha'
    ELSE 'Unknown'
  END as total_used_units
FROM workorder_items wi
LEFT JOIN supplies s ON s.id = wi.supply_id
ORDER BY s.unit_id, wi.workorder_id;

-- ========================================
-- VERIFICACIÓN DE STOCKS
-- ========================================
SELECT '=== VERIFICACIÓN DE STOCKS ===' as info;

-- Verificar stocks cargados por proyecto
SELECT '=== STOCKS POR PROYECTO ===' as info;
SELECT 
  s.project_id,
  p.name as project_name,
  COUNT(*) as total_stock_records,
  MIN(s.close_date) as first_inventory_date,
  MAX(s.close_date) as last_inventory_date
FROM stocks s
LEFT JOIN projects p ON p.id = s.project_id
GROUP BY s.project_id, p.name
ORDER BY s.project_id;

-- Verificar stocks por tipo de insumo
SELECT '=== STOCKS POR TIPO DE INSUMO ===' as info;
SELECT 
  s.supply_id,
  sup.name as supply_name,
  CASE sup.unit_id
    WHEN 1 THEN 'Lt (unit_liters)'
    WHEN 2 THEN 'Kg (unit_kilos)'
    WHEN 3 THEN 'Ha (unit_hectares)'
    ELSE 'Unknown'
  END as unit_description,
  COUNT(*) as stock_records,
  SUM(s.initial_units) as total_initial_units,
  SUM(s.real_stock_units) as total_real_stock_units,
  SUM(s.initial_units) - SUM(s.real_stock_units) as total_consumed_units
FROM stocks s
LEFT JOIN supplies sup ON sup.id = s.supply_id
GROUP BY s.supply_id, sup.name, sup.unit_id
ORDER BY sup.unit_id, s.supply_id;

-- Verificar último inventario por proyecto (para campaign_closing)
SELECT '=== ÚLTIMO INVENTARIO POR PROYECTO ===' as info;
SELECT 
  s.project_id,
  p.name as project_name,
  MAX(s.close_date) as last_stock_count_date,
  COUNT(DISTINCT s.supply_id) as different_supplies_inventoried,
  SUM(s.real_stock_units) as total_remaining_stock
FROM stocks s
LEFT JOIN projects p ON p.id = s.project_id
WHERE s.close_date = (
  SELECT MAX(s2.close_date) 
  FROM stocks s2 
  WHERE s2.project_id = s.project_id
)
GROUP BY s.project_id, p.name
ORDER BY s.project_id;