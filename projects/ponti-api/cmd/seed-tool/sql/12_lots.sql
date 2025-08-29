-- Crear lotes de prueba para el sistema
-- Los lotes se crean asociando campos y cultivos existentes
-- Incluye las nuevas columnas: variety, sowing_date, version

INSERT INTO lots (name, field_id, hectares, previous_crop_id, current_crop_id, season, variety, sowing_date, created_by, updated_by, created_at, updated_at)
SELECT 
  'Lot ' || generate_series(1, 10),
  generate_series(1, 10),
  generate_series(1, 10) * 5.0,
  generate_series(1, 10),
  generate_series(1, 10),
  '2025',
  'V' || LPAD(generate_series(1, 10)::text, 2, '0'),
  CURRENT_DATE - (generate_series(1, 10) * interval '15 days'),
  2, 2, NOW(), NOW();

-- Verificar inserción
SELECT '✅ Lotes creados:' as status, COUNT(*) as total FROM lots; 