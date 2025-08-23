-- Crear managers de prueba para el sistema
-- Los managers se crean para gestionar proyectos agrícolas
-- Incluye diferentes tipos de managers con especializaciones técnicas y comerciales

INSERT INTO managers (name, created_by, updated_by, created_at, updated_at)
VALUES 
  ('Manager Norte', 2, 2, NOW(), NOW()),
  ('Manager Sur', 2, 2, NOW(), NOW()),
  ('Manager Este', 2, 2, NOW(), NOW()),
  ('Manager Oeste', 2, 2, NOW(), NOW()),
  ('Manager Central', 2, 2, NOW(), NOW()),
  ('Manager Especializado', 2, 2, NOW(), NOW()),
  ('Manager Senior', 2, 2, NOW(), NOW()),
  ('Manager Junior', 2, 2, NOW(), NOW()),
  ('Manager Técnico', 2, 2, NOW(), NOW()),
  ('Manager Comercial', 2, 2, NOW(), NOW())
ON CONFLICT (name) DO NOTHING;

-- Verificar inserción
SELECT '✅ Managers creados:' as status, COUNT(*) as total FROM managers; 