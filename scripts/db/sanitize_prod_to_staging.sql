-- Sanitizacion basica PROD -> STAGING.
-- Objetivo: mantener consistencia referencial pero remover PII.
--
-- Nota: este script esta pensado para ejecutarse en STAGING DESPUES de importar un dump de PROD.
-- Evitar dependencias en extensiones; usar SQL/PLpgSQL basico.

BEGIN;

-- 1) Users / AuthN
DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM information_schema.tables
    WHERE table_schema = 'public' AND table_name = 'users'
  ) THEN
    -- Columnas legacy: email/username/password/token_hash/refresh_tokens
    -- Columnas nuevas: idp_email/idp_sub (Identity Platform)
    EXECUTE $q$
      UPDATE public.users
      SET
        email = 'user_' || id::text || '@example.invalid',
        username = 'user_' || id::text,
        password = md5(random()::text),
        token_hash = md5(random()::text),
        refresh_tokens = ARRAY[]::text[],
        is_verified = FALSE,
        active = TRUE
    $q$;

    IF EXISTS (
      SELECT 1 FROM information_schema.columns
      WHERE table_schema = 'public' AND table_name = 'users' AND column_name = 'idp_email'
    ) THEN
      EXECUTE $q$
        UPDATE public.users
        SET idp_email = 'user_' || id::text || '@example.invalid'
        WHERE idp_email IS NOT NULL
      $q$;
    END IF;

    IF EXISTS (
      SELECT 1 FROM information_schema.columns
      WHERE table_schema = 'public' AND table_name = 'users' AND column_name = 'idp_sub'
    ) THEN
      -- Mantener unicidad pero no exponer sub real.
      EXECUTE $q$
        UPDATE public.users
        SET idp_sub = 'san_' || id::text || '_' || md5(random()::text)
        WHERE idp_sub IS NOT NULL
      $q$;
    END IF;
  END IF;
END
$$;

-- 2) Nombres de entidades (no PII directo pero suelen identificar)
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema='public' AND table_name='customers') THEN
    EXECUTE 'UPDATE public.customers SET name = ''Customer '' || id::text WHERE name IS NOT NULL';
  END IF;
  IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema='public' AND table_name='investors') THEN
    EXECUTE 'UPDATE public.investors SET name = ''Investor '' || id::text WHERE name IS NOT NULL';
  END IF;
  IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema='public' AND table_name='providers') THEN
    EXECUTE 'UPDATE public.providers SET name = ''Provider '' || id::text WHERE name IS NOT NULL';
  END IF;
  IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema='public' AND table_name='managers') THEN
    EXECUTE 'UPDATE public.managers SET name = ''Manager '' || id::text WHERE name IS NOT NULL';
  END IF;
END
$$;

-- 3) Documentos / textos libres (si existieran en el schema)
-- Intentamos sanitizar columnas tipicas si existen, sin fallar si no estan.
DO $$
DECLARE
  r record;
BEGIN
  FOR r IN
    SELECT table_schema, table_name, column_name
    FROM information_schema.columns
    WHERE table_schema = 'public'
      AND column_name IN ('notes', 'note', 'observations', 'observation', 'comments', 'comment', 'address', 'direccion', 'phone', 'telefono', 'cuit', 'dni')
  LOOP
    EXECUTE format('UPDATE %I.%I SET %I = NULL WHERE %I IS NOT NULL', r.table_schema, r.table_name, r.column_name, r.column_name);
  END LOOP;
END
$$;

COMMIT;

