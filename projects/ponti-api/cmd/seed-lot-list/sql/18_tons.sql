-- =======================
-- TONELADAS PARA LIST LOT
-- =======================
-- Actualizar toneladas de los lotes (rendimiento 8 tn/ha)

UPDATE lots SET tons = 20.0 WHERE id = 1;  -- Parcela A1: 2.5 ha × 8 tn/ha = 20 tn
UPDATE lots SET tons = 24.0 WHERE id = 2;  -- Parcela A2: 3.0 ha × 8 tn/ha = 24 tn
UPDATE lots SET tons = 9.6 WHERE id = 3;  -- Parcela B1: 1.2 ha × 8 tn/ha = 9.6 tn

-- Verificar actualización
SELECT '✅ Toneladas actualizadas' as status, id, name, hectares, tons, ROUND(tons/hectares, 2) as rendimiento_tn_ha FROM lots ORDER BY id;
