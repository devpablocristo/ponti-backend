-- ========================================
-- MIGRACIÓN 000080: CREAR VISTA v3_dashboard (UP)
-- ========================================
-- 
-- Objetivo: Vista de dashboard v3 basada en vistas base de costos e ingresos
-- Fecha: 2025-09-12
-- Autor: Sistema
-- 
-- Nota: Código en inglés, comentarios en español.

CREATE OR REPLACE VIEW public.v3_dashboard AS
WITH lots_base AS (
  SELECT
    l.id          AS lot_id,
    f.project_id  AS project_id,
    l.hectares    AS hectares,
    l.tons        AS tons,
    l.sowing_date AS sowing_date
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  WHERE l.deleted_at IS NULL
),
w_min AS (
  SELECT project_id, MIN(date) AS start_date
  FROM public.workorders
  WHERE deleted_at IS NULL
  GROUP BY project_id
),
w_max AS (
  SELECT project_id, MAX(date) AS end_date
  FROM public.workorders
  WHERE deleted_at IS NULL
  GROUP BY project_id
)
SELECT
  p.customer_id,
  p.id            AS project_id,
  p.campaign_id,
  -- Progresos operativos (si no hay lotes, da 0)
  COALESCE(SUM(CASE WHEN lb.sowing_date IS NOT NULL
                    THEN lb.hectares ELSE 0 END), 0)::numeric(14,2) AS sowing_hectares,
  COALESCE(SUM(lb.hectares), 0)::numeric(14,2)                    AS sowing_total_hectares,
  calc.percentage(
    COALESCE(SUM(CASE WHEN lb.sowing_date IS NOT NULL THEN lb.hectares ELSE 0 END), 0)::numeric,
    COALESCE(SUM(lb.hectares), 0)::numeric
  )::numeric(6,2)                                                 AS sowing_progress_pct,

  COALESCE(SUM(CASE WHEN lb.tons IS NOT NULL AND lb.tons > 0
                    THEN lb.hectares ELSE 0 END), 0)::numeric(14,2) AS harvest_hectares,
  COALESCE(SUM(lb.hectares), 0)::numeric(14,2)                       AS harvest_total_hectares,
  calc.percentage(
    COALESCE(SUM(CASE WHEN lb.tons IS NOT NULL AND lb.tons > 0 THEN lb.hectares ELSE 0 END), 0)::numeric,
    COALESCE(SUM(lb.hectares), 0)::numeric
  )::numeric(6,2)                                                    AS harvest_progress_pct,

  -- Costos e ingresos agregados calculados vía funciones SSOT (calc.*)
  COALESCE(SUM(calc.direct_cost_for_lot(lb.lot_id)), 0)::numeric(14,2) AS executed_costs_usd,
  COALESCE(p.admin_cost, 0)::numeric(14,2)                              AS budget_cost_usd,
  calc.percentage(
    COALESCE(SUM(calc.direct_cost_for_lot(lb.lot_id)), 0)::numeric,
    COALESCE(p.admin_cost, 0)::numeric
  )::numeric(6,2)                                                       AS costs_progress_pct,

  COALESCE(SUM(calc.income_net_total_for_lot(lb.lot_id)), 0)::numeric(14,2) AS income_usd,
  COALESCE(SUM(calc.operating_result_per_ha_for_lot(lb.lot_id) * lb.hectares), 0)::numeric(14,2)
                                                                             AS operating_result_usd,
  COALESCE(SUM(calc.direct_cost_for_lot(lb.lot_id)), 0)::numeric(14,2)        AS operating_result_total_costs_usd,
  calc.renta_pct(
    COALESCE(SUM(calc.operating_result_per_ha_for_lot(lb.lot_id) * lb.hectares), 0)::double precision,
    COALESCE(SUM(calc.direct_cost_for_lot(lb.lot_id)), 0)::double precision
  )::numeric(6,2)                                                       AS operating_result_pct,

  -- Fechas operativas
  w_min.start_date,
  w_max.end_date,
  public.calculate_campaign_closing_date(w_max.end_date)                AS campaign_closing_date

FROM public.projects p
LEFT JOIN lots_base lb ON lb.project_id = p.id
LEFT JOIN w_min ON w_min.project_id = p.id
LEFT JOIN w_max ON w_max.project_id = p.id
WHERE p.deleted_at IS NULL
GROUP BY
  p.customer_id, p.id, p.campaign_id, p.admin_cost,
  w_min.start_date, w_max.end_date;


