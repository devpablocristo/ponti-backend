-- 0) Dummy user for audit FKs (self-referential)
INSERT INTO users (
  id, email, username, password, token_hash, refresh_tokens,
  id_rol, is_verified, active, created_by, updated_by, created_at, updated_at
) VALUES (
  1, 'seed@example.com', 'seeduser', 'passwordhash', 'tokenhash',
  ARRAY[]::TEXT[], 1, TRUE, TRUE, 1, 1, now(), now()
)
ON CONFLICT (id) DO NOTHING;

-- 1) Customers
INSERT INTO customers (id, name, created_by, updated_by) VALUES
  (1, 'DemoCustomer', 1, 1)
ON CONFLICT (id) DO NOTHING;

-- 2) Campaigns
INSERT INTO campaigns (id, name, created_by, updated_by) VALUES
  (1, 'DemoCampaign', 1, 1)
ON CONFLICT (id) DO NOTHING;

-- 3) Investors
INSERT INTO investors (
  id, name, created_by, updated_by, created_at, updated_at
) VALUES
  (1, 'DemoInvestor',    1, 1, now(), now()),
  (10,'DemoInvestor-10', 1, 1, now(), now())
ON CONFLICT (id) DO NOTHING;

-- 4) Crops
INSERT INTO crops (id, name, created_by, updated_by) VALUES
  (1, 'DemoCrop',   1, 1),
  (4, 'DemoCrop-4', 1, 1)
ON CONFLICT (id) DO NOTHING;

-- 5) Lease types
INSERT INTO lease_types (id, name, created_by, updated_by) VALUES
  (1, 'DemoLease', 1, 1)
ON CONFLICT (id) DO NOTHING;

-- 6) Projects (add both id=1 and id=2)
INSERT INTO projects (
  id, name, customer_id, campaign_id, admin_cost, created_by, updated_by
) VALUES
  (1, 'DemoProject-1', 1, 1, 1000.00, 1, 1),
  (2, 'DemoProject-2', 1, 1, 1000.00, 1, 1)
ON CONFLICT (id) DO NOTHING;

-- 7) Fields
INSERT INTO fields (
  id, name, project_id, lease_type_id, lease_type_percent, lease_type_value, created_by, updated_by
) VALUES
  (1, 'DemoField-1', 1, 1, 10.0, 100.0, 1, 1),
  (2, 'DemoField-2', 1, 1, 20.0, 200.0, 1, 1)
ON CONFLICT (id) DO NOTHING;

-- 8) Lots
INSERT INTO lots (
  id, name, field_id, hectares, previous_crop_id, current_crop_id, season, created_by, updated_by
) VALUES
  (1, 'DemoLot-1', 1, 50.0, 1, 1, '2025', 1, 1),
  (3, 'DemoLot-3', 2, 75.0, 1, 4, '2025', 1, 1)
ON CONFLICT (id) DO NOTHING;

-- 9) Labor types & categories
INSERT INTO labor_types (id, name, created_by, updated_by) VALUES
  (1, 'DemoLaborType', 1, 1)
ON CONFLICT (id) DO NOTHING;
INSERT INTO labor_categories (id, name, type_id, created_by, updated_by) VALUES
  (1, 'DemoLaborCat', 1, 1, 1)
ON CONFLICT (id) DO NOTHING;

-- 10) Labors
INSERT INTO labors (
  id, project_id, name, category_id, price, contractor_name, created_by, updated_by
) VALUES
  (1, 1, 'DemoLabor-1', 1, 500.00, 'DemoContractor', 1, 1),
  (5, 1, 'DemoLabor-5', 1, 550.00, 'Proveedor X',    1, 1)
ON CONFLICT (id) DO NOTHING;

-- 11) Types & Categories for supplies
INSERT INTO types (id, name, created_by, updated_by) VALUES
  (1, 'DemoType', 1, 1)
ON CONFLICT (id) DO NOTHING;
INSERT INTO categories (id, name, type_id, created_by, updated_by) VALUES
  (1, 'DemoCategory', 1, 1, 1)
ON CONFLICT (id) DO NOTHING;

