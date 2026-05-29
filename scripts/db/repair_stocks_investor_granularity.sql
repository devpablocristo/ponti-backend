-- ============================================================================
-- Reparación one-shot de datos de stocks dañados por bugs históricos.
--
-- Bugs originales (presentes desde la creación del módulo, ~sept 2025):
--   1) supply_movements pickaba stock por (project, supply) sin investor →
--      stock_id terminó apuntando al stock del investor equivocado.
--   2) UpdateCloseDateByProject cerraba N stocks pero replicaba sólo 1 al
--      período siguiente → quedan triplets sin stock activo.
--   3) DeleteSupplyMovement borraba stocks por (project, supply) → pisaba
--      filas de otros investors (no genera huérfanos directos pero amplifica
--      los anteriores).
--
-- Esos bugs ya están corregidos en código forward. Este script repara
-- ÚNICAMENTE los datos históricos dañados. No es idempotente en el sentido
-- estricto, pero los WHERE filtran solo casos rotos, así que correrlo dos
-- veces no genera daño.
--
-- USO:
--   1) Backup de la DB destino (siempre).
--   2) Correr el bloque de auditoría de abajo y guardar los conteos.
--   3) BEGIN; \i repair_stocks_investor_granularity.sql; → revisar el
--      bloque de verificación final → COMMIT o ROLLBACK.
-- ============================================================================

-- ----------------------------------------------------------------------------
-- AUDITORÍA PRE-REPARACIÓN (correr suelto, no es parte de la transacción)
-- ----------------------------------------------------------------------------
-- SELECT 'mismatched_movements' AS metric, COUNT(*) AS value
-- FROM supply_movements sm
-- JOIN stocks s ON s.id = sm.stock_id
-- WHERE sm.deleted_at IS NULL AND s.deleted_at IS NULL
--   AND (sm.project_id <> s.project_id
--     OR sm.supply_id <> s.supply_id
--     OR sm.investor_id <> s.investor_id);
--
-- WITH mismatched AS (
--   SELECT DISTINCT sm.project_id, sm.supply_id, sm.investor_id
--   FROM supply_movements sm
--   JOIN stocks s ON s.id = sm.stock_id
--   WHERE sm.deleted_at IS NULL AND s.deleted_at IS NULL
--     AND (sm.project_id <> s.project_id
--       OR sm.supply_id <> s.supply_id
--       OR sm.investor_id <> s.investor_id))
-- SELECT 'missing_triplets' AS metric, COUNT(*) AS value
-- FROM mismatched m
-- LEFT JOIN stocks s
--   ON s.project_id = m.project_id AND s.supply_id = m.supply_id
--   AND s.investor_id = m.investor_id AND s.deleted_at IS NULL
-- WHERE s.id IS NULL;
--
-- WITH latest_close AS (
--   SELECT project_id, MAX(close_date) AS close_date
--   FROM stocks WHERE deleted_at IS NULL AND close_date IS NOT NULL
--   GROUP BY project_id),
-- closed_triplets AS (
--   SELECT s.project_id, s.supply_id, s.investor_id
--   FROM stocks s JOIN latest_close lc
--     ON lc.project_id = s.project_id AND lc.close_date = s.close_date
--   WHERE s.deleted_at IS NULL),
-- active_triplets AS (
--   SELECT project_id, supply_id, investor_id FROM stocks
--   WHERE deleted_at IS NULL AND close_date IS NULL)
-- SELECT ct.project_id,
--        COUNT(*) - COUNT(at.project_id) AS missing_active_triplets
-- FROM closed_triplets ct LEFT JOIN active_triplets at
--   ON at.project_id = ct.project_id AND at.supply_id = ct.supply_id
--   AND at.investor_id = ct.investor_id
-- GROUP BY ct.project_id
-- HAVING COUNT(*) <> COUNT(at.project_id)
-- ORDER BY ct.project_id;

-- ============================================================================
-- TRANSACCIÓN: ejecutar dentro de BEGIN; ... COMMIT;
-- ============================================================================

