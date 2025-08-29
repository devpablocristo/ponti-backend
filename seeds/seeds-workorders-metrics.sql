ROLLBACK;

-- =========================================================
-- CLEAN
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

-- =========================================================
-- BASE / PARENTS
-- =========================================================
INSERT INTO customers (id, name, deleted_at) VALUES (1, 'Cliente Demo', NULL);
INSERT INTO campaigns (id, name, deleted_at) VALUES (1, 'Campaña Demo', NULL);

INSERT INTO projects (id, customer_id, campaign_id, name, admin_cost, deleted_at) VALUES
  (1, 1, 1, 'Proyecto 1', 0, NULL),
  (2, 1, 1, 'Proyecto 2', 0, NULL),
  (3, 1, 1, 'Proyecto 3', 0, NULL);

INSERT INTO lease_types (id, name, deleted_at) VALUES (1, 'Arrendamiento Demo', NULL);

INSERT INTO crops (id, name, deleted_at) VALUES
  (1, 'Trigo', NULL),
  (2, 'Maíz', NULL),
  (3, 'Soja', NULL);

INSERT INTO investors (id, name, deleted_at) VALUES
  (1, 'Inversor A', NULL),
  (2, 'Inversor B', NULL),
  (3, 'Inversor C', NULL);

INSERT INTO fields (id, project_id, name, lease_type_id, deleted_at) VALUES
  (101, 1, 'Field 101', 1, NULL),
  (102, 1, 'Field 102', 1, NULL),
  (201, 2, 'Field 201', 1, NULL),
  (301, 3, 'Field 301', 1, NULL),
  (302, 3, 'Field 302', 1, NULL);

INSERT INTO lots (id, field_id, name, hectares, previous_crop_id, current_crop_id, season, deleted_at) VALUES
  (1001, 101, 'Lot 1001', 50, 1, 2, '2025', NULL),
  (1002, 102, 'Lot 1002', 20, 2, 3, '2025', NULL),
  (2001, 201, 'Lot 2001', 40, 3, 1, '2025', NULL),
  (3001, 301, 'Lot 3001', 15, 1, 2, '2025', NULL),
  (3002, 302, 'Lot 3002', 60, 2, 3, '2025', NULL);

-- =========================================================
-- TYPES & CATEGORIES/home/pablo/Projects/Pablo/github.com/alphacodinggroup/ponti-backend/projects/ponti-api/migrations/000046_fix_workorder_metrics.up.sql
--   type_id: 1=Labores, 2=Insumos
--   Para diferenciar litros vs kilos:
--     - Litros: Herbicida(2), Fungicida(5)
--     - Kilogramos: Fertilizante(3), Semilla(4)
-- =========================================================
INSERT INTO types (id, name, deleted_at) VALUES
  (1, 'Labores', NULL),
  (2, 'Insumos', NULL);

INSERT INTO categories (id, name, type_id, deleted_at) VALUES
  (1, 'Labores Generales', 1, NULL),
  (2, 'Herbicida',         2, NULL), -- L
  (3, 'Fertilizante',      2, NULL), -- Kg
  (4, 'Semilla',           2, NULL), -- Kg
  (5, 'Fungicida',         2, NULL); -- L

-- =========================================================
-- LABORS (precio por Ha)
-- =========================================================
INSERT INTO labors (id, project_id, name, category_id, price, contractor_name, deleted_at) VALUES
  (1, 1, 'Siembra',       1, 100, 'Contratista A', NULL),
  (2, 1, 'Fertilización', 1, 150, 'Contratista A', NULL),
  (3, 1, 'Cosecha',       1, 200, 'Contratista A', NULL),
  (11, 2, 'Siembra',       1, 105, 'Contratista B', NULL),
  (12, 2, 'Fertilización', 1, 155, 'Contratista B', NULL),
  (13, 2, 'Cosecha',       1, 205, 'Contratista B', NULL),
  (21, 3, 'Siembra',       1, 110, 'Contratista C', NULL),
  (22, 3, 'Fertilización', 1, 160, 'Contratista C', NULL),
  (23, 3, 'Cosecha',       1, 210, 'Contratista C', NULL);

