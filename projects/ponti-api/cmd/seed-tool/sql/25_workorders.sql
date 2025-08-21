-- Crear órdenes de trabajo de prueba para el sistema
-- Las órdenes de trabajo se crean asociando proyectos, campos, lotes, cultivos, labores e inversores existentes
-- Se crean órdenes de trabajo solo para el primer proyecto con entidades válidas
-- Incluye información de contratista, observaciones y área efectiva

INSERT INTO workorders (number, project_id, field_id, lot_id, crop_id, labor_id, contractor, observations, date, investor_id, effective_area, created_by, updated_by, created_at, updated_at)
SELECT 
  'WO-' || LPAD(generate_series(1, 5)::text, 4, '0'),
  1,
  1,
  1,
  1,
  1,
  'Contractor ' || generate_series(1, 5),
  'Observación para orden ' || generate_series(1, 5),
  CURRENT_DATE - (generate_series(1, 5) * interval '1 day'),
  1,
  5.0,
  2, 2, NOW(), NOW();

-- Verificar inserción
SELECT '✅ Órdenes de trabajo creadas:' as status, COUNT(*) as total FROM workorders; 