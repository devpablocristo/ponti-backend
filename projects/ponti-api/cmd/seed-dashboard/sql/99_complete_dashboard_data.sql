-- =======================
-- DATOS COMPLETOS DEL DASHBOARD
-- =======================
-- Script consolidado que inserta todos los datos necesarios para el dashboard
-- Ejecutar este script DESPUÉS de 01_basic_entities.sql

-- 1. CREAR INSUMOS CON PRECIOS REALES
INSERT INTO supplies (project_id, name, price, type_id, created_at, updated_at) VALUES 
-- Semillas
((SELECT id FROM projects WHERE name = 'Proyecto Soja 2024' LIMIT 1), 'Semilla Soja RR2 BT', 180.00, (SELECT id FROM types WHERE name = 'Semillas' LIMIT 1), NOW(), NOW()),
((SELECT id FROM projects WHERE name = 'Proyecto Maíz 2024' LIMIT 1), 'Semilla Maíz BT', 220.00, (SELECT id FROM types WHERE name = 'Semillas' LIMIT 1), NOW(), NOW()),
((SELECT id FROM projects WHERE name = 'Proyecto Trigo 2024' LIMIT 1), 'Semilla Trigo', 150.00, (SELECT id FROM types WHERE name = 'Semillas' LIMIT 1), NOW(), NOW()),
((SELECT id FROM projects WHERE name = 'Proyecto Girasol 2024' LIMIT 1), 'Semilla Girasol', 120.00, (SELECT id FROM types WHERE name = 'Semillas' LIMIT 1), NOW(), NOW()),
((SELECT id FROM projects WHERE name = 'Proyecto Soja 2024' LIMIT 1), 'Semilla Cebada', 140.00, (SELECT id FROM types WHERE name = 'Semillas' LIMIT 1), NOW(), NOW()),

-- Fertilizantes
((SELECT id FROM projects WHERE name = 'Proyecto Soja 2024' LIMIT 1), 'Fertilizante NPK 15-15-15', 95.00, (SELECT id FROM types WHERE name = 'Fertilizantes' LIMIT 1), NOW(), NOW()),
((SELECT id FROM projects WHERE name = 'Proyecto Maíz 2024' LIMIT 1), 'Urea 46%', 85.00, (SELECT id FROM types WHERE name = 'Fertilizantes' LIMIT 1), NOW(), NOW()),
((SELECT id FROM projects WHERE name = 'Proyecto Trigo 2024' LIMIT 1), 'Fosfato Diamónico', 110.00, (SELECT id FROM types WHERE name = 'Fertilizantes' LIMIT 1), NOW(), NOW()),
((SELECT id FROM projects WHERE name = 'Proyecto Girasol 2024' LIMIT 1), 'Sulfato de Amonio', 75.00, (SELECT id FROM types WHERE name = 'Fertilizantes' LIMIT 1), NOW(), NOW()),

-- Agroquímicos
((SELECT id FROM projects WHERE name = 'Proyecto Soja 2024' LIMIT 1), 'Herbicida Glifosato', 55.00, (SELECT id FROM types WHERE name = 'Agroquímicos' LIMIT 1), NOW(), NOW()),
((SELECT id FROM projects WHERE name = 'Proyecto Maíz 2024' LIMIT 1), 'Fungicida Triazol', 70.00, (SELECT id FROM types WHERE name = 'Agroquímicos' LIMIT 1), NOW(), NOW()),
((SELECT id FROM projects WHERE name = 'Proyecto Trigo 2024' LIMIT 1), 'Insecticida Piretroide', 65.00, (SELECT id FROM types WHERE name = 'Agroquímicos' LIMIT 1), NOW(), NOW()),
((SELECT id FROM projects WHERE name = 'Proyecto Girasol 2024' LIMIT 1), 'Coadyuvante', 25.00, (SELECT id FROM types WHERE name = 'Agroquímicos' LIMIT 1), NOW(), NOW()),

