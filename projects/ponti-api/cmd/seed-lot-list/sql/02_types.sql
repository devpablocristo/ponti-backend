-- =======================
-- TIPOS BASE PARA LIST LOT
-- =======================
-- Tipos necesarios para supplies y labors

INSERT INTO types (id, name, created_at, updated_at) VALUES
(1, 'Semilla', NOW(), NOW()),
(2, 'Fertilizante', NOW(), NOW()),
(3, 'Herbicida', NOW(), NOW()),
(4, 'Labor', NOW(), NOW());

-- Verificar inserción
SELECT '✅ Tipos base creados' as status, COUNT(*) as total FROM types;
