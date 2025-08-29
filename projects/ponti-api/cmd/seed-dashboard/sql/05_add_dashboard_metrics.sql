-- =======================
-- AGREGAR MÉTRICAS REALES AL DASHBOARD
-- =======================
-- Script para agregar datos reales que permitan probar las métricas del dashboard
-- Incluye: workorders, supplies, labors, stocks, investors, invoices

-- 1. AGREGAR INSUMOS (SUPPLIES) - Para métricas de costos
INSERT INTO supplies (project_id, name, price, type_id, created_at, updated_at) VALUES 
(1, 'Semilla Soja', 150.00, 1, NOW(), NOW()),
(1, 'Fertilizante NPK', 80.00, 4, NOW(), NOW()),
(1, 'Herbicida', 45.00, 5, NOW(), NOW()),
(1, 'Fungicida', 60.00, 6, NOW(), NOW())
ON CONFLICT DO NOTHING;

-- 2. AGREGAR LABORES (LABORS) - Para métricas de costos
INSERT INTO labors (name, labor_type_id, created_at, updated_at) VALUES 
('Siembra directa', 1, NOW(), NOW()),
('Fertilización', 2, NOW(), NOW()),
('Pulverización', 3, NOW(), NOW()),
('Cosecha', 4, NOW(), NOW())
ON CONFLICT DO NOTHING;

-- 3. AGREGAR INVERSORES - Para métricas de contribuciones
INSERT INTO investors (name, email, created_at, updated_at) VALUES 
('Inversor A', 'inversor.a@email.com', NOW(), NOW()),
('Inversor B', 'inversor.b@email.com', NOW(), NOW()),
('Inversor C', 'inversor.c@email.com', NOW(), NOW())
ON CONFLICT DO NOTHING;

-- 4. VINCULAR INVERSORES AL PROYECTO - Para métricas de contribuciones
INSERT INTO project_investors (project_id, investor_id, investment_amount, created_at, updated_at) VALUES 
(1, 1, 5000.00, NOW(), NOW()),
(1, 2, 3000.00, NOW(), NOW()),
(1, 3, 2000.00, NOW(), NOW())
ON CONFLICT DO NOTHING;

-- 5. AGREGAR WORKORDERS - Para métricas operativas y costos
-- Nota: workorders ya tiene labor_id directamente, no necesitamos workorder_items para labores
INSERT INTO workorders (number, project_id, field_id, lot_id, crop_id, labor_id, contractor, date, investor_id, effective_area, created_at, updated_at) VALUES 
('WO-001', 1, 1, 1, 1, 1, 'Contratista A', '2024-01-15', 1, 10.5, NOW(), NOW()), -- Siembra Soja
('WO-002', 1, 1, 1, 1, 2, 'Contratista B', '2024-02-01', 1, 10.5, NOW(), NOW()), -- Fertilización Soja
('WO-003', 1, 1, 1, 1, 3, 'Contratista C', '2024-03-01', 1, 10.5, NOW(), NOW()), -- Pulverización Soja
('WO-004', 1, 1, 1, 1, 4, 'Contratista D', '2024-06-01', 1, 10.5, NOW(), NOW()), -- Cosecha Soja
('WO-005', 1, 2, 2, 2, 1, 'Contratista A', '2024-01-20', 1, 8.5, NOW(), NOW()),  -- Siembra Maíz
('WO-006', 1, 2, 2, 2, 2, 'Contratista B', '2024-02-05', 1, 8.5, NOW(), NOW()),  -- Fertilización Maíz
('WO-007', 1, 3, 3, 3, 1, 'Contratista A', '2024-01-25', 1, 12.0, NOW(), NOW()), -- Siembra Trigo
('WO-008', 1, 4, 4, 4, 1, 'Contratista A', '2024-02-01', 1, 15.0, NOW(), NOW())  -- Siembra Girasol
ON CONFLICT DO NOTHING;

-- 6. AGREGAR DETALLES DE WORKORDERS PARA SUPPLIES - Para métricas de costos de insumos
INSERT INTO workorder_items (workorder_id, supply_id, total_used, final_dose, created_at, updated_at) VALUES 
-- WO-001: Siembra Soja - Semilla
(1, 1, 10.5, 150.00, NOW(), NOW()), -- Semilla: 10.5 ha * $150/ha = $1,575

-- WO-002: Fertilización Soja - Fertilizante
(2, 2, 10.5, 80.00, NOW(), NOW()), -- Fertilizante: 10.5 ha * $80/ha = $840

