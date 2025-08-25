-- =======================
-- LOT DATES PARA LIST LOT
-- =======================
-- Fechas de siembra y cosecha para los lotes

INSERT INTO lot_dates (lot_id, sowing_date, harvest_date, sequence, created_by, updated_by, created_at, updated_at) VALUES
(1, '2025-06-15 00:00:00', '2025-12-15 00:00:00', 1, 123, 123, NOW(), NOW()),
(2, '2025-12-15 00:00:00', '2025-06-15 00:00:00', 1, 123, 123, NOW(), NOW()),
(3, '2025-03-15 00:00:00', '2025-09-15 00:00:00', 1, 123, 123, NOW(), NOW());

-- Verificar inserción
SELECT '✅ Lot dates creadas' as status, COUNT(*) as total FROM lot_dates;
