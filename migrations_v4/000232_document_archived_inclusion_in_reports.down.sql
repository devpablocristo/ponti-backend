-- 000232_document_archived_inclusion_in_reports.down.sql
-- Reverse: drop the comments. The views and functions themselves are not
-- touched (their definitions live in earlier migrations).

BEGIN;

COMMENT ON VIEW v4_report.workorder_list IS NULL;
COMMENT ON VIEW v4_report.labor_list IS NULL;
COMMENT ON FUNCTION v4_ssot.labor_cost_for_lot(bigint) IS NULL;
COMMENT ON FUNCTION v4_ssot.labor_cost_pre_harvest_for_lot(bigint) IS NULL;
COMMENT ON FUNCTION v4_ssot.seeded_area_for_lot(bigint) IS NULL;
COMMENT ON FUNCTION v4_ssot.supply_cost_for_lot(bigint) IS NULL;

COMMIT;
