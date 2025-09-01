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
TRUNCATE investors RESTART IDENTITY CASCADE;

-- ========================================
-- UN SOLO PROYECTO SIMPLE
-- ========================================
INSERT INTO customers (id, name) VALUES (1, 'Cliente Demo');

INSERT INTO projects (id, customer_id, campaign_id, name, admin_cost) VALUES
  (1, 1, 1, 'Proyecto Demo', 1000);

-- NOTA: Los lease_types ya existen por la migración 000002
-- ID 1 = % INGRESO NETO, ID 2 = % UTILIDAD, ID 3 = ARRIENDO FIJO, ID 4 = ARRIENDO FIJO + % INGRESO NETO

-- NOTA: Los cultivos ya existen por la migración 000002
-- ID 1 = Soja, ID 2 = Maíz, ID 3 = Trigo, etc.

INSERT INTO investors (id, name) VALUES (1, 'Inversor Demo');

-- ========================================
-- UN CAMPO (usando lease_type existente)
-- ========================================
INSERT INTO fields (id, project_id, name, lease_type_id) VALUES
  (101, 1, 'Campo Principal', 1);  -- ID 1 = % INGRESO NETO

-- ========================================
-- DOS LOTES (usando cultivos existentes)
-- ========================================
INSERT INTO lots (id, field_id, name, hectares, previous_crop_id, current_crop_id, season, tons) VALUES
  (1001, 101, 'Lote A', 100, 1, 2, '2024-2025', 200),  -- Anterior: Soja, Actual: Maíz
  (1002, 101, 'Lote B', 150, 2, 1, '2024-2025', 300);   -- Anterior: Maíz, Actual: Soja

-- ========================================
-- TRES LABORS (Siembra, Cosecha y Fertilización)
-- ========================================
INSERT INTO labors (id, project_id, name, category_id, price, contractor_name) VALUES
  (1, 1, 'Siembra', 9, 100, 'Contratista Demo'),
  (2, 1, 'Cosecha', 13, 200, 'Contratista Demo'),
  (3, 1, 'Fertilización', 8, 150, 'Contratista Demo');  -- ID 8 = Fertilizantes

-- ========================================
-- CINCO INSUMOS (con precios redondos)
-- ========================================
INSERT INTO supplies (id, project_id, name, type_id, category_id, price) VALUES
  (1, 1, 'Fertilizante NPK', 3, 8, 100),  -- ID 3 = Fertilizantes, ID 8 = Fertilizantes, precio por kilo
  (2, 1, 'Semilla de Soja', 1, 1, 200),  -- ID 1 = Semillas, ID 1 = Semilla, precio por kilo
  (3, 1, 'Herbicida Glifosato', 2, 4, 150),  -- ID 2 = Agroquímicos, ID 4 = Herbicidas, precio por litro
  (4, 1, 'Fungicida Triazol', 2, 6, 250),  -- ID 2 = Agroquímicos, ID 6 = Fungicidas, precio por litro
  (5, 1, 'Insecticida Piretroide', 2, 5, 300);  -- ID 2 = Agroquímicos, ID 5 = Insecticidas, precio por litro

-- ========================================
-- OCHO WORKORDERS (con áreas variadas pero redondas)
-- ========================================
INSERT INTO workorders (id, project_id, field_id, lot_id, crop_id, labor_id, investor_id, date, effective_area) VALUES
  (1, 1, 101, 1001, 2, 1, 1, '2024-10-15', 100),  -- Siembra Lote A (Maíz) - 100 ha
  (2, 1, 101, 1001, 2, 2, 1, '2025-03-20', 100),  -- Cosecha Lote A (Maíz) - 100 ha
  (3, 1, 101, 1002, 1, 1, 1, '2024-10-18', 150),  -- Siembra Lote B (Soja) - 150 ha
  (4, 1, 101, 1002, 1, 2, 1, '2025-03-25', 150),  -- Cosecha Lote B (Soja) - 150 ha
  (5, 1, 101, 1001, 2, 3, 1, '2024-11-10', 100),  -- Fertilización Lote A (Maíz) - 100 ha
  (6, 1, 101, 1002, 1, 3, 1, '2024-11-15', 150),  -- Fertilización Lote B (Soja) - 150 ha
  (7, 1, 101, 1001, 2, 3, 1, '2024-12-01', 200),  -- Fertilización Lote A - 200 ha
  (8, 1, 101, 1002, 1, 3, 1, '2024-12-05', 200);  -- Fertilización Lote B - 200 ha

-- ========================================
-- WORKORDER_ITEMS (con dosis redondas y fáciles de calcular)
-- ========================================
INSERT INTO workorder_items (workorder_id, supply_id, final_dose, total_used) VALUES
  (5, 1, 100, 10000),  -- Workorder 5: 100 kg/ha × 100 ha = 10,000 kg total
  (6, 2, 200, 30000),  -- Workorder 6: 200 kg/ha × 150 ha = 30,000 kg total
  (7, 3, 150, 30000),  -- Workorder 7: 150 L/ha × 200 ha = 30,000 L total
  (8, 4, 250, 50000);  -- Workorder 8: 250 L/ha × 200 ha = 50,000 L total

-- ========================================
-- VERIFICAR
-- ========================================
/* SELECT '=== RESUMEN ===' as info;
SELECT 'Customers' as tabla, COUNT(*) as total FROM customers
UNION ALL
SELECT 'Projects', COUNT(*) FROM projects
UNION ALL
SELECT 'Fields', COUNT(*) FROM fields
UNION ALL
SELECT 'Lots', COUNT(*) FROM lots
UNION ALL
SELECT 'Labors', COUNT(*) FROM labors
UNION ALL
SELECT 'Supplies', COUNT(*) FROM supplies
UNION ALL
SELECT 'Workorders', COUNT(*) FROM workorders
UNION ALL
SELECT 'Workorder Items', COUNT(*) FROM workorder_items;

SELECT '=== PROYECTO ===' as info;
SELECT p.name, c.name as customer, p.admin_cost
FROM projects p
JOIN customers c ON p.customer_id = c.id;

SELECT '=== LOTES CON CULTIVOS ===' as info;
SELECT l.name as lote, l.hectares, 
       c1.name as cultivo_anterior, c2.name as cultivo_actual, l.tons
FROM lots l
JOIN crops c1 ON l.previous_crop_id = c1.id
JOIN crops c2 ON l.current_crop_id = c2.id
ORDER BY l.id;

SELECT '=== CAMPO CON TIPO DE ARRENDAMIENTO ===' as info;
SELECT f.name as campo, lt.name as tipo_arrendamiento
FROM fields f
JOIN lease_types lt ON f.lease_type_id = lt.id; */

/* SELECT * FROM workorders;

SELECT * FROM labors;

SELECT * FROM supplies; */

/* SELECT '=== WORKORDER CON INSUMO ===' as info;
SELECT w.id as workorder_id, w.date, l.name as labor, s.name as supply, 
       wi.final_dose as dosis_kg_ha, wi.total_used as total_kg, w.effective_area as hectareas
FROM workorders w
JOIN labors l ON w.labor_id = l.id
JOIN workorder_items wi ON w.id = wi.workorder_id
JOIN supplies s ON wi.supply_id = s.id
WHERE w.id = 5; */

SELECT * FROM  workorder_metrics_view
ORDER BY project_id, field_id, customer_id, campaign_id;

/* ### **1. 📏 Superficie Ejecutada (`surface_ha`):**
```
Suma todas las áreas de todas las OTs del campo
= 100 + 100 + 150 + 150 + 100 + 150 + 200 + 200 = 1,150 ha
```

### **2. �� Litros (`liters`):**
```
Suma solo las dosis de insumos líquidos (Herbicidas, Fungicidas, Insecticidas, Coadyuvantes)
= 150 + 250 = 400 L
```

### **3. ⚖️ Kilogramos (`kilograms`):**
```
Suma solo las dosis de insumos sólidos (Fertilizantes, Semillas, Curasemillas)
= 100 + 200 = 300 kg
```

### **4. 💰 Costo Directo (`direct_cost`):**
```
Suma los precios base de labor + precios base de insumos
= (100 + 200 + 150) + (100 + 200 + 150 + 250 + 300) = 450 + 1,000 = $1,450
``` */
