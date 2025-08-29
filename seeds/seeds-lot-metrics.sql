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

INSERT INTO projects (id, customer_id, campaign_id, name, admin_cost) VALUES
  (1, 1, 1, 'Proyecto 1', 0),
  (2, 1, 1, 'Proyecto 2', 0),
  (3, 1, 1, 'Proyecto 3', 0),
  (4, 1, 1, 'Proyecto 4', 0),
  (5, 1, 1, 'Proyecto 5', 0);

INSERT INTO lease_types (id, name) VALUES (1, 'Arrendamiento Demo');

INSERT INTO crops (id, name) VALUES
  (1, 'Trigo'),
  (2, 'Maíz'),
  (3, 'Soja');

INSERT INTO investors (id, name) VALUES
  (1, 'Inversor A'),
  (2, 'Inversor B'),
  (3, 'Inversor C');

-- ========================================
-- FIELDS
-- ========================================
INSERT INTO fields (id, project_id, name, lease_type_id) VALUES
  (101, 1, 'Field 101', 1),
  (102, 1, 'Field 102', 1),
  (201, 2, 'Field 201', 1),
  (301, 3, 'Field 301', 1),
  (302, 3, 'Field 302', 1),
  (401, 4, 'Field 401', 1),
  (402, 4, 'Field 402', 1),
  (501, 5, 'Field 501', 1);

-- ========================================
-- LOTES (con toneladas manuales)
-- ========================================
INSERT INTO lots (id, field_id, name, hectares, previous_crop_id, current_crop_id, season, tons) VALUES
  -- Proyecto 1
  (1001, 101, 'Lot 1001', 50, 1, 2, '2025', 120),
  (1002, 101, 'Lot 1002', 30, 2, 1, '2025',  60),
  (1003, 102, 'Lot 1003', 40, 1, 3, '2025',  80),
  (1004, 102, 'Lot 1004', 25, 3, 1, '2025',  55),
  -- Proyecto 2
  (2001, 201, 'Lot 2001', 60, 2, 3, '2025', 150),
  (2002, 201, 'Lot 2002', 35, 3, 2, '2025',  70),
  -- Proyecto 3
  (3001, 301, 'Lot 3001', 45, 1, 2, '2025', 110),
  (3002, 302, 'Lot 3002', 55, 2, 3, '2025', 140),
  -- Proyecto 4
  (4001, 401, 'Lot 4001', 30, 3, 1, '2025',  75),
  (4002, 402, 'Lot 4002', 25, 1, 2, '2025',  65),
  -- Proyecto 5
  (5001, 501, 'Lot 5001', 70, 2, 3, '2025', 180);

-- ========================================
-- LABORS (9=Siembra, 13=Cosecha, varío precios por proyecto)
-- ========================================
INSERT INTO labors (id, project_id, name, category_id, price, contractor_name) VALUES
  -- Proyecto 1
  (1,  1, 'Siembra',  9, 100, 'Contratista A'),
  (2,  1, 'Cosecha', 13, 200, 'Contratista B'),
  -- Proyecto 2
  (11, 2, 'Siembra',  9, 110, 'Contratista C'),
  (12, 2, 'Cosecha', 13, 210, 'Contratista D'),
  -- Proyecto 3
  (21, 3, 'Siembra',  9, 120, 'Contratista E'),
  (22, 3, 'Cosecha', 13, 220, 'Contratista F'),
  -- Proyecto 4
  (31, 4, 'Siembra',  9, 130, 'Contratista G'),
  (32, 4, 'Cosecha', 13, 230, 'Contratista H'),
  -- Proyecto 5
  (41, 5, 'Siembra',  9, 140, 'Contratista I'),
  (42, 5, 'Cosecha', 13, 240, 'Contratista J');

-- ========================================
-- WORKORDERS
-- ========================================
-- Proyecto 1
INSERT INTO workorders (id, project_id, field_id, lot_id, crop_id, labor_id, investor_id, date, effective_area) VALUES
  (1, 1, 101, 1001, 2,  1, 1, '2025-01-10', 40), -- Siembra
  (2, 1, 101, 1001, 2,  2, 1, '2025-03-15', 35), -- Cosecha
  (3, 1, 101, 1002, 1,  1, 1, '2025-01-12', 20), -- Siembra
  (4, 1, 101, 1002, 1,  2, 1, '2025-04-20', 25), -- Cosecha
  (5, 1, 102, 1003, 3,  1, 1, '2025-01-18', 30), -- Siembra
  (6, 1, 102, 1003, 3,  2, 1, '2025-04-10', 28), -- Cosecha
  (7, 1, 102, 1004, 1,  1, 1, '2025-01-22', 18), -- Siembra
  (8, 1, 102, 1004, 1,  2, 1, '2025-04-25', 17); -- Cosecha

-- Proyecto 2
INSERT INTO workorders (id, project_id, field_id, lot_id, crop_id, labor_id, investor_id, date, effective_area) VALUES
  (9,  2, 201, 2001, 3, 11, 2, '2025-01-15', 50), -- Siembra
  (10, 2, 201, 2001, 3, 12, 2, '2025-04-18', 45), -- Cosecha
  (11, 2, 201, 2002, 2, 11, 2, '2025-01-28', 25), -- Siembra
  (12, 2, 201, 2002, 2, 12, 2, '2025-05-03', 20); -- Cosecha

-- Proyecto 3
INSERT INTO workorders (id, project_id, field_id, lot_id, crop_id, labor_id, investor_id, date, effective_area) VALUES
  (13, 3, 301, 3001, 2, 21, 1, '2025-01-09', 40), -- Siembra
  (14, 3, 301, 3001, 2, 22, 1, '2025-04-11', 38), -- Cosecha
  (15, 3, 302, 3002, 3, 21, 1, '2025-01-16', 50), -- Siembra
  (16, 3, 302, 3002, 3, 22, 1, '2025-04-22', 47); -- Cosecha

-- Proyecto 4
INSERT INTO workorders (id, project_id, field_id, lot_id, crop_id, labor_id, investor_id, date, effective_area) VALUES
  (17, 4, 401, 4001, 1, 31, 2, '2025-01-20', 25), -- Siembra
  (18, 4, 401, 4001, 1, 32, 2, '2025-04-29', 22), -- Cosecha
  (19, 4, 402, 4002, 2, 31, 2, '2025-01-25', 20), -- Siembra
  (20, 4, 402, 4002, 2, 32, 2, '2025-05-06', 18); -- Cosecha

-- Proyecto 5
INSERT INTO workorders (id, project_id, field_id, lot_id, crop_id, labor_id, investor_id, date, effective_area) VALUES
  (21, 5, 501, 5001, 3, 41, 3, '2025-01-30', 60), -- Siembra
  (22, 5, 501, 5001, 3, 42, 3, '2025-05-15', 55); -- Cosecha

-- ========================================
-- TEST
-- ========================================
SELECT * FROM lot_metrics_view
ORDER BY project_id, field_id, current_crop_id;
