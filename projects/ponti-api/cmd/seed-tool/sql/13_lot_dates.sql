-- Crear fechas de lotes de prueba para el sistema
-- Las fechas se crean asociando lotes existentes con fechas de siembra y cosecha
-- Cada lote puede tener múltiples secuencias de fechas (limitado a 1-3 por restricción CHECK)

INSERT INTO lot_dates (lot_id, sowing_date, harvest_date, sequence, created_by, updated_by, created_at, updated_at)
SELECT 
  generate_series(1, 5),
  CURRENT_DATE - (generate_series(1, 5) * interval '30 days'),
  CURRENT_DATE + (generate_series(1, 5) * interval '90 days'),
  (generate_series(1, 5) - 1) % 3 + 1,
  2, 2, NOW(), NOW()
ON CONFLICT (lot_id, sequence) DO NOTHING;

-- Verificar inserción
SELECT '✅ Fechas de lotes creadas:' as status, COUNT(*) as total FROM lot_dates;
