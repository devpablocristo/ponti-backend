-- ===========================================================
-- MIGRATION 000128 V4 REPORT WORKORDER METRICS WRAPPER (DOWN)
-- ===========================================================
-- Nota: SQL en inglés, comentarios en español

BEGIN;

CREATE OR REPLACE VIEW v4_report.workorder_metrics AS
WITH lot_ids AS (
  SELECT DISTINCT
    w.project_id,
    w.field_id,
    w.lot_id
  FROM public.workorders w
  WHERE w.deleted_at IS NULL
)
SELECT
  project_id,
  field_id,
  lot_id,
  v4_ssot.surface_for_lot(lot_id) AS surface_ha,
  v4_ssot.liters_for_lot(lot_id) AS liters,
  v4_ssot.kilograms_for_lot(lot_id) AS kilograms,
  v4_ssot.labor_cost_for_lot(lot_id) AS labor_cost_usd,
  v4_ssot.supply_cost_for_lot(lot_id) AS supplies_cost_usd,
  v4_ssot.labor_cost_for_lot(lot_id) + v4_ssot.supply_cost_for_lot(lot_id) AS direct_cost_usd,
  v4_core.per_ha(
    v4_ssot.labor_cost_for_lot(lot_id) + v4_ssot.supply_cost_for_lot(lot_id),
    v4_ssot.surface_for_lot(lot_id)
  ) AS avg_cost_per_ha_usd,
  v4_core.per_ha(
    v4_ssot.liters_for_lot(lot_id),
    v4_ssot.surface_for_lot(lot_id)
  ) AS liters_per_ha,
  v4_core.per_ha(
    v4_ssot.kilograms_for_lot(lot_id),
    v4_ssot.surface_for_lot(lot_id)
  ) AS kilograms_per_ha
FROM lot_ids li;

COMMIT;
