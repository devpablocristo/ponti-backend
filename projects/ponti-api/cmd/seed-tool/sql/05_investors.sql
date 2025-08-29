-- Crear inversores para el sistema
-- Crea inversores para el sistema

INSERT INTO investors (name, created_at, updated_at, created_by, updated_by)
VALUES 
  ('Inversor Principal', NOW(), NOW(), 2, 2),
  ('Inversor Secundario', NOW(), NOW(), 2, 2),
  ('Inversor Técnico', NOW(), NOW(), 2, 2),
  ('Inversor Comercial', NOW(), NOW(), 2, 2),
  ('Inversor Estratégico', NOW(), NOW(), 2, 2),
  ('Inversor Local', NOW(), NOW(), 2, 2),
  ('Inversor Internacional', NOW(), NOW(), 2, 2),
  ('Inversor Institucional', NOW(), NOW(), 2, 2),
  ('Inversor Privado', NOW(), NOW(), 2, 2),
  ('Inversor Corporativo', NOW(), NOW(), 2, 2)
ON CONFLICT (name) DO NOTHING;

-- Reiniciar secuencia de investors

-- Verificar inserción
SELECT '✅ Inversores creados:' as status, COUNT(*) as total FROM investors; 