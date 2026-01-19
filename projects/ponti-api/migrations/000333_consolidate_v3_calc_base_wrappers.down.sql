-- ========================================
-- MIGRACIÓN 000333: Consolidar base v3_calc -> v3_core_ssot (DOWN)
-- ========================================
--
-- Propósito: Revertir wrappers y restaurar definiciones locales en v3_calc.
-- Nota: Comentarios en español, código en inglés.
--
BEGIN;

-- Operaciones matemáticas seguras
CREATE OR REPLACE FUNCTION v3_calc.coalesce0(numeric)
RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT COALESCE($1, 0)
$$;

CREATE OR REPLACE FUNCTION v3_calc.coalesce0(double precision)
RETURNS double precision
LANGUAGE sql IMMUTABLE AS $$
  SELECT COALESCE($1, 0)
$$;

CREATE OR REPLACE FUNCTION v3_calc.safe_div(numeric, numeric)
RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT CASE WHEN $2 IS NULL OR $2 = 0 THEN 0 ELSE $1 / $2 END
$$;

CREATE OR REPLACE FUNCTION v3_calc.safe_div_dp(double precision, double precision)
RETURNS double precision
LANGUAGE sql IMMUTABLE AS $$
  SELECT CASE WHEN $2 IS NULL OR $2 = 0 THEN 0 ELSE $1 / $2 END
$$;

CREATE OR REPLACE FUNCTION v3_calc.percentage(numeric, numeric)
RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT v3_calc.safe_div($1, $2) * 100
$$;

CREATE OR REPLACE FUNCTION v3_calc.percentage_capped(numeric, numeric)
RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT LEAST(v3_calc.safe_div($1, $2) * 100, 100)
$$;

CREATE OR REPLACE FUNCTION v3_calc.percentage_rounded(numeric, numeric)
RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT v3_calc.safe_div($1, $2) * 100
$$;

-- Conversiones por hectárea
CREATE OR REPLACE FUNCTION v3_calc.per_ha(numeric, numeric)
RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT v3_calc.safe_div($1, $2)
$$;

CREATE OR REPLACE FUNCTION v3_calc.per_ha_dp(double precision, double precision)
RETURNS double precision
LANGUAGE sql IMMUTABLE AS $$
  SELECT v3_calc.safe_div_dp($1, $2)
$$;

COMMIT;
