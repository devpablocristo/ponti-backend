-- Elimina vista de consumos de insumos por proyecto
BEGIN;
DROP VIEW IF EXISTS v4_report.stock_consumed_by_supply;
COMMIT;
