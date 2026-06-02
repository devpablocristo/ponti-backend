# implementation-status.md — feature-003 · be-multitenant-db-hardening

## Estado global
- **Estado:** COMPLETA a nivel SQL (las dos migraciones están escritas, son coherentes y reversibles) / **PARCIAL a nivel sistema** (el código que escribe `tenant_id` vive en feature-001 y no está en esta flist).
- **% completitud (de lo que corresponde a esta feature, o sea las 4 migraciones):** ~95%. El 5% restante es la **decisión de numeración/ordenamiento** que no se puede cerrar sin coordinación.

## Estado en este repo (BE)
- 4 migraciones presentes en SOURCE (777e5f6a), **ausentes en develop** (hueco 223–228).
- 224 (aditiva): completa, idempotente (`IF NOT EXISTS`, `ON CONFLICT DO NOTHING`, `to_regclass`). 248 líneas, leída completa.
- 225 (estricta): completa; maneja nulls (abort), FK (validate), NOT NULL, y unicidad por tenant con fallback seguro (NOTICE+CONTINUE) ante duplicados. Leída completa.
- Downs: ambas reversibles; 224 down borra tablas de seguridad + seed, 225 down vuelve a fase aditiva. NO borran datos de dominio.

## Estado en el otro repo (FE)
- N/A. Sin cambios FE. No hay carpeta de esta feature en FE.

## Tests
- **Ninguno** en la flist. No hay tests Go ni fixtures asociados a estas migraciones.
- Validación disponible vía herramientas de DB del repo: `make db-verify` (db-reset + db-migrate-up + db-validate + db-schema-snapshot + db-schema-diff). Ver validation.md.

## Pendientes
### BLOQUEANTE-para-mergear
1. **Resolver numeración/orden** vs 229/230 ya aplicadas en develop (Camino A o B de extraction-plan). Sin esto, `migrate up` no aplica 224/225 sobre DB en versión 230.
2. **Coordinar feature-001** (código tenancy). 225 lo asume explícitamente; mergear 225 sin 001 puede romper inserts en runtime (violación de unicidad por tenant) o dejar el endurecimiento sin efecto en la app.
3. **Auditar duplicados de nombre activos** antes de aplicar 225 (si los hay, 225 NO crea la unicidad por tenant para esa tabla y lo hace en silencio con NOTICE).

### Mejora-futura
- Añadir un test/seed que ejercite multi-tenant real (hoy todo cae al tenant `default`).
- Considerar convertir el `RAISE NOTICE + CONTINUE` de 225 en un reporte explícito (las tablas saltadas quedan sin unicidad por tenant; conviene listarlas).

### Deuda-aceptable
- Tablas de seguridad (`tenant_invites`, `security_audit_events`, `auth_session_events`) creadas pero sin consumidor en esta flist. Es deuda esperada: el consumidor llega con 001/018.
- Doble-seed potencial de roles/permisos entre 224 y 001: aceptable por idempotencia (`ON CONFLICT`).

### Duda-humana
- ¿La DB destino (develop) ya está en versión 230 o se resetea? Define Camino A vs B.
- ¿Se porta el bloque 223–228 completo en orden, o solo 224/225 renumerados?

## Bugs / fragilidades observadas
- **No es bug, es diseño:** 225 con duplicados activos salta la tabla (NOTICE + CONTINUE). Silencioso. Riesgo de "endurecimiento parcial" no advertido.
- **Fragilidad de ordenamiento:** el hueco 223–228 + 229/230 ya presentes es un síntoma de extracción/porteo previo desordenado. No es culpa de esta feature, pero la afecta.

## Confianza
- Contenido de migraciones: **ALTA** (leídas completas).
- Estado en develop/SOURCE: **ALTA** (`git ls-tree`, `git diff`).
- Impacto del ordenamiento real sobre la DB productiva: **MEDIA** (depende de la versión actual de `schema_migrations` en cada entorno, no verificable desde el repo).
