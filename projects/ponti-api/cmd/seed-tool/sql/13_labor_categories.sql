-- Crear categorías de labor para el sistema
-- Las categorías se crean asociando tipos de labor existentes
-- Incluye diferentes tipos de labores agrícolas y servicios especializados

INSERT INTO labor_categories (name, type_id, created_by, updated_by, created_at, updated_at)
SELECT 
  'Siembra', t.id, 2, 2, NOW(), NOW()
FROM labor_types t 
WHERE t.name = 'Labores';

INSERT INTO labor_categories (name, type_id, created_by, updated_by, created_at, updated_at)
SELECT 
  'Cosecha', t.id, 2, 2, NOW(), NOW()
FROM labor_types t 
WHERE t.name = 'Labores';

INSERT INTO labor_categories (name, type_id, created_by, updated_by, created_at, updated_at)
SELECT 
  'Fertilización', t.id, 2, 2, NOW(), NOW()
FROM labor_types t 
WHERE t.name = 'Fertilizantes';

INSERT INTO labor_categories (name, type_id, created_by, updated_by, created_at, updated_at)
SELECT 
  'Aplicación de Herbicidas', t.id, 2, 2, NOW(), NOW()
FROM labor_types t 
WHERE t.name = 'Herbicidas';

INSERT INTO labor_categories (name, type_id, created_by, updated_by, created_at, updated_at)
SELECT 
  'Aplicación de Fungicidas', t.id, 2, 2, NOW(), NOW()
FROM labor_types t 
WHERE t.name = 'Fungicidas';

INSERT INTO labor_categories (name, type_id, created_by, updated_by, created_at, updated_at)
SELECT 
  'Aplicación de Insecticidas', t.id, 2, 2, NOW(), NOW()
FROM labor_types t 
WHERE t.name = 'Insecticidas';

INSERT INTO labor_categories (name, type_id, created_by, updated_by, created_at, updated_at)
SELECT 
  'Alquiler de Maquinaria', t.id, 2, 2, NOW(), NOW()
FROM labor_types t 
WHERE t.name = 'Maquinaria';

INSERT INTO labor_categories (name, type_id, created_by, updated_by, created_at, updated_at)
SELECT 
  'Mano de Obra Especializada', t.id, 2, 2, NOW(), NOW()
FROM labor_types t 
WHERE t.name = 'Mano de Obra';

INSERT INTO labor_categories (name, type_id, created_by, updated_by, created_at, updated_at)
SELECT 
  'Transporte de Cosecha', t.id, 2, 2, NOW(), NOW()
FROM labor_types t 
WHERE t.name = 'Transporte';

INSERT INTO labor_categories (name, type_id, created_by, updated_by, created_at, updated_at)
SELECT 
  'Otros Servicios', t.id, 2, 2, NOW(), NOW()
FROM labor_types t 
WHERE t.name = 'Otros';

-- Verificar inserción
SELECT '✅ Categorías de labor creadas:' as status, COUNT(*) as total FROM labor_categories; 