ROLLBACK;

-- ========================================
-- CLEAN
-- ========================================
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

-- ========================================
-- BASE / PARENTS
-- ========================================
INSERT INTO customers (id, name) VALUES (1, 'Cliente Demo');
INSERT INTO campaigns (id, name) VALUES (1, 'Campaña Demo');
INSERT INTO projects (id, customer_id, campaign_id, name, admin_cost)
VALUES (1, 1, 1, 'Proyecto 1', 0);

INSERT INTO lease_types (id, name) VALUES (1, 'Arrendamiento Demo');
INSERT INTO crops (id, name) VALUES (1, 'Trigo'), (2, 'Maíz');
INSERT INTO investors (id, name) VALUES (1, 'Inversor A');

INSERT INTO fields (id, project_id, name, lease_type_id)
VALUES (101, 1, 'Field 101', 1);

INSERT INTO lots (id, field_id, name, hectares, previous_crop_id, current_crop_id, season, tons)
VALUES
  (1001, 101, 'Lot 1001', 50, 1, 2, '2025', 120), -- 120 toneladas declaradas
  (1002, 101, 'Lot 1002', 30, 2, 1, '2025', 60);  -- 60 toneladas declaradas

-- ========================================
-- LABORS (usando categorías correctas: 9=Siembra, 13=Cosecha)
-- ========================================
INSERT INTO labors (id, project_id, name, category_id, price, contractor_name)
VALUES
  (1, 1, 'Siembra', 9, 100, 'Contratista A'),
  (2, 1, 'Cosecha', 13, 200, 'Contratista B');

-- ========================================
-- WORKORDERS
-- ========================================
-- Lote 1001: sembrado 40 ha, cosechado 35 ha
INSERT INTO workorders (id, project_id, field_id, lot_id, crop_id, labor_id, investor_id, date, effective_area)
VALUES
  (1, 1, 101, 1001, 1, 1, 1, '2025-01-10', 40),  -- Siembra
  (2, 1, 101, 1001, 1, 2, 1, '2025-03-15', 35);  -- Cosecha

-- Lote 1002: sembrado 20 ha, cosechado 25 ha
INSERT INTO workorders (id, project_id, field_id, lot_id, crop_id, labor_id, investor_id, date, effective_area)
VALUES
  (3, 1, 101, 1002, 2, 1, 1, '2025-01-12', 20),  -- Siembra
  (4, 1, 101, 1002, 2, 2, 1, '2025-04-20', 25);  -- Cosecha

-- ========================================
-- TEST
-- ========================================
SELECT * FROM lot_metrics_view
ORDER BY field_id, previous_crop_id, current_crop_id;
