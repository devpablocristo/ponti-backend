-- Crear categorías de prueba para el sistema
-- Las categorías se crean asociando tipos existentes
-- Se utilizan para clasificar suministros y otros elementos del sistema agrícola

INSERT INTO categories (name, type_id, created_by, updated_by, created_at, updated_at)
SELECT 
  'Categoría ' || generate_series(1, 4),
  generate_series(1, 4),
  2, 2, NOW(), NOW();

-- Verificar inserción
SELECT '✅ Categorías creadas:' as status, COUNT(*) as total FROM categories; 