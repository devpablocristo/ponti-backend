-- Habilita extensión unaccent para búsquedas sin acentos
BEGIN;
CREATE EXTENSION IF NOT EXISTS unaccent;
COMMIT;
