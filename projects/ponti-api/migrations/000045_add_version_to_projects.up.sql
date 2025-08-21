-- Agregar campo version a la tabla projects
ALTER TABLE projects ADD COLUMN version BIGINT NOT NULL DEFAULT 1;

-- Crear índice para optimizar búsquedas por version
CREATE INDEX idx_projects_version ON projects(version);
