-- 000232_document_archived_inclusion_in_reports.up.sql
--
-- Document the INTENTIONAL inclusion of archived labors/supplies in report
-- views and SSOT functions. The behaviour was introduced by migration
-- 000196_allow_archived_labors_supplies_in_views (workorders históricas
-- preservan la labor/insumo que tuvieron aunque después se archiven) and
-- has been a source of confusion in the schema since: a fresh reader of
-- v4_report.workorder_list or v4_ssot.labor_cost_for_lot has no signal in
-- the catalog that the JOINs to `labors`/`supplies` skip deleted_at on
-- purpose. This migration only adds COMMENT metadata — no view or function
-- is rewritten.

BEGIN;

COMMENT ON VIEW v4_report.workorder_list IS
'Listado detallado de workorders publicadas y drafts digitales. NOTA: incluye '
'labors e insumos archivados (deleted_at IS NOT NULL) intencionalmente — ver '
'migration 000196_allow_archived_labors_supplies_in_views. Una OT histórica '
'preserva la labor/insumo que se le cargó aunque después se archive en su '
'catálogo. Lo opuesto sería ocultar reportes históricos, comportamiento que '
'el negocio rechaza.';

COMMENT ON VIEW v4_report.labor_list IS
'Listado de labores con sus workorders asociadas. NOTA: incluye labors '
'archivadas en el contexto de OTs históricas — mismo criterio que '
'v4_report.workorder_list (ver migration 000196).';

COMMENT ON FUNCTION v4_ssot.labor_cost_for_lot(bigint) IS
'Costo total de labor para un lote. NOTA: suma costos de labores archivadas '
'intencionalmente — ver migration 000196. Las métricas históricas de '
'proyectos preservan los costos reales aunque la labor se archive después.';

COMMENT ON FUNCTION v4_ssot.labor_cost_pre_harvest_for_lot(bigint) IS
'Costo de labor pre-cosecha para un lote. NOTA: incluye labores archivadas — '
'mismo criterio que labor_cost_for_lot (ver migration 000196).';

COMMENT ON FUNCTION v4_ssot.seeded_area_for_lot(bigint) IS
'Superficie sembrada total para un lote, calculada vía workorders. NOTA: '
'incluye labores archivadas en el contexto de OTs históricas (ver migration '
'000196).';

COMMENT ON FUNCTION v4_ssot.supply_cost_for_lot(bigint) IS
'Costo total de insumos para un lote. NOTA: incluye supplies archivados en el '
'contexto de OTs históricas — mismo criterio que labor_cost_for_lot (ver '
'migration 000196).';

COMMIT;