-- =========================================================
-- SUPPLIES por proyecto (precio por unidad)
-- =========================================================
-- P1
INSERT INTO supplies (id, project_id, name, type_id, category_id, price, deleted_at) VALUES
  (11, 1, 'Glifosato', 2, 2, 8.00,  NULL), -- L
  (12, 1, 'Urea',      2, 3, 0.50,  NULL), -- Kg
  (13, 1, 'Semilla',   2, 4, 2.00,  NULL), -- Kg
  (14, 1, 'Fungicida', 2, 5, 6.00,  NULL); -- L
-- P2
INSERT INTO supplies (id, project_id, name, type_id, category_id, price, deleted_at) VALUES
  (21, 2, 'Glifosato', 2, 2, 8.20,  NULL), -- L
  (22, 2, 'Urea',      2, 3, 0.55,  NULL), -- Kg
  (23, 2, 'Semilla',   2, 4, 2.10,  NULL), -- Kg
  (24, 2, 'Fungicida', 2, 5, 6.10,  NULL); -- L
-- P3
INSERT INTO supplies (id, project_id, name, type_id, category_id, price, deleted_at) VALUES
  (31, 3, 'Glifosato', 2, 2, 7.90,  NULL), -- L
  (32, 3, 'Urea',      2, 3, 0.48,  NULL), -- Kg
  (33, 3, 'Semilla',   2, 4, 1.95,  NULL), -- Kg
  (34, 3, 'Fungicida', 2, 5, 5.80,  NULL); -- L

-- =========================================================
-- WORKORDERS (área efectiva por WO)
-- =========================================================
INSERT INTO workorders (id, project_id, field_id, lot_id, crop_id, labor_id, investor_id, date, effective_area, deleted_at) VALUES
  (1, 1, 101, 1001, 1,  1, 1, '2025-01-10', 50, NULL),
  (2, 1, 101, 1001, 2,  2, 1, '2025-02-15', 30, NULL),
  (3, 1, 102, 1002, 3,  3, 1, '2025-03-20', 20, NULL),
  (4, 2, 201, 2001, 1, 11, 2, '2025-01-12', 40, NULL),
  (5, 2, 201, 2001, 2, 12, 2, '2025-02-18', 25, NULL),
  (6, 3, 301, 3001, 3, 23, 3, '2025-03-05', 15, NULL),
  (7, 3, 302, 3002, 1, 21, 3, '2025-04-01', 60, NULL);

-- =========================================================
-- WORKORDER ITEMS (final_dose por Ha) + total_used
--   L -> Herbicida/Fungicida ; Kg -> Fertilizante/Semilla
-- =========================================================
-- P1 / Field 101
INSERT INTO workorder_items (workorder_id, supply_id, final_dose, total_used, deleted_at) VALUES
  (1, 11, 2.0, 100.0, NULL),  -- Glifosato L: 2 x 50 = 100 L
  (1, 12, 1.5,  75.0, NULL),  -- Urea Kg: 1.5 x 50 = 75 Kg
  (2, 14, 1.0,  30.0, NULL);  -- Fungicida L: 1 x 30 = 30 L
-- P1 / Field 102
INSERT INTO workorder_items (workorder_id, supply_id, final_dose, total_used, deleted_at) VALUES
  (3, 13, 3.0, 60.0, NULL);   -- Semilla Kg: 3 x 20 = 60 Kg
-- P2 / Field 201
INSERT INTO workorder_items (workorder_id, supply_id, final_dose, total_used, deleted_at) VALUES
  (4, 21, 1.2, 48.0, NULL),   -- Glifosato L: 1.2 x 40 = 48 L
  (4, 22, 1.0, 40.0, NULL),   -- Urea Kg: 1 x 40 = 40 Kg
  (5, 23, 2.5, 62.5, NULL);   -- Semilla Kg: 2.5 x 25 = 62.5 Kg
