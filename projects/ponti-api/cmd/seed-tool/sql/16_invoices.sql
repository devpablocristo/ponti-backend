-- Crear facturas de prueba para el sistema
-- Las facturas se crean asociando órdenes de trabajo existentes
-- Incluye diferentes estados: Pagada, Pendiente
-- Cada factura tiene un número único y está asociada a una empresa

INSERT INTO invoices (work_order_id, number, company, date, status, created_by, updated_by, created_at, updated_at)
SELECT 
  wo.id, 'INV-001', 'AgroServicios SRL', '2025-06-01'::timestamp, 'Pagada', 2, 2, NOW(), NOW()
FROM workorders wo
WHERE wo.number = 'WO-001'
ON CONFLICT (work_order_id) DO NOTHING;

INSERT INTO invoices (work_order_id, number, company, date, status, created_by, updated_by, created_at, updated_at)
SELECT 
  wo.id, 'INV-002', 'Cosechadoras del Norte', '2025-12-15'::timestamp, 'Pendiente', 2, 2, NOW(), NOW()
FROM workorders wo
WHERE wo.number = 'WO-002'
ON CONFLICT (work_order_id) DO NOTHING;

INSERT INTO invoices (work_order_id, number, company, date, status, created_by, updated_by, created_at, updated_at)
SELECT 
  wo.id, 'INV-003', 'Fertilizantes del Sur', '2025-06-15'::timestamp, 'Pagada', 2, 2, NOW(), NOW()
FROM workorders wo
WHERE wo.number = 'WO-003'
ON CONFLICT (work_order_id) DO NOTHING;

INSERT INTO invoices (work_order_id, number, company, date, status, created_by, updated_by, created_at, updated_at)
SELECT 
  wo.id, 'INV-004', 'Transportes Rápidos', '2025-12-20'::timestamp, 'Pendiente', 2, 2, NOW(), NOW()
FROM workorders wo
WHERE wo.number = 'WO-004'
ON CONFLICT (work_order_id) DO NOTHING;

INSERT INTO invoices (work_order_id, number, company, date, status, created_by, updated_by, created_at, updated_at)
SELECT 
  wo.id, 'INV-005', 'Maquinaria Agrícola', '2025-07-01'::timestamp, 'Pagada', 2, 2, NOW(), NOW()
FROM workorders wo
WHERE wo.number = 'WO-005'
ON CONFLICT (work_order_id) DO NOTHING;

-- Verificar inserción
SELECT '✅ Facturas creadas:' as status, COUNT(*) as total FROM invoices; 