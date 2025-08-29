-- =======================
-- SEED DASHBOARD MINIMAL
-- =======================
-- Datos mínimos necesarios para que el dashboard funcione
-- Crea: 1 cultivo, 1 cliente, 1 campaña, 1 proyecto, 1 campo, 1 lote

-- 1. CULTIVO
INSERT INTO crops (name, created_at, updated_at) VALUES 
('Soja', NOW(), NOW());

-- 2. CLIENTE
INSERT INTO customers (name, created_at, updated_at) VALUES 
('Cliente A', NOW(), NOW());

-- 3. CAMPAÑA
INSERT INTO campaigns (name, created_at, updated_at) VALUES 
('2024-2025', NOW(), NOW());

-- 4. PROYECTO
INSERT INTO projects (name, customer_id, campaign_id, admin_cost, created_at, updated_at) VALUES 
('Proyecto Soja', 1, 1, 1000, NOW(), NOW());

-- 5. CAMPO
INSERT INTO fields (name, project_id, lease_type_id, created_at, updated_at) VALUES 
('Campo Norte', 1, 1, NOW(), NOW());

-- 6. LOTE CON SOJA (previous_crop_id es obligatorio, uso el mismo cultivo)
INSERT INTO lots (name, field_id, hectares, previous_crop_id, current_crop_id, season, created_at, updated_at) VALUES 
('Lote A1', 1, 10.5, 1, 1, '2024-2025', NOW(), NOW());
