-- Backfill: los borradores digitales (creados desde mobile) podían quedar con un
-- campaign_id distinto al del proyecto (incluso uno inexistente), lo que los dejaba
-- fuera de los listados filtrados por campaña y disparaba 400 en ResolveProjectIDs.
-- La campaña la determina el proyecto (projects.campaign_id es 1:1), así que la alineamos.
BEGIN;

UPDATE public.work_order_drafts wod
SET campaign_id = p.campaign_id
FROM public.projects p
WHERE p.id = wod.project_id
  AND wod.is_digital = true
  AND wod.deleted_at IS NULL
  AND wod.campaign_id IS DISTINCT FROM p.campaign_id;

COMMIT;
