-- Migration: 000326_fix_summary_results_rent_bug (DOWN)
DROP VIEW IF EXISTS v4_report.summary_results;
CREATE VIEW v4_report.summary_results AS SELECT * FROM v3_report_summary_results_view;
