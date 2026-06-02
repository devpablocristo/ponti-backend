# spec — multitenant-db-hardening (feature-003)

> **Spec definitivo** re-baselineado contra `develop` real (tip `19b96dc4`). Fuente: `777e5f6a`
> (= `develop-problematico~1`). NO se implementa acá; es el contrato de qué traer y cómo.

- **id / slug:** feature-003 · `multitenant-db-hardening`
- **tipo:** migration (4 archivos SQL; **sin código Go ni FE**)
- **fuente:** `777e5f6a` · **destino:** `develop` (`19b96dc4`)
- **orden en el cluster tenants/users:** **2º** (después de 001 código; antes de 007/008)

---

## 1. Propósito

Llevar el aislamiento multi-tenant a la DB: `tenant_id` obligatorio (FK + índices) en ~32 tablas de
dominio, modelo de roles/permisos por tenant, y tablas de seguridad — en dos migraciones por fases (224
aditiva/reversible, 225 estricta).

## 2. Estado vs `develop` (diff real re-baselineado)

- `git ls-tree develop -- migrations_v4/ | grep 00022[45]` → **vacío**: 224/225 **no existen** en develop.
- Las migraciones que traen `tenant_id` y todo el resto del bloque **faltan por completo**. La feature
  está al 0% en develop.
- **Lo que cambió desde el baseline de los docs (003a9b8f → 19b96dc4):** develop ya **no** está solo en
  229/230 — ahora llega a **231**. Nombres reales verificados:

  | nº | develop (`19b96dc4`) | source (`777e5f6a`) | ¿colisión? |
  |----|----------------------|----------------------|-----------|
  | 229 | `dashboard_active_total_and_lot_yield` | `dashboard_active_total_and_lot_yield` | iguales ✓ (no re-portear) |
  | 230 | `workorders_is_digital_origin` | `workorders_is_digital_origin` | iguales ✓ (no re-portear) |
  | **231** | **`lot_yield_over_total_hectares`** | **`consolidate_actor_archived_at`** | **❌ COLISIÓN DURA** |

  develop ganó su `000231_lot_yield_over_total_hectares` en #125/#127 (extracto/excel-lotes), **posterior** a
  los docs. El source usa el slot 231 para `consolidate_actor_archived_at` (de 007/027). ⇒ El "Camino A" de
  los docs (portar 223–228 antes de pasar 222) está **muerto**: develop ya está en 231.

## 3. Alcance / archivos

**4 archivos, whole-file, sin partial-hunks, sin compartidos** (`git checkout 777e5f6a -- <path>`):

```
migrations_v4/000224_tenant_security_foundation.up.sql      (248 líneas)
migrations_v4/000224_tenant_security_foundation.down.sql
migrations_v4/000225_tenant_constraints_validation.up.sql
migrations_v4/000225_tenant_constraints_validation.down.sql
```

Qué hace cada una:

- **224 (aditiva, reversible, idempotente):** `CREATE EXTENSION pgcrypto`; resuelve/crea tenant `default`;
  sobre 32 tablas tenant-owned agrega `tenant_id uuid` nullable + backfill al `default` + índices
  (`idx_<t>_tenant_id`, `_tenant_id_id`, `_tenant_name`) + FK a `auth_tenants(id) ON DELETE RESTRICT`
  **NOT VALID**; re-deriva `tenant_id` de tablas hijas desde el padre; siembra roles
  (`saas_superadmin`/`tenant_owner`/`tenant_admin`/`tenant_manager`/`tenant_viewer`) + permisos por recurso
  (`ON CONFLICT DO NOTHING`, mapeando a roles base `admin/manager/viewer` de migr 180); crea
  `tenant_invites`, `security_audit_events`, `auth_session_events`.
- **225 (estricta):** backfill final + abort (`RAISE EXCEPTION`) si quedan nulls; `VALIDATE CONSTRAINT` de
  la FK; `SET NOT NULL`; reemplaza unicidad global de `name` por unicidad por tenant
  (`uq_<t>_tenant_name_active` parcial si hay `deleted_at`, si no `uq_<t>_tenant_name`). **Si hay duplicados
  activos → `RAISE NOTICE` + `CONTINUE`** (salta la tabla, **no** falla → endurecimiento parcial silencioso).
- **downs:** 224 down dropea tablas de seguridad + seed + columnas/FK/idx; 225 down vuelve a fase aditiva
  (dropea `uq_*`, `tenant_id` nullable). **Ninguno borra datos de dominio.**

### EXCLUIR (no son 003)
- 223 (`actors_safe_migration`), 226 (`customer_actor_master_link`), 227/228 (crudar), 231–234: pertenecen a
  007/004/009/027. Aquí solo 224/225.
- Cualquier `.go` (eso es 001). El consumidor de `tenant_invites`/etc. está en 008/018.

## 4. Migraciones — riesgo de numeración (BLOQUEANTE, re-baselineado)

**El "Camino A" (portar el bloque 223–228 antes de pasar 222) ya no es viable** porque develop está en 231.
Queda **Camino B (renumerar) obligatorio**, y ahora más complejo que en los docs por la colisión de 231:

- `golang-migrate` (`migrate/migrate:v4.17.1`, `scripts/db/db_migrate_up.sh`) trackea **una versión entera**.
  Sobre una DB en versión 231, `migrate up` **no aplica** 224/225 (números < 231) → quedarían muertas.
- El slot **231 ya está ocupado** por develop con otro contenido → no se puede traer el source tal cual.
- **Decisión transversal (afecta 003, 007, 004, 009, 027):** renumerar **todo el bloque del source**
  (223–228 + 231–234; 229/230 ya están y son idénticas, no se tocan) a números **> 231** (máximo actual de
  develop), **preservando el orden relativo** `223 < 224 < 225 < 226 < 227 < 228 < 231 < 232 < 233 < 234`.
  Mapeo sugerido (ajustar al cerrar la decisión global, p.ej. en el spec de 007):

  | source | feature | → nuevo nº sugerido |
  |--------|---------|---------------------|
  | 223 actors_safe_migration | 007 | 232 |
  | **224 tenant_security_foundation** | **003** | **233** |
  | **225 tenant_constraints_validation** | **003** | **234** |
  | 226 customer_actor_master_link | 004/009 | 235 |
  | 227 crudar_archive_batches | 009 | 236 |
  | 228 crudar_remaining_archive_metadata | 009 | 237 |
  | 231 consolidate_actor_archived_at | 007/027 | 238 |
  | 232 document_archived_inclusion_in_reports | 027 | 239 |
  | 233 archived_invariant_triggers | 027 | 240 |
  | 234 actor_unique_normalized_name | 007/009 | 241 |

  **Invariante para 003:** sus dos migr deben quedar **después** del máximo de develop **y antes** de
  226/231/234 renumeradas (que dependen de `tenant_id` que crea 224). El mapeo de arriba lo cumple
  (233/234 < 235/238/241).

> Renumerar = `git mv` de los 4 archivos al nuevo nº. golang-migrate descubre por nombre de archivo (no hay
> índice que editar).

## 5. Dependencias

- **Prerequisito ya en develop:** migr **180** (`authn_authz_mvp`: `auth_tenants`/`auth_roles`/
  `auth_permissions`/`auth_role_permissions` + roles base) y las tablas de dominio → **presentes**.
- **001 (código tenancy) — FUERTE:** el header de 225 dice literalmente "run only after every tenant-owned
  repository is tenant-scoped". Sin 001, mergear 225 puede romper inserts en runtime (23502 not-null /
  23505 unique-por-tenant) o dejar el endurecimiento sin efecto. **Mergear 003 junto/después de 001.**
- **003 desbloquea:** 226 (004/009), 231/234 (007/027) y todo repo Go que lea/escriba `tenant_id`.
- **Cross-repo (FE):** ninguno.

## 6. Plan de implementación (pasos, sin ejecutar acá)

1. **Decidir la renumeración global** (§4) — idealmente fijada al recuperar 007 (dueña de la mayor parte del
   bloque). Para 003: 224→233, 225→234 (o el nº que resulte).
2. Confirmar 001 en el mismo tren.
3. `git checkout 777e5f6a -- <los 4 archivos>` y `git mv` a la nueva numeración.
4. **Auditar datos antes de 225** (queries §7): nulls de `tenant_id` y duplicados de `name` activos por
   tenant. Si hay duplicados, 225 los salta en silencio.
5. `make db-verify` en DB efímera; revisar logs por `RAISE NOTICE` de tablas saltadas.
6. PR a develop.

## 7. Validación

```bash
make db-verify   # db-reset + db-migrate-up + db-validate + db-schema-snapshot + db-schema-diff
```
- 224 luego 225 aplican sin `RAISE EXCEPTION`.
- Sin `RAISE NOTICE 'tenant strict validation skipped ...'` (si aparece → tabla sin unicidad por tenant).
- Post: las 32 tablas con `tenant_id NOT NULL`; FKs `convalidated=true`; índices `uq_*_tenant_name*`
  creados; roles `saas_superadmin`/`tenant_*` y tablas `tenant_invites`/`security_audit_events`/
  `auth_session_events` presentes.
- Reversibilidad: down 225 → nullable + sin `uq_*`; down 224 → esquema previo. Sin pérdida de datos.

**Auditoría previa (antes de 225), repetir por cada tabla con `name`+`deleted_at`:**
```sql
SELECT count(*) FROM public.customers WHERE tenant_id IS NULL;                          -- nulls
SELECT tenant_id, lower(btrim(name)) n, count(*) FROM public.customers
  WHERE deleted_at IS NULL GROUP BY 1,2 HAVING count(*)>1;                              -- duplicados activos
```

## 8. Riesgos y decisiones pendientes

- **Numeración (ALTO, bloqueante):** colisión dura en 231 (develop=`lot_yield` vs source=`consolidate_actor
  _archived_at`). Renumerar todo el bloque a >231 (§4). Es decisión de orquestación que excede 003.
- **003-sin-001 (ALTO):** mergear el endurecimiento sin el código que setea `tenant_id` rompe runtime. El
  "merge solo este repo" problemático no es FE-vs-BE, es 003-sin-001.
- **Backfill al tenant `default` (MEDIO):** toda fila con `tenant_id IS NULL` cae al `default`; si ya
  conviven datos de varios tenants reales se colapsan. Auditar antes (supuesto razonable: un solo tenant hoy).
- **Endurecimiento parcial silencioso (MEDIO):** 225 ante duplicados activos hace NOTICE+CONTINUE → tabla
  sin unicidad por tenant, sin error visible. Auditar duplicados antes; revisar logs.
- **`schema_migrations` dirty (MEDIO):** si una migración intermedia falla, migrate marca `dirty` y bloquea.
  Probar en DB efímera; tener `migrate force <v>` listo.
- **Decisiones para humano:** ¿DB destino se resetea o ya está en 231 (define cuán agresiva la renumeración)?
  ¿001 va antes/junto? ¿hay duplicados de nombre activos hoy?
