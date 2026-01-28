-- ========================================
-- MIGRACIÓN 000352: FIX ADMIN COST PRORATION (DOWN)
-- ========================================
--
-- Revierte a lógica anterior:
-- - admin_cost_per_ha_for_lot: admin_cost sin prorrateo
-- - admin_cost_total_for_project: admin_cost * total_hectares
--
-- Nota: Código en inglés, comentarios en español

BEGIN;

CREATE OR REPLACE FUNCTION v4_ssot.admin_cost_per_ha_for_lot(p_lot_id bigint)
RETURNS numeric
LANGUAGE sql STABLE AS $$
  -- Retorna admin_cost del proyecto tal cual
  SELECT COALESCE(p.admin_cost, 0)::numeric
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  JOIN public.projects p ON p.id = f.project_id AND p.deleted_at IS NULL
  WHERE l.id = p_lot_id AND l.deleted_at IS NULL
$$;

CREATE OR REPLACE FUNCTION v4_ssot.admin_cost_total_for_project(p_project_id bigint)
RETURNS double precision
LANGUAGE sql STABLE AS $$
  -- Retorna admin_cost × total_hectares para el Dashboard (costo total del proyecto)
  SELECT COALESCE(
    (SELECT p.admin_cost * v4_ssot.total_hectares_for_project(p_project_id)
     FROM public.projects p
     WHERE p.id = p_project_id AND p.deleted_at IS NULL)
  , 0)::double precision
$$;

COMMIT;
