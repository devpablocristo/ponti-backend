-- ========================================
-- MIGRATION 000216 WORKORDER SEQUENCE DAY BACKEND ONLY (UP)
-- ========================================
-- Agrega un numero secuencial por proyecto+fecha para desempatar workorders
-- del mismo dia sin depender del frontend.

BEGIN;

ALTER TABLE public.workorders
  ADD COLUMN IF NOT EXISTS sequence_day integer;

WITH ranked AS (
  SELECT
    w.id,
    ROW_NUMBER() OVER (
      PARTITION BY w.project_id, w.date
      ORDER BY w.created_at ASC, w.id ASC
    )::integer AS next_sequence_day
  FROM public.workorders w
)
UPDATE public.workorders w
SET sequence_day = ranked.next_sequence_day
FROM ranked
WHERE w.id = ranked.id
  AND (w.sequence_day IS NULL OR w.sequence_day <> ranked.next_sequence_day);

ALTER TABLE public.workorders
  ALTER COLUMN sequence_day SET NOT NULL;

ALTER TABLE public.workorders
  DROP CONSTRAINT IF EXISTS chk_workorders_sequence_day_positive;

ALTER TABLE public.workorders
  ADD CONSTRAINT chk_workorders_sequence_day_positive
  CHECK (sequence_day > 0);

CREATE UNIQUE INDEX IF NOT EXISTS uq_workorders_project_date_sequence_day
  ON public.workorders (project_id, date, sequence_day);

CREATE OR REPLACE FUNCTION public.assign_workorder_sequence_day()
RETURNS trigger
LANGUAGE plpgsql
AS $$
DECLARE
  v_next_sequence integer;
BEGIN
  IF NEW.project_id IS NULL OR NEW.date IS NULL THEN
    RETURN NEW;
  END IF;

  IF TG_OP = 'UPDATE'
     AND NEW.project_id = OLD.project_id
     AND NEW.date = OLD.date
     AND COALESCE(OLD.sequence_day, 0) > 0 THEN
    NEW.sequence_day = OLD.sequence_day;
    RETURN NEW;
  END IF;

  PERFORM pg_advisory_xact_lock(
    hashtextextended(format('workorders:%s:%s', NEW.project_id, NEW.date), 0)
  );

  SELECT COALESCE(MAX(w.sequence_day), 0) + 1
  INTO v_next_sequence
  FROM public.workorders w
  WHERE w.project_id = NEW.project_id
    AND w.date = NEW.date
    AND w.id <> COALESCE(NEW.id, -1);

  NEW.sequence_day = v_next_sequence;
  RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS trg_assign_workorder_sequence_day ON public.workorders;

CREATE TRIGGER trg_assign_workorder_sequence_day
BEFORE INSERT OR UPDATE OF project_id, date
ON public.workorders
FOR EACH ROW
EXECUTE FUNCTION public.assign_workorder_sequence_day();

CREATE OR REPLACE FUNCTION v4_ssot.first_workorder_id_for_project(p_project_id bigint)
RETURNS bigint
LANGUAGE sql STABLE AS $$
  SELECT id
  FROM public.workorders
  WHERE project_id = p_project_id
    AND deleted_at IS NULL
  ORDER BY date ASC, sequence_day ASC, id ASC
  LIMIT 1
$$;

CREATE OR REPLACE FUNCTION v4_ssot.last_workorder_id_for_project(p_project_id bigint)
RETURNS bigint
LANGUAGE sql STABLE AS $$
  SELECT id
  FROM public.workorders
  WHERE project_id = p_project_id
    AND deleted_at IS NULL
  ORDER BY date DESC, sequence_day DESC, id DESC
  LIMIT 1
$$;

CREATE OR REPLACE FUNCTION v4_ssot.first_workorder_number_for_project(p_project_id bigint)
RETURNS text
LANGUAGE sql STABLE AS $$
  SELECT number::text
  FROM public.workorders
  WHERE project_id = p_project_id
    AND deleted_at IS NULL
  ORDER BY date ASC, sequence_day ASC, id ASC
  LIMIT 1
$$;

CREATE OR REPLACE FUNCTION v4_ssot.last_workorder_number_for_project(p_project_id bigint)
RETURNS text
LANGUAGE sql STABLE AS $$
  SELECT number::text
  FROM public.workorders
  WHERE project_id = p_project_id
    AND deleted_at IS NULL
  ORDER BY date DESC, sequence_day DESC, id DESC
  LIMIT 1
$$;