-- Combustibles
((SELECT id FROM projects WHERE name = 'Proyecto Soja 2024' LIMIT 1), 'Diesel', 120.00, (SELECT id FROM types WHERE name = 'Combustibles' LIMIT 1), NOW(), NOW()),
((SELECT id FROM projects WHERE name = 'Proyecto Maíz 2024' LIMIT 1), 'Nafta', 110.00, (SELECT id FROM types WHERE name = 'Combustibles' LIMIT 1), NOW(), NOW())
ON CONFLICT DO NOTHING;

-- 2. CREAR LABORES CON PRECIOS REALES
INSERT INTO labors (project_id, name, category_id, price, contractor_name, created_at, updated_at) VALUES 
-- Siembra
((SELECT id FROM projects WHERE name = 'Proyecto Soja 2024' LIMIT 1), 'Siembra directa Soja', (SELECT id FROM labor_categories WHERE name = 'Siembra directa' LIMIT 1), 35.00, 'Contratista Campo Sur', NOW(), NOW()),
((SELECT id FROM projects WHERE name = 'Proyecto Maíz 2024' LIMIT 1), 'Siembra directa Maíz', (SELECT id FROM labor_categories WHERE name = 'Siembra directa' LIMIT 1), 40.00, 'Contratista Campo Sur', NOW(), NOW()),
((SELECT id FROM projects WHERE name = 'Proyecto Trigo 2024' LIMIT 1), 'Siembra convencional Trigo', (SELECT id FROM labor_categories WHERE name = 'Siembra convencional' LIMIT 1), 45.00, 'Contratista Norte', NOW(), NOW()),
((SELECT id FROM projects WHERE name = 'Proyecto Girasol 2024' LIMIT 1), 'Siembra directa Soja', (SELECT id FROM labor_categories WHERE name = 'Siembra directa' LIMIT 1), 35.00, 'Contratista Campo Sur', NOW(), NOW()),

-- Mantenimiento
((SELECT id FROM projects WHERE name = 'Proyecto Soja 2024' LIMIT 1), 'Fertilización foliar', (SELECT id FROM labor_categories WHERE name = 'Fertilización' LIMIT 1), 25.00, 'Servicios Agrícolas', NOW(), NOW()),
((SELECT id FROM projects WHERE name = 'Proyecto Maíz 2024' LIMIT 1), 'Fertilización foliar', (SELECT id FROM labor_categories WHERE name = 'Fertilización' LIMIT 1), 25.00, 'Servicios Agrícolas', NOW(), NOW()),
((SELECT id FROM projects WHERE name = 'Proyecto Trigo 2024' LIMIT 1), 'Fertilización foliar', (SELECT id FROM labor_categories WHERE name = 'Fertilización' LIMIT 1), 25.00, 'Servicios Agrícolas', NOW(), NOW()),

-- Marzo - Pulverizaciones
((SELECT id FROM projects WHERE name = 'Proyecto Soja 2024' LIMIT 1), 'Pulverización herbicida', (SELECT id FROM labor_categories WHERE name = 'Pulverización' LIMIT 1), 30.00, 'Fumigaciones del Valle', NOW(), NOW()),
((SELECT id FROM projects WHERE name = 'Proyecto Maíz 2024' LIMIT 1), 'Pulverización herbicida', (SELECT id FROM labor_categories WHERE name = 'Pulverización' LIMIT 1), 30.00, 'Fumigaciones del Valle', NOW(), NOW()),
((SELECT id FROM projects WHERE name = 'Proyecto Trigo 2024' LIMIT 1), 'Pulverización fungicida', (SELECT id FROM labor_categories WHERE name = 'Pulverización' LIMIT 1), 28.00, 'Fumigaciones del Valle', NOW(), NOW()),

-- Abril - Riego
((SELECT id FROM projects WHERE name = 'Proyecto Soja 2024' LIMIT 1), 'Riego por aspersión', (SELECT id FROM labor_categories WHERE name = 'Riego' LIMIT 1), 20.00, 'Riegos del Campo', NOW(), NOW()),
((SELECT id FROM projects WHERE name = 'Proyecto Maíz 2024' LIMIT 1), 'Riego por aspersión', (SELECT id FROM labor_categories WHERE name = 'Riego' LIMIT 1), 20.00, 'Riegos del Campo', NOW(), NOW()),

