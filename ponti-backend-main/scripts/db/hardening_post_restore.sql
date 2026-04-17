-- Hardening post-restore for Cloud SQL export/import.
--
-- Problem:
-- - Cloud SQL import can leave report/calc views owned by `cloudsqlsuperuser`.
-- - Those views may select from base tables owned by the app user (e.g. `soalen-db-v3`).
-- - Since view execution uses the view owner's privileges, reads can fail with:
--     "permission denied for table <...>"
--
-- Fix:
-- - Grant the minimum read privileges needed to `cloudsqlsuperuser` on `public`.
-- - Keep defaults so new tables/sequences created by the app user stay readable.
--
-- This is safe in DEV/STG/PROD because it only grants read access to the Cloud SQL
-- maintenance role, not to anonymous users.

GRANT USAGE ON SCHEMA public TO cloudsqlsuperuser;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO cloudsqlsuperuser;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO cloudsqlsuperuser;

ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT ON TABLES TO cloudsqlsuperuser;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT USAGE, SELECT ON SEQUENCES TO cloudsqlsuperuser;

