-- Crear clientes de prueba para el sistema
-- Los clientes se crean para asociar con proyectos agrícolas
-- Incluye 10 clientes con nombres únicos para diferentes tipos de proyectos

INSERT INTO customers (name, created_by, updated_by, created_at, updated_at)
VALUES 
  ('Customer 1', 2, 2, NOW(), NOW()),
  ('Customer 2', 2, 2, NOW(), NOW()),
  ('Customer 3', 2, 2, NOW(), NOW()),
  ('Customer 4', 2, 2, NOW(), NOW()),
  ('Customer 5', 2, 2, NOW(), NOW()),
  ('Customer 6', 2, 2, NOW(), NOW()),
  ('Customer 7', 2, 2, NOW(), NOW()),
  ('Customer 8', 2, 2, NOW(), NOW()),
  ('Customer 9', 2, 2, NOW(), NOW()),
  ('Customer 10', 2, 2, NOW(), NOW())
ON CONFLICT (name) DO NOTHING;

-- Verificar inserción
SELECT '✅ Clientes creados:' as status, COUNT(*) as total FROM customers; 