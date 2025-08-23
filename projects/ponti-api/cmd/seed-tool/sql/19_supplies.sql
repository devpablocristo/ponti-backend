-- Crear suministros de prueba para el sistema
-- Los suministros se crean asociando proyectos, unidades, categorías y tipos existentes
-- Incluye la relación con la tabla units

INSERT INTO supplies (project_id, name, price, unit_id, category_id, type_id, created_by, updated_by, created_at, updated_at)
SELECT 
  generate_series(1, 10),
  'Supply ' || generate_series(1, 10),
  generate_series(1, 10) * 100.0,
  generate_series(1, 3),
  generate_series(1, 4),
  generate_series(1, 4),
  2, 2, NOW(), NOW();

-- Verificar inserción
SELECT '✅ Suministros creados:' as status, COUNT(*) as total FROM supplies; 