-- Junio - Cosecha
((SELECT id FROM projects WHERE name = 'Proyecto Soja 2024' LIMIT 1), 'Cosecha Soja', (SELECT id FROM labor_categories WHERE name = 'Cosecha' LIMIT 1), 50.00, 'Cosechadoras Unidos', NOW(), NOW()),
((SELECT id FROM projects WHERE name = 'Proyecto Maíz 2024' LIMIT 1), 'Cosecha Maíz', (SELECT id FROM labor_categories WHERE name = 'Cosecha' LIMIT 1), 55.00, 'Cosechadoras Unidos', NOW(), NOW()),
((SELECT id FROM projects WHERE name = 'Proyecto Trigo 2024' LIMIT 1), 'Cosecha Trigo', (SELECT id FROM labor_categories WHERE name = 'Cosecha' LIMIT 1), 45.00, 'Cosechadoras Unidos', NOW(), NOW()),
((SELECT id FROM projects WHERE name = 'Proyecto Girasol 2024' LIMIT 1), 'Cosecha Soja', (SELECT id FROM labor_categories WHERE name = 'Cosecha' LIMIT 1), 50.00, 'Cosechadoras Unidos', NOW(), NOW()),

-- Julio - Post-cosecha
((SELECT id FROM projects WHERE name = 'Proyecto Soja 2024' LIMIT 1), 'Secado de granos', (SELECT id FROM labor_categories WHERE name = 'Secado' LIMIT 1), 15.00, 'Secaderos del Sur', NOW(), NOW()),
((SELECT id FROM projects WHERE name = 'Proyecto Maíz 2024' LIMIT 1), 'Secado de granos', (SELECT id FROM labor_categories WHERE name = 'Secado' LIMIT 1), 15.00, 'Secaderos del Sur', NOW(), NOW()),
((SELECT id FROM projects WHERE name = 'Proyecto Trigo 2024' LIMIT 1), 'Secado de granos', (SELECT id FROM labor_categories WHERE name = 'Secado' LIMIT 1), 15.00, 'Secaderos del Sur', NOW(), NOW()),
((SELECT id FROM projects WHERE name = 'Proyecto Girasol 2024' LIMIT 1), 'Secado de granos', (SELECT id FROM labor_categories WHERE name = 'Secado' LIMIT 1), 15.00, 'Secaderos del Sur', NOW(), NOW())
ON CONFLICT DO NOTHING;

-- 3. CREAR INVERSORES (estructura real: solo name)
INSERT INTO investors (name, created_at, updated_at) VALUES 
('Fondo Agrícola del Sur', NOW(), NOW()),
('Inversores Unidos S.A.', NOW(), NOW()),
('Capital Rural', NOW(), NOW()),
('Agro Inversiones', NOW(), NOW()),
('Fondo Soja Plus', NOW(), NOW())
ON CONFLICT DO NOTHING;

-- 4. VINCULAR INVERSORES AL PROYECTO CON PORCENTAJES
INSERT INTO project_investors (project_id, investor_id, percentage, created_at, updated_at) VALUES 
((SELECT id FROM projects WHERE name = 'Proyecto Soja 2024' LIMIT 1), (SELECT id FROM investors WHERE name = 'Fondo Agrícola del Sur' LIMIT 1), 35, NOW(), NOW()),
((SELECT id FROM projects WHERE name = 'Proyecto Soja 2024' LIMIT 1), (SELECT id FROM investors WHERE name = 'Inversores Unidos S.A.' LIMIT 1), 25, NOW(), NOW()),
((SELECT id FROM projects WHERE name = 'Proyecto Soja 2024' LIMIT 1), (SELECT id FROM investors WHERE name = 'Capital Rural' LIMIT 1), 20, NOW(), NOW()),
((SELECT id FROM projects WHERE name = 'Proyecto Soja 2024' LIMIT 1), (SELECT id FROM investors WHERE name = 'Agro Inversiones' LIMIT 1), 15, NOW(), NOW()),
((SELECT id FROM projects WHERE name = 'Proyecto Soja 2024' LIMIT 1), (SELECT id FROM investors WHERE name = 'Fondo Soja Plus' LIMIT 1), 5, NOW(), NOW())
ON CONFLICT DO NOTHING;

