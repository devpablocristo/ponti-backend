-- Crear stocks de suministros de prueba para el sistema
-- Los stocks se crean asociando proyectos, suministros e inversores existentes
-- Incluye información de stock real, inicial y período
-- NOTA: La columna field_id fue eliminada en la migración 000031

INSERT INTO stocks (project_id, supply_id, investor_id, close_date, real_stock_units, initial_units, year_period, month_period, created_by, updated_by, created_at, updated_at) VALUES
(1, 1, 1, '2024-12-31', 1000.000, 1000.000, 2024, 1, 1, 1, NOW(), NOW()),
(1, 2, 1, '2024-12-31', 500.000, 500.000, 2024, 1, 1, 1, NOW(), NOW()),
(1, 3, 1, '2024-12-31', 750.000, 750.000, 2024, 1, 1, 1, NOW(), NOW()),
(1, 4, 1, '2024-12-31', 300.000, 300.000, 2024, 1, 1, 1, NOW(), NOW()),
(1, 5, 1, '2024-12-31', 600.000, 600.000, 2024, 1, 1, 1, NOW(), NOW()),
(2, 6, 2, '2024-12-31', 400.000, 400.000, 2024, 1, 1, 1, NOW(), NOW()),
(2, 7, 2, '2024-12-31', 800.000, 800.000, 2024, 1, 1, 1, NOW(), NOW()),
(2, 8, 2, '2024-12-31', 350.000, 350.000, 2024, 1, 1, 1, NOW(), NOW()),
(3, 9, 3, '2024-12-31', 450.000, 450.000, 2024, 1, 1, 1, NOW(), NOW()),
(3, 10, 3, '2024-12-31', 550.000, 550.000, 2024, 1, 1, 1, NOW(), NOW());

-- Verificar inserción
SELECT '✅ Stocks creados:' as status, COUNT(*) as total FROM stocks;
