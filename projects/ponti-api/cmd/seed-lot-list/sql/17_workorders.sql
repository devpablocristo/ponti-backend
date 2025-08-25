-- =======================
-- WORKORDERS PARA LIST LOT
-- =======================
-- Órdenes de trabajo para los lotes (fechas corregidas según estaciones)

-- Lote 1 - Parcela A1 (Invierno 2025)
INSERT INTO workorders (id, number, project_id, field_id, lot_id, crop_id, labor_id, investor_id, date, effective_area, created_at, updated_at) VALUES
(1, 'WO-SOWING-A1-001', 1, 1, 1, 2, 1, 1, '2025-06-15', 2.5, NOW(), NOW()),
(2, 'WO-HARVEST-A1-001', 1, 1, 1, 2, 2, 1, '2025-12-15', 2.5, NOW(), NOW());

-- Lote 2 - Parcela A2 (Verano 2025)
INSERT INTO workorders (id, number, project_id, field_id, lot_id, crop_id, labor_id, investor_id, date, effective_area, created_at, updated_at) VALUES
(3, 'WO-SOWING-A2-001', 1, 1, 2, 2, 1, 1, '2025-12-15', 3.0, NOW(), NOW()),
(4, 'WO-HARVEST-A2-001', 1, 1, 2, 2, 2, 1, '2025-06-15', 3.0, NOW(), NOW());

-- Lote 3 - Parcela B1 (Otoño 2025)
INSERT INTO workorders (id, number, project_id, field_id, lot_id, crop_id, labor_id, investor_id, date, effective_area, created_at, updated_at) VALUES
(5, 'WO-SOWING-B1-001', 1, 2, 3, 2, 1, 1, '2025-03-15', 1.2, NOW(), NOW()),
(6, 'WO-HARVEST-B1-001', 1, 2, 3, 2, 2, 1, '2025-09-15', 1.2, NOW(), NOW());

-- Verificar inserción
SELECT '✅ Workorders creados' as status, COUNT(*) as total FROM workorders;