-- 12) Supplies
INSERT INTO supplies (
  id, project_id, name, price, unit_id, category_id, type_id, created_by, updated_by
) VALUES
  (1, 1, 'DemoSupply', 20.00, NULL, 1, 1, 1, 1)
ON CONFLICT (id) DO NOTHING;

-- 13) Workorder headers (IDs 1..12), variando project_id y field_id
INSERT INTO workorders (
  id, number,   project_id, field_id, lot_id,
  crop_id, labor_id, contractor, observations,
  date, investor_id, effective_area,
  created_by, updated_by, created_at, updated_at
) VALUES
  -- 3 con project=1 & field=1
  ( 1, 'WO-001', 1, 1, 1, 1, 1, 'C1-F1', 'Obs1', '2025-08-01', 1, 10.0, 1, 1, now(), now()),
  ( 2, 'WO-002', 1, 1, 1, 1, 1, 'C1-F1', 'Obs2', '2025-08-02', 1, 10.0, 1, 1, now(), now()),
  ( 3, 'WO-003', 1, 1, 1, 1, 1, 'C1-F1', 'Obs3', '2025-08-03', 1, 10.0, 1, 1, now(), now()),

  -- 3 con project=1 & field!=1  ⇒ project=1 total=6
  ( 4, 'WO-004', 1, 2, 1, 1, 1, 'C1-F2', 'Obs4', '2025-08-04', 1, 10.0, 1, 1, now(), now()),
  ( 5, 'WO-005', 1, 2, 1, 1, 1, 'C1-F2', 'Obs5', '2025-08-05', 1, 10.0, 1, 1, now(), now()),
  ( 6, 'WO-006', 1, 2, 1, 1, 1, 'C1-F2', 'Obs6', '2025-08-06', 1, 10.0, 1, 1, now(), now()),

  -- 3 con project!=1 & field=1  ⇒ field=1 total=6
  ( 7, 'WO-007', 2, 1, 1, 1, 1, 'C2-F1', 'Obs7', '2025-08-07', 1, 10.0, 1, 1, now(), now()),
  ( 8, 'WO-008', 2, 1, 1, 1, 1, 'C2-F1', 'Obs8', '2025-08-08', 1, 10.0, 1, 1, now(), now()),
  ( 9, 'WO-009', 2, 1, 1, 1, 1, 'C2-F1', 'Obs9', '2025-08-09', 1, 10.0, 1, 1, now(), now()),

  -- 3 con project!=1 & field!=1  ⇒ resto sin restricciones
  (10, 'WO-010', 2, 2, 1, 1, 1, 'C2-F2', 'Obs10','2025-08-10', 1, 10.0, 1, 1, now(), now()),
  (11, 'WO-011', 2, 2, 1, 1, 1, 'C2-F2', 'Obs11','2025-08-11', 1, 10.0, 1, 1, now(), now()),
  (12, 'WO-012', 2, 2, 1, 1, 1, 'C2-F2', 'Obs12','2025-08-12', 1, 10.0, 1, 1, now(), now())
ON CONFLICT (id) DO NOTHING;

-- 14) Workorder items (uno por workorder, IDs 1..12)
INSERT INTO workorder_items (
  id, workorder_id, supply_id, total_used, final_dose, created_at, updated_at
) VALUES
  ( 1,  1, 1,  10.0, 0.2, now(), now()),
  ( 2,  2, 1,  20.0, 0.4, now(), now()),
  ( 3,  3, 1,  30.0, 0.6, now(), now()),
  ( 4,  4, 1,  40.0, 0.8, now(), now()),
  ( 5,  5, 1,  50.0, 1.0, now(), now()),
  ( 6,  6, 1,  60.0, 1.2, now(), now()),
  ( 7,  7, 1,  70.0, 1.4, now(), now()),
  ( 8,  8, 1,  80.0, 1.6, now(), now()),
  ( 9,  9, 1,  90.0, 1.8, now(), now()),
  (10, 10, 1, 100.0, 2.0, now(), now()),
  (11, 11, 1, 110.0, 2.2, now(), now()),
  (12, 12, 1, 120.0, 2.4, now(), now())
ON CONFLICT (id) DO NOTHING;
