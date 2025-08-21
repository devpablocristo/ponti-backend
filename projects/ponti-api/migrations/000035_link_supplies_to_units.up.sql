-- Seeds idempotentes
INSERT INTO units (name) VALUES ('Lts') ON CONFLICT (name) DO NOTHING;
INSERT INTO units (name) VALUES ('Kg')  ON CONFLICT (name) DO NOTHING;
INSERT INTO units (name) VALUES ('Ha')  ON CONFLICT (name) DO NOTHING;

-- Si había una FK anterior en supplies.unit_id, la removemos primero
DO $$
BEGIN
  IF EXISTS (
    SELECT 1
    FROM information_schema.table_constraints
    WHERE table_name = 'supplies'
      AND constraint_type = 'FOREIGN KEY'
      AND constraint_name = 'supplies_unit_id_fkey'
  ) THEN
    ALTER TABLE supplies DROP CONSTRAINT supplies_unit_id_fkey;
  END IF;
END$$;

-- Crear FK hacia units (la columna unit_id ya existe y es NULLABLE para compatibilidad)
ALTER TABLE supplies
  ADD CONSTRAINT supplies_unit_id_fkey
  FOREIGN KEY (unit_id) REFERENCES units(id)
  ON UPDATE CASCADE ON DELETE RESTRICT;

CREATE INDEX IF NOT EXISTS idx_supplies_unit_id ON supplies(unit_id);
