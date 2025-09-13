-- ========================================
-- MIGRATION 000001: DROP V3 SCHEMA (DOWN)
-- ========================================
-- 
-- Purpose: Drop all tables and functions for v3 system
-- Date: 2025-09-13
-- Author: System
-- 
-- Note: Code in English, comments in Spanish.

BEGIN;

-- Eliminar funciones
DROP FUNCTION IF EXISTS public.calculate_campaign_closing_date(date);
DROP FUNCTION IF EXISTS public.get_project_dollar_value(bigint, character varying);
DROP FUNCTION IF EXISTS public.get_iva_percentage();
DROP FUNCTION IF EXISTS public.get_default_fx_rate();
DROP FUNCTION IF EXISTS public.get_campaign_closure_days();
DROP FUNCTION IF EXISTS public.get_app_parameter_integer(character varying);
DROP FUNCTION IF EXISTS public.get_app_parameter_decimal(character varying);
DROP FUNCTION IF EXISTS public.get_app_parameter(character varying);
DROP FUNCTION IF EXISTS public.update_timestamp();

-- Eliminar tablas en orden inverso (respetando FKs)
DROP TABLE IF EXISTS supply_movements;
DROP TABLE IF EXISTS stocks;
DROP TABLE IF EXISTS invoices;
DROP TABLE IF EXISTS project_managers;
DROP TABLE IF EXISTS project_investors;
DROP TABLE IF EXISTS workorder_items;
DROP TABLE IF EXISTS workorders;
DROP TABLE IF EXISTS crop_commercializations;
DROP TABLE IF EXISTS lots;
DROP TABLE IF EXISTS fields;
DROP TABLE IF EXISTS projects;
DROP TABLE IF EXISTS labors;
DROP TABLE IF EXISTS labor_categories;
DROP TABLE IF EXISTS labor_types;
DROP TABLE IF EXISTS supplies;
DROP TABLE IF EXISTS categories;
DROP TABLE IF EXISTS types;
DROP TABLE IF EXISTS crops;
DROP TABLE IF EXISTS lease_types;
DROP TABLE IF EXISTS investors;
DROP TABLE IF EXISTS managers;
DROP TABLE IF EXISTS providers;
DROP TABLE IF EXISTS customers;
DROP TABLE IF EXISTS campaigns;
DROP TABLE IF EXISTS project_dollar_values;
DROP TABLE IF EXISTS app_parameters;
DROP TABLE IF EXISTS fx_rates;
DROP TABLE IF EXISTS schema_migrations;
DROP TABLE IF EXISTS users;

-- Eliminar tipos
DROP TYPE IF EXISTS public.movement_type;

COMMIT;