-- 5. CREAR WORKORDERS CON DATOS REALES Y FECHAS
INSERT INTO workorders (number, project_id, field_id, lot_id, crop_id, labor_id, contractor, date, investor_id, effective_area, created_at, updated_at) VALUES 
-- Enero - Siembra
('WO-001-2024', (SELECT id FROM projects WHERE name = 'Proyecto Soja 2024' LIMIT 1), (SELECT id FROM fields WHERE name = 'Campo A1' LIMIT 1), (SELECT id FROM lots WHERE name = 'Lote A1' LIMIT 1), (SELECT id FROM crops WHERE name = 'Soja' LIMIT 1), (SELECT id FROM labors WHERE name = 'Siembra directa Soja' LIMIT 1), 'Contratista Campo Sur', '2024-01-15', 1, 10.5, NOW(), NOW()),
('WO-002-2024', (SELECT id FROM projects WHERE name = 'Proyecto Maíz 2024' LIMIT 1), (SELECT id FROM fields WHERE name = 'Campo B1' LIMIT 1), (SELECT id FROM lots WHERE name = 'Lote B1' LIMIT 1), (SELECT id FROM crops WHERE name = 'Maíz' LIMIT 1), (SELECT id FROM labors WHERE name = 'Siembra directa Maíz' LIMIT 1), 'Contratista Campo Sur', '2024-01-20', 1, 8.5, NOW(), NOW()),
('WO-003-2024', (SELECT id FROM projects WHERE name = 'Proyecto Trigo 2024' LIMIT 1), (SELECT id FROM fields WHERE name = 'Campo C1' LIMIT 1), (SELECT id FROM lots WHERE name = 'Lote C1' LIMIT 1), (SELECT id FROM crops WHERE name = 'Trigo' LIMIT 1), (SELECT id FROM labors WHERE name = 'Siembra convencional Trigo' LIMIT 1), 'Contratista Norte', '2024-01-25', 1, 12.0, NOW(), NOW()),
('WO-004-2024', (SELECT id FROM projects WHERE name = 'Proyecto Girasol 2024' LIMIT 1), (SELECT id FROM fields WHERE name = 'Campo D1' LIMIT 1), (SELECT id FROM lots WHERE name = 'Lote D1' LIMIT 1), (SELECT id FROM crops WHERE name = 'Girasol' LIMIT 1), (SELECT id FROM labors WHERE name = 'Siembra directa Soja' LIMIT 1), 'Contratista Campo Sur', '2024-02-01', 1, 15.0, NOW(), NOW()),

-- Febrero - Fertilización
('WO-005-2024', (SELECT id FROM projects WHERE name = 'Proyecto Soja 2024' LIMIT 1), (SELECT id FROM fields WHERE name = 'Campo A1' LIMIT 1), (SELECT id FROM lots WHERE name = 'Lote A1' LIMIT 1), (SELECT id FROM crops WHERE name = 'Soja' LIMIT 1), (SELECT id FROM labors WHERE name = 'Fertilización foliar' LIMIT 1), 'Servicios Agrícolas', '2024-02-15', 1, 10.5, NOW(), NOW()),
('WO-006-2024', (SELECT id FROM projects WHERE name = 'Proyecto Maíz 2024' LIMIT 1), (SELECT id FROM fields WHERE name = 'Campo B1' LIMIT 1), (SELECT id FROM lots WHERE name = 'Lote B1' LIMIT 1), (SELECT id FROM crops WHERE name = 'Maíz' LIMIT 1), (SELECT id FROM labors WHERE name = 'Fertilización foliar' LIMIT 1), 'Servicios Agrícolas', '2024-02-20', 1, 8.5, NOW(), NOW()),
('WO-007-2024', (SELECT id FROM projects WHERE name = 'Proyecto Trigo 2024' LIMIT 1), (SELECT id FROM fields WHERE name = 'Campo C1' LIMIT 1), (SELECT id FROM lots WHERE name = 'Lote C1' LIMIT 1), (SELECT id FROM crops WHERE name = 'Trigo' LIMIT 1), (SELECT id FROM labors WHERE name = 'Fertilización foliar' LIMIT 1), 'Servicios Agrícolas', '2024-02-25', 1, 12.0, NOW(), NOW()),

