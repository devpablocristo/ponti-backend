ROLLBACK;

-- ========================================
-- CLEAN
-- ========================================
TRUNCATE workorder_items RESTART IDENTITY CASCADE;
TRUNCATE invoices RESTART IDENTITY CASCADE;
TRUNCATE workorders RESTART IDENTITY CASCADE;
TRUNCATE labors RESTART IDENTITY CASCADE;
TRUNCATE lots RESTART IDENTITY CASCADE;
TRUNCATE fields RESTART IDENTITY CASCADE;
TRUNCATE projects RESTART IDENTITY CASCADE;
TRUNCATE customers RESTART IDENTITY CASCADE;
TRUNCATE campaigns RESTART IDENTITY CASCADE;
TRUNCATE lease_types RESTART IDENTITY CASCADE;
TRUNCATE crops RESTART IDENTITY CASCADE;
TRUNCATE investors RESTART IDENTITY CASCADE;

-- ========================================
-- MIN PARENTS
-- ========================================
INSERT INTO customers (id, name, deleted_at) VALUES
  (1, 'Cliente Demo', NULL);

INSERT INTO campaigns (id, name, deleted_at) VALUES
  (1, 'Campaña Demo', NULL);

INSERT INTO projects (id, customer_id, campaign_id, name, admin_cost, deleted_at)
VALUES
  (1, 1, 1, 'Proyecto 1', 0, NULL),
  (2, 1, 1, 'Proyecto 2', 0, NULL),
  (3, 1, 1, 'Proyecto 3', 0, NULL);

INSERT INTO lease_types (id, name, deleted_at) VALUES
  (1, 'Arrendamiento Demo', NULL);

INSERT INTO crops (id, name, deleted_at) VALUES
  (1, 'Trigo', NULL),
  (2, 'Maíz', NULL),
  (3, 'Soja', NULL);

INSERT INTO investors (id, name, deleted_at) VALUES
  (1, 'Inversor A', NULL),
  (2, 'Inversor B', NULL),
  (3, 'Inversor C', NULL);

INSERT INTO fields (id, project_id, name, lease_type_id, deleted_at)
VALUES
  (101, 1, 'Field 101', 1, NULL),
  (102, 1, 'Field 102', 1, NULL),
  (201, 2, 'Field 201', 1, NULL),
  (301, 3, 'Field 301', 1, NULL),
  (302, 3, 'Field 302', 1, NULL);

INSERT INTO lots (id, field_id, name, hectares, previous_crop_id, current_crop_id, season, deleted_at)
VALUES
  (1001, 101, 'Lot 1001', 50, 1, 2, '2025', NULL),
  (1002, 102, 'Lot 1002', 20, 2, 3, '2025', NULL),
  (2001, 201, 'Lot 2001', 40, 3, 1, '2025', NULL),
  (3001, 301, 'Lot 3001', 15, 1, 2, '2025', NULL),
  (3002, 302, 'Lot 3002', 60, 2, 3, '2025', NULL);

-- ========================================
-- LABORS
-- ========================================
INSERT INTO labors (id, project_id, name, category_id, price, contractor_name, deleted_at)
VALUES
  (1, 1, 'Siembra',       1, 100, 'Contratista A', NULL),
  (2, 1, 'Fertilización', 1, 150, 'Contratista A', NULL),
  (3, 1, 'Cosecha',       1, 200, 'Contratista A', NULL),
  (11, 2, 'Siembra',       1, 105, 'Contratista B', NULL),
  (12, 2, 'Fertilización', 1, 155, 'Contratista B', NULL),
  (13, 2, 'Cosecha',       1, 205, 'Contratista B', NULL),
  (21, 3, 'Siembra',       1, 110, 'Contratista C', NULL),
  (22, 3, 'Fertilización', 1, 160, 'Contratista C', NULL),
  (23, 3, 'Cosecha',       1, 210, 'Contratista C', NULL);

-- ========================================
-- WORKORDERS
-- ========================================
INSERT INTO workorders (id, project_id, field_id, lot_id, crop_id, labor_id, investor_id, date, effective_area, deleted_at)
VALUES
  (1, 1, 101, 1001, 1,  1, 1, '2025-01-10', 50, NULL),
  (2, 1, 101, 1001, 2,  2, 1, '2025-02-15', 30, NULL),
  (3, 1, 102, 1002, 3,  3, 1, '2025-03-20', 20, NULL),
  (4, 2, 201, 2001, 1, 11, 2, '2025-01-12', 40, NULL),
  (5, 2, 201, 2001, 2, 12, 2, '2025-02-18', 25, NULL),
  (6, 3, 301, 3001, 3, 23, 3, '2025-03-05', 15, NULL),
  (7, 3, 302, 3002, 1, 21, 3, '2025-04-01', 60, NULL);

-- ========================================
-- TEST
-- ========================================
SELECT * FROM labor_cards_cube_view
ORDER BY level, project_id, field_id;
