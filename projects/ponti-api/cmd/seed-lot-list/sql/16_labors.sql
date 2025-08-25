-- =======================
-- LABORS PARA LIST LOT
-- =======================
-- Labores para el proyecto (categorías corregidas: 9=Siembra, 13=Cosecha)

INSERT INTO labors (id, name, project_id, price, category_id, contractor_name, created_at, updated_at) VALUES
(1, 'Siembra Maíz', 1, 25.50, 9, 'Juan Pérez', NOW(), NOW()),
(2, 'Cosecha Maíz', 1, 45.00, 13, 'María García', NOW(), NOW());

-- Verificar inserción
SELECT '✅ Labors creadas' as status, COUNT(*) as total FROM labors;
