-- ========================================
-- MIGRATION 000219 FAST LOT REPORTING VIEWS (UP)
-- ========================================
-- Reemplaza los cálculos por-lote de lot_base_costs por agregaciones set-based.
-- La interfaz de la vista se mantiene igual para que v4_report.lot_metrics y
-- v4_report.lot_list sigan siendo el contrato estable de reporting.

BEGIN;

WITH ranked AS (
  SELECT
    id,
    ROW_NUMBER() OVER (
      PARTITION BY project_id, crop_id
      ORDER BY updated_at DESC NULLS LAST, id DESC
    ) AS row_number
  FROM public.crop_commercializations
  WHERE deleted_at IS NULL
)
UPDATE public.crop_commercializations cc
SET
  deleted_at = NOW(),
  deleted_by = COALESCE(deleted_by, 'migration:000219'),
  updated_at = NOW(),
  updated_by = COALESCE(updated_by, 'migration:000219')
FROM ranked
WHERE ranked.id = cc.id
  AND ranked.row_number > 1;

CREATE UNIQUE INDEX IF NOT EXISTS uq_crop_commercializations_project_crop_active
  ON public.crop_commercializations (project_id, crop_id)
  WHERE deleted_at IS NULL;

CREATE OR REPLACE VIEW v4_calc.lot_base_costs AS
WITH
raw AS (
  SELECT
    f.project_id,
    f.id AS field_id,
    f.lease_type_id,
    f.lease_type_percent,
    f.lease_type_value,
    COALESCE(p.admin_cost, 0)::numeric AS admin_cost_per_ha_usd,
    l.current_crop_id,
    l.id AS lot_id,
    l.name AS lot_name,
    COALESCE(l.hectares, 0)::numeric AS hectares,
    COALESCE(l.tons, 0)::numeric AS tons,
    l.sowing_date,
    COALESCE(cc.net_price, 0)::numeric AS net_price_usd_tn
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  JOIN public.projects p ON p.id = f.project_id AND p.deleted_at IS NULL
  LEFT JOIN public.crop_commercializations cc
    ON cc.project_id = f.project_id
   AND cc.crop_id = l.current_crop_id
   AND cc.deleted_at IS NULL
  WHERE l.deleted_at IS NULL
),
workorder_base AS (
  SELECT
    w.id AS workorder_id,
    w.lot_id,
    w.effective_area,
    COALESCE(lb.price, 0)::numeric AS labor_price,
    lb.category_id
  FROM public.workorders w
  JOIN raw r ON r.lot_id = w.lot_id
  JOIN public.labors lb ON lb.id = w.labor_id
  WHERE w.deleted_at IS NULL
    AND w.effective_area IS NOT NULL
    AND w.effective_area > 0
),
workorder_totals AS (
  SELECT
    lot_id,
    SUM(CASE WHEN category_id = 9 THEN effective_area ELSE 0 END)::numeric AS seeded_area_ha,
    SUM(labor_price * effective_area)::numeric AS labor_cost_usd
  FROM workorder_base
  GROUP BY lot_id
),
supply_totals AS (
  SELECT
    wb.lot_id,
    SUM(COALESCE(wi.total_used, 0) * COALESCE(s.price, 0))::numeric AS supplies_cost_usd
  FROM workorder_base wb
  JOIN public.workorder_items wi ON wi.workorder_id = wb.workorder_id AND wi.deleted_at IS NULL
  LEFT JOIN public.supplies s ON s.id = wi.supply_id
  GROUP BY wb.lot_id
),
base AS (
  SELECT
    r.project_id,
    r.field_id,
    r.current_crop_id,
    r.lot_id,
    r.lot_name,
    r.hectares,
    r.tons,
    r.sowing_date,
    r.lease_type_id,
    r.lease_type_percent,
    r.lease_type_value,
    r.admin_cost_per_ha_usd,
    COALESCE(wt.seeded_area_ha, 0)::numeric AS seeded_area_ha,
    v4_core.harvested_area(r.tons, r.hectares)::numeric AS harvested_area_ha,
    v4_core.per_ha(r.tons, r.hectares)::numeric AS yield_tn_per_ha,
    COALESCE(wt.labor_cost_usd, 0)::numeric AS labor_cost_usd,
    COALESCE(st.supplies_cost_usd, 0)::numeric AS supplies_cost_usd,
    (COALESCE(wt.labor_cost_usd, 0) + COALESCE(st.supplies_cost_usd, 0))::numeric AS direct_cost_usd,
    (r.tons * r.net_price_usd_tn)::numeric AS income_net_total_usd
  FROM raw r
  LEFT JOIN workorder_totals wt ON wt.lot_id = r.lot_id
  LEFT JOIN supply_totals st ON st.lot_id = r.lot_id
),
derived AS (
  SELECT
    b.*,
    v4_core.per_ha(b.income_net_total_usd, b.hectares)::numeric AS income_net_per_ha_usd,
    v4_core.per_ha(b.direct_cost_usd, b.hectares)::numeric AS direct_cost_per_ha_usd
  FROM base b
)
SELECT
  project_id,
  field_id,
  current_crop_id,
  lot_id,
  lot_name,
  hectares,
  tons,
  sowing_date,
  seeded_area_ha,
  harvested_area_ha,
  yield_tn_per_ha,
  labor_cost_usd,
  supplies_cost_usd,
  direct_cost_usd,
  income_net_total_usd,
  income_net_per_ha_usd,
  direct_cost_per_ha_usd,
  v4_core.rent_per_ha(
    lease_type_id,
    lease_type_percent,
    lease_type_value,
    income_net_per_ha_usd,
    direct_cost_per_ha_usd,
    admin_cost_per_ha_usd
  )::numeric AS rent_per_ha_usd,
  CASE
    WHEN lease_type_id IN (3, 4) THEN COALESCE(lease_type_value, 0)
    ELSE 0
  END::numeric AS rent_fixed_per_ha_usd,
  admin_cost_per_ha_usd,
  seeded_area_ha AS sowed_area_ha,
  seeded_area_ha AS sown_area_ha
FROM derived;

COMMIT;
