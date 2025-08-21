-- Crear tipos fundamentales para el sistema
-- Los tipos se crean para categorizar diferentes elementos del sistema
-- Incluye semillas, agroquímicos, fertilizantes y labores para clasificación

INSERT INTO types (name, created_by, updated_by, created_at, updated_at)
VALUES 
  ('Semilla', 2, 2, NOW(), NOW()),
  ('Agroquímicos', 2, 2, NOW(), NOW()),
  ('Fertilizantes', 2, 2, NOW(), NOW()),
  ('Labores', 2, 2, NOW(), NOW())
ON CONFLICT (name) DO NOTHING;

-- Verificar inserción
SELECT '✅ Tipos creados:' as status, COUNT(*) as total FROM types; 