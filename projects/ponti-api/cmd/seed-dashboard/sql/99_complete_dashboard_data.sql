-- =======================
-- DATOS COMPLETOS DEL DASHBOARD
-- =======================
-- Script consolidado que inserta todos los datos necesarios para el dashboard
-- Ejecutar este script DESPUÉS de los scripts base (01, 02, 03, 04)

-- 1. CREAR TIPOS BASE (si no existen)
INSERT INTO types (name, created_at, updated_at) VALUES 
('Tipo Semillas', NOW(), NOW()),
('Tipo Fertilizantes', NOW(), NOW()),
('Tipo Agroquímicos', NOW(), NOW()),
('Tipo Labores', NOW(), NOW())
ON CONFLICT DO NOTHING;

-- 2. CREAR CATEGORÍAS (si no existen)
INSERT INTO categories (name, type_id, created_at, updated_at) VALUES 
('Semillas', (SELECT id FROM types WHERE name = 'Tipo Semillas' LIMIT 1), NOW(), NOW()),
('Fertilizantes', (SELECT id FROM types WHERE name = 'Tipo Fertilizantes' LIMIT 1), NOW(), NOW()),
('Agroquímicos', (SELECT id FROM types WHERE name = 'Tipo Agroquímicos' LIMIT 1), NOW(), NOW()),
('Labores', (SELECT id FROM types WHERE name = 'Tipo Labores' LIMIT 1), NOW(), NOW())
ON CONFLICT DO NOTHING;

-- 3. CREAR TIPOS DE LABOR (si no existen)
INSERT INTO labor_types (name, created_at, updated_at) VALUES 
('Tipo Siembra', NOW(), NOW()),
('Tipo Mantenimiento', NOW(), NOW()),
('Tipo Cosecha', NOW(), NOW())
ON CONFLICT DO NOTHING;

-- 4. CREAR CATEGORÍAS DE LABOR (si no existen)
INSERT INTO labor_categories (name, type_id, created_at, updated_at) VALUES 
('Siembra', (SELECT id FROM labor_types WHERE name = 'Tipo Siembra' LIMIT 1), NOW(), NOW()),
('Mantenimiento', (SELECT id FROM labor_types WHERE name = 'Tipo Mantenimiento' LIMIT 1), NOW(), NOW()),
('Cosecha', (SELECT id FROM labor_types WHERE name = 'Tipo Cosecha' LIMIT 1), NOW(), NOW())
ON CONFLICT DO NOTHING;

-- 5. CREAR LABORES CON PRECIOS
INSERT INTO labors (project_id, name, category_id, price, contractor_name, created_at, updated_at) VALUES 
(1, 'Siembra directa', (SELECT id FROM labor_categories WHERE name = 'Siembra' LIMIT 1), 25.00, 'Contratista A', NOW(), NOW()),
(1, 'Fertilización', (SELECT id FROM labor_categories WHERE name = 'Mantenimiento' LIMIT 1), 15.00, 'Contratista B', NOW(), NOW()),
(1, 'Pulverización', (SELECT id FROM labor_categories WHERE name = 'Mantenimiento' LIMIT 1), 20.00, 'Contratista C', NOW(), NOW()),
(1, 'Cosecha', (SELECT id FROM labor_categories WHERE name = 'Cosecha' LIMIT 1), 30.00, 'Contratista D', NOW(), NOW())
ON CONFLICT DO NOTHING;

-- 6. CREAR INSUMOS CON PRECIOS
INSERT INTO supplies (project_id, name, price, type_id, created_at, updated_at) VALUES 
(1, 'Semilla Soja', 150.00, (SELECT id FROM types WHERE name = 'Tipo Semillas' LIMIT 1), NOW(), NOW()),
(1, 'Fertilizante NPK', 80.00, (SELECT id FROM types WHERE name = 'Tipo Fertilizantes' LIMIT 1), NOW(), NOW()),
(1, 'Herbicida', 45.00, (SELECT id FROM types WHERE name = 'Tipo Agroquímicos' LIMIT 1), NOW(), NOW()),
(1, 'Fungicida', 60.00, (SELECT id FROM types WHERE name = 'Tipo Agroquímicos' LIMIT 1), NOW(), NOW())
ON CONFLICT DO NOTHING;

