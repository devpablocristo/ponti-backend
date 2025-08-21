-- Agregar columna de versión para bloqueo optimista
ALTER TABLE labors
  ADD COLUMN IF NOT EXISTS version BIGINT NOT NULL DEFAULT 1;

-- Índice compuesto para WHERE id, version (opcional pero recomendado)
CREATE INDEX IF NOT EXISTS idx_labors_id_version_notdel
  ON labors (id, version)
  WHERE deleted_at IS NULL;

-- Normalizar valores iniciales (si había NULLs por alguna razón histórica)
UPDATE labors SET version = 1 WHERE version IS NULL;
