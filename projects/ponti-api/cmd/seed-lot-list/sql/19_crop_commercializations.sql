-- =======================
-- CROP COMMERCIALIZATIONS PARA LIST LOT
-- =======================
-- Precios de comercialización para los cultivos

INSERT INTO crop_commercializations (id, project_id, crop_id, board_price, freight_cost, commercial_cost, net_price, created_at, updated_at) VALUES
(1, 1, 2, 600.00, 50.00, 100.00, 450.00, NOW(), NOW()),
(2, 1, 1, 800.00, 50.00, 100.00, 650.00, NOW(), NOW());

-- Verificar inserción
SELECT '✅ Crop commercializations creadas' as status, COUNT(*) as total FROM crop_commercializations;
