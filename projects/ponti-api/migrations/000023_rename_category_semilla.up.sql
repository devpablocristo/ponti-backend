-- 000023_rename_category_semilla.up.sql
UPDATE categories
  SET name = 'Variedad Demo Semilla'
  WHERE name = 'Semilla';
