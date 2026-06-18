-- field_investors_to_lessees.sql
-- Migra los arrendatarios históricos: hoy viven (mal etiquetados) en field_investors;
-- pasan a field_lessees, que es la fuente que consume el form nuevo.
--
-- DEPENDENCIA DE ORDEN (importante): este script corre DESPUÉS del backfill de actores
-- (cmd/backfill-actors), que es quien sella investors.actor_id. Si se corre antes, el
-- puente directo no encuentra actor y migra 0 filas.
--
-- Idempotente: ON CONFLICT DO NOTHING en ambos INSERT. Relevamiento 2026-06-16 sobre
-- datos de PROD: 22/22 filas cubiertas tras el backfill (sin_actor = 0).
BEGIN;

-- 1) Puente directo: field_investors -> field_lessees vía investors.actor_id.
--    Todo field_investors es, en la práctica, un arrendatario (único escritor: el
--    selector de campo mal etiquetado). Se omiten filas sin actor (requerirían que el
--    backfill haya corrido) y las soft-deleted.
INSERT INTO public.field_lessees (field_id, actor_id, percentage, created_by, updated_by)
SELECT fi.field_id, i.actor_id, fi.percentage, fi.created_by, fi.updated_by
FROM public.field_investors fi
JOIN public.investors i ON i.id = fi.investor_id
WHERE i.actor_id IS NOT NULL
  AND fi.deleted_at IS NULL
ON CONFLICT (field_id, actor_id) DO NOTHING;

-- 2) Sellar rol lessee (OBLIGATORIO): el backfill crea esos actores con rol investor,
--    no lessee. Sin esto, el dato migra pero no aparece en el selector (que filtra lessee).
INSERT INTO public.actor_roles (actor_id, role)
SELECT DISTINCT actor_id, 'lessee'
FROM public.field_lessees
ON CONFLICT (actor_id, role) DO NOTHING;

COMMIT;

-- Verificación rápida (no transaccional):
--   SELECT count(*) FROM field_lessees;                                  -- esperado: filas migradas
--   SELECT count(*) FROM field_investors fi JOIN investors i ON i.id=fi.investor_id
--     WHERE i.actor_id IS NULL AND fi.deleted_at IS NULL;                -- esperado: 0 (cobertura total)
