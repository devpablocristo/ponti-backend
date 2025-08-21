-- Instalar extensión unaccent para funcionalidad de búsqueda de texto
-- Esta extensión es requerida por el sistema de sugerencia de palabras para búsquedas insensibles a acentos

-- Instalar la extensión unaccent
CREATE EXTENSION IF NOT EXISTS unaccent;

-- Verificar que la extensión fue instalada
SELECT extname, extversion FROM pg_extension WHERE extname = 'unaccent';
