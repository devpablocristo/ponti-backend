-- Agregar columna de versión para bloqueo optimista
ALTER TABLE lots
  ADD COLUMN IF NOT EXISTS version BIGINT NOT NULL DEFAULT 1;

-- Índice compuesto para WHERE id, version (opcional pero recomendado)
CREATE INDEX IF NOT EXISTS idx_lots_id_version_notdel
  ON lots (id, version)
  WHERE deleted_at IS NULL;

-- Normalizar valores iniciales (si había NULLs por alguna razón histórica)
UPDATE lots SET version = 1 WHERE version IS NULL;
