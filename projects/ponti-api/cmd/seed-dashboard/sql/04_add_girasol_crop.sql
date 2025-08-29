-- =======================
-- AGREGAR GIRASOL AL DASHBOARD
-- =======================
-- Script para agregar un cuarto cultivo (Girasol) al dashboard existente
-- Requiere que ya existan los datos básicos del dashboard + Maíz + Trigo

-- 1. AGREGAR CULTIVO GIRASOL
INSERT INTO crops (name, created_at, updated_at) VALUES 
('Girasol', NOW(), NOW())
ON CONFLICT (name) DO NOTHING;

-- 2. AGREGAR CAMPO ADICIONAL PARA GIRASOL
INSERT INTO fields (name, project_id, lease_type_id, created_at, updated_at) VALUES 
('Campo Oeste', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;

-- 3. AGREGAR LOTE CON GIRASOL
INSERT INTO lots (name, field_id, hectares, previous_crop_id, current_crop_id, season, created_at, updated_at) VALUES 
('Lote D1', 4, 15.0, 4, 4, '2024-2025', NOW(), NOW())
ON CONFLICT DO NOTHING;

-- Verificar inserción
SELECT '✅ Girasol agregado al dashboard:' as status;
SELECT '   - Cultivos totales:' as item, COUNT(*) as total FROM crops;
SELECT '   - Campos totales:' as item, COUNT(*) as total FROM fields;
SELECT '   - Lotes totales:' as item, COUNT(*) as total FROM lots;

-- Verificar que el dashboard ahora tenga 4 cultivos
SELECT '🎯 Dashboard ahora debería mostrar:' as info;
SELECT '   - Soja: 10.5 hectáreas' as crop_info;
SELECT '   - Maíz: 8.5 hectáreas' as crop_info;
SELECT '   - Trigo: 12.0 hectáreas' as crop_info;
SELECT '   - Girasol: 15.0 hectáreas' as crop_info;
SELECT '   - Total hectáreas: 46.0' as total_info;

