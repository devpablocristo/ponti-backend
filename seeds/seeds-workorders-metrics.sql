ROLLBACK;

-- =========================================================
-- CLEAN
-- =========================================================
TRUNCATE workorder_items RESTART IDENTITY CASCADE;
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
INSERT INTO customers (id, name) VALUES (1, 'Cliente Demo');
INSERT INTO campaigns (id, name) VALUES (1, 'Campaña Demo');

INSERT INTO projects (id, customer_id, campaign_id, name, admin_cost)
VALUES (1, 1, 1, 'Proyecto Demo', 1000);

INSERT INTO lease_types (id, name) VALUES (1, 'Arrendamiento Demo');

INSERT INTO crops (id, name) VALUES
  (1, 'Trigo'),
  (2, 'Maíz');

INSERT INTO investors (id, name) VALUES
  (1, 'Inversor A'),
  (2, 'Inversor B');

INSERT INTO project_investors (project_id, investor_id, percentage) VALUES
  (1, 1, 60.0),
  (1, 2, 40.0);

INSERT INTO fields (id, project_id, name, lease_type_id)
VALUES (101, 1, 'Field Demo', 1);

-- 1 lote con cultivo
INSERT INTO lots (id, field_id, name, hectares, previous_crop_id, current_crop_id, season)
VALUES (1001, 101, 'Lote Demo', 50, 1, 2, '2025');

-- =========================================================
-- TYPES & CATEGORIES (mínimo viable)
-- =========================================================
INSERT INTO types (id, name) VALUES
  (1, 'Labores'),
  (2, 'Insumos');

INSERT INTO categories (id, name, type_id) VALUES
  (1, 'Labores Generales', 1),
  (2, 'Herbicida',         2),
  (3, 'Fertilizante',      2),
  (4, 'Semilla',           2);

-- =========================================================
-- LABORS (precio por Ha)
-- =========================================================
INSERT INTO labors (id, project_id, name, category_id, price, contractor_name)
VALUES (1, 1, 'Siembra', 1, 100, 'Contratista Demo');

-- =========================================================
-- SUPPLIES
-- =========================================================
INSERT INTO supplies (id, project_id, name, type_id, category_id, price) VALUES
  (1, 1, 'Glifosato', 2, 2, 10.00),   -- L
  (2, 1, 'Urea',      2, 3, 1.00),    -- Kg
  (3, 1, 'Semilla',   2, 4, 2.00);    -- Kg

-- =========================================================
-- WORKORDERS
-- =========================================================
INSERT INTO workorders (id, project_id, field_id, lot_id, crop_id, labor_id, investor_id, date, effective_area)
VALUES
  (1, 1, 101, 1001, 2, 1, 1, '2025-03-01', 50);

-- =========================================================
-- WORKORDER ITEMS
-- =========================================================
-- Con esto hay litros, kilos y semilla → cubre todas las métricas
INSERT INTO workorder_items (workorder_id, supply_id, final_dose, total_used) VALUES
  (1, 1, 2.0, 100.0),  -- Glifosato: 2 L/Ha * 50 = 100 L
  (1, 2, 1.0, 50.0),   -- Urea: 1 Kg/Ha * 50 = 50 Kg
  (1, 3, 3.0, 150.0);  -- Semilla: 3 Kg/Ha * 50 = 150 Kg
