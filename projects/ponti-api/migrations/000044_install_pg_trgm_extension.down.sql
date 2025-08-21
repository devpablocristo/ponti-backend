-- Remover extensión pg_trgm
-- Nota: Esto solo funcionará si ningún objeto depende de la extensión

DROP EXTENSION IF EXISTS pg_trgm;