CREATE OR REPLACE VIEW v4_report.dashboard_operational_indicators_field AS
SELECT
  p.id AS project_id,
  p.customer_id,
  p.campaign_id,
  f.id AS field_id,
  (SELECT w2.date
   FROM public.workorders w2
   WHERE w2.field_id = f.id
     AND w2.deleted_at IS NULL
   ORDER BY w2.date ASC, w2.sequence_day ASC, w2.id ASC
   LIMIT 1
  ) AS start_date,
  (SELECT w2.date
   FROM public.workorders w2
   WHERE w2.field_id = f.id
     AND w2.deleted_at IS NULL
   ORDER BY w2.date DESC, w2.sequence_day DESC, w2.id DESC
   LIMIT 1
  ) AS end_date,
  NULL::date AS campaign_closing_date,
  (SELECT w2.number::text
   FROM public.workorders w2
   WHERE w2.field_id = f.id
     AND w2.deleted_at IS NULL
   ORDER BY w2.date ASC, w2.sequence_day ASC, w2.id ASC
   LIMIT 1
  ) AS first_workorder_id,
  (SELECT w2.number::text
   FROM public.workorders w2
   WHERE w2.field_id = f.id
     AND w2.deleted_at IS NULL
   ORDER BY w2.date DESC, w2.sequence_day DESC, w2.id DESC
   LIMIT 1
  ) AS last_workorder_id,
  v4_ssot.last_manual_stock_entry_date_for_project(p.id) AS last_stock_count_date
FROM public.projects p
JOIN public.fields f ON f.project_id = p.id AND f.deleted_at IS NULL
WHERE p.deleted_at IS NULL;

-- La vista puede venir de ramas con columnas extra (por ejemplo is_digital/status).
-- CREATE OR REPLACE VIEW no permite "reducir" la forma de una vista existente,
-- así que la recreamos explícitamente para mantener el mismo estado final.
DROP VIEW IF EXISTS v4_report.workorder_list;

CREATE VIEW v4_report.workorder_list AS
SELECT
  w.id,
  w.number,
  w.project_id,
  w.field_id,
  p.name  AS project_name,
  f.name  AS field_name,
  l.name  AS lot_name,
  w.date,
  c.name  AS crop_name,
  lb.name AS labor_name,
  cat_lb.name AS labor_category_name,
  t.name  AS type_name,
  w.contractor,
  w.effective_area AS surface_ha,
  s.name AS supply_name,
  wi.total_used AS consumption,
  cat.name AS category_name,
  wi.final_dose AS dose_per_ha,
  s.price AS unit_price,
  CASE
    WHEN wi.final_dose IS NOT NULL AND s.price IS NOT NULL
    THEN v4_core.cost_per_ha(
      (wi.final_dose::numeric * s.price)::numeric,
      1::numeric
    )
    ELSE 0
  END AS supply_cost_per_ha,
  v4_core.supply_cost(
    wi.final_dose::numeric,
    s.price::numeric,
    w.effective_area::numeric
  ) AS supply_total_cost,
  w.sequence_day
FROM public.workorders w
JOIN public.projects p ON p.id = w.project_id AND p.deleted_at IS NULL
JOIN public.fields f ON f.id = w.field_id AND f.deleted_at IS NULL
JOIN public.lots l ON l.id = w.lot_id AND l.deleted_at IS NULL
JOIN public.crops c ON c.id = w.crop_id AND c.deleted_at IS NULL
JOIN public.labors lb ON lb.id = w.labor_id
LEFT JOIN public.categories cat_lb ON cat_lb.id = lb.category_id AND cat_lb.deleted_at IS NULL
LEFT JOIN public.workorder_items wi ON wi.workorder_id = w.id AND wi.deleted_at IS NULL
LEFT JOIN public.supplies s ON s.id = wi.supply_id
LEFT JOIN public.types t ON t.id = s.type_id AND t.deleted_at IS NULL
LEFT JOIN public.categories cat ON cat.id = s.category_id AND cat.deleted_at IS NULL
WHERE w.deleted_at IS NULL

UNION ALL

SELECT
  w.id,
  w.number,
  w.project_id,
  w.field_id,
  p.name  AS project_name,
  f.name  AS field_name,
  l.name  AS lot_name,
  w.date,
  c.name  AS crop_name,
  lb.name AS labor_name,
  cat_lb.name AS labor_category_name,
  'Labor'::varchar(250) AS type_name,
  w.contractor,
  w.effective_area AS surface_ha,
  lb.name::varchar(100) AS supply_name,
  0::numeric(18,6) AS consumption,
  cat_lb.name::varchar(250) AS category_name,
  0::numeric(18,6) AS dose_per_ha,
  lb.price::numeric(18,6) AS unit_price,
  lb.price::numeric AS supply_cost_per_ha,
  (lb.price * w.effective_area)::numeric AS supply_total_cost,
  w.sequence_day
FROM public.workorders w
JOIN public.projects p ON p.id = w.project_id AND p.deleted_at IS NULL
JOIN public.fields f ON f.id = w.field_id AND f.deleted_at IS NULL
JOIN public.lots l ON l.id = w.lot_id AND l.deleted_at IS NULL
JOIN public.crops c ON c.id = w.crop_id AND c.deleted_at IS NULL
JOIN public.labors lb ON lb.id = w.labor_id
LEFT JOIN public.categories cat_lb ON cat_lb.id = lb.category_id AND cat_lb.deleted_at IS NULL
WHERE w.deleted_at IS NULL
  AND w.effective_area IS NOT NULL
  AND w.effective_area > 0
  AND lb.price IS NOT NULL;

COMMIT;
