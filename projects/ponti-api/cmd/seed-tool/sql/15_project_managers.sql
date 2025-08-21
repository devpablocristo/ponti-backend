-- Crear relaciones entre proyectos y managers de prueba
-- Las relaciones se crean asociando proyectos y managers existentes
-- Cada proyecto puede tener uno o más managers asignados para su gestión

INSERT INTO project_managers (project_id, manager_id, created_by, updated_by, created_at, updated_at)
SELECT 
  p.id, m.id, 2, 2, NOW(), NOW()
FROM projects p, managers m
WHERE p.name = 'Project 1' AND m.name = 'Manager Norte'
ON CONFLICT (project_id, manager_id) DO NOTHING;

INSERT INTO project_managers (project_id, manager_id, created_by, updated_by, created_at, updated_at)
SELECT 
  p.id, m.id, 2, 2, NOW(), NOW()
FROM projects p, managers m
WHERE p.name = 'Project 2' AND m.name = 'Manager Sur'
ON CONFLICT (project_id, manager_id) DO NOTHING;

INSERT INTO project_managers (project_id, manager_id, created_by, updated_by, created_at, updated_at)
SELECT 
  p.id, m.id, 2, 2, NOW(), NOW()
FROM projects p, managers m
WHERE p.name = 'Project 3' AND m.name = 'Manager Este'
ON CONFLICT (project_id, manager_id) DO NOTHING;

INSERT INTO project_managers (project_id, manager_id, created_by, updated_by, created_at, updated_at)
SELECT 
  p.id, m.id, 2, 2, NOW(), NOW()
FROM projects p, managers m
WHERE p.name = 'Project 4' AND m.name = 'Manager Oeste'
ON CONFLICT (project_id, manager_id) DO NOTHING;

INSERT INTO project_managers (project_id, manager_id, created_by, updated_by, created_at, updated_at)
SELECT 
  p.id, m.id, 2, 2, NOW(), NOW()
FROM projects p, managers m
WHERE p.name = 'Project 5' AND m.name = 'Manager Central'
ON CONFLICT (project_id, manager_id) DO NOTHING;

-- Verificar inserción
SELECT '✅ Relaciones proyecto-managers creadas:' as status, COUNT(*) as total FROM project_managers; 