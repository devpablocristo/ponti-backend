-- Crear movimientos de suministros de prueba para el sistema
-- Los movimientos se crean asociando stocks, proyectos, suministros, inversores y proveedores existentes
-- Incluye diferentes tipos de movimientos: Stock, Movimiento interno
-- NOTA: La columna field_id fue eliminada en la migración 000032

INSERT INTO supply_movements (stock_id, quantity, movement_type, movement_date, reference_number, is_entry, project_id, project_destination_id, supply_id, investor_id, provider_id, created_by, updated_by, created_at, updated_at) VALUES
(1, 100.000, 'Stock', '2024-01-15 10:00:00', 'REF-001', true, 1, 1, 1, 1, 1, 1, 1, NOW(), NOW()),
(1, -25.000, 'Movimiento interno', '2024-01-20 14:30:00', 'REF-002', false, 1, 1, 1, 1, 1, 1, 1, NOW(), NOW()),
(2, 50.000, 'Stock', '2024-01-16 09:15:00', 'REF-003', true, 1, 1, 2, 1, 2, 1, 1, NOW(), NOW()),
(2, -10.000, 'Movimiento interno', '2024-01-25 16:45:00', 'REF-004', false, 1, 1, 2, 1, 2, 1, 1, NOW(), NOW()),
(3, 75.000, 'Stock', '2024-01-17 11:20:00', 'REF-005', true, 1, 1, 3, 1, 3, 1, 1, NOW(), NOW()),
(3, -15.000, 'Movimiento interno', '2024-01-30 13:10:00', 'REF-006', false, 1, 1, 3, 1, 3, 1, 1, NOW(), NOW()),
(4, 30.000, 'Stock', '2024-01-18 08:30:00', 'REF-007', true, 1, 1, 4, 1, 4, 1, 1, NOW(), NOW()),
(4, -5.000, 'Movimiento interno', '2024-02-05 15:20:00', 'REF-008', false, 1, 1, 4, 1, 4, 1, 1, NOW(), NOW()),
(5, 60.000, 'Stock', '2024-01-19 12:45:00', 'REF-009', true, 1, 1, 5, 1, 5, 1, 1, NOW(), NOW()),
(5, -12.000, 'Movimiento interno', '2024-02-10 17:30:00', 'REF-010', false, 1, 1, 5, 1, 5, 1, 1, NOW(), NOW());

-- Verificar inserción
SELECT '✅ Movimientos de suministros creados:' as status, COUNT(*) as total FROM supply_movements;
