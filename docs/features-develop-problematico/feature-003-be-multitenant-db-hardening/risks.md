# risks.md — feature-003 · be-multitenant-db-hardening

## 1. Riesgos de migración / ordenamiento (ALTO)
| riesgo | severidad | detalle | mitigación |
|--------|-----------|---------|------------|
| **Colisión de versiones golang-migrate** | ALTA | develop ya aplicó migraciones 229/230 (idénticas a SOURCE). golang-migrate (`migrate/migrate:v4.17.1`) trackea una sola versión entera. Sobre una DB en versión 230, `migrate up` **no aplica** 224/225 (números menores) -> migraciones muertas. | Camino A: portar bloque 223–228 en orden ANTES de pasar de 222 (o reset). Camino B: renumerar 224/225 a >230 manteniendo 224<226/231/234. Ver extraction-plan.md. |
| **Estado `dirty` de schema_migrations** | MEDIA | Si una migración intermedia falla, migrate deja la versión marcada `dirty` y bloquea las siguientes. | Probar en DB efímera con `make db-verify` antes; tener listo `migrate force <v>` en entorno de prueba. |
| **Hueco 223–228 en develop** | MEDIA | develop salta de 222 a 229. Indica porteo previo desordenado. 226 (no en flist) depende de 224. | Coordinar el tren de migraciones con dueños de 004/009/027. |

## 2. Riesgos de datos / backfill (MEDIO-ALTO)
| riesgo | severidad | detalle | mitigación |
|--------|-----------|---------|------------|
| **Backfill al tenant `default`** | MEDIA | 224/225 mandan TODA fila con `tenant_id IS NULL` al tenant `default`. Si ya conviven datos de varios tenants reales, se colapsan en uno. | Auditar `SELECT count(*) ... WHERE tenant_id IS NULL` y la realidad multi-tenant ANTES. En este stack el supuesto razonable es un único tenant existente. |
| **Abort por nulls remanentes en 225** | MEDIA | Si una tabla tiene `tenant_id` pero queda algún null tras el `UPDATE` (p.ej. fila insertada concurrentemente), `RAISE EXCEPTION` aborta la migración entera. | Correr 225 sin tráfico concurrente; verificar nulls previo. |
| **DROP de índices/constraints únicos de `name`** | MEDIA | 225 dropea unicidad global de `name` para reemplazarla por unicidad por tenant. Si la nueva no se crea (duplicados), queda SIN unicidad de nombre. | Auditar duplicados activos antes (query en validation.md). |

## 3. Riesgos funcionales (MEDIO)
| riesgo | severidad | detalle | mitigación |
|--------|-----------|---------|------------|
| **Endurecimiento parcial silencioso** | MEDIA | 225 ante nombres duplicados activos: `RAISE NOTICE` + `CONTINUE`. La tabla queda sin `uq_*_tenant_name`. No falla, no se ve salvo en logs. | Revisar logs de la migración; correr query de duplicados antes; convertir a reporte explícito (mejora futura). |
| **Inserts rotos en runtime sin feature-001** | MEDIA | Tras 225, `tenant_id NOT NULL` + unicidad por tenant. Si el código (001) no setea `tenant_id` o reusa nombres, los inserts fallan (23502 / 23505) ante el usuario. | Mergear 001 antes/junto. NO mergear 225 solo. |

## 4. Riesgos técnicos (BAJO-MEDIO)
| riesgo | severidad | detalle | mitigación |
|--------|-----------|---------|------------|
| FK NOT VALID -> VALIDATE | BAJA | 224 crea FK NOT VALID, 225 la valida. Si entre ambas se insertan filas con `tenant_id` huérfano, VALIDATE falla. | Aplicar 224 y 225 juntas/seguidas sin ventana de escritura. |
| `pgcrypto` ausente | BAJA | `gen_random_uuid()` de las tablas de seguridad. 224 hace `CREATE EXTENSION IF NOT EXISTS`. | Requiere permiso para crear extensión; ya se asume en el entorno. |

## 5. Riesgos de integración / cross-repo
- **Cross-repo:** NULO. Solo-BE; sin FE.
- **Integración con 001:** ver "Inserts rotos" arriba. Es el acoplamiento más importante.

## 6. Riesgos de archivos compartidos
- **Archivos:** NINGUNO compartido en la flist (sin wire/cmd/go.mod/Makefile/shared).
- **Recurso compartido real:** el espacio de numeración de `migrations_v4/` (cubierto en sección 1) y la siembra de roles/permisos (idempotente, bajo riesgo por `ON CONFLICT`).

## 7. Riesgo de extracción parcial
- BAJO dentro de la flist (son 4 archivos enteros; o están los 4 o ninguno).
- El riesgo real de "parcial" es **sistémico**: traer 224/225 SIN 001 (código), SIN 226 (que depende de 224), o SIN resolver el orden. Eso deja el sistema a medias.

## 8. Riesgo de mergear solo este repo
- **Mergear solo BE (esta feature) es lo esperado** (Solo-BE), PERO mergear 003 sin feature-001 es peligroso (sección 3). El "solo este repo" problemático no es FE-vs-BE sino **003-sin-001** dentro del mismo BE.
- **Mergear solo el otro repo (FE):** N/A — no hay cambios FE.

## Mitigación transversal (orden seguro)
1. Auditar datos (nulls + duplicados de nombre activos).
2. Confirmar versión actual de `schema_migrations` en cada entorno -> elegir Camino A/B.
3. Mergear/portar feature-001 antes o junto.
4. Aplicar 224 y 225 seguidas, sin escritura concurrente.
5. Revisar logs por NOTICE de tablas saltadas.
6. `make db-verify` verde antes del PR.
