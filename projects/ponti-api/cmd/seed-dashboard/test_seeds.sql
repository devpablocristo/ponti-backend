-- =======================
-- SCRIPT DE PRUEBA PARA VERIFICAR SEEDS
-- =======================
-- Ejecutar después de cargar los seeds para verificar que funcionaron

-- Verificar datos base
SELECT '🔍 VERIFICANDO DATOS BASE:' as section;
SELECT 'Types:' as tabla, COUNT(*) as total FROM types;
SELECT 'Categories:' as tabla, COUNT(*) as total FROM categories;
SELECT 'Labor types:' as tabla, COUNT(*) as total FROM labor_types;
SELECT 'Labor categories:' as tabla, COUNT(*) as total FROM labor_categories;
SELECT 'Providers:' as tabla, COUNT(*) as total FROM providers;
SELECT 'Users:' as tabla, COUNT(*) as total FROM users;

-- Verificar datos del dashboard
SELECT '🔍 VERIFICANDO DATOS DEL DASHBOARD:' as section;
SELECT 'Supplies:' as tabla, COUNT(*) as total FROM supplies;
SELECT 'Labors:' as tabla, COUNT(*) as total FROM labors;
SELECT 'Workorders:' as tabla, COUNT(*) as total FROM workorders;
SELECT 'Investors:' as tabla, COUNT(*) as total FROM investors;
SELECT 'Project investors:' as tabla, COUNT(*) as total FROM project_investors;
SELECT 'Stocks:' as tabla, COUNT(*) as total FROM stocks;
SELECT 'Invoices:' as tabla, COUNT(*) as total FROM invoices;

-- Verificar métricas específicas
SELECT '🔍 VERIFICANDO MÉTRICAS:' as section;
SELECT 'Lotes sembrados:' as metrica, COUNT(*) as total FROM lots WHERE sowing_date IS NOT NULL;
SELECT 'Total hectáreas sembradas:' as metrica, SUM(hectares) as total FROM lots WHERE sowing_date IS NOT NULL;
SELECT 'Total toneladas cosechadas:' as metrica, SUM(tons) as total FROM lots WHERE tons IS NOT NULL AND tons > 0;
SELECT 'Stock total valorado:' as metrica, SUM(total_cost) as total FROM stocks;
SELECT 'Facturas pagadas:' as metrica, SUM(amount) as total FROM invoices WHERE status = 'paid';
SELECT 'Facturas pendientes:' as metrica, SUM(amount) as total FROM invoices WHERE status = 'pending';

-- Verificar workorders por mes
SELECT '🔍 VERIFICANDO WORKORDERS POR MES:' as section;
SELECT 
    EXTRACT(MONTH FROM date) as mes,
    COUNT(*) as total_workorders,
    SUM(effective_area) as total_hectareas
FROM workorders 
GROUP BY EXTRACT(MONTH FROM date) 
ORDER BY mes;

-- Verificar cultivos por lote
SELECT '🔍 VERIFICANDO CULTIVOS POR LOTE:' as section;
SELECT 
    l.name as lote,
    c.name as cultivo,
    l.hectares,
    l.sowing_date,
    l.tons
FROM lots l
JOIN crops c ON l.current_crop_id = c.id
ORDER BY l.id;

-- Verificar inversores y porcentajes
SELECT '🔍 VERIFICANDO INVERSORES:' as section;
SELECT 
    i.name as inversor,
    pi.percentage as porcentaje,
    (p.budget_cost * pi.percentage / 100) as monto_invertido
FROM project_investors pi
JOIN investors i ON pi.investor_id = i.id
JOIN projects p ON pi.project_id = p.id
ORDER BY pi.percentage DESC;

-- Verificar que el dashboard funcione
SELECT '🔍 VERIFICANDO FUNCIÓN DEL DASHBOARD:' as section;
SELECT 'Función get_dashboard_payload existe:' as status, 
       CASE WHEN EXISTS (SELECT 1 FROM pg_proc WHERE proname = 'get_dashboard_payload') 
            THEN '✅ SÍ' ELSE '❌ NO' END as resultado;

-- Resumen final
SELECT '🎯 RESUMEN FINAL:' as section;
SELECT 'Si todo está correcto, el dashboard debería mostrar:' as info;
SELECT '   - Siembra: 46.0 ha (100%)' as metrica;
SELECT '   - Cosecha: 19.0 toneladas' as metrica;
SELECT '   - Costos: ~$1,200 / $25,000 (4.8%)' as metrica;
SELECT '   - Ingresos: $80,000' as metrica;
SELECT '   - Stock: ~$700,000' as metrica;
SELECT '   - Contribuciones: 100% (5 inversores)' as metrica;
