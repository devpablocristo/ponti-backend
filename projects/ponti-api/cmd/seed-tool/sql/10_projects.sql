-- Crear proyectos de prueba para el sistema
-- Los proyectos se crean asociando clientes y campañas existentes
-- Incluye la nueva columna version para bloqueo optimista

INSERT INTO projects (name, customer_id, campaign_id, admin_cost, created_by, updated_by, created_at, updated_at)
SELECT 
  'Project ' || generate_series(1, 10),
  generate_series(1, 10),
  generate_series(1, 10),
  1000,
  2, 2, NOW(), NOW();

-- Verificar inserción
SELECT '✅ Proyectos creados:' as status, COUNT(*) as total FROM projects; 