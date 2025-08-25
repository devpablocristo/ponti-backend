-- =======================
-- FIELDS PARA LIST LOT
-- =======================
-- Campos para el proyecto

INSERT INTO fields (id, name, project_id, lease_type_id, lease_type_value, lease_type_percent, created_at, updated_at) VALUES
(1, 'Campo A', 1, 1, 100, NULL, NOW(), NOW()),
(2, 'Campo B', 1, 2, NULL, 15, NOW(), NOW());

-- Verificar inserción
SELECT '✅ Fields creados' as status, COUNT(*) as total FROM fields;
