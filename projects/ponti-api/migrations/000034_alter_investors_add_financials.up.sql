DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_name='investors' AND column_name='contributions'
  ) THEN
    ALTER TABLE investors ADD COLUMN contributions NUMERIC(10,2) NOT NULL DEFAULT 0;
  END IF;

  IF NOT EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_name='investors' AND column_name='contribution_date'
  ) THEN
    ALTER TABLE investors ADD COLUMN contribution_date TIMESTAMP NULL;
  END IF;

  IF NOT EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_name='investors' AND column_name='percentage'
  ) THEN
    ALTER TABLE investors ADD COLUMN percentage INTEGER NOT NULL DEFAULT 0;
  END IF;
END $$; 