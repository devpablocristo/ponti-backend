-- Crear tipos de arriendo para el sistema
-- Los tipos se crean para definir diferentes modalidades de arriendo de tierras
-- Incluye arriendos simples, con opción de compra, por hectárea, por tonelada, etc.

INSERT INTO lease_types (name, created_by, updated_by, created_at, updated_at)
VALUES 
  ('Arriendo Simple', 2, 2, NOW(), NOW()),
  ('Arriendo con Opción de Compra', 2, 2, NOW(), NOW()),
  ('Arriendo con Participación', 2, 2, NOW(), NOW()),
  ('Arriendo por Hectárea', 2, 2, NOW(), NOW()),
  ('Arriendo por Tonelada', 2, 2, NOW(), NOW()),
  ('Arriendo Mixto', 2, 2, NOW(), NOW()),
  ('Arriendo por Temporada', 2, 2, NOW(), NOW()),
  ('Arriendo con Inversión', 2, 2, NOW(), NOW()),
  ('Arriendo por Rendimiento', 2, 2, NOW(), NOW()),
  ('Arriendo por Calidad', 2, 2, NOW(), NOW())
ON CONFLICT (name) DO NOTHING;

-- Verificar inserción
SELECT '✅ Tipos de arriendo creados:' as status, COUNT(*) as total FROM lease_types; 