-- 7. CREAR INVERSORES
INSERT INTO investors (name, email, created_at, updated_at) VALUES 
('Inversor A', 'inversor.a@email.com', NOW(), NOW()),
('Inversor B', 'inversor.b@email.com', NOW(), NOW()),
('Inversor C', 'inversor.c@email.com', NOW(), NOW())
ON CONFLICT DO NOTHING;

-- 8. VINCULAR INVERSORES AL PROYECTO
INSERT INTO project_investors (project_id, investor_id, investment_amount, created_at, updated_at) VALUES 
(1, (SELECT id FROM investors WHERE name = 'Inversor A' LIMIT 1), 5000.00, NOW(), NOW()),
(1, (SELECT id FROM investors WHERE name = 'Inversor B' LIMIT 1), 3000.00, NOW(), NOW()),
(1, (SELECT id FROM investors WHERE name = 'Inversor C' LIMIT 1), 2000.00, NOW(), NOW())
ON CONFLICT DO NOTHING;

-- 9. CREAR WORKORDERS CON DATOS REALES
INSERT INTO workorders (number, project_id, field_id, lot_id, crop_id, labor_id, contractor, date, investor_id, effective_area, created_at, updated_at) VALUES 
('WO-001', 1, 1, 1, 1, (SELECT id FROM labors WHERE name = 'Siembra directa' LIMIT 1), 'Contratista A', '2024-01-15', 1, 10.5, NOW(), NOW()),
('WO-002', 1, 1, 1, 1, (SELECT id FROM labors WHERE name = 'Fertilización' LIMIT 1), 'Contratista B', '2024-02-01', 1, 10.5, NOW(), NOW()),
('WO-003', 1, 1, 1, 1, (SELECT id FROM labors WHERE name = 'Pulverización' LIMIT 1), 'Contratista C', '2024-03-01', 1, 10.5, NOW(), NOW()),
('WO-004', 1, 1, 1, 1, (SELECT id FROM labors WHERE name = 'Cosecha' LIMIT 1), 'Contratista D', '2024-06-01', 1, 10.5, NOW(), NOW()),
('WO-005', 1, 2, 2, 2, (SELECT id FROM labors WHERE name = 'Siembra directa' LIMIT 1), 'Contratista A', '2024-01-20', 1, 8.5, NOW(), NOW()),
('WO-006', 1, 2, 2, 2, (SELECT id FROM labors WHERE name = 'Fertilización' LIMIT 1), 'Contratista B', '2024-02-05', 1, 8.5, NOW(), NOW()),
('WO-007', 1, 3, 3, 3, (SELECT id FROM labors WHERE name = 'Siembra directa' LIMIT 1), 'Contratista A', '2024-01-25', 1, 12.0, NOW(), NOW()),
('WO-008', 1, 4, 4, 4, (SELECT id FROM labors WHERE name = 'Siembra directa' LIMIT 1), 'Contratista A', '2024-02-01', 1, 15.0, NOW(), NOW())
ON CONFLICT DO NOTHING;

-- 10. CREAR DETALLES DE WORKORDERS PARA INSUMOS
INSERT INTO workorder_items (workorder_id, supply_id, total_used, final_dose, created_at, updated_at) VALUES 
-- WO-001: Siembra Soja - Semilla
((SELECT id FROM workorders WHERE number = 'WO-001' LIMIT 1), (SELECT id FROM supplies WHERE name = 'Semilla Soja' LIMIT 1), 10.5, 150.00, NOW(), NOW()),

-- WO-002: Fertilización Soja - Fertilizante
((SELECT id FROM workorders WHERE number = 'WO-002' LIMIT 1), (SELECT id FROM supplies WHERE name = 'Fertilizante NPK' LIMIT 1), 10.5, 80.00, NOW(), NOW()),

-- WO-003: Pulverización Soja - Herbicida
((SELECT id FROM workorders WHERE number = 'WO-003' LIMIT 1), (SELECT id FROM supplies WHERE name = 'Herbicida' LIMIT 1), 10.5, 45.00, NOW(), NOW()),