-- ----------------------------------------------------------------------------
-- PASO 1: crear stocks faltantes para los movimientos mal ligados.
-- Deriva el período (close_date + month + year) del stock al que el movimiento
-- estaba mal apuntado (mejor heurística disponible sin metadata adicional).
-- ----------------------------------------------------------------------------
WITH bad_links AS (
    SELECT
        sm.project_id,
        sm.supply_id,
        sm.investor_id,
        tpl.close_date    AS template_close_date,
        tpl.month_period  AS template_month_period,
        tpl.year_period   AS template_year_period,
        MIN(sm.created_at) OVER (
            PARTITION BY sm.project_id, sm.supply_id, sm.investor_id,
                         tpl.close_date, tpl.month_period, tpl.year_period
        ) AS first_created_at
    FROM public.supply_movements sm
    JOIN public.stocks tpl
      ON tpl.id = sm.stock_id AND tpl.deleted_at IS NULL
    WHERE sm.deleted_at IS NULL
      AND (sm.project_id <> tpl.project_id
        OR sm.supply_id  <> tpl.supply_id
        OR sm.investor_id <> tpl.investor_id)
),
missing_period_stocks AS (
    SELECT DISTINCT
        bl.project_id, bl.supply_id, bl.investor_id,
        bl.template_close_date, bl.template_month_period,
        bl.template_year_period, bl.first_created_at
    FROM bad_links bl
    LEFT JOIN public.stocks s
      ON s.deleted_at IS NULL
     AND s.project_id  = bl.project_id
     AND s.supply_id   = bl.supply_id
     AND s.investor_id = bl.investor_id
     AND s.close_date  IS NOT DISTINCT FROM bl.template_close_date
     AND s.month_period = bl.template_month_period
     AND s.year_period  = bl.template_year_period
    WHERE s.id IS NULL
)
INSERT INTO public.stocks (
    project_id, supply_id, investor_id, close_date,
    real_stock_units, initial_units, year_period, month_period,
    created_at, updated_at, units_entered, units_consumed, has_real_stock_count
)
SELECT
    mps.project_id, mps.supply_id, mps.investor_id, mps.template_close_date,
    0, 0, mps.template_year_period, mps.template_month_period,
    COALESCE(mps.first_created_at, CURRENT_TIMESTAMP),
    COALESCE(mps.first_created_at, CURRENT_TIMESTAMP),
    0, 0, false
FROM missing_period_stocks mps;

-- ----------------------------------------------------------------------------
-- PASO 2: reapuntar movimientos al stock correcto del mismo período.
-- ----------------------------------------------------------------------------
WITH bad_links AS (
    SELECT
        sm.id AS movement_id,
        sm.project_id, sm.supply_id, sm.investor_id,
        tpl.close_date    AS template_close_date,
        tpl.month_period  AS template_month_period,
        tpl.year_period   AS template_year_period
    FROM public.supply_movements sm
    JOIN public.stocks tpl
      ON tpl.id = sm.stock_id AND tpl.deleted_at IS NULL
    WHERE sm.deleted_at IS NULL
      AND (sm.project_id <> tpl.project_id
        OR sm.supply_id  <> tpl.supply_id
        OR sm.investor_id <> tpl.investor_id)
)
UPDATE public.supply_movements sm
SET stock_id   = target.id,
    updated_at = CURRENT_TIMESTAMP
FROM bad_links bl
JOIN public.stocks target
  ON target.deleted_at IS NULL
 AND target.project_id  = bl.project_id
 AND target.supply_id   = bl.supply_id
 AND target.investor_id = bl.investor_id
 AND target.close_date  IS NOT DISTINCT FROM bl.template_close_date
 AND target.month_period = bl.template_month_period
 AND target.year_period  = bl.template_year_period
WHERE sm.id = bl.movement_id
  AND sm.stock_id <> target.id;

