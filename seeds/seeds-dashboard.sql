ROLLBACK;

-- =========================================================
-- CLEAN (reset para pruebas mínimas)
-- =========================================================
TRUNCATE workorder_items RESTART IDENTITY CASCADE;
TRUNCATE invoices RESTART IDENTITY CASCADE;
TRUNCATE workorders RESTART IDENTITY CASCADE;
TRUNCATE labors RESTART IDENTITY CASCADE;
TRUNCATE supplies RESTART IDENTITY CASCADE;
TRUNCATE lots RESTART IDENTITY CASCADE;
TRUNCATE fields RESTART IDENTITY CASCADE;
TRUNCATE projects RESTART IDENTITY CASCADE;
TRUNCATE customers RESTART IDENTITY CASCADE;
TRUNCATE campaigns RESTART IDENTITY CASCADE;
TRUNCATE lease_types RESTART IDENTITY CASCADE;
TRUNCATE crops RESTART IDENTITY CASCADE;
TRUNCATE investors RESTART IDENTITY CASCADE;
TRUNCATE types RESTART IDENTITY CASCADE;
TRUNCATE categories RESTART IDENTITY CASCADE;
TRUNCATE project_investors RESTART IDENTITY CASCADE;

-- =========================================================
-- BASE
-- =========================================================
INSERT INTO customers (id, name, deleted_at) VALUES (1, 'Cliente Demo', NULL);
INSERT INTO campaigns (id, name, deleted_at) VALUES (1, 'Campaña Demo', NULL);

-- Un proyecto con admin_cost para probar cards de costos
INSERT INTO projects (id, customer_id, campaign_id, name, admin_cost, deleted_at) 
VALUES (1, 1, 1, 'Proyecto Único', 1000, NULL);

INSERT INTO lease_types (id, name, deleted_at) VALUES (1, 'Arrendamiento Demo', NULL);

-- Crops (al menos 2 para poder setear previous/current)
INSERT INTO crops (id, name, deleted_at) VALUES 
  (1, 'Trigo', NULL),
  (2, 'Maíz',  NULL);

-- Investor + participación (para contributions)
INSERT INTO investors (id, name, deleted_at) VALUES (1, 'Inversor A', NULL);
INSERT INTO project_investors (project_id, investor_id, percentage, deleted_at) 
VALUES (1, 1, 100, NULL);

-- Field
INSERT INTO fields (id, project_id, name, lease_type_id, deleted_at) 
VALUES (101, 1, 'Campo Demo', 1, NULL);

-- Lot (¡corregido previous_crop_id NOT NULL!)
-- Incluimos sowing_date y tons para que haya siembra/cosecha/ingreso
INSERT INTO lots (
  id, field_id, name, hectares, previous_crop_id, current_crop_id, season, sowing_date, tons, deleted_at
) VALUES
  (1001, 101, 'Lote 1', 50, 2, 1, '2025', '2025-01-05', 100.0, NULL);

-- =========================================================
-- TYPES & CATEGORIES (mínimo para L vs Kg)
-- =========================================================
INSERT INTO types (id, name, deleted_at) VALUES 
  (1, 'Labores', NULL),
  (2, 'Insumos', NULL);

-- Convención: 2=Herbicida (L), 3=Fertilizante (Kg)
INSERT INTO categories (id, name, type_id, deleted_at) VALUES
  (1, 'Labores Generales', 1, NULL),
  (2, 'Herbicida',         2, NULL), -- L
  (3, 'Fertilizante',      2, NULL); -- Kg

-- =========================================================
-- LABORS (precio por Ha) para costos ejecutados
-- =========================================================
INSERT INTO labors (id, project_id, name, category_id, price, contractor_name, deleted_at) VALUES
  (1, 1, 'Siembra',       1, 100, 'Contratista A', NULL),
  (2, 1, 'Fertilización', 1, 150, 'Contratista A', NULL);

-- =========================================================
-- SUPPLIES (uno en L y otro en Kg) para costos y métricas
-- =========================================================
INSERT INTO supplies (id, project_id, name, type_id, category_id, price, deleted_at) VALUES
  (11, 1, 'Glifosato', 2, 2, 8.00,  NULL),  -- L
  (12, 1, 'Urea',      2, 3, 0.50,  NULL);  -- Kg

-- =========================================================
-- WORKORDERS (dos fechas distintas para first/last)
-- =========================================================
INSERT INTO workorders (
  id, project_id, field_id, lot_id, crop_id, labor_id, investor_id, date, effective_area, deleted_at
) VALUES
  (1, 1, 101, 1001, 1, 1, 1, '2025-01-10', 50, NULL),  -- Siembra 50 ha
  (2, 1, 101, 1001, 2, 2, 1, '2025-02-20', 30, NULL);  -- Fertilización 30 ha

-- =========================================================
-- WORKORDER ITEMS (L y Kg para diferenciar unidades)
-- =========================================================
INSERT INTO workorder_items (workorder_id, supply_id, final_dose, total_used, deleted_at) VALUES
  (1, 11, 2.0, 100.0, NULL),   -- 2 L/Ha * 50 Ha = 100 L
  (2, 12, 1.5, 45.0,  NULL);   -- 1.5 Kg/Ha * 30 Ha = 45 Kg


SELECT *
FROM dashboard_view
WHERE customer_id IS NULL
  AND project_id IS NULL
  AND campaign_id IS NULL
  AND field_id IS NULL;