-- Strict tenant validation. Run only after every tenant-owned repository is tenant-scoped
-- and the golden master is green.

DO $$
DECLARE
    default_tenant uuid;
    t text;
    constraint_name text;
    idx_name text;
    null_count bigint;
    duplicate_name_count bigint;
    has_name boolean;
    has_deleted_at boolean;
    has_id boolean;
    tenant_tables text[] := ARRAY[
        'customers',
        'projects',
        'campaigns',
        'fields',
        'lots',
        'lot_dates',
        'workorders',
        'workorder_items',
        'workorder_investor_splits',
        'workorder_supply_items',
        'work_order_drafts',
        'project_managers',
        'project_investors',
        'admin_cost_investors',
        'field_investors',
        'labors',
        'supplies',
        'supply_movements',
        'stock_movements',
        'stocks',
        'invoices',
        'investors',
        'managers',
        'providers',
        'crops',
        'categories',
        'class_types',
        'lease_types',
        'business_parameters',
        'crop_commercializations',
        'project_dollar_values'
    ];
BEGIN
    SELECT id INTO default_tenant
    FROM public.auth_tenants
    WHERE name = 'default'
    ORDER BY created_at NULLS LAST
    LIMIT 1;

    IF default_tenant IS NULL THEN
        INSERT INTO public.auth_tenants (name, created_at, updated_at)
        VALUES ('default', now(), now())
        RETURNING id INTO default_tenant;
    END IF;

    FOREACH t IN ARRAY tenant_tables LOOP
        IF to_regclass('public.' || t) IS NOT NULL
           AND EXISTS (
                SELECT 1
                FROM information_schema.columns
                WHERE table_schema = 'public' AND table_name = t AND column_name = 'tenant_id'
           )
        THEN
            EXECUTE format('UPDATE public.%I SET tenant_id = $1 WHERE tenant_id IS NULL', t) USING default_tenant;
            EXECUTE format('SELECT COUNT(*) FROM public.%I WHERE tenant_id IS NULL', t) INTO null_count;
            IF null_count > 0 THEN
                RAISE EXCEPTION 'tenant strict validation failed: %.tenant_id has % null rows', t, null_count;
            END IF;

            BEGIN
                EXECUTE format('ALTER TABLE public.%I VALIDATE CONSTRAINT %I', t, t || '_tenant_id_fkey');
            EXCEPTION WHEN undefined_object THEN
                NULL;
            END;

            EXECUTE format('ALTER TABLE public.%I ALTER COLUMN tenant_id SET NOT NULL', t);
            SELECT EXISTS (
                SELECT 1 FROM information_schema.columns
                WHERE table_schema = 'public' AND table_name = t AND column_name = 'id'
            ) INTO has_id;
            IF has_id THEN
                EXECUTE format('CREATE INDEX IF NOT EXISTS idx_%I_tenant_id_id ON public.%I (tenant_id, id)', t, t);
            END IF;

            SELECT EXISTS (
                SELECT 1 FROM information_schema.columns
                WHERE table_schema = 'public' AND table_name = t AND column_name = 'name'
            ) INTO has_name;
            SELECT EXISTS (
                SELECT 1 FROM information_schema.columns
                WHERE table_schema = 'public' AND table_name = t AND column_name = 'deleted_at'
            ) INTO has_deleted_at;

            IF has_name THEN
                FOR constraint_name IN
                    SELECT c.conname
                    FROM pg_constraint c
                    JOIN pg_class rel ON rel.oid = c.conrelid
                    JOIN pg_namespace nsp ON nsp.oid = rel.relnamespace
                    WHERE nsp.nspname = 'public'
                      AND rel.relname = t
                      AND c.contype = 'u'
                      AND (
                          SELECT array_agg(att.attname ORDER BY ord.ordinality)
                          FROM unnest(c.conkey) WITH ORDINALITY AS ord(attnum, ordinality)
                          JOIN pg_attribute att ON att.attrelid = rel.oid AND att.attnum = ord.attnum
                      ) = ARRAY['name']::name[]
                LOOP
                    EXECUTE format('ALTER TABLE public.%I DROP CONSTRAINT IF EXISTS %I', t, constraint_name);
                END LOOP;

                FOR idx_name IN
                    SELECT idx.relname
                    FROM pg_index i
                    JOIN pg_class tbl ON tbl.oid = i.indrelid
                    JOIN pg_namespace nsp ON nsp.oid = tbl.relnamespace
                    JOIN pg_class idx ON idx.oid = i.indexrelid
                    WHERE nsp.nspname = 'public'
                      AND tbl.relname = t
                      AND i.indisunique
                      AND NOT i.indisprimary
                      AND (
                          SELECT array_agg(att.attname ORDER BY ord.ordinality)
                          FROM unnest(i.indkey) WITH ORDINALITY AS ord(attnum, ordinality)
                          JOIN pg_attribute att ON att.attrelid = tbl.oid AND att.attnum = ord.attnum
                      ) = ARRAY['name']::name[]
                LOOP
                    EXECUTE format('DROP INDEX IF EXISTS public.%I', idx_name);
                END LOOP;

                IF has_deleted_at THEN
                    EXECUTE format(
                        'SELECT COUNT(*) FROM (
                            SELECT tenant_id, lower(btrim(name)) AS normalized_name
                            FROM public.%I
                            WHERE deleted_at IS NULL
                            GROUP BY tenant_id, lower(btrim(name))
                            HAVING COUNT(*) > 1
                        ) duplicates',
                        t
                    ) INTO duplicate_name_count;

                    IF duplicate_name_count > 0 THEN
                        RAISE NOTICE 'tenant strict validation skipped tenant/name unique index for %. % duplicate active name groups found', t, duplicate_name_count;
                        CONTINUE;
                    END IF;

                    EXECUTE format(
                        'CREATE UNIQUE INDEX IF NOT EXISTS uq_%I_tenant_name_active ON public.%I (tenant_id, lower(btrim(name))) WHERE deleted_at IS NULL',
                        t,
                        t
                    );
                ELSE
                    EXECUTE format(
                        'SELECT COUNT(*) FROM (
                            SELECT tenant_id, lower(btrim(name)) AS normalized_name
                            FROM public.%I
                            GROUP BY tenant_id, lower(btrim(name))
                            HAVING COUNT(*) > 1
                        ) duplicates',
                        t
                    ) INTO duplicate_name_count;

                    IF duplicate_name_count > 0 THEN
                        RAISE NOTICE 'tenant strict validation skipped tenant/name unique index for %. % duplicate name groups found', t, duplicate_name_count;
                        CONTINUE;
                    END IF;

                    EXECUTE format(
                        'CREATE UNIQUE INDEX IF NOT EXISTS uq_%I_tenant_name ON public.%I (tenant_id, lower(btrim(name)))',
                        t,
                        t
                    );
                END IF;
            END IF;
        END IF;
    END LOOP;
END $$;
