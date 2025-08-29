-- =======================
-- DATOS BASE NECESARIOS
-- =======================
-- Script para poblar las tablas base necesarias para el dashboard
-- Debe ejecutarse ANTES de los otros scripts

-- 1. AGREGAR TIPOS BASE
INSERT INTO types (name, created_at, updated_at) VALUES 
('Semillas', NOW(), NOW()),
('Fertilizantes', NOW(), NOW()),
('Agroquímicos', NOW(), NOW()),
('Combustibles', NOW(), NOW()),
('Labores', NOW(), NOW()),
('Maquinaria', NOW(), NOW())
ON CONFLICT DO NOTHING;

-- 2. AGREGAR CATEGORÍAS DE INSUMOS
INSERT INTO categories (name, type_id, created_at, updated_at) VALUES 
('Semillas', (SELECT id FROM types WHERE name = 'Semillas' LIMIT 1), NOW(), NOW()),
('Fertilizantes', (SELECT id FROM types WHERE name = 'Fertilizantes' LIMIT 1), NOW(), NOW()),
('Herbicidas', (SELECT id FROM types WHERE name = 'Agroquímicos' LIMIT 1), NOW(), NOW()),
('Fungicidas', (SELECT id FROM types WHERE name = 'Agroquímicos' LIMIT 1), NOW(), NOW()),
('Insecticidas', (SELECT id FROM types WHERE name = 'Agroquímicos' LIMIT 1), NOW(), NOW()),
('Diesel', (SELECT id FROM types WHERE name = 'Combustibles' LIMIT 1), NOW(), NOW()),
('Nafta', (SELECT id FROM types WHERE name = 'Combustibles' LIMIT 1), NOW(), NOW())
ON CONFLICT DO NOTHING;

-- 3. AGREGAR TIPOS DE LABOR
INSERT INTO labor_types (name, created_at, updated_at) VALUES 
('Siembra', NOW(), NOW()),
('Mantenimiento', NOW(), NOW()),
('Cosecha', NOW(), NOW()),
('Post-cosecha', NOW(), NOW()),
('Transporte', NOW(), NOW()),
('Almacenamiento', NOW(), NOW())
ON CONFLICT DO NOTHING;

-- 4. AGREGAR CATEGORÍAS DE LABOR
INSERT INTO labor_categories (name, type_id, created_at, updated_at) VALUES 
('Siembra directa', (SELECT id FROM labor_types WHERE name = 'Siembra' LIMIT 1), NOW(), NOW()),
('Siembra convencional', (SELECT id FROM labor_types WHERE name = 'Siembra' LIMIT 1), NOW(), NOW()),
('Fertilización', (SELECT id FROM labor_types WHERE name = 'Mantenimiento' LIMIT 1), NOW(), NOW()),
('Pulverización', (SELECT id FROM labor_types WHERE name = 'Mantenimiento' LIMIT 1), NOW(), NOW()),
('Riego', (SELECT id FROM labor_types WHERE name = 'Mantenimiento' LIMIT 1), NOW(), NOW()),
('Cosecha', (SELECT id FROM labor_types WHERE name = 'Cosecha' LIMIT 1), NOW(), NOW()),
('Secado', (SELECT id FROM labor_types WHERE name = 'Post-cosecha' LIMIT 1), NOW(), NOW()),
('Transporte', (SELECT id FROM labor_types WHERE name = 'Transporte' LIMIT 1), NOW(), NOW())
ON CONFLICT DO NOTHING;

-- 5. AGREGAR TIPOS DE ARRENDAMIENTO
INSERT INTO lease_types (name, created_at, updated_at) VALUES 
('Arrendamiento', NOW(), NOW()),
('Propio', NOW(), NOW()),
('Comodato', NOW(), NOW()),
('Mediería', NOW(), NOW()),
('Aparcería', NOW(), NOW())
ON CONFLICT DO NOTHING;

-- 6. AGREGAR PROVEEDORES (estructura real: solo name)
INSERT INTO providers (name, created_at, updated_at) VALUES 
('AgroSeed S.A.', NOW(), NOW()),
('Fertilizantes del Sur', NOW(), NOW()),
('AgroQuímica Plus', NOW(), NOW()),
('Combustibles Rurales', NOW(), NOW()),
('Maquinaria Agrícola', NOW(), NOW()),
('Transporte Campo', NOW(), NOW())
ON CONFLICT DO NOTHING;

-- 7. AGREGAR USUARIOS (estructura real: email, username, password, token_hash, id_rol, created_by, updated_by)
INSERT INTO users (email, username, password, token_hash, id_rol, created_by, updated_by, created_at, updated_at) VALUES 
('admin@ponti.com', 'admin', 'hashed_password', 'token_hash_admin', 1, 1, 1, NOW(), NOW()),
('manager@ponti.com', 'manager', 'hashed_password', 'token_hash_manager', 2, 1, 1, NOW(), NOW()),
('operador1@ponti.com', 'operador1', 'hashed_password', 'token_hash_op1', 3, 1, 1, NOW(), NOW()),
('operador2@ponti.com', 'operador2', 'hashed_password', 'token_hash_op2', 3, 1, 1, NOW(), NOW()),
('contador@ponti.com', 'contador', 'hashed_password', 'token_hash_cont', 4, 1, 1, NOW(), NOW()),
('tecnico@ponti.com', 'tecnico', 'hashed_password', 'token_hash_tec', 5, 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;

-- Verificar inserción de datos base
SELECT '✅ Datos base cargados:' as status;
SELECT '   - Types:' as item, COUNT(*) as total FROM types;
SELECT '   - Categories:' as item, COUNT(*) as total FROM categories;
SELECT '   - Labor types:' as item, COUNT(*) as total FROM labor_types;
SELECT '   - Labor categories:' as item, COUNT(*) as total FROM labor_categories;
SELECT '   - Lease types:' as item, COUNT(*) as total FROM lease_types;
SELECT '   - Providers:' as item, COUNT(*) as total FROM providers;
SELECT '   - Users:' as item, COUNT(*) as total FROM users;
