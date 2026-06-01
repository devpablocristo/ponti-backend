# validation.md — feature-003 · be-multitenant-db-hardening

## Checklist pre-PR
- [ ] `cat /tmp/flists/be-003.txt` -> exactamente 4 líneas (224/225 up+down). Sin `.go`, sin FE.
- [ ] Los 4 archivos provienen de `develop-problematico~1` (SHA 777e5f6a) sin modificaciones de contenido (comparar con `git show`).
- [ ] `git diff --check` limpio (sin trailing whitespace).
- [ ] Decidido Camino A o B de numeración (extraction-plan.md). Si Camino B, los 4 archivos renumerados con `git mv` y el orden 224<226/231/234 preservado.
- [ ] feature-001 (código tenancy) mergeada o en el mismo tren.
- [ ] Auditoría de datos hecha (nulls + duplicados de nombre) — queries abajo.

## Validación de DB (local, sobre DB efímera con datos representativos)
Herramientas del repo (`make`, ver `scripts/db/`):
```bash
make db-verify       # db-reset + db-migrate-up + db-validate + db-schema-snapshot + db-schema-diff
# o por partes:
make db-reset
make db-migrate-up   # corre migrate/migrate:v4.17.1 up sobre migrations_v4/
make db-validate
make db-schema-snapshot
make db-schema-diff
```
- [ ] `make db-migrate-up` aplica 224 luego 225 SIN abortar (`RAISE EXCEPTION`).
- [ ] Revisar la salida por `RAISE NOTICE 'tenant strict validation skipped ...'` -> si aparece, alguna tabla quedó SIN unicidad por tenant (duplicados activos). Anotar cuáles.

## Queries de auditoría (correr ANTES de 225)
```sql
-- 1) nulls de tenant_id por tabla (ejemplo customers; repetir por las 32)
SELECT 'customers' AS t, count(*) FROM public.customers WHERE tenant_id IS NULL;

-- 2) duplicados de nombre ACTIVOS por tenant (los que harían saltar 225)
SELECT tenant_id, lower(btrim(name)) AS n, count(*)
FROM public.customers
WHERE deleted_at IS NULL
GROUP BY 1,2 HAVING count(*) > 1;
-- repetir por cada tabla con `name` + `deleted_at`
```

## Verificación post-migración
```sql
-- a) tenant_id NOT NULL en las 32 tablas presentes
SELECT table_name, is_nullable
FROM information_schema.columns
WHERE table_schema='public' AND column_name='tenant_id'
ORDER BY table_name;   -- todas deben ser NO

-- b) FKs validadas (no quedan NOT VALID)
SELECT conrelid::regclass, conname, convalidated
FROM pg_constraint WHERE conname LIKE '%_tenant_id_fkey' AND NOT convalidated;
-- esperado: 0 filas

-- c) índices únicos por tenant creados
SELECT indexname FROM pg_indexes
WHERE schemaname='public' AND (indexname LIKE 'uq_%_tenant_name%');

-- d) roles y tablas de seguridad
SELECT name FROM public.auth_roles
WHERE name IN ('saas_superadmin','tenant_owner','tenant_admin','tenant_manager','tenant_viewer');
SELECT to_regclass('public.tenant_invites'),
       to_regclass('public.security_audit_events'),
       to_regclass('public.auth_session_events');  -- ninguno NULL
```

## Validación de reversibilidad (down)
```bash
# en DB efímera, tras up:
# migrate down 1  -> aplica 000225 down: dropea uq_*, tenant_id vuelve nullable. Sin pérdida de datos.
# migrate down 1  -> aplica 000224 down: dropea tablas seguridad + seed, columnas/FK/idx tenant_id.
```
- [ ] Down de 225 deja `tenant_id` nullable y sin `uq_*`, datos de dominio intactos.
- [ ] Down de 224 deja el esquema como antes (sin columna tenant_id ni tablas de seguridad).

## Tests sugeridos
- **BE (Go):** ninguno en la flist. No hay paquete que testear directamente por esta feature. Cuando vaya con feature-001, correr `go test ./internal/...` de los repos tenant-scoped.
- **FE:** N/A (sin cambios).

## Casos borde a probar
- Tabla del array que NO existe en el entorno -> 224/225 la saltan (`to_regclass IS NULL`). Verificar que no aborta.
- Tabla con `name` pero SIN `deleted_at` -> 225 usa `uq_<t>_tenant_name` (no parcial). Comprobar rama.
- Re-ejecución (idempotencia): correr 224 dos veces no debe duplicar columnas/índices/roles (gracias a `IF NOT EXISTS` / `ON CONFLICT`).
- Aplicar sobre DB ya en versión 230 -> confirmar que migrate NO aplica 224/225 (esto VALIDA el riesgo de ordenamiento; si pasa, elegir Camino A/B).

## Qué revisar en UI/API/DB/env
- **UI:** nada.
- **API:** nada en esta flist. (Con 001: confirmar que crear customer/project/etc. setea `tenant_id` y respeta unicidad por tenant.)
- **DB:** queries de verificación arriba.
- **env:** la migración crea `pgcrypto` -> el rol de DB necesita permiso `CREATE EXTENSION`.

## Qué validar en el otro repo
- Nada (Solo-BE). En el cross-repo-map del FE: "feature-003 sin cambios FE".

## Señales de incompletitud / incompatibilidad
- `migrate up` no avanza (DB en versión > 225) -> problema de ordenamiento sin resolver.
- NOTICE de "skipped tenant/name unique index" -> endurecimiento parcial; faltan limpiar duplicados.
- Errores 23502 (not-null) / 23505 (unique) en runtime tras mergear -> feature-001 ausente o desincronizada.
- FK con `convalidated = false` tras 225 -> VALIDATE falló silenciosamente (revisar bloque `EXCEPTION WHEN undefined_object`).
