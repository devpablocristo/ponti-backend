-- ========================================
-- MIGRACIÓN 000067: DEPRECAR FUNCIONES DUPLICADAS
-- Entidad: views (Deprecar funciones duplicadas)
-- Funcionalidad: Marcar como deprecadas las funciones duplicadas y asegurar que se usen las versiones más recientes
-- ========================================

-- ========================================
-- 1. DEPRECAR VISTAS LABOR_CARDS_CUBE_VIEW DUPLICADAS
-- ========================================
-- Las vistas en migraciones 000045, 000050, 000061 están duplicadas
-- La versión correcta está en migración 000066 como labor_cards_cube_view_v2

-- Comentario: Las vistas labor_cards_cube_view en migraciones 000045, 000050, 000061
-- están deprecadas. Usar labor_cards_cube_view_v2 de la migración 000066.

-- ========================================
-- 2. DEPRECAR VISTAS WORKORDER_METRICS_VIEW DUPLICADAS
-- ========================================
-- Las vistas en migraciones 000042, 000046, 000049 están duplicadas
-- La versión correcta está en migración 000066 como workorder_metrics_view_v2

-- Comentario: Las vistas workorder_metrics_view en migraciones 000042, 000046, 000049
-- están deprecadas. Usar workorder_metrics_view_v2 de la migración 000066.

-- ========================================
-- 3. DEPRECAR VISTAS DASHBOARD DUPLICADAS
-- ========================================
-- Las vistas en migraciones 000055, 000059, 000060 están duplicadas
-- Las versiones correctas están en migraciones 000065 y 000066 como *_v2

-- Comentario: Las vistas dashboard_operating_result_view en migración 000055,
-- dashboard_management_balance_view en migración 000059, y
-- dashboard_crop_incidence_view en migración 000060 están deprecadas.
-- Usar las versiones *_v2 de las migraciones 000065 y 000066.

-- ========================================
-- 4. VERIFICAR QUE LAS VISTAS CORRECTAS EXISTAN
-- ========================================
-- Verificar que las vistas base existan (migración 000064)
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.views WHERE table_name = 'base_direct_costs_view') THEN
        RAISE EXCEPTION 'Vista base_direct_costs_view no existe. Ejecutar migración 000064 primero.';
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM information_schema.views WHERE table_name = 'base_yield_calculations_view') THEN
        RAISE EXCEPTION 'Vista base_yield_calculations_view no existe. Ejecutar migración 000064 primero.';
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM information_schema.views WHERE table_name = 'base_income_net_view') THEN
        RAISE EXCEPTION 'Vista base_income_net_view no existe. Ejecutar migración 000064 primero.';
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM information_schema.views WHERE table_name = 'base_admin_costs_view') THEN
        RAISE EXCEPTION 'Vista base_admin_costs_view no existe. Ejecutar migración 000064 primero.';
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM information_schema.views WHERE table_name = 'base_lease_calculations_view') THEN
        RAISE EXCEPTION 'Vista base_lease_calculations_view no existe. Ejecutar migración 000064 primero.';
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM information_schema.views WHERE table_name = 'base_active_total_view') THEN
        RAISE EXCEPTION 'Vista base_active_total_view no existe. Ejecutar migración 000064 primero.';
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM information_schema.views WHERE table_name = 'base_operating_result_view') THEN
        RAISE EXCEPTION 'Vista base_operating_result_view no existe. Ejecutar migración 000064 primero.';
    END IF;
END $$;

-- Verificar que las vistas corregidas existan (migraciones 000065 y 000066)
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.views WHERE table_name = 'labor_cards_cube_view_v2') THEN
        RAISE EXCEPTION 'Vista labor_cards_cube_view_v2 no existe. Ejecutar migración 000066 primero.';
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM information_schema.views WHERE table_name = 'workorder_metrics_view_v2') THEN
        RAISE EXCEPTION 'Vista workorder_metrics_view_v2 no existe. Ejecutar migración 000066 primero.';
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM information_schema.views WHERE table_name = 'dashboard_operating_result_view_v2') THEN
        RAISE EXCEPTION 'Vista dashboard_operating_result_view_v2 no existe. Ejecutar migraciones 000065 y 000066 primero.';
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM information_schema.views WHERE table_name = 'dashboard_management_balance_view_v2') THEN
        RAISE EXCEPTION 'Vista dashboard_management_balance_view_v2 no existe. Ejecutar migraciones 000065 y 000066 primero.';
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM information_schema.views WHERE table_name = 'dashboard_crop_incidence_view_v2') THEN
        RAISE EXCEPTION 'Vista dashboard_crop_incidence_view_v2 no existe. Ejecutar migraciones 000065 y 000066 primero.';
    END IF;
END $$;

-- ========================================
-- 5. CREAR COMENTARIOS DE DEPRECACIÓN
-- ========================================
-- Agregar comentarios a las vistas deprecadas para documentar su estado (solo si existen)
DO $$
BEGIN
    -- Solo agregar comentarios si las vistas existen
    IF EXISTS (SELECT 1 FROM information_schema.views WHERE table_name = 'labor_cards_cube_view') THEN
        COMMENT ON VIEW labor_cards_cube_view IS 'DEPRECATED: Usar labor_cards_cube_view_v2 de la migración 000066';
    END IF;
    
    IF EXISTS (SELECT 1 FROM information_schema.views WHERE table_name = 'workorder_metrics_view') THEN
        COMMENT ON VIEW workorder_metrics_view IS 'DEPRECATED: Usar workorder_metrics_view_v2 de la migración 000066';
    END IF;
    
    IF EXISTS (SELECT 1 FROM information_schema.views WHERE table_name = 'dashboard_operating_result_view') THEN
        COMMENT ON VIEW dashboard_operating_result_view IS 'DEPRECATED: Usar dashboard_operating_result_view_v2 de las migraciones 000065/000066';
    END IF;
    
    IF EXISTS (SELECT 1 FROM information_schema.views WHERE table_name = 'dashboard_management_balance_view') THEN
        COMMENT ON VIEW dashboard_management_balance_view IS 'DEPRECATED: Usar dashboard_management_balance_view_v2 de las migraciones 000065/000066';
    END IF;
    
    IF EXISTS (SELECT 1 FROM information_schema.views WHERE table_name = 'dashboard_crop_incidence_view') THEN
        COMMENT ON VIEW dashboard_crop_incidence_view IS 'DEPRECATED: Usar dashboard_crop_incidence_view_v2 de las migraciones 000065/000066';
    END IF;
END $$;

-- ========================================
-- 6. DOCUMENTAR JERARQUÍA DE PRIORIDAD
-- ========================================
-- Crear tabla de documentación para la jerarquía de prioridad
CREATE TABLE IF NOT EXISTS migration_priority_documentation (
    id SERIAL PRIMARY KEY,
    view_name VARCHAR(255) NOT NULL,
    deprecated_migrations TEXT[] NOT NULL,
    active_migration VARCHAR(20) NOT NULL,
    active_view_name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insertar documentación de prioridad
INSERT INTO migration_priority_documentation (view_name, deprecated_migrations, active_migration, active_view_name) VALUES
('labor_cards_cube_view', ARRAY['000045', '000050', '000061'], '000066', 'labor_cards_cube_view_v2'),
('workorder_metrics_view', ARRAY['000042', '000046', '000049'], '000066', 'workorder_metrics_view_v2'),
('dashboard_operating_result_view', ARRAY['000055'], '000065/000066', 'dashboard_operating_result_view_v2'),
('dashboard_management_balance_view', ARRAY['000059'], '000065/000066', 'dashboard_management_balance_view_v2'),
('dashboard_crop_incidence_view', ARRAY['000060'], '000065/000066', 'dashboard_crop_incidence_view_v2');

-- Comentario en la tabla de documentación
COMMENT ON TABLE migration_priority_documentation IS 'Documentación de jerarquía de prioridad para vistas deprecadas vs activas';
