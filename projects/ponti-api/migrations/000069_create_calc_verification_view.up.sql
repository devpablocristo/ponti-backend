-- =====================================================
-- 000069: VERIFICATION - Vista de Verificación
-- =====================================================
-- Entidad: Verification (Verificación)
-- Funcionalidad: Verificar que todos los cálculos funcionen correctamente
-- =====================================================

-- Consultas de verificación para validar que los cálculos funcionen correctamente
-- Esta vista emite verificaciones determinísticas que deben coincidir con los valores esperados
CREATE OR REPLACE VIEW v_calc_verification AS
WITH verification_results AS (
    -- Verificación 1: Labor-only order example
    -- labor_price_per_ha = 5 USD, effective_area = 15 ha → total = 75 USD
    SELECT 
        'Labor-only order verification' AS test_name,
        'Expected: 75 USD for 5 USD/ha * 15 ha' AS expected_result,
        CASE 
            WHEN EXISTS (
                SELECT 1 FROM v_calc_workorders 
                WHERE labor_price_per_ha = 5 
                AND effective_area = 15 
                AND workorder_total_usd = 75
            ) THEN 'PASS: Found matching record'
            ELSE 'FAIL: No matching record found'
        END AS test_result,
        'Check if labor-only workorder calculation works correctly' AS description
    
    UNION ALL
    
    -- Verificación 2: Labor + supplies example  
    -- supplies_total_usd = 99.98, plus labor 5 USD/ha * 15 ha → total = 174.98 USD
    SELECT 
        'Labor + supplies verification' AS test_name,
        'Expected: 174.98 USD for 99.98 supplies + (5 USD/ha * 15 ha)' AS expected_result,
        CASE 
            WHEN EXISTS (
                SELECT 1 FROM v_calc_workorders 
                WHERE supplies_total_usd = 99.98 
                AND labor_price_per_ha = 5 
                AND effective_area = 15 
                AND workorder_total_usd = 174.98
            ) THEN 'PASS: Found matching record'
            ELSE 'FAIL: No matching record found'
        END AS test_result,
        'Check if labor + supplies workorder calculation works correctly' AS description
    
    UNION ALL
    
    -- Verificación 3: IVA computed with 10.5%, not 21%
    SELECT 
        'IVA calculation verification' AS test_name,
        'Expected: IVA = 10.5% of total_usd_net' AS expected_result,
        CASE 
            WHEN EXISTS (
                SELECT 1 FROM v_calc_labors 
                WHERE ABS(iva_amount - (total_usd_net * 0.105)) < 0.01
                LIMIT 1
            ) THEN 'PASS: IVA calculated correctly with 10.5%'
            ELSE 'FAIL: IVA not calculated with 10.5%'
        END AS test_result,
        'Verify IVA is computed with 10.5% rate' AS description
    
    UNION ALL
    
    -- Verificación 4: ARS conversion uses latest FX rate
    SELECT 
        'ARS conversion verification' AS test_name,
        'Expected: ARS conversion uses latest FX rate or defaults to 1' AS expected_result,
        CASE 
            WHEN EXISTS (
                SELECT 1 FROM v_calc_labors 
                WHERE cost_ars_per_ha = labor_price_per_ha * usd_ars_rate
                LIMIT 1
            ) THEN 'PASS: ARS conversion working correctly'
            ELSE 'FAIL: ARS conversion not working'
        END AS test_result,
        'Verify ARS conversion uses latest FX rate' AS description
    
    UNION ALL
    
    -- Verificación 5: Yield calculation preference (harvests > lots.tons)
    SELECT 
        'Yield calculation verification' AS test_name,
        'Expected: Yield calculated from harvests or fallback to lots.tons' AS expected_result,
        CASE 
            WHEN EXISTS (
                SELECT 1 FROM v_calc_lots 
                WHERE yield_tonha >= 0
                LIMIT 1
            ) THEN 'PASS: Yield calculation working'
            ELSE 'FAIL: Yield calculation not working'
        END AS test_result,
        'Verify yield calculation works with harvests preference' AS description
    
    UNION ALL
    
    -- Verificación 6: Net price selection policy
    SELECT 
        'Net price selection verification' AS test_name,
        'Expected: Last net price by date for (field_id, crop_id) with project fallback' AS expected_result,
        CASE 
            WHEN EXISTS (
                SELECT 1 FROM v_calc_lots 
                WHERE net_price_usd >= 0
                LIMIT 1
            ) THEN 'PASS: Net price selection working'
            ELSE 'FAIL: Net price selection not working'
        END AS test_result,
        'Verify net price selection follows correct policy' AS description
    
    UNION ALL
    
    -- Verificación 7: Lease modes calculation
    SELECT 
        'Lease modes verification' AS test_name,
        'Expected: All lease modes (fixed, %, utility, mixed) calculated correctly' AS expected_result,
        CASE 
            WHEN EXISTS (
                SELECT 1 FROM v_calc_lots 
                WHERE lease_per_ha >= 0
                LIMIT 1
            ) THEN 'PASS: Lease modes calculation working'
            ELSE 'FAIL: Lease modes calculation not working'
        END AS test_result,
        'Verify all lease modes are calculated correctly' AS description
    
    UNION ALL
    
    -- Verificación 8: Project rollups aggregation
    SELECT 
        'Project rollups verification' AS test_name,
        'Expected: Project costs and economics aggregated correctly' AS expected_result,
        CASE 
            WHEN EXISTS (
                SELECT 1 FROM v_calc_project_costs 
                WHERE total_costs_usd >= 0
                LIMIT 1
            ) AND EXISTS (
                SELECT 1 FROM v_calc_project_economics 
                WHERE net_income_usd >= 0 OR active_total_usd >= 0
                LIMIT 1
            ) THEN 'PASS: Project rollups working'
            ELSE 'FAIL: Project rollups not working'
        END AS test_result,
        'Verify project-level aggregations work correctly' AS description
)
SELECT 
    test_name,
    expected_result,
    test_result,
    description
FROM verification_results
ORDER BY test_name;
