-- =======================
-- MANAGERS PARA LIST LOT
-- =======================
-- Managers para el proyecto

INSERT INTO managers (id, name, created_at, updated_at) VALUES
(1, 'María López', NOW(), NOW()),
(2, 'Juan Pérez', NOW(), NOW());

-- Verificar inserción
SELECT '✅ Managers creados' as status, COUNT(*) as total FROM managers;
