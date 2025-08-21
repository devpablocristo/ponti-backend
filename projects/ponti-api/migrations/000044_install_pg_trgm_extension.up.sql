-- Instalar extensión pg_trgm para operaciones de similitud de trigramas
-- Esta extensión es requerida por el sistema de sugerencia de palabras para búsquedas de similitud de texto

-- Instalar la extensión pg_trgm
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Verificar que la extensión fue instalada
SELECT extname, extversion FROM pg_extension WHERE extname = 'pg_trgm';
