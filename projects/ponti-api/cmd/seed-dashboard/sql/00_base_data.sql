-- =======================
-- DATOS BASE NECESARIOS
-- =======================
-- Script para poblar las tablas base necesarias para el dashboard
-- Debe ejecutarse ANTES de los otros scripts

-- 1. AGREGAR TIPOS DE ARRENDAMIENTO
INSERT INTO lease_types (name, created_at, updated_at) VALUES 
('Arrendamiento', NOW(), NOW()),
('Propio', NOW(), NOW()),
('Comodato', NOW(), NOW())
ON CONFLICT DO NOTHING;

-- 2. AGREGAR CATEGORÍAS DE LABOR
INSERT INTO labor_categories (name, created_at, updated_at) VALUES 
('Siembra', NOW(), NOW()),
('Mantenimiento', NOW(), NOW()),
('Cosecha', NOW(), NOW()),
('Post-cosecha', NOW(), NOW())
ON CONFLICT DO NOTHING;

-- 3. AGREGAR TIPOS DE LABOR
INSERT INTO labor_types (name, labor_category_id, created_at, updated_at) VALUES 
('Siembra directa', 1, NOW(), NOW()),
('Fertilización', 2, NOW(), NOW()),
('Pulverización', 2, NOW(), NOW()),
('Cosecha', 3, NOW(), NOW()),
('Secado', 4, NOW(), NOW())
ON CONFLICT DO NOTHING;

-- 4. AGREGAR CATEGORÍAS DE INSUMOS
INSERT INTO categories (name, created_at, updated_at) VALUES 
('Semillas', NOW(), NOW()),
('Fertilizantes', NOW(), NOW()),
('Agroquímicos', NOW(), NOW()),
('Combustibles', NOW(), NOW())
ON CONFLICT DO NOTHING;

-- 5. AGREGAR TIPOS DE INSUMOS
INSERT INTO types (name, category_id, created_at, updated_at) VALUES 
('Semilla Soja', 1, NOW(), NOW()),
('Semilla Maíz', 1, NOW(), NOW()),
('Semilla Trigo', 1, NOW(), NOW()),
('Fertilizante NPK', 2, NOW(), NOW()),
('Herbicida', 3, NOW(), NOW()),
('Fungicida', 3, NOW(), NOW())
ON CONFLICT DO NOTHING;

-- 6. AGREGAR PROVEEDORES
INSERT INTO providers (name, email, phone, created_at, updated_at) VALUES 
('Proveedor A', 'proveedor.a@email.com', '+54 11 1234-5678', NOW(), NOW()),
('Proveedor B', 'proveedor.b@email.com', '+54 11 2345-6789', NOW(), NOW()),
('Proveedor C', 'proveedor.c@email.com', '+54 11 3456-7890', NOW(), NOW())
ON CONFLICT DO NOTHING;

-- 7. AGREGAR USUARIOS (si no existen)
INSERT INTO users (name, email, password, created_at, updated_at) VALUES 
('Admin', 'admin@ponti.com', 'hashed_password', NOW(), NOW()),
('Manager', 'manager@ponti.com', 'hashed_password', NOW(), NOW()),
('Operator', 'operator@ponti.com', 'hashed_password', NOW(), NOW())
ON CONFLICT DO NOTHING;

-- Verificar inserción de datos base
SELECT '✅ Datos base cargados:' as status;
SELECT '   - Lease types:' as item, COUNT(*) as total FROM lease_types;
SELECT '   - Labor categories:' as item, COUNT(*) as total FROM labor_categories;
SELECT '   - Labor types:' as item, COUNT(*) as total FROM labor_types;
SELECT '   - Categories:' as item, COUNT(*) as total FROM categories;
SELECT '   - Types:' as item, COUNT(*) as total FROM types;
SELECT '   - Providers:' as item, COUNT(*) as total FROM providers;
SELECT '   - Users:' as item, COUNT(*) as total FROM users;
