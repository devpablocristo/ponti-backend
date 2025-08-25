-- =======================
-- PROJECT INVESTORS PARA LIST LOT
-- =======================
-- Asociar investors con el proyecto

INSERT INTO project_investors (project_id, investor_id, percentage, created_at, updated_at) VALUES
(1, 1, 50, NOW(), NOW()),
(1, 2, 50, NOW(), NOW());

-- Verificar inserción
SELECT '✅ Project investors creados' as status, COUNT(*) as total FROM project_investors;
