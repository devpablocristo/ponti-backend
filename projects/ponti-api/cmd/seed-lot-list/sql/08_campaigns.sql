-- =======================
-- CAMPAIGNS PARA LIST LOT
-- =======================
-- Campañas para el proyecto

INSERT INTO campaigns (id, name, created_at, updated_at) VALUES
(1, 'Campaña Loteo 2025', NOW(), NOW());

-- Verificar inserción
SELECT '✅ Campaigns creadas' as status, COUNT(*) as total FROM campaigns;
