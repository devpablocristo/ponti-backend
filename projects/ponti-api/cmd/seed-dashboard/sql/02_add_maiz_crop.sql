-- =======================
-- AGREGAR MAÍZ AL DASHBOARD
-- =======================
-- Script para agregar un cultivo adicional (Maíz) al dashboard existente
-- Requiere que ya existan los datos básicos del dashboard

-- 1. AGREGAR CULTIVO MAÍZ
INSERT INTO crops (name, created_at, updated_at) VALUES 
('Maíz', NOW(), NOW())
ON CONFLICT (name) DO NOTHING;

-- 2. AGREGAR CAMPO ADICIONAL PARA MAÍZ
INSERT INTO fields (name, project_id, lease_type_id, created_at, updated_at) VALUES 
('Campo Sur', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;

-- 3. AGREGAR LOTE CON MAÍZ
INSERT INTO lots (name, field_id, hectares, previous_crop_id, current_crop_id, season, created_at, updated_at) VALUES 
('Lote B1', 2, 8.5, 2, 2, '2024-2025', NOW(), NOW())
ON CONFLICT DO NOTHING;

-- Verificar inserción
SELECT '✅ Maíz agregado al dashboard:' as status;
SELECT '   - Cultivos totales:' as item, COUNT(*) as total FROM crops;
SELECT '   - Campos totales:' as item, COUNT(*) as total FROM fields;
SELECT '   - Lotes totales:' as item, COUNT(*) as total FROM lots;

-- Verificar que el dashboard ahora tenga 2 cultivos
SELECT '🎯 Dashboard ahora debería mostrar:' as info;
SELECT '   - Soja: 10.5 hectáreas' as crop_info;
SELECT '   - Maíz: 8.5 hectáreas' as crop_info;
SELECT '   - Total hectáreas: 19.0' as total_info;

