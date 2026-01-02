-- Migration: 000325_optimize_summary_results_ssot (DOWN)
-- Rollback: Volver a la versión 000324 (copia de v3 con fix)
DROP VIEW IF EXISTS v4_report.summary_results;
