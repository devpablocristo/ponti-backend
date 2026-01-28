-- ==========================================
-- MIGRATION 000121 V4 ADMIN COST TOTAL (DOWN)
-- ==========================================
-- Nota: SQL en inglés, comentarios en español

BEGIN;

-- Revertir a admin_cost como costo por ha
CREATE OR REPLACE FUNCTION v4_ssot.admin_cost_per_ha_for_lot(p_lot_id bigint)
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(p.admin_cost, 0)::numeric
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  JOIN public.projects p ON p.id = f.project_id AND p.deleted_at IS NULL
  WHERE l.id = p_lot_id AND l.deleted_at IS NULL
$$;

CREATE OR REPLACE FUNCTION v4_ssot.admin_cost_total_for_project(p_project_id bigint)
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    (SELECT p.admin_cost * v4_ssot.total_hectares_for_project(p_project_id)
     FROM public.projects p
     WHERE p.id = p_project_id AND p.deleted_at IS NULL)
  , 0)::double precision
$$;

CREATE OR REPLACE FUNCTION v4_ssot.operating_result_total_for_project(p_project_id bigint)
RETURNS double precision
LANGUAGE sql STABLE AS $$
  WITH project_totals AS (
    SELECT
      p.id,
      p.admin_cost,
      COALESCE(SUM(l.hectares), 0)::double precision AS total_hectares
    FROM public.projects p
    LEFT JOIN public.fields f ON f.project_id = p.id AND f.deleted_at IS NULL
    LEFT JOIN public.lots l ON l.field_id = f.id AND l.deleted_at IS NULL
    WHERE p.id = p_project_id AND p.deleted_at IS NULL
    GROUP BY p.id, p.admin_cost
  ),
  lease_cost AS (
    SELECT
      COALESCE(
        SUM(v4_ssot.rent_per_ha_for_lot(l.id) * l.hectares),
        0
      )::double precision AS total_lease
    FROM public.lots l
    JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
    WHERE f.project_id = p_project_id AND l.deleted_at IS NULL
  )
  SELECT COALESCE(
    (SELECT COALESCE(SUM(v4_ssot.income_net_total_for_lot(l.id)), 0)::double precision
     FROM public.lots l
     JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
     WHERE f.project_id = p_project_id AND l.deleted_at IS NULL)
    -
    v4_ssot.direct_costs_total_for_project(p_project_id)
    -
    (SELECT total_lease FROM lease_cost)
    -
    (SELECT COALESCE(admin_cost * total_hectares, 0)::double precision FROM project_totals)
  , 0)::double precision
$$;

CREATE OR REPLACE FUNCTION v4_ssot.total_invested_cost_for_project(p_project_id bigint)
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    (SELECT COALESCE(SUM(v4_ssot.direct_cost_for_lot(l.id)), 0)::double precision
     FROM public.lots l
     JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
     WHERE f.project_id = p_project_id AND l.deleted_at IS NULL)
    +
    (SELECT COALESCE(SUM(v4_ssot.rent_per_ha_for_lot(l.id) * l.hectares), 0)::double precision
     FROM public.lots l
     JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
     WHERE f.project_id = p_project_id AND l.deleted_at IS NULL)
    +
    (SELECT COALESCE(SUM(v4_ssot.admin_cost_per_ha_for_lot(l.id) * l.hectares), 0)::double precision
     FROM public.lots l
     JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
     WHERE f.project_id = p_project_id AND l.deleted_at IS NULL)
  , 0)::double precision
$$;

CREATE OR REPLACE FUNCTION v4_ssot.total_budget_cost_for_project(p_project_id bigint)
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(p.admin_cost * 10, 0)::numeric
  FROM public.projects p
  WHERE p.id = p_project_id AND p.deleted_at IS NULL
$$;

COMMIT;
