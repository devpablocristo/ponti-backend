# notes-for-future-agent.md — feature-003 · be-multitenant-db-hardening

## Resumen corto
Feature **Solo-BE, tipo migration**. Son **4 archivos**: el par up/down de `000224_tenant_security_foundation` y de `000225_tenant_constraints_validation`. 224 agrega `tenant_id` (nullable + backfill + FK NOT VALID) a 32 tablas y siembra roles/permisos por tenant + 3 tablas de seguridad; 225 endurece (`NOT NULL`, valida FK, unicidad de `name` por tenant). No hay código Go ni FE en esta flist.

## Qué está en FE y en BE
- **FE:** NADA. Sin carpeta, sin cambios. Marcar "sin cambios FE" en su cross-repo-map.
- **BE:** las 4 migraciones (este recorte). El CÓDIGO que escribe `tenant_id` NO está acá — vive en **feature-001 (be-platform-tenancy-refactor)**.

## Archivos esenciales / peligrosos / mezclados
- **Esenciales (los 4, whole-file):**
  - `migrations_v4/000224_tenant_security_foundation.up.sql` (248 líneas)
  - `migrations_v4/000224_tenant_security_foundation.down.sql`
  - `migrations_v4/000225_tenant_constraints_validation.up.sql`
  - `migrations_v4/000225_tenant_constraints_validation.down.sql`
- **Peligroso:** la **numeración**. develop ya aplicó 229/230 y tiene hueco 223–228. golang-migrate no aplicará números menores a la versión actual. Ver extraction-plan.md (Camino A/B).
- **Mezclados:** NINGUNO. No hay archivos compartidos (ni wire/cmd/go.mod/Makefile/shared) en esta flist.

## Decisiones ya tomadas
- Extracción = whole-file para los 4 (no hay hunks parciales).
- DECISIÓN de spec: "extraer tal cual, pero arreglar el ordenamiento antes + coordinar con 001".
- No tocar 223/226/227/228 (otras features) salvo que se elija portar el bloque completo en orden.

## Dudas abiertas (para el humano)
1. ¿La DB destino ya está en versión 230 o se resetea? Define si hay que renumerar 224/225 (Camino B) o portar 223–228 en orden (Camino A).
2. ¿feature-001 va antes/junto? 225 lo asume.
3. ¿Hay nombres duplicados activos en alguna de las 32 tablas? Si sí, 225 salta esa tabla (NOTICE + CONTINUE) y queda sin unicidad por tenant.

## Comandos para mirar primero (read-only)
```bash
cat /tmp/flists/be-003.txt
# contenido real:
git -C <repo> show develop-problematico~1:migrations_v4/000224_tenant_security_foundation.up.sql | head -160
git -C <repo> show develop-problematico~1:migrations_v4/000225_tenant_constraints_validation.up.sql
# confirmar ausencia en develop y el hueco:
git -C <repo> ls-tree --name-only 003a9b8f migrations_v4/ | grep -E "00022[2-9]|0023[0-4]"
# confirmar que 229/230 en develop == SOURCE (por eso NO se re-portan):
git -C <repo> diff 003a9b8f 777e5f6a -- migrations_v4/000229_dashboard_active_total_and_lot_yield.up.sql
```
(repo = `/home/pablocristo/Proyectos/pablo/ponti/core`)

## Errores a evitar
- NO usar `develop-problematico` tip (vacío/restore). Usar `develop-problematico~1` (777e5f6a).
- NO mergear 003 sin feature-001: rompe inserts en runtime (23502/23505) o deja el endurecimiento sin efecto.
- NO aplicar 225 sin auditar duplicados de nombre activos (endurecimiento parcial silencioso).
- NO asumir que `migrate up` aplicará 224/225 si la DB ya está en 230 — no lo hará.
- NO re-portar 229/230 (ya están en develop, idénticos).

## Camino más seguro
1. Auditar datos (nulls + duplicados de nombre activos) en la DB destino.
2. Confirmar versión actual de `schema_migrations` -> elegir Camino A (portar bloque 223–228 en orden) o B (renumerar 224/225 > 230, manteniendo 224 < 226/231/234).
3. Portar/mergear feature-001 antes o junto.
4. Traer los 4 archivos enteros (`git checkout develop-problematico~1 -- <paths>`).
5. `make db-verify` en DB efímera; revisar logs por NOTICE.
6. PR a develop.

## Qué PR del otro repo debe ir antes/después
- **FE:** ninguno (Solo-BE).
- **BE intra-repo:** feature-001 ANTES o junto. Las migraciones 226 (feature-004) / 231 / 234 (feature-009/027) van DESPUÉS (dependen de `tenant_id` que crea 224).
