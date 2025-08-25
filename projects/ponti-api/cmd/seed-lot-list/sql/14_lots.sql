-- =======================
-- LOTS PARA LIST LOT
-- =======================
-- Lotes para los campos

INSERT INTO lots (id, name, field_id, hectares, previous_crop_id, current_crop_id, season, variety, created_at, updated_at) VALUES
(1, 'Parcela A1', 1, 2.5, 1, 2, 'Invierno 2025', 'DK 72-10', NOW(), NOW()),
(2, 'Parcela A2', 1, 3.0, 1, 2, 'Verano 2025', 'DK 72-10', NOW(), NOW()),
(3, 'Parcela B1', 2, 1.2, 1, 2, 'Otoño 2025', 'DK 72-10', NOW(), NOW());

-- Verificar inserción
SELECT '✅ Lots creados' as status, COUNT(*) as total FROM lots;
