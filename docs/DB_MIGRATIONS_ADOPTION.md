TLDR: Estrategia para adoptar `migrations_v4/` sin romper STG/PROD.

# Adopción de migrations_v4

## Nota importante (no ejecutar ahora)
Estos pasos son **solo para cuando decidas aplicar cambios en remoto**.  
Por pedido actual: **no ejecutar en DEV/STG/PROD** hasta nuevo aviso.

## Objetivo
Adoptar el nuevo set de migraciones manteniendo compatibilidad total con el schema y datos existentes.

## Opción A — Baseline + v4 incrementales (recomendada)
1. Generar snapshot del schema actual en cada ambiente (STG/PROD):
   - `pg_dump --schema-only` del estado real.
2. Crear una migración `migrations_v4/000000_baseline_schema.up.sql` que:
   - Cree **exactamente** el schema actual (snapshot).
   - **No** toque datos.
3. Aplicar baseline en el ambiente existente:
   - `scripts/db/db_adopt_baseline.sh <DB_HOST> <DB_NAME> <DB_USER> <DB_PORT> <DB_SSL_MODE>`
4. Registrar en el control de migraciones la versión `000000` como aplicada.
   - `migrate -path migrations_v4 -database "postgres://<DB_USER>@<DB_HOST>:<DB_PORT>/<DB_NAME>?sslmode=<DB_SSL_MODE>" force 0`
5. Aplicar `migrations_v4` desde `000001` en adelante.
   - `migrate -path migrations_v4 -database "postgres://<DB_USER>@<DB_HOST>:<DB_PORT>/<DB_NAME>?sslmode=<DB_SSL_MODE>" up`
5. Verificar con `scripts/db/db_validate.sh` y `db_schema_diff.sh`.

## Opción B — Transición con detección de estado
1. Crear migraciones de transición que:
   - Detecten columnas/vistas existentes.
   - Ajusten a canonical (por ejemplo, creando aliases de compatibilidad).
2. Aplicar transiciones en STG primero.
3. Validar con `db_validate.sh` y snapshot.
4. Repetir en PROD.

## Compatibilidad de nombres (seeded_area)
- Canonical: `seeded_area_ha`.
- Compatibilidad: aliases `sowed_area_ha` y `sown_area_ha` en vistas clave.
- No se elimina ninguna columna/vista pública usada por consumidores.

## Validación mínima obligatoria
1. `make db-verify` en local.
2. En STG: migrar + `db_validate.sh`.
3. Verificar que `v4_calc` y `v4_report` compilen y consulten sin errores.

## Notas de seguridad
- No hacer `DROP` ni `RENAME` sin capa de compatibilidad.
- El baseline debe ser **exacto** al schema real del ambiente.
