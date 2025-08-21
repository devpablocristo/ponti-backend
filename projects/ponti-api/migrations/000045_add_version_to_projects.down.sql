-- Remover campo version de la tabla projects
DROP INDEX IF EXISTS idx_projects_version;
ALTER TABLE projects DROP COLUMN IF EXISTS version;
