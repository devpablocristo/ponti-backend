-- Grants para app_stg en new_ponti_db_staging
-- Conectar: gcloud sql connect new-ponti-db-dev --user=postgres --database=postgres --project=new-ponti-dev
-- Luego: \i docs/grants_app_stg.sql  (o pegar el contenido)

-- 1) CONNECT a la DB
GRANT CONNECT ON DATABASE new_ponti_db_staging TO app_stg;
\c new_ponti_db_staging

-- 2) Schema public
GRANT USAGE ON SCHEMA public TO app_stg;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO app_stg;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO app_stg;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO app_stg;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT USAGE, SELECT ON SEQUENCES TO app_stg;

-- 3) Schemas v4 (si existen)
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.schemata WHERE schema_name = 'v4_core') THEN
    EXECUTE 'GRANT USAGE ON SCHEMA v4_core TO app_stg';
    EXECUTE 'GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA v4_core TO app_stg';
  END IF;
  IF EXISTS (SELECT 1 FROM information_schema.schemata WHERE schema_name = 'v4_ssot') THEN
    EXECUTE 'GRANT USAGE ON SCHEMA v4_ssot TO app_stg';
    EXECUTE 'GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA v4_ssot TO app_stg';
  END IF;
  IF EXISTS (SELECT 1 FROM information_schema.schemata WHERE schema_name = 'v4_calc') THEN
    EXECUTE 'GRANT USAGE ON SCHEMA v4_calc TO app_stg';
    EXECUTE 'GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA v4_calc TO app_stg';
    EXECUTE 'GRANT SELECT ON ALL TABLES IN SCHEMA v4_calc TO app_stg';
  END IF;
  IF EXISTS (SELECT 1 FROM information_schema.schemata WHERE schema_name = 'v4_report') THEN
    EXECUTE 'GRANT USAGE ON SCHEMA v4_report TO app_stg';
    EXECUTE 'GRANT SELECT ON ALL TABLES IN SCHEMA v4_report TO app_stg';
  END IF;
END $$;