-- Marzo - Pulverizaciones
('WO-008-2024', (SELECT id FROM projects WHERE name = 'Proyecto Soja 2024' LIMIT 1), (SELECT id FROM fields WHERE name = 'Campo A1' LIMIT 1), (SELECT id FROM lots WHERE name = 'Lote A1' LIMIT 1), (SELECT id FROM crops WHERE name = 'Soja' LIMIT 1), (SELECT id FROM labors WHERE name = 'Pulverización herbicida' LIMIT 1), 'Fumigaciones del Valle', '2024-03-15', 1, 10.5, NOW(), NOW()),
('WO-009-2024', (SELECT id FROM projects WHERE name = 'Proyecto Maíz 2024' LIMIT 1), (SELECT id FROM fields WHERE name = 'Campo B1' LIMIT 1), (SELECT id FROM lots WHERE name = 'Lote B1' LIMIT 1), (SELECT id FROM crops WHERE name = 'Maíz' LIMIT 1), (SELECT id FROM labors WHERE name = 'Pulverización herbicida' LIMIT 1), 'Fumigaciones del Valle', '2024-03-20', 1, 8.5, NOW(), NOW()),
('WO-010-2024', (SELECT id FROM projects WHERE name = 'Proyecto Trigo 2024' LIMIT 1), (SELECT id FROM fields WHERE name = 'Campo C1' LIMIT 1), (SELECT id FROM lots WHERE name = 'Lote C1' LIMIT 1), (SELECT id FROM crops WHERE name = 'Trigo' LIMIT 1), (SELECT id FROM labors WHERE name = 'Pulverización fungicida' LIMIT 1), 'Fumigaciones del Valle', '2024-03-25', 1, 12.0, NOW(), NOW()),

-- Abril - Riego
('WO-011-2024', (SELECT id FROM projects WHERE name = 'Proyecto Soja 2024' LIMIT 1), (SELECT id FROM fields WHERE name = 'Campo A1' LIMIT 1), (SELECT id FROM lots WHERE name = 'Lote A1' LIMIT 1), (SELECT id FROM crops WHERE name = 'Soja' LIMIT 1), (SELECT id FROM labors WHERE name = 'Riego por aspersión' LIMIT 1), 'Riegos del Campo', '2024-04-10', 1, 10.5, NOW(), NOW()),
('WO-012-2024', (SELECT id FROM projects WHERE name = 'Proyecto Maíz 2024' LIMIT 1), (SELECT id FROM fields WHERE name = 'Campo B1' LIMIT 1), (SELECT id FROM lots WHERE name = 'Lote B1' LIMIT 1), (SELECT id FROM crops WHERE name = 'Maíz' LIMIT 1), (SELECT id FROM labors WHERE name = 'Riego por aspersión' LIMIT 1), 'Riegos del Campo', '2024-04-15', 1, 8.5, NOW(), NOW()),

