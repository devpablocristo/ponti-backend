-- Crear tipos de labor para el sistema
-- Los tipos se crean para categorizar diferentes labores agrícolas
-- Incluye labores, fertilizantes, semillas, agroquímicos y servicios especializados

INSERT INTO labor_types (name, created_by, updated_by, created_at, updated_at)
VALUES 
  ('Labores', 2, 2, NOW(), NOW()),
  ('Fertilizantes', 2, 2, NOW(), NOW()),
  ('Semilla', 2, 2, NOW(), NOW()),
  ('Herbicidas', 2, 2, NOW(), NOW()),
  ('Fungicidas', 2, 2, NOW(), NOW()),
  ('Insecticidas', 2, 2, NOW(), NOW()),
  ('Maquinaria', 2, 2, NOW(), NOW()),
  ('Mano de Obra', 2, 2, NOW(), NOW()),
  ('Transporte', 2, 2, NOW(), NOW()),
  ('Otros', 2, 2, NOW(), NOW())
ON CONFLICT (name) DO NOTHING;

-- Verificar inserción
SELECT '✅ Tipos de labor creados:' as status, COUNT(*) as total FROM labor_types; 