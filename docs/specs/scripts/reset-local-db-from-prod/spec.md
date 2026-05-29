# Spec — `scripts/db/reset-local-db-from-prod.sh`

> Resetea la base de datos **local** y la rellena con los **datos reales de PROD**
> (data-only), dejando un entorno local funcional y verificado contra producción.

/ Comando: `make reset-local-db-from-prod`

---

## 1. Propósito

Dejar la DB local idéntica en **datos** a PROD, sin clonar el esquema de prod y sin
tocar prod más que para leer. Útil para reproducir bugs/datos reales en local.

- **Origen (PROD):** `new_ponti_db_prod` — **read-only** (`pg_dump` + `SELECT`/introspección).
- **Destino (local):** la DB de `.env` (debe ser local; hay guardas que lo fuerzan).
- **Tratamiento:** `data-only` (sin schema, sin renames). El esquema local lo crean las **migraciones**, no el dump.

---

## 2. Flujo (orden real, ver `### Main`)

1. `validate_destination_safety` — guardas de seguridad del destino (ver §4).
2. `validate_required_tools` — `python3 psql pg_dump pg_restore pg_isready docker`.
3. `resolve_source_connection` — resuelve host de PROD (gcloud o Cloud SQL Proxy).
4. `validate_destination_connection`.
5. **Captura conteos críticos de PROD** (fuente de verdad para la verificación final).
6. *(si `DRY_RUN=1` → imprime plan y termina, sin destruir nada).*
7. `generate_or_reuse_dump` — `pg_dump` data-only de PROD → `DUMP_FILE` (o reusa si `SKIP_DUMP=1`).
8. `validate_dump_manifest` — confirma que el dump trae las tablas críticas.
9. `reset_local_database` — `DROP DATABASE` + `CREATE DATABASE` (vacía).
10. `run_local_migrations_to_target` — **aplica migraciones** (`migrate up` por default; ver §5).
11. `detect_trigger_strategy` / `truncate_public_tables`.
12. `restore_data_only` — `pg_restore --data-only --schema=public`, omitiendo `schema_migrations`
    y cualquier entrada del dump que **no exista** en el esquema local migrado.
13. `run_post_restore_steps` — backfills post-restore, **best-effort** (ver §6).
14. `sync_sequences` — `setval` de cada secuencia al `MAX(id)` real.
15. `compare_critical_counts` — **PROD vs local deben coincidir**; si no, falla.

---

## 3. Variables de entorno

### Destino local (desde `.env`, obligatorias)
| Var | Default | Notas |
|---|---|---|
| `DB_USER` `DB_PASSWORD` `DB_NAME` `DB_PORT` | — | requeridas |
| `DB_HOST` | `127.0.0.1` | **debe** ser `localhost`/`127.0.0.1` |
| `DB_SSL_MODE` | `disable` | |

### Origen PROD (read-only)
| Var | Default |
|---|---|
| `SRC_USER` | `DB_USER_PROD` / `soalen-db-v3` |
| `SRC_PASS` | env, o Secret Manager (`db-password-dev` @ `new-ponti-dev`) vía gcloud |
| `SRC_DB` | `DB_NAME_PROD` / `new_ponti_db_prod` |
| `SRC_HOST` `SRC_PORT` `SRC_SSL` | inferido por gcloud / `5432` / `disable` |
| `USE_CLOUDSQL_PROXY` | `auto` (`auto`\|`1`\|`0`) |

### Control de comportamiento
| Var | Default | Qué hace |
|---|---|---|
| `MIGRATE_TARGET_VERSION` | *(vacío)* | vacío ⇒ `migrate up` (TODAS). Si se setea, hace `goto <ver>` (la versión **debe existir** en `migrations_v4`). |
| `DRY_RUN` | `0` | `1` = solo imprime el plan, no destruye nada. |
| `SKIP_DUMP` | `0` | `1` = reusa `DUMP_FILE` existente (no toca PROD). |
| `DUMP_FILE` | `/tmp/prod_to_local_<ts>.dump` | |
| `RESTORE_MODE` | `data-only` | único soportado. |
| `RESET_LOCAL_DB` | `1` | requerido en `1`. |
| `TRUNCATE_BEFORE_RESTORE` | `1` | |
| `DISABLE_TRIGGERS` | `1` | requiere superuser; se auto-desactiva si no lo es. |
| `POST_RESTORE_TENANT_BACKFILL` | `1` | best-effort (ver §6). |
| `ACTORS_BACKFILL_SYNC` | `1` | best-effort (ver §6). |
| `RUN_FINAL_MIGRATIONS` | `1` | `migrate up` final (no-op si ya se aplicaron todas). |

---

## 4. Guardas de seguridad (no se puede apuntar a prod por error)

- `DB_HOST` debe ser `localhost`/`127.0.0.1`.
- `DB_NAME` solo `[A-Za-z0-9_]` y **bloqueado** si contiene `production|prod|prd|live`.
- Solo `RESTORE_MODE=data-only` y `RESET_LOCAL_DB=1`.
- PROD se usa **solo lectura** (`pg_dump` + `SELECT`, `default_transaction_read_only=on`).

---

## 5. Migraciones (clave)

El dump es **data-only**: antes de cargar datos hay que **crear el esquema** aplicando
migraciones de `migrations_v4`.

- Por **default** (`MIGRATE_TARGET_VERSION` vacío) hace **`migrate up`** = aplica todas las
  migraciones disponibles del set. **No fija un número** — así funciona con cualquier set.
- Si se necesita un punto intermedio (p. ej. esquema legacy compatible con datos viejos de
  prod), se puede pasar `MIGRATE_TARGET_VERSION=<n>` **siempre que esa versión exista** como
  archivo en `migrations_v4`. **Cuidado:** la numeración cambió entre versiones del repo
  (un mismo número no significa lo mismo, y algunos números no existen).

---

## 6. Pasos post-restore (best-effort)

Son normalizaciones **opcionales** que dependen de features que pueden no existir en todos
los esquemas. **No abortan el restore**: si fallan o no aplican, avisan con `WARN` y siguen.

- **Tenant backfill** (`POST_RESTORE_TENANT_BACKFILL=1`): corre
  `migrations_v4/000224_tenant_security_foundation.up.sql`. Se **saltea** si el archivo no
  existe en el set actual.
- **Final migrations** (`RUN_FINAL_MIGRATIONS=1`): `migrate up` final.
- **Actors backfill** (`ACTORS_BACKFILL_SYNC=1`): corre `scripts/db/actors_backfill_sync.sql`.
  Best-effort: si el esquema no tiene la tabla `actors`, avisa y continúa.

---

## 7. Uso

```bash
# Normal (dump fresco de PROD + reset + datos):
make reset-local-db-from-prod

# Ver el plan sin destruir nada:
DRY_RUN=1 ./scripts/db/reset-local-db-from-prod.sh

# Reusar un dump ya bajado (no toca PROD):
SKIP_DUMP=1 DUMP_FILE=/tmp/prod_to_local_XXXX.dump make reset-local-db-from-prod

# Forzar un target de migración puntual (debe existir en migrations_v4):
MIGRATE_TARGET_VERSION=224 make reset-local-db-from-prod
```

---

## 8. Caveats

- El dump es **data-only**: el esquema sale de las migraciones locales, no de prod.
- Si una tabla de prod **no existe** en el esquema local migrado, sus datos se **omiten**
  (se loguea como `WARN ... Entradas omitidas`). Ej. típico: `user_logins`, `schema_migrations`.
- La carga es **verificada**: al final compara conteos de tablas críticas PROD vs local y
  **falla** si no coinciden.
- PROD (`new_ponti_db_prod`) corre el código vigente en prod; sus datos encajan en el esquema
  que produzcan las migraciones de **esa misma versión** del código. Si se corre con código de
  otra época, puede haber columnas/tablas que no matcheen (de ahí los `WARN` y los best-effort).

---

## 9. Decisiones de diseño / correcciones

- **`run_pg_cmd` propaga el exit code real.** Antes tomaba `$?` después de un `if` (donde ya
  valía 0) y **siempre retornaba 0**, tragando errores de `psql`/`pg_restore`. Corregido para
  que los fallos sean visibles.
- **`MIGRATE_TARGET_VERSION` por default vacío (`migrate up`)** en vez de un número fijo, que
  podía no existir en el set de migraciones y rompía el flujo.
- **Backfills post-restore best-effort:** se saltean/avisan en vez de abortar cuando el archivo
  o la tabla destino no existen. El paso crítico (datos + verificación de conteos) manda.

---

## 10. Dependencias

`python3`, `psql`, `pg_dump`, `pg_restore`, `pg_isready`, `docker` (para `migrate/migrate` y,
si aplica, Cloud SQL Proxy), y acceso de lectura a PROD (gcloud / Secret Manager o `SRC_PASS`).