-- WO-003: Pulverización Soja - Herbicida
(3, 3, 10.5, 45.00, NOW(), NOW()), -- Herbicida: 10.5 ha * $45/ha = $472.50

-- WO-005: Siembra Maíz - Semilla
(5, 1, 8.5, 120.00, NOW(), NOW()), -- Semilla: 8.5 ha * $120/ha = $1,020

-- WO-006: Fertilización Maíz - Fertilizante
(6, 2, 8.5, 90.00, NOW(), NOW()), -- Fertilizante: 8.5 ha * $90/ha = $765

-- WO-007: Siembra Trigo - Semilla
(7, 1, 12.0, 100.00, NOW(), NOW()), -- Semilla: 12.0 ha * $100/ha = $1,200

-- WO-008: Siembra Girasol - Semilla
(8, 1, 15.0, 80.00, NOW(), NOW())  -- Semilla: 15.0 ha * $80/ha = $1,200
ON CONFLICT DO NOTHING;

-- 7. AGREGAR STOCK - Para métricas de inventario
INSERT INTO stocks (supply_id, quantity, unit_cost, total_cost, created_at, updated_at) VALUES 
(1, 100.0, 150.00, 15000.00, NOW(), NOW()), -- Semilla Soja
(2, 50.0, 80.00, 4000.00, NOW(), NOW()),   -- Fertilizante
(3, 25.0, 45.00, 1125.00, NOW(), NOW()),   -- Herbicida
(4, 20.0, 60.00, 1200.00, NOW(), NOW())    -- Fungicida
ON CONFLICT DO NOTHING;

-- 8. AGREGAR FACTURAS - Para métricas de ingresos
INSERT INTO invoices (number, project_id, amount, status, due_date, created_at, updated_at) VALUES 
('INV-001', 1, 5000.00, 'paid', '2024-06-30', NOW(), NOW()),
('INV-002', 1, 3000.00, 'pending', '2024-07-30', NOW(), NOW()),
('INV-003', 1, 2000.00, 'draft', '2024-08-30', NOW(), NOW())
ON CONFLICT DO NOTHING;

-- 9. ACTUALIZAR LOTS CON DATOS DE SIEMBRA REALES
UPDATE lots SET 
    sowed_area = hectares,
    sowed_date = '2024-01-15',
    harvest_date = '2024-06-01'
WHERE id = 1; -- Lote A1 (Soja)

UPDATE lots SET 
    sowed_area = hectares,
    sowed_date = '2024-01-20',
    harvest_date = '2024-06-15'
WHERE id = 2; -- Lote B1 (Maíz)

UPDATE lots SET 
    sowed_area = hectares,
    sowed_date = '2024-01-25',
    harvest_date = '2024-07-01'
WHERE id = 3; -- Lote C1 (Trigo)

UPDATE lots SET 
    sowed_area = hectares,
    sowed_date = '2024-02-01',
    harvest_date = '2024-07-15'
WHERE id = 4; -- Lote D1 (Girasol)

-- 10. ACTUALIZAR PROJECTS CON COSTOS REALES
UPDATE projects SET 
    budget_cost = 15000.00,
    admin_cost = 2000.00
WHERE id = 1;

-- Verificar inserción de métricas
SELECT '✅ Métricas del dashboard cargadas:' as status;
SELECT '   - Workorders totales:' as item, COUNT(*) as total FROM workorders;
SELECT '   - Supplies totales:' as item, COUNT(*) as total FROM supplies;
SELECT '   - Labors totales:' as item, COUNT(*) as total FROM labors;
SELECT '   - Investors totales:' as item, COUNT(*) as total FROM investors;
SELECT '   - Stock total:' as item, SUM(total_cost) as total FROM stocks;
SELECT '   - Invoices total:' as item, SUM(amount) as total FROM invoices;

-- Verificar métricas esperadas
SELECT '🎯 Dashboard ahora debería mostrar:' as info;
SELECT '   - Siembra: 46.0 ha / 46.0 ha (100%)' as sowing_info;
SELECT '   - Costos ejecutados: ~$7,000 / $15,000 (46.7%)' as costs_info;
SELECT '   - Ingresos: $5,000 (facturas pagadas)' as income_info;
SELECT '   - Stock: ~$21,325' as stock_info;
SELECT '   - Contribuciones: $10,000 (3 inversores)' as contrib_info;