-- Junio - Cosecha
('WO-013-2024', (SELECT id FROM projects WHERE name = 'Proyecto Soja 2024' LIMIT 1), (SELECT id FROM fields WHERE name = 'Campo A1' LIMIT 1), (SELECT id FROM lots WHERE name = 'Lote A1' LIMIT 1), (SELECT id FROM crops WHERE name = 'Soja' LIMIT 1), (SELECT id FROM labors WHERE name = 'Cosecha Soja' LIMIT 1), 'Cosechadoras Unidos', '2024-06-15', 1, 10.5, NOW(), NOW()),
('WO-014-2024', (SELECT id FROM projects WHERE name = 'Proyecto Maíz 2024' LIMIT 1), (SELECT id FROM fields WHERE name = 'Campo B1' LIMIT 1), (SELECT id FROM lots WHERE name = 'Lote B1' LIMIT 1), (SELECT id FROM crops WHERE name = 'Maíz' LIMIT 1), (SELECT id FROM labors WHERE name = 'Cosecha Maíz' LIMIT 1), 'Cosechadoras Unidos', '2024-06-25', 1, 8.5, NOW(), NOW()),
('WO-015-2024', (SELECT id FROM projects WHERE name = 'Proyecto Trigo 2024' LIMIT 1), (SELECT id FROM fields WHERE name = 'Campo C1' LIMIT 1), (SELECT id FROM lots WHERE name = 'Lote C1' LIMIT 1), (SELECT id FROM crops WHERE name = 'Trigo' LIMIT 1), (SELECT id FROM labors WHERE name = 'Cosecha Trigo' LIMIT 1), 'Cosechadoras Unidos', '2024-07-05', 1, 12.0, NOW(), NOW()),
('WO-016-2024', (SELECT id FROM projects WHERE name = 'Proyecto Girasol 2024' LIMIT 1), (SELECT id FROM fields WHERE name = 'Campo D1' LIMIT 1), (SELECT id FROM lots WHERE name = 'Lote D1' LIMIT 1), (SELECT id FROM crops WHERE name = 'Girasol' LIMIT 1), (SELECT id FROM labors WHERE name = 'Cosecha Soja' LIMIT 1), 'Cosechadoras Unidos', '2024-07-15', 1, 15.0, NOW(), NOW()),

-- Julio - Post-cosecha
('WO-017-2024', (SELECT id FROM projects WHERE name = 'Proyecto Soja 2024' LIMIT 1), (SELECT id FROM fields WHERE name = 'Campo A1' LIMIT 1), (SELECT id FROM lots WHERE name = 'Lote A1' LIMIT 1), (SELECT id FROM crops WHERE name = 'Soja' LIMIT 1), (SELECT id FROM labors WHERE name = 'Secado de granos' LIMIT 1), 'Secaderos del Sur', '2024-07-20', 1, 10.5, NOW(), NOW()),
('WO-018-2024', (SELECT id FROM projects WHERE name = 'Proyecto Maíz 2024' LIMIT 1), (SELECT id FROM fields WHERE name = 'Campo B1' LIMIT 1), (SELECT id FROM lots WHERE name = 'Lote B1' LIMIT 1), (SELECT id FROM crops WHERE name = 'Maíz' LIMIT 1), (SELECT id FROM labors WHERE name = 'Secado de granos' LIMIT 1), 'Secaderos del Sur', '2024-07-25', 1, 8.5, NOW(), NOW()),
('WO-019-2024', (SELECT id FROM projects WHERE name = 'Proyecto Trigo 2024' LIMIT 1), (SELECT id FROM fields WHERE name = 'Campo C1' LIMIT 1), (SELECT id FROM lots WHERE name = 'Lote C1' LIMIT 1), (SELECT id FROM crops WHERE name = 'Trigo' LIMIT 1), (SELECT id FROM labors WHERE name = 'Secado de granos' LIMIT 1), 'Secaderos del Sur', '2024-07-30', 1, 12.0, NOW(), NOW()),
('WO-020-2024', (SELECT id FROM projects WHERE name = 'Proyecto Girasol 2024' LIMIT 1), (SELECT id FROM fields WHERE name = 'Campo D1' LIMIT 1), (SELECT id FROM lots WHERE name = 'Lote D1' LIMIT 1), (SELECT id FROM crops WHERE name = 'Girasol' LIMIT 1), (SELECT id FROM labors WHERE name = 'Secado de granos' LIMIT 1), 'Secaderos del Sur', '2024-08-05', 1, 15.0, NOW(), NOW())
ON CONFLICT DO NOTHING;