-- P3 / Field 301 y 302
INSERT INTO workorder_items (workorder_id, supply_id, final_dose, total_used, deleted_at) VALUES
  (6, 34, 0.8, 12.0, NULL),   -- Fungicida L: 0.8 x 15 = 12 L
  (7, 31, 1.5, 90.0, NULL),   -- Glifosato L: 1.5 x 60 = 90 L
  (7, 32, 2.0,120.0, NULL);   -- Urea Kg: 2 x 60 = 120 Kg

-- =========================================================
-- VIEW: workorder_metrics_view (ACTUALIZADA para diferenciar L vs Kg)
-- =========================================================
DROP VIEW IF EXISTS workorder_metrics_view;

CREATE VIEW workorder_metrics_view AS
WITH
workorder_base AS (
  SELECT
    w.id              AS workorder_id,
    w.project_id,
    w.field_id,
    w.effective_area,
    COALESCE(lb.price, 0) AS labor_price_per_ha,
    COALESCE(lb.price, 0) * w.effective_area AS labor_cost_per_wo
  FROM workorders w
  JOIN labors lb ON lb.id = w.labor_id
  WHERE w.deleted_at IS NULL
    AND lb.deleted_at IS NULL
    AND w.effective_area IS NOT NULL
    AND w.effective_area > 0
),
supply_aggregation AS (
  SELECT
    w.id         AS workorder_id,
    w.project_id,
    w.field_id,
    -- costo total de insumos
    SUM(COALESCE(wi.final_dose, 0) * COALESCE(s.price, 0) * w.effective_area) AS total_supplies_cost,
    -- Diferenciar litros y kilogramos por categoría
    SUM(
      CASE WHEN s.category_id IN (2,5) -- Herbicida/Fungicida -> L
           THEN COALESCE(wi.final_dose,0) * w.effective_area
           ELSE 0 END
    ) AS total_liters,
    SUM(
      CASE WHEN s.category_id IN (3,4) -- Fertilizante/Semilla -> Kg
           THEN COALESCE(wi.final_dose,0) * w.effective_area
           ELSE 0 END
    ) AS total_kilograms
  FROM workorders w
  JOIN workorder_items wi ON wi.workorder_id = w.id AND wi.deleted_at IS NULL
  JOIN supplies s         ON s.id = wi.supply_id   AND s.deleted_at IS NULL
  WHERE w.deleted_at IS NULL
    AND w.effective_area IS NOT NULL
    AND w.effective_area > 0
  GROUP BY w.id, w.project_id, w.field_id
),
field_metrics AS (
  SELECT
    wb.project_id,
    wb.field_id,
    SUM(wb.effective_area) AS total_surface_ha,
    SUM(wb.labor_cost_per_wo) AS total_labor_cost,
    SUM(COALESCE(sa.total_supplies_cost, 0)) AS total_supplies_cost,
    SUM(wb.labor_cost_per_wo + COALESCE(sa.total_supplies_cost, 0)) AS total_direct_cost,
    SUM(COALESCE(sa.total_liters, 0)) AS total_liters,
    SUM(COALESCE(sa.total_kilograms, 0)) AS total_kilograms,
    COUNT(DISTINCT wb.workorder_id) AS total_workorders
  FROM workorder_base wb
  LEFT JOIN supply_aggregation sa ON sa.workorder_id = wb.workorder_id
  GROUP BY wb.project_id, wb.field_id
)
SELECT
  fm.project_id,
  fm.field_id,
  fm.total_surface_ha AS surface_ha,
  COALESCE(fm.total_liters, 0) AS liters,
  COALESCE(fm.total_kilograms, 0) AS kilograms,
  fm.total_direct_cost AS direct_cost
FROM field_metrics fm
WHERE fm.total_surface_ha > 0;

-- =========================================================
-- TESTS
-- =========================================================
-- Por campo
SELECT * FROM workorder_metrics_view ORDER BY project_id, field_id;

-- Global (deberán diferir liters vs kilograms)
SELECT
  SUM(surface_ha)  AS surface_ha,
  SUM(liters)      AS liters,
  SUM(kilograms)   AS kilograms,
  SUM(direct_cost) AS direct_cost
FROM workorder_metrics_view;
