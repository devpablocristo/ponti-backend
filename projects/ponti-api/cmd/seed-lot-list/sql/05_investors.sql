-- =======================
-- INVESTORS PARA LIST LOT
-- =======================
-- Inversores para el proyecto

INSERT INTO investors (id, name, created_at, updated_at) VALUES
(1, 'Fondo Capital Innovador', NOW(), NOW()),
(2, 'Grupo Inversor del Sur', NOW(), NOW());

-- Verificar inserción
SELECT '✅ Investors creados' as status, COUNT(*) as total FROM investors;
