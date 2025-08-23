-- Crear valores del dólar para proyectos de prueba
-- Los valores se crean asociando proyectos existentes con datos mensuales
-- Incluye valor inicial, final y promedio mensual para seguimiento de cotización

INSERT INTO project_dollar_values (project_id, year, month, start_value, end_value, average_value, created_by, updated_by, created_at, updated_at)
SELECT 
  p.id, 2025, 'Enero', 850.00, 870.00, 860.00, 2, 2, NOW(), NOW()
FROM projects p
WHERE p.name = 'Project 1'
ON CONFLICT (project_id, year, month) DO NOTHING;

INSERT INTO project_dollar_values (project_id, year, month, start_value, end_value, average_value, created_by, updated_by, created_at, updated_at)
SELECT 
  p.id, 2025, 'Febrero', 870.00, 890.00, 880.00, 2, 2, NOW(), NOW()
FROM projects p
WHERE p.name = 'Project 1'
ON CONFLICT (project_id, year, month) DO NOTHING;

INSERT INTO project_dollar_values (project_id, year, month, start_value, end_value, average_value, created_by, updated_by, created_at, updated_at)
SELECT 
  p.id, 2025, 'Marzo', 890.00, 920.00, 905.00, 2, 2, NOW(), NOW()
FROM projects p
WHERE p.name = 'Project 1'
ON CONFLICT (project_id, year, month) DO NOTHING;

INSERT INTO project_dollar_values (project_id, year, month, start_value, end_value, average_value, created_by, updated_by, created_at, updated_at)
SELECT 
  p.id, 2025, 'Abril', 920.00, 950.00, 935.00, 2, 2, NOW(), NOW()
FROM projects p
WHERE p.name = 'Project 1'
ON CONFLICT (project_id, year, month) DO NOTHING;

INSERT INTO project_dollar_values (project_id, year, month, start_value, end_value, average_value, created_by, updated_by, created_at, updated_at)
SELECT 
  p.id, 2025, 'Mayo', 950.00, 980.00, 965.00, 2, 2, NOW(), NOW()
FROM projects p
WHERE p.name = 'Project 1'
ON CONFLICT (project_id, year, month) DO NOTHING;

INSERT INTO project_dollar_values (project_id, year, month, start_value, end_value, average_value, created_by, updated_by, created_at, updated_at)
SELECT 
  p.id, 2025, 'Junio', 980.00, 1000.00, 990.00, 2, 2, NOW(), NOW()
FROM projects p
WHERE p.name = 'Project 1'
ON CONFLICT (project_id, year, month) DO NOTHING;

INSERT INTO project_dollar_values (project_id, year, month, start_value, end_value, average_value, created_by, updated_by, created_at, updated_at)
SELECT 
  p.id, 2025, 'Julio', 1000.00, 1020.00, 1010.00, 2, 2, NOW(), NOW()
FROM projects p
WHERE p.name = 'Project 1'
ON CONFLICT (project_id, year, month) DO NOTHING;

INSERT INTO project_dollar_values (project_id, year, month, start_value, end_value, average_value, created_by, updated_by, created_at, updated_at)
SELECT 
  p.id, 2025, 'Agosto', 1020.00, 1050.00, 1035.00, 2, 2, NOW(), NOW()
FROM projects p
WHERE p.name = 'Project 1'
ON CONFLICT (project_id, year, month) DO NOTHING;

INSERT INTO project_dollar_values (project_id, year, month, start_value, end_value, average_value, created_by, updated_by, created_at, updated_at)
SELECT 
  p.id, 2025, 'Septiembre', 1050.00, 1080.00, 1065.00, 2, 2, NOW(), NOW()
FROM projects p
WHERE p.name = 'Project 1'
ON CONFLICT (project_id, year, month) DO NOTHING;

INSERT INTO project_dollar_values (project_id, year, month, start_value, end_value, average_value, created_by, updated_by, created_at, updated_at)
SELECT 
  p.id, 2025, 'Octubre', 1080.00, 1100.00, 1090.00, 2, 2, NOW(), NOW()
FROM projects p
WHERE p.name = 'Project 1'
ON CONFLICT (project_id, year, month) DO NOTHING;

INSERT INTO project_dollar_values (project_id, year, month, start_value, end_value, average_value, created_by, updated_by, created_at, updated_at)
SELECT 
  p.id, 2025, 'Noviembre', 1100.00, 1120.00, 1110.00, 2, 2, NOW(), NOW()
FROM projects p
WHERE p.name = 'Project 1'
ON CONFLICT (project_id, year, month) DO NOTHING;

INSERT INTO project_dollar_values (project_id, year, month, start_value, end_value, average_value, created_by, updated_by, created_at, updated_at)
SELECT 
  p.id, 2025, 'Diciembre', 1120.00, 1150.00, 1135.00, 2, 2, NOW(), NOW()
FROM projects p
WHERE p.name = 'Project 1'
ON CONFLICT (project_id, year, month) DO NOTHING;

-- Verificar inserción
SELECT '✅ Valores del dólar por proyecto creados:' as status, COUNT(*) as total FROM project_dollar_values; 