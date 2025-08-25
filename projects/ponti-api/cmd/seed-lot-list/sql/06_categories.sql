-- =======================
-- CATEGORIES PARA LIST LOT
-- =======================
-- Categorías necesarias para supplies y labors

INSERT INTO categories (id, name, type_id, created_at, updated_at) VALUES
(1, 'Semilla', 1, NOW(), NOW()),
(9, 'Siembra', 4, NOW(), NOW()),
(13, 'Cosecha', 4, NOW(), NOW());

-- Verificar inserción
SELECT '✅ Categories creadas' as status, COUNT(*) as total FROM categories;
