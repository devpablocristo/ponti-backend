-- =======================
-- PROJECTS PARA LIST LOT
-- =======================
-- Proyecto principal para list lot

INSERT INTO projects (id, name, admin_cost, customer_id, campaign_id, created_at, updated_at) VALUES
(1, 'Construcción Torre Norte', 15000, 4, 1, NOW(), NOW());

-- Verificar inserción
SELECT '✅ Projects creados' as status, COUNT(*) as total FROM projects;
