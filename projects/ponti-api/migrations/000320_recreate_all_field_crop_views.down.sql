-- Rollback: las vistas se recrearán con las migraciones anteriores
DROP VIEW IF EXISTS v4_report.field_crop_metrics CASCADE;
DROP VIEW IF EXISTS v4_report.field_crop_rentabilidad CASCADE;
DROP VIEW IF EXISTS v4_report.field_crop_economicos CASCADE;
DROP VIEW IF EXISTS v4_report.field_crop_insumos CASCADE;
DROP VIEW IF EXISTS v4_report.field_crop_labores CASCADE;
DROP VIEW IF EXISTS v4_report.field_crop_cultivos CASCADE;
