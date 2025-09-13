-- ========================================
-- MIGRATION 000084: DROP v3_report_field_crop_metrics_view VIEW (DOWN)
-- ========================================
-- 
-- Purpose: Drop the view created in the UP migration
-- Date: 2025-09-12
-- Author: System
-- 
-- Note: Code in English, comments in Spanish.

-- -------------------------------------------------------------------
-- v3_report_field_crop_metrics_view: rollback elimina la vista
-- -------------------------------------------------------------------
DROP VIEW IF EXISTS public.v3_report_field_crop_metrics_view;


