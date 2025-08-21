-- Crear datos de comercialización de cultivos de prueba para el sistema
-- Los datos se crean asociando proyectos y cultivos existentes con precios y costos

INSERT INTO crop_commercializations (project_id, crop_id, board_price, freight_cost, commercial_cost, net_price, created_by, updated_by, created_at, updated_at)
SELECT 
  generate_series(1, 3),
  generate_series(1, 3),
  generate_series(100, 300, 100)::NUMERIC(12,2),
  generate_series(10, 30, 10)::NUMERIC(12,2),
  generate_series(5, 15, 5)::DOUBLE PRECISION,
  generate_series(85, 255, 85)::NUMERIC(12,2),
  2, 2, NOW(), NOW()
ON CONFLICT DO NOTHING;

-- Verificar inserción
SELECT '✅ Comercialización de cultivos creada:' as status, COUNT(*) as total FROM crop_commercializations;