-- WO-005: Siembra Maíz - Semilla
((SELECT id FROM workorders WHERE number = 'WO-005' LIMIT 1), (SELECT id FROM supplies WHERE name = 'Semilla Soja' LIMIT 1), 8.5, 120.00, NOW(), NOW()),

-- WO-006: Fertilización Maíz - Fertilizante
((SELECT id FROM workorders WHERE number = 'WO-006' LIMIT 1), (SELECT id FROM supplies WHERE name = 'Fertilizante NPK' LIMIT 1), 8.5, 90.00, NOW(), NOW()),

-- WO-007: Siembra Trigo - Semilla
((SELECT id FROM workorders WHERE number = 'WO-007' LIMIT 1), (SELECT id FROM supplies WHERE name = 'Semilla Soja' LIMIT 1), 12.0, 100.00, NOW(), NOW()),

-- WO-008: Siembra Girasol - Semilla
((SELECT id FROM workorders WHERE number = 'WO-008' LIMIT 1), (SELECT id FROM supplies WHERE name = 'Semilla Soja' LIMIT 1), 15.0, 80.00, NOW(), NOW())
ON CONFLICT DO NOTHING;

-- 11. CREAR STOCK CON INVENTARIO
INSERT INTO stocks (supply_id, quantity, unit_cost, total_cost, created_at, updated_at) VALUES 
((SELECT id FROM supplies WHERE name = 'Semilla Soja' LIMIT 1), 100.0, 150.00, 15000.00, NOW(), NOW()),
((SELECT id FROM supplies WHERE name = 'Fertilizante NPK' LIMIT 1), 50.0, 80.00, 4000.00, NOW(), NOW()),
((SELECT id FROM supplies WHERE name = 'Herbicida' LIMIT 1), 25.0, 45.00, 1125.00, NOW(), NOW()),
((SELECT id FROM supplies WHERE name = 'Fungicida' LIMIT 1), 20.0, 60.00, 1200.00, NOW(), NOW())
ON CONFLICT DO NOTHING;

-- 12. CREAR FACTURAS CON INGRESOS
INSERT INTO invoices (number, project_id, amount, status, due_date, created_at, updated_at) VALUES 
('INV-001', 1, 5000.00, 'paid', '2024-06-30', NOW(), NOW()),
('INV-002', 1, 3000.00, 'pending', '2024-07-30', NOW(), NOW()),
('INV-003', 1, 2000.00, 'draft', '2024-08-30', NOW(), NOW())
ON CONFLICT DO NOTHING;

-- 13. ACTUALIZAR LOTS CON DATOS DE SIEMBRA REALES
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

-- 14. ACTUALIZAR PROJECTS CON COSTOS REALES
UPDATE projects SET 
    budget_cost = 15000.00,
    admin_cost = 2000.00
WHERE id = 1;

-- VERIFICAR INSERCIÓN DE DATOS
SELECT '✅ Datos completos del dashboard cargados:' as status;
SELECT '   - Workorders totales:' as item, COUNT(*) as total FROM workorders;
SELECT '   - Supplies totales:' as item, COUNT(*) as total FROM supplies;
SELECT '   - Labors totales:' as item, COUNT(*) as total FROM labors;
SELECT '   - Investors totales:' as item, COUNT(*) as total FROM investors;
SELECT '   - Stock total:' as item, SUM(total_cost) as total FROM stocks;
SELECT '   - Invoices total:' as item, SUM(amount) as total FROM invoices;
SELECT '   - Lotes sembrados:' as item, COUNT(*) as total FROM lots WHERE sowed_area > 0;

-- VERIFICAR MÉTRICAS ESPERADAS
SELECT '🎯 Dashboard ahora debería mostrar:' as info;
SELECT '   - Siembra: 46.0 ha / 46.0 ha (100%)' as sowing_info;
SELECT '   - Costos ejecutados: ~$7,000 / $15,000 (46.7%)' as costs_info;
SELECT '   - Ingresos: $5,000 (facturas pagadas)' as income_info;
SELECT '   - Stock: ~$21,325' as stock_info;
SELECT '   - Contribuciones: $10,000 (3 inversores)' as contrib_info;