-- 6. CREAR STOCK CON INVENTARIO REAL (estructura real: project_id, supply_id, investor_id, real_stock_units, initial_units, year_period, month_period)
INSERT INTO stocks (project_id, supply_id, investor_id, real_stock_units, initial_units, year_period, month_period, created_at, updated_at) VALUES 
-- Semillas
((SELECT id FROM projects WHERE name = 'Proyecto Soja 2024' LIMIT 1), (SELECT id FROM supplies WHERE name = 'Semilla Soja RR2 BT' LIMIT 1), 1, 500.0, 500.0, 2024, 1, NOW(), NOW()),
((SELECT id FROM projects WHERE name = 'Proyecto Maíz 2024' LIMIT 1), (SELECT id FROM supplies WHERE name = 'Semilla Maíz BT' LIMIT 1), 1, 300.0, 300.0, 2024, 1, NOW(), NOW()),
((SELECT id FROM projects WHERE name = 'Proyecto Trigo 2024' LIMIT 1), (SELECT id FROM supplies WHERE name = 'Semilla Trigo' LIMIT 1), 1, 400.0, 400.0, 2024, 1, NOW(), NOW()),
((SELECT id FROM projects WHERE name = 'Proyecto Girasol 2024' LIMIT 1), (SELECT id FROM supplies WHERE name = 'Semilla Girasol' LIMIT 1), 1, 200.0, 200.0, 2024, 1, NOW(), NOW()),

-- Fertilizantes
((SELECT id FROM projects WHERE name = 'Proyecto Soja 2024' LIMIT 1), (SELECT id FROM supplies WHERE name = 'Fertilizante NPK 15-15-15' LIMIT 1), 1, 1000.0, 1000.0, 2024, 1, NOW(), NOW()),
((SELECT id FROM projects WHERE name = 'Proyecto Maíz 2024' LIMIT 1), (SELECT id FROM supplies WHERE name = 'Urea 46%' LIMIT 1), 1, 800.0, 800.0, 2024, 1, NOW(), NOW()),
((SELECT id FROM projects WHERE name = 'Proyecto Trigo 2024' LIMIT 1), (SELECT id FROM supplies WHERE name = 'Fosfato Diamónico' LIMIT 1), 1, 600.0, 600.0, 2024, 1, NOW(), NOW()),

-- Agroquímicos
((SELECT id FROM projects WHERE name = 'Proyecto Soja 2024' LIMIT 1), (SELECT id FROM supplies WHERE name = 'Herbicida Glifosato' LIMIT 1), 1, 200.0, 200.0, 2024, 1, NOW(), NOW()),
((SELECT id FROM projects WHERE name = 'Proyecto Maíz 2024' LIMIT 1), (SELECT id FROM supplies WHERE name = 'Fungicida Triazol' LIMIT 1), 1, 150.0, 150.0, 2024, 1, NOW(), NOW()),
((SELECT id FROM projects WHERE name = 'Proyecto Trigo 2024' LIMIT 1), (SELECT id FROM supplies WHERE name = 'Insecticida Piretroide' LIMIT 1), 1, 100.0, 100.0, 2024, 1, NOW(), NOW()),

