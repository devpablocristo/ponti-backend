BEGIN;

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
    cat.name AS labor_category_name,
    cat.type_id AS labor_category_type_id
  FROM public.workorders w
  JOIN raw r ON r.lot_id = w.lot_id
  JOIN public.labors lb ON lb.id = w.labor_id
  LEFT JOIN public.categories cat ON cat.id = lb.category_id
  WHERE w.deleted_at IS NULL
    AND w.effective_area IS NOT NULL
    AND w.effective_area > 0
),
workorder_totals AS (
  SELECT
    lot_id,
    SUM(
      CASE
        WHEN labor_category_name = 'Siembra' AND labor_category_type_id = 4
        THEN effective_area
        ELSE 0
      END
    )::numeric AS seeded_area_ha,
    SUM(
      CASE
        WHEN labor_category_name = 'Cosecha' AND labor_category_type_id = 4
        THEN effective_area
        ELSE 0
      END
    )::numeric AS harvested_area_ha,
    SUM(labor_price * effective_area)::numeric AS labor_cost_usd
  FROM workorder_base
  GROUP BY lot_id
),
supply_totals AS (
  SELECT
    wb.lot_id,
    SUM(COALESCE(wi.total_used, 0) * COALESCE(s.price, 0))::numeric AS supplies_cost_us