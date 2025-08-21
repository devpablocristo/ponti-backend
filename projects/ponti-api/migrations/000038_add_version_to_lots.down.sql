DROP INDEX IF EXISTS idx_lots_id_version_notdel;
ALTER TABLE lots DROP COLUMN IF EXISTS version;
