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

DROP INDEX IF EXISTS idx_supplies_unit_id;
-- Nota: no eliminamos filas de units para evitar violaciones de FK si ya hay datos.
