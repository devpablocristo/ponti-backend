-- =======================
-- PROJECT MANAGERS PARA LIST LOT
-- =======================
-- Asociar managers con el proyecto

INSERT INTO project_managers (project_id, manager_id, created_at, updated_at) VALUES
(1, 1, NOW(), NOW()),
(1, 2, NOW(), NOW());

-- Verificar inserción
SELECT '✅ Project managers creados' as status, COUNT(*) as total FROM project_managers;
