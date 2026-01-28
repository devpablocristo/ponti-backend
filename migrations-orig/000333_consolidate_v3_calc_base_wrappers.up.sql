-- ========================================
-- MIGRACIÓN 000333: Consolidar base v3_calc -> v3_core_ssot (UP)
-- ========================================
--
-- Propósito: Evitar duplicación de funciones base en v3_calc.
-- Enfoque: v3_calc queda como wrapper de v3_core_ssot.
-- Nota: Comentarios en español, código en inglés.
--
BEGIN;

-- Operaciones matemáticas seguras
CREATE OR REPLACE FUNCTION v3_calc.coalesce0(numeric)
RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT v3_core_ssot.coalesce0($1)
$$;

CREATE OR REPLACE FUNCTION v3_calc.coalesce0(double precision)
RETURNS double precision
LANGUAGE sql IMMUTABLE AS $$
  SELECT v3_core_ssot.coalesce0($1)
$$;

CREATE OR REPLACE FUNCTION v3_calc.safe_div(numeric, numeric)
RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT v3_core_ssot.safe_div($1, $2)
$$;

CREATE OR REPLACE FUNCTION v3_calc.safe_div_dp(double precision, double precision)
RETURNS double precision
LANGUAGE sql IMMUTABLE AS $$
  SELECT v3_core_ssot.safe_div_dp($1, $2)
$$;

CREATE OR REPLACE FUNCTION v3_calc.percentage(numeric, numeric)
RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT v3_core_ssot.percentage($1, $2)
$$;

CREATE OR REPLACE FUNCTION v3_calc.percentage_capped(numeric, numeric)
RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT v3_core_ssot.percentage_capped($1, $2)
$$;

CREATE OR REPLACE FUNCTION v3_calc.percentage_rounded(numeric, numeric)
RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT v3_core_ssot.percentage_rounded($1, $2)
$$;

-- Conversiones por hectárea
CREATE OR REPLACE FUNCTION v3_calc.per_ha(numeric, numeric)
RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT v3_core_ssot.per_ha($1, $2)
$$;

CREATE OR REPLACE FUNCTION v3_calc.per_ha_dp(double precision, double precision)
RETURNS double precision
LANGUAGE sql IMMUTABLE AS $$
  SELECT v3_core_ssot.per_ha_dp($1, $2)
$$;

COMMIT;
