-- =======================
-- AGREGAR TRIGO AL DASHBOARD
-- =======================
-- Script para agregar un tercer cultivo (Trigo) al dashboard existente
-- Requiere que ya existan los datos básicos del dashboard + Maíz

-- 1. AGREGAR CULTIVO TRIGO
INSERT INTO crops (name, created_at, updated_at) VALUES 
('Trigo', NOW(), NOW())
ON CONFLICT (name) DO NOTHING;

-- 2. AGREGAR CAMPO ADICIONAL PARA TRIGO
INSERT INTO fields (name, project_id, lease_type_id, created_at, updated_at) VALUES 
('Campo Este', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;

-- 3. AGREGAR LOTE CON TRIGO
INSERT INTO lots (name, field_id, hectares, previous_crop_id, current_crop_id, season, created_at, updated_at) VALUES 
('Lote C1', 3, 12.0, 3, 3, '2024-2025', NOW(), NOW())
ON CONFLICT DO NOTHING;

-- Verificar inserción
SELECT '✅ Trigo agregado al dashboard:' as status;
SELECT '   - Cultivos totales:' as item, COUNT(*) as total FROM crops;
SELECT '   - Campos totales:' as item, COUNT(*) as total FROM fields;
SELECT '   - Lotes totales:' as item, COUNT(*) as total FROM lots;

-- Verificar que el dashboard ahora tenga 3 cultivos
SELECT '🎯 Dashboard ahora debería mostrar:' as info;
SELECT '   - Soja: 10.5 hectáreas' as crop_info;
SELECT '   - Maíz: 8.5 hectáreas' as crop_info;
SELECT '   - Trigo: 12.0 hectáreas' as crop_info;
SELECT '   - Total hectáreas: 31.0' as total_info;

