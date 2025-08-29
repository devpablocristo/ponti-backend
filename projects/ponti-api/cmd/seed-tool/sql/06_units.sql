-- Crear unidades de medida para el sistema
-- Las unidades se crean para estandarizar medidas en el sistema
-- Incluye litros, kilogramos y hectáreas para diferentes tipos de insumos

-- INSERT INTO units (name, created_by, updated_by, created_at, updated_at)
-- VALUES 
--   ('Lts', 2, 2, NOW(), NOW()),
--   ('Kg', 2, 2, NOW(), NOW()),
--   ('Ha', 2, 2, NOW(), NOW())
-- ON CONFLICT (name) DO NOTHING;

-- Verificar inserción
-- SELECT '✅ Unidades creadas:' as status, COUNT(*) as total FROM units; 