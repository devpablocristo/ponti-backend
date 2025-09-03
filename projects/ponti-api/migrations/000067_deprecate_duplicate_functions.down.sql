-- ========================================
-- MIGRACIÓN 000067: REVERTIR DEPRECACIÓN DE FUNCIONES DUPLICADAS
-- Entidad: views (Revertir deprecación)
-- Funcionalidad: Revertir la deprecación de funciones duplicadas
-- ========================================

-- ========================================
-- 1. ELIMINAR TABLA DE DOCUMENTACIÓN
-- ========================================
DROP TABLE IF EXISTS migration_priority_documentation;

-- ========================================
-- 2. ELIMINAR COMENTARIOS DE DEPRECACIÓN
-- ========================================
COMMENT ON VIEW labor_cards_cube_view IS NULL;
COMMENT ON VIEW workorder_metrics_view IS NULL;
COMMENT ON VIEW dashboard_operating_result_view IS NULL;
COMMENT ON VIEW dashboard_management_balance_view IS NULL;
COMMENT ON VIEW dashboard_crop_incidence_view IS NULL;