-- ----------------------------------------------------------------------------
-- PASO 3: regenerar stocks activos faltantes a partir del último cierre de
-- cada proyecto (replica los triplets cerrados que no tienen contraparte
-- activa). Período nuevo = mes+1 del último cierre.
-- ----------------------------------------------------------------------------
WITH latest_close AS (
    SELECT project_id, MAX(close_date) AS latest_close_date
    FROM public.stocks
    WHERE deleted_at IS NULL AND close_date IS NOT NULL
    GROUP BY project_id
),
latest_closed_triplets AS (
    SELECT DISTINCT s.project_id, s.supply_id, s.investor_id, lc.latest_close_date
    FROM public.stocks s
    JOIN latest_close lc
      ON lc.project_id = s.project_id AND lc.latest_close_date = s.close_date
    WHERE s.deleted_at IS NULL
),
missing_active AS (
    SELECT lct.project_id, lct.supply_id, lct.investor_id, lct.latest_close_date
    FROM latest_closed_triplets lct
    LEFT JOIN public.stocks active
      ON active.deleted_at IS NULL
     AND active.project_id  = lct.project_id
     AND active.supply_id   = lct.supply_id
     AND active.investor_id = lct.investor_id
     AND active.close_date  IS NULL
    WHERE active.id IS NULL
)
INSERT INTO public.stocks (
    project_id, supply_id, investor_id, close_date,
    real_stock_units, initial_units, year_period, month_period,
    created_at, updated_at, units_entered, units_consumed, has_real_stock_count
)
SELECT
    ma.project_id, ma.supply_id, ma.investor_id, NULL,
    0, 0,
    CASE WHEN EXTRACT(MONTH FROM ma.latest_close_date)::int = 12
         THEN EXTRACT(YEAR FROM ma.latest_close_date)::int + 1
         ELSE EXTRACT(YEAR FROM ma.latest_close_date)::int END,
    CASE WHEN EXTRACT(MONTH FROM ma.latest_close_date)::int = 12
         THEN 1
         ELSE EXTRACT(MONTH FROM ma.latest_close_date)::int + 1 END,
    CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 0, 0, false
FROM missing_active ma;

-- ----------------------------------------------------------------------------
-- VERIFICACIÓN POST-REPARACIÓN: las 3 métricas deben dar 0.
-- Si alguna no da 0, ROLLBACK e investigar antes de COMMIT.
-- ----------------------------------------------------------------------------
WITH mismatched AS (
    SELECT COUNT(*) AS c
    FROM supply_movements sm
    JOIN stocks s ON s.id = sm.stock_id
    WHERE sm.deleted_at IS NULL AND s.deleted_at IS NULL
      AND (sm.project_id <> s.project_id
        OR sm.supply_id <> s.supply_id
        OR sm.investor_id <> s.investor_id)
),
missing_triplet AS (
    WITH m AS (
        SELECT DISTINCT sm.project_id, sm.supply_id, sm.investor_id
        FROM supply_movements sm
        JOIN stocks s ON s.id = sm.stock_id
        WHERE sm.deleted_at IS NULL AND s.deleted_at IS NULL
          AND (sm.project_id <> s.project_id
            OR sm.supply_id <> s.supply_id
            OR sm.investor_id <> s.investor_id))
    SELECT COUNT(*) AS c
    FROM m LEFT JOIN stocks s
      ON s.project_id = m.project_id AND s.supply_id = m.supply_id
     AND s.investor_id = m.investor_id AND s.deleted_at IS NULL
    WHERE s.id IS NULL
),
missing_active AS (
    WITH lc AS (
        SELECT project_id, MAX(close_date) AS d FROM stocks
        WHERE deleted_at IS NULL AND close_date IS NOT NULL
        GROUP BY project_id),
    ct AS (
        SELECT s.project_id, s.supply_id, s.investor_id
        FROM stocks s JOIN lc ON lc.project_id = s.project_id AND lc.d = s.close_date
        WHERE s.deleted_at IS NULL),
    at AS (
        SELECT project_id, supply_id, investor_id FROM stocks
        WHERE deleted_at IS NULL AND close_date IS NULL)
    SELECT COALESCE(SUM(CASE WHEN at.project_id IS NULL THEN 1 ELSE 0 END), 0) AS c
    FROM ct LEFT JOIN at
      ON at.project_id = ct.project_id AND at.supply_id = ct.supply_id
     AND at.investor_id = ct.investor_id
)
SELECT
    (SELECT c FROM mismatched)      AS mismatched_movements_after,
    (SELECT c FROM missing_triplet) AS missing_triplets_after,
    (SELECT c FROM missing_active)  AS missing_active_after;
