-- Crear relaciones entre proyectos e inversores de prueba
-- Las relaciones se crean asociando proyectos e inversores existentes
-- Incluye porcentaje de participación de cada inversor en el proyecto

INSERT INTO project_investors (project_id, investor_id, percentage, created_by, updated_by, created_at, updated_at)
SELECT 
  p.id, i.id, 40, 2, 2, NOW(), NOW()
FROM projects p, investors i
WHERE p.name = 'Project 1' AND i.name = 'Inversor Principal'
ON CONFLICT (project_id, investor_id) DO NOTHING;

INSERT INTO project_investors (project_id, investor_id, percentage, created_by, updated_by, created_at, updated_at)
SELECT 
  p.id, i.id, 30, 2, 2, NOW(), NOW()
FROM projects p, investors i
WHERE p.name = 'Project 1' AND i.name = 'Inversor Secundario'
ON CONFLICT (project_id, investor_id) DO NOTHING;

INSERT INTO project_investors (project_id, investor_id, percentage, created_by, updated_by, created_at, updated_at)
SELECT 
  p.id, i.id, 30, 2, 2, NOW(), NOW()
FROM projects p, investors i
WHERE p.name = 'Project 1' AND i.name = 'Inversor Estratégico'
ON CONFLICT (project_id, investor_id) DO NOTHING;

INSERT INTO project_investors (project_id, investor_id, percentage, created_by, updated_by, created_at, updated_at)
SELECT 
  p.id, i.id, 50, 2, 2, NOW(), NOW()
FROM projects p, investors i
WHERE p.name = 'Project 2' AND i.name = 'Inversor Principal'
ON CONFLICT (project_id, investor_id) DO NOTHING;

INSERT INTO project_investors (project_id, investor_id, percentage, created_by, updated_by, created_at, updated_at)
SELECT 
  p.id, i.id, 50, 2, 2, NOW(), NOW()
FROM projects p, investors i
WHERE p.name = 'Project 2' AND i.name = 'Inversor Secundario'
ON CONFLICT (project_id, investor_id) DO NOTHING;

INSERT INTO project_investors (project_id, investor_id, percentage, created_by, updated_by, created_at, updated_at)
SELECT 
  p.id, i.id, 100, 2, 2, NOW(), NOW()
FROM projects p, investors i
WHERE p.name = 'Project 3' AND i.name = 'Inversor Principal'
ON CONFLICT (project_id, investor_id) DO NOTHING;

-- Verificar inserción
SELECT '✅ Relaciones proyecto-inversores creadas:' as status, COUNT(*) as total FROM project_investors; 