-- Crear campos de prueba para el sistema
-- Los campos se crean asociando proyectos y tipos de arriendo existentes

INSERT INTO fields (name, project_id, lease_type_id, lease_type_percent, lease_type_value, created_by, updated_by, created_at, updated_at)
SELECT 
  'Field ' || generate_series(1, 10),
  generate_series(1, 10),
  1,
  10.0,
  500.0,
  2, 2, NOW(), NOW();

-- Verificar inserción
SELECT '✅ Campos creados:' as status, COUNT(*) as total FROM fields;
