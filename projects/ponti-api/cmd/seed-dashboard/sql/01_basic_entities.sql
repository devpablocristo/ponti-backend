-- =======================
-- ENTIDADES BÁSICAS NECESARIAS
-- =======================
-- Script para crear customers, campaigns, projects, fields, lots y crops
-- Debe ejecutarse DESPUÉS de 00_base_data.sql y ANTES de 99_complete_dashboard_data.sql

-- 1. CREAR CUSTOMERS (estructura real: solo name)
INSERT INTO customers (name, created_at, updated_at) VALUES 
('AgroEmpresa S.A.', NOW(), NOW()),
('Campo del Sur', NOW(), NOW()),
('Finca Los Pinos', NOW(), NOW())
ON CONFLICT DO NOTHING;

-- 2. CREAR CAMPAIGNS (estructura real: solo name)
INSERT INTO campaigns (name, created_at, updated_at) VALUES 
('Campaña 2024/25', NOW(), NOW()),
('Campaña 2023/24', NOW(), NOW()),
('Campaña 2022/23', NOW(), NOW())
ON CONFLICT DO NOTHING;

-- 3. CREAR PROJECTS
INSERT INTO projects (name, customer_id, campaign_id, admin_cost, created_at, updated_at) VALUES 
('Proyecto Soja 2024', 1, 1, 8000.00, NOW(), NOW()),
('Proyecto Maíz 2024', 1, 1, 6000.00, NOW(), NOW()),
('Proyecto Trigo 2024', 2, 1, 5000.00, NOW(), NOW()),
('Proyecto Girasol 2024', 3, 1, 4000.00, NOW(), NOW())
ON CONFLICT DO NOTHING;

-- 4. CREAR FIELDS (estructura real: name, project_id, lease_type_id, lease_type_percent, lease_type_value)
-- Usar los IDs reales de projects que se generaron
INSERT INTO fields (name, project_id, lease_type_id, lease_type_percent, lease_type_value, created_at, updated_at) VALUES 
('Campo A1', (SELECT id FROM projects WHERE name = 'Proyecto Soja 2024' LIMIT 1), (SELECT id FROM lease_types WHERE name = 'Propio' LIMIT 1), 100.0, 0.0, NOW(), NOW()),
('Campo B1', (SELECT id FROM projects WHERE name = 'Proyecto Soja 2024' LIMIT 1), (SELECT id FROM lease_types WHERE name = 'Propio' LIMIT 1), 100.0, 0.0, NOW(), NOW()),
('Campo C1', (SELECT id FROM projects WHERE name = 'Proyecto Maíz 2024' LIMIT 1), (SELECT id FROM lease_types WHERE name = 'Arrendamiento' LIMIT 1), 80.0, 200.0, NOW(), NOW()),
('Campo D1', (SELECT id FROM projects WHERE name = 'Proyecto Trigo 2024' LIMIT 1), (SELECT id FROM lease_types WHERE name = 'Arrendamiento' LIMIT 1), 75.0, 150.0, NOW(), NOW()),
('Campo E1', (SELECT id FROM projects WHERE name = 'Proyecto Girasol 2024' LIMIT 1), (SELECT id FROM lease_types WHERE name = 'Propio' LIMIT 1), 100.0, 0.0, NOW(), NOW())
ON CONFLICT DO NOTHING;

-- 5. CREAR CROPS
INSERT INTO crops (name, created_at, updated_at) VALUES 
('Soja', NOW(), NOW()),
('Maíz', NOW(), NOW()),
('Trigo', NOW(), NOW()),
('Girasol', NOW(), NOW()),
('Cebada', NOW(), NOW()),
('Sorgo', NOW(), NOW())
ON CONFLICT DO NOTHING;

-- 6. CREAR LOTS (estructura real: name, field_id, hectares, previous_crop_id, current_crop_id, season)
-- Usar los IDs reales de fields que se generaron
INSERT INTO lots (name, field_id, hectares, previous_crop_id, current_crop_id, season, created_at, updated_at) VALUES 
('Lote A1', (SELECT id FROM fields WHERE name = 'Campo A1' LIMIT 1), 10.5, 1, 1, '2024/25', NOW(), NOW()),
('Lote B1', (SELECT id FROM fields WHERE name = 'Campo B1' LIMIT 1), 8.5, 2, 2, '2024/25', NOW(), NOW()),
('Lote C1', (SELECT id FROM fields WHERE name = 'Campo C1' LIMIT 1), 12.0, 3, 3, '2024/25', NOW(), NOW()),
('Lote D1', (SELECT id FROM fields WHERE name = 'Campo D1' LIMIT 1), 15.0, 4, 4, '2024/25', NOW(), NOW()),
('Lote E1', (SELECT id FROM fields WHERE name = 'Campo E1' LIMIT 1), 20.0, 1, 1, '2024/25', NOW(), NOW())
ON CONFLICT DO NOTHING;

-- Verificar inserción de entidades básicas
SELECT '✅ Entidades básicas creadas:' as status;
SELECT '   - Customers:' as item, COUNT(*) as total FROM customers;
SELECT '   - Campaigns:' as item, COUNT(*) as total FROM campaigns;
SELECT '   - Projects:' as item, COUNT(*) as total FROM projects;
SELECT '   - Fields:' as item, COUNT(*) as total FROM fields;
SELECT '   - Crops:' as item, COUNT(*) as total FROM crops;
SELECT '   - Lots:' as item, COUNT(*) as total FROM lots;