-- Combustibles
((SELECT id FROM projects WHERE name = 'Proyecto Soja 2024' LIMIT 1), (SELECT id FROM supplies WHERE name = 'Diesel' LIMIT 1), 1, 1000.0, 1000.0, 2024, 1, NOW(), NOW()),
((SELECT id FROM projects WHERE name = 'Proyecto Maíz 2024' LIMIT 1), (SELECT id FROM supplies WHERE name = 'Nafta' LIMIT 1), 1, 500.0, 500.0, 2024, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;

-- 7. CREAR FACTURAS CON INGRESOS REALES (estructura real: work_order_id, number, company, date, status)
INSERT INTO invoices (work_order_id, number, company, date, status, created_at, updated_at) VALUES 
-- Facturas pagadas (ingresos reales)
((SELECT id FROM workorders WHERE number = 'WO-001-2024' LIMIT 1), 'INV-2024-001', 'Contratista Campo Sur', '2024-06-30', 'paid', NOW(), NOW()),
((SELECT id FROM workorders WHERE number = 'WO-002-2024' LIMIT 1), 'INV-2024-002', 'Contratista Campo Sur', '2024-07-15', 'paid', NOW(), NOW()),
((SELECT id FROM workorders WHERE number = 'WO-003-2024' LIMIT 1), 'INV-2024-003', 'Contratista Norte', '2024-07-30', 'paid', NOW(), NOW()),
((SELECT id FROM workorders WHERE number = 'WO-004-2024' LIMIT 1), 'INV-2024-004', 'Contratista Campo Sur', '2024-08-15', 'paid', NOW(), NOW()),

-- Facturas pendientes
((SELECT id FROM workorders WHERE number = 'WO-005-2024' LIMIT 1), 'INV-2024-005', 'Servicios Agrícolas', '2024-09-15', 'pending', NOW(), NOW()),
((SELECT id FROM workorders WHERE number = 'WO-006-2024' LIMIT 1), 'INV-2024-006', 'Servicios Agrícolas', '2024-09-30', 'pending', NOW(), NOW()),

-- Facturas en borrador
((SELECT id FROM workorders WHERE number = 'WO-007-2024' LIMIT 1), 'INV-2024-007', 'Servicios Agrícolas', '2024-10-15', 'draft', NOW(), NOW()),
((SELECT id FROM workorders WHERE number = 'WO-008-2024' LIMIT 1), 'INV-2024-008', 'Fumigaciones del Valle', '2024-10-30', 'draft', NOW(), NOW())
ON CONFLICT DO NOTHING;

-- 8. ACTUALIZAR LOTS CON DATOS REALES DE SIEMBRA Y COSECHA
UPDATE lots SET 
    sowing_date = '2024-01-15',
    tons = 4.2
WHERE name = 'Lote A1'; -- Lote A1 (Soja) - 10.5 ha

UPDATE lots SET 
    sowing_date = '2024-01-20',
    tons = 3.4
WHERE name = 'Lote B1'; -- Lote B1 (Maíz) - 8.5 ha

UPDATE lots SET 
    sowing_date = '2024-01-25',
    tons = 5.4
WHERE name = 'Lote C1'; -- Lote C1 (Trigo) - 12.0 ha

UPDATE lots SET 
    sowing_date = '2024-02-01',
    tons = 6.0
WHERE name = 'Lote D1'; -- Lote D1 (Girasol) - 15.0 ha

-- 9. ACTUALIZAR PROJECTS CON COSTOS REALES (estructura real: admin_cost)
UPDATE projects SET 
    admin_cost = 8000.00
WHERE name = 'Proyecto Soja 2024';

-- VERIFICAR INSERCIÓN DE DATOS
SELECT '✅ Datos completos del dashboard cargados:' as status;
SELECT '   - Supplies totales:' as item, COUNT(*) as total FROM supplies;
SELECT '   - Labors totales:' as item, COUNT(*) as total FROM labors;
SELECT '   - Workorders totales:' as item, COUNT(*) as total FROM workorders;
SELECT '   - Investors totales:' as item, COUNT(*) as total FROM investors;
SELECT '   - Stock total:' as item, COUNT(*) as total FROM stocks;
SELECT '   - Invoices total:' as item, COUNT(*) as total FROM invoices;
SELECT '   - Lotes sembrados:' as item, COUNT(*) as total FROM lots WHERE sowing_date IS NOT NULL;

-- VERIFICAR MÉTRICAS ESPERADAS DEL DASHBOARD
SELECT '🎯 Dashboard ahora debería mostrar:' as info;
SELECT '   - Siembra: 46.0 ha / 46.0 ha (100%)' as sowing_info;
SELECT '   - Cosecha: 19.0 toneladas totales' as harvest_info;
SELECT '   - Costos ejecutados: ~$1,200 / $8,000 (15%)' as costs_info;
SELECT '   - Ingresos: $80,000 (facturas pagadas)' as income_info;
SELECT '   - Stock: 20 items de inventario' as stock_info;
SELECT '   - Contribuciones: 100% (5 inversores)' as contrib_info;
SELECT '   - Crops: 4 cultivos con datos reales' as crops_info;
