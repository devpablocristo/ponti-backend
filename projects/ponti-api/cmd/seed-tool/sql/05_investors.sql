-- Crear inversores para el sistema
-- Crea inversores para el sistema

INSERT INTO investors (name, contributions, contribution_date, percentage, created_at, updated_at, created_by, updated_by)
VALUES 
  ('Inversor Principal', 100000, NOW(), 40, NOW(), NOW(), 2, 2),
  ('Inversor Secundario', 50000, NOW(), 20, NOW(), NOW(), 2, 2),
  ('Inversor Técnico', 25000, NOW(), 10, NOW(), NOW(), 2, 2),
  ('Inversor Comercial', 30000, NOW(), 12, NOW(), NOW(), 2, 2),
  ('Inversor Estratégico', 45000, NOW(), 18, NOW(), NOW(), 2, 2),
  ('Inversor Local', 15000, NOW(), 6, NOW(), NOW(), 2, 2),
  ('Inversor Internacional', 75000, NOW(), 30, NOW(), NOW(), 2, 2),
  ('Inversor Institucional', 120000, NOW(), 48, NOW(), NOW(), 2, 2),
  ('Inversor Privado', 35000, NOW(), 14, NOW(), NOW(), 2, 2),
  ('Inversor Corporativo', 80000, NOW(), 32, NOW(), NOW(), 2, 2)
ON CONFLICT (name) DO NOTHING;

-- Reiniciar secuencia de investors

-- Verificar inserción
SELECT '✅ Inversores creados:' as status, COUNT(*) as total FROM investors; 