-- Revertir adición de columna version
DROP INDEX IF EXISTS idx_supplies_id_version_notdel;
ALTER TABLE supplies DROP COLUMN IF EXISTS version;
