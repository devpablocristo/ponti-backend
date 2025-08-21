-- Crear elementos de orden de trabajo de prueba para el sistema
-- Los elementos se crean asociando órdenes de trabajo e insumos existentes
-- Incluye cantidad total usada y dosis final para cada insumo
-- NOTA: La tabla no tiene columnas created_by y updated_by

INSERT INTO workorder_items (workorder_id, supply_id, total_used, final_dose, created_at, updated_at)
SELECT 
  generate_series(1, 5),
  generate_series(1, 5),
  generate_series(1, 5) * 10.0,
  generate_series(1, 5) * 2.5,
  NOW(), NOW();

-- Verificar inserción
SELECT '✅ Items de órdenes de trabajo creados:' as status, COUNT(*) as total FROM workorder_items; 