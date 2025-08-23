-- Crear labores de prueba para el sistema
-- Las labores se crean asociando proyectos y categorías existentes
-- Incluye precio y nombre del contratista

INSERT INTO labors (project_id, name, category_id, price, contractor_name, created_by, updated_by, created_at, updated_at)
SELECT 
  generate_series(1, 10),
  'Labor ' || generate_series(1, 10),
  (generate_series(1, 10) - 1) % 4 + 1,
  generate_series(1, 10) * 50.0,
  'Contractor ' || generate_series(1, 10),
  2, 2, NOW(), NOW();

-- Verificar inserción
SELECT '✅ Labores creadas:' as status, COUNT(*) as total FROM labors; 