-- =======================
-- CUSTOMERS PARA LIST LOT
-- =======================
-- Clientes para el proyecto

INSERT INTO customers (id, name, created_at, updated_at) VALUES
(4, 'Inmobiliaria Buenos Aires S.A.', NOW(), NOW());

-- Verificar inserción
SELECT '✅ Customers creados' as status, COUNT(*) as total FROM customers;
