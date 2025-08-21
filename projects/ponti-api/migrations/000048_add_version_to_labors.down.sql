-- Revertir adición de columna version
DROP INDEX IF EXISTS idx_labors_id_version_notdel;
ALTER TABLE labors DROP COLUMN IF EXISTS version;
