-- ========================================
-- MIGRACIÓN 000352: FIX ADMIN COST PRORATION (UP)
-- ========================================
--
-- Propósito: Alinear admin_cost con la lógica v3 (prorrateo por hectáreas)
-- - admin_cost_per_ha_for_lot: admin_cost / total_hectares
-- - admin_cost_total_for_project: admin_cost (total del proyecto)
--
-- Nota: Código en inglés, comentarios en español

BEGIN;

-- Admin cost por ha: prorratea por hectáreas totales del proyecto
CREATE OR REPLACE FUNCTION v4_ssot.admin_cost_per_ha_for_lot(p_lot_id bigint)
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT CASE WHEN t.total_hectares > 0
              THEN COALESCE(p.admin_cost, 0)::numeric / t.total_hectares
              ELSE 0::numeric END
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  JOIN public.projects p ON p.id = f.project_id AND p.deleted_at IS NULL
  CROSS JOIN LATERAL (
    SELECT v4_ssot.total_hectares_for_project(f.project_id) AS total_hectares
  ) t
  WHERE l.id = p_lot_id AND l.deleted_at IS NULL
$$;

-- Admin cost total del proyecto: usa admin_cost sin multiplicar
CREATE OR REPLACE FUNCTION v4_ssot.admin_cost_total_for_project(p_project_id bigint)
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(p.admin_cost, 0)::double precision
  FROM public.projects p
  WHERE p.id = p_project_id AND p.deleted_at IS NULL
$$;

COMMIT;
