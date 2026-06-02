# Dependency Map — descomposición de `develop-problematico`

> **Doc global** del análisis de descomposición. Grafo textual de dependencias entre features (intra-repo y cross-repo).
>
> | | |
> |---|---|
> | **Fecha del análisis** | 2026-05-30 |
> | **Repo de este análisis** | Backend Go `ponti-backend` (`core`) — `/home/pablocristo/Proyectos/pablo/ponti/core` |
> | **Rango fuente** | `0972e565..777e5f6a` |
> | **SOURCE de extracción** | `develop-problematico~1` (SHA `777e5f6a`) — **NUNCA** el tip |
> | **Destino del merge** | `develop` |
>
> **Contexto:** `develop-problematico` fue una rama de integración (new-cns3 + projects + admin + ...) cuyo **último** commit es un *restore* que la vacía. Por eso la fuente real es `develop-problematico~1` (el pico de contenido). Todo `git restore -p <SHA> -- <path>` debe partir de **`777e5f6a`**, jamás del tip.
>
> **Guardrails:** este doc describe dependencias y sugiere orden de merge. Los comandos `git` que aparezcan son **sugerencias**; ningún cambio de código se deriva de este doc.

---

## 0. Cómo leer este mapa

- **A -> B** = "B depende de A" = A debe mergearse **antes** que B (A bloquea a B).
- Fuerza de la arista:
  - **fuerte** = dependencia de compilación/contrato; ignorarla rompe build o runtime.
  - **débil** = acoplamiento de producto/UX o de orden recomendado; compila sin ella pero el feature no funciona end-to-end o queda inconsistente.
  - **incierta** = sospecha por archivos compartidos / no confirmada a nivel de símbolo; tratar como fuerte hasta verificar.
- **`[BE]`** vive en este repo (`core`). **`[FE]`** vive en el repo frontend (`web` + BFF `api`). **`[FULL]`** cruza ambos.
- **DONE** = ya está en `develop`, sin paquete; se excluye del trabajo pero puede aparecer como pre-requisito ya satisfecho.

Las dependencias declaradas (`deps=...`) salen de la tabla autoritativa de features. Las que agrego como **débiles/inciertas** están marcadas explícitamente y justificadas.

---

## 1. Grafo raíz (camino crítico del BE)

Cadena espinal del backend, de la raíz hacia las hojas:

```
develop
  └─ 001 be-platform-tenancy-refactor   [BE] (raíz BE — drop MaybeTenantScope -> tenancy.Scope, ~23 repos)
       ├─ 003 be-multitenant-db-hardening      [BE] (migr 224/225: backfill -> constraints)
       ├─ 027 be-cleanup-domain-purity         [BE] (staticcheck + remove governance/jwt legacy)
       ├─ 025 be-test-coverage                 [BE] (valida 001) [+ tambien dep 002,009]
       └─ 007 actor-system                     [FULL] (necesita 001,002,003,004,006)

  └─ 002 be-crudar-lifecycle-framework  [BE] (internal/shared/lifecycle + migr 227/228/232/233 — funda 009)
       ├─ 009 crudar-archive-surface          [BE] (CONTRATO DELETE -> archive/hard/archived, ~20 dominios)
       └─ 007 actor-system                     [FULL]

  └─ 004 shared-text-propername         [BE] (util chico — bloquea 007)
       └─ 007 actor-system                     [FULL]

  └─ 005 be-config-modularization       [BE] (cmd/config split + .env.example — funda 012 y 023)
       ├─ 012 ai-companion-integration         [FULL]
       └─ 023 be-wire-di                        [BE]
```

Camino crítico ejemplar (el más largo del BE):

```
develop -> 001 -> 002 -> 007 -> 008
                              └-> 010
develop -> 001 -> 003 -> 007 (007 tambien depende de 002,004,006)
develop -> 005 -> 012 -> 023
```

---

## 2. Grafo raíz del FE (cross-repo)

El FE tiene su propia raíz, **006 fe-design-system**, que no vive en este repo:

```
006 fe-design-system   [FE] (raíz FE — primitivos, lib, router/main shell; router.tsx/main.tsx MEZCLADOS)
  ├─ 007 actor-system               [FULL]  (FE useActors + master-data/actors)
  ├─ 014 fe-master-data-pages       [FE]    (familia 212 archivos; tambien dep 007,009)
  ├─ 015 fe-dashboard-consolidation [FE]
  ├─ 016 fe-access-notifications    [FE]
  ├─ 017 fe-dollar-commerce-forms   [FE]
  └─ 026 fe-test-infra              [FE]
```

Ejemplo de raíz FE (de la consigna): `006 -> 014`.

---

## 3. Dependencias declaradas (de la tabla) — aristas fuertes

Lista normalizada `prereq -> feature` con la fuerza por defecto **fuerte** salvo nota.

| Feature | deps declaradas | Aristas (prereq -> feature) | Notas de fuerza |
|---|---|---|---|
| 001 | — | `develop -> 001` | raíz BE |
| 002 | — | `develop -> 002` | raíz BE (lifecycle) |
| 003 | 001 | `001 -> 003` | **fuerte** (tenancy.Scope ya presente para constraints) |
| 004 | — | `develop -> 004` | util independiente |
| 005 | — | `develop -> 005` | infra config independiente |
| 006 | — | `develop(FE) -> 006` | raíz FE |
| 007 | 001,002,003,004,006 | `001->007`, `002->007`, `003->007`, `004->007`, `006->007` | 001/002/003/004 **fuertes** (BE compila/migra). 006 **fuerte para la mitad FE** |
| 008 | 007 | `007 -> 008` | **fuerte** (me_context se apoya en actors/tenants de 007) |
| 009 | 002 | `002 -> 009` | **fuerte** (archive usa lifecycle de 002) |
| 010 | 007,009 | `007 -> 010`, `009 -> 010` | **fuerte** (project-bridge sobre actors + archive) |
| 011 | — | `develop -> 011` | bugfix de shape; sin deps de build |
| 012 | 005 | `005 -> 012` | **fuerte** (axis/companion config sale de cmd/config) |
| 013 | — | `develop -> 013` | csvexport independiente |
| 014 | 006,007,009 | `006->014`, `007->014`, `009->014` | **fuerte** 006; **fuerte/cross-repo** 007,009 (contratos BE) |
| 015 | 006 | `006 -> 015` | **fuerte** (usa design-system) |
| 016 | 006 | `006 -> 016` | **fuerte** |
| 017 | 006 | `006 -> 017` | **fuerte** |
| 018 | — | `develop -> 018` | data-integrity; ver excepción tentative-prices abajo |
| 019 | — | `develop -> 019` | tooling/scripts, bajo riesgo |
| 020 | — | `develop -> 020` | CI por repo |
| 021 | — | `develop -> 021` | build/deploy por repo; ver excepción dep-bumps abajo |
| 022 | — | `develop -> 022` | hooks, opcional |
| 023 | 001,005,007,008,009,012 | `001->023`, `005->023`, `007->023`, `008->023`, `009->023`, `012->023` | **fuerte** (wire/cmd-api referencian símbolos de todos esos módulos) |
| 024 | — | `develop -> 024` | docs |
| 025 | 001,002,009 | `001->025`, `002->025`, `009->025` | **fuerte para compilar tests** (valida esos módulos) |
| 026 | 006 | `006 -> 026` | **fuerte** (smoke/e2e contra el shell de 006) |
| 027 | 001 | `001 -> 027` | **fuerte** (limpieza sobre la base tenancy) |

---

## 4. Aristas débiles e inciertas (no declaradas en la tabla, pero reales)

Estas **no** están en `deps=` pero condicionan el orden o el funcionamiento end-to-end. No bloquean compilación salvo donde se diga.

### 4.1 Débiles (producto / UX / coherencia)

- **011 campaign-dto-projectid** está marcado `merge=coordinado (shape change)`. Arista débil bidireccional **FE<->BE de 011**: si BE serializa `project_id/id/name` en minúscula y el FE no se actualiza (o viceversa), el dropdown de campañas queda vacío. No rompe build; rompe la feature. **Mergear FE y BE de 011 juntos.**
  - `011[BE] <≈> 011[FE]` (débil, coordinado)
  - Posible relación débil **010 -> 011**: projects introduce `project_id`; si 010 aún no está, el `project_id` de campañas no tiene a qué apuntar a nivel producto. *Incierta* — verificar si el shape de 011 referencia entidades de 010.

- **018 data-integrity-admin** (`merge=coordinado`): FE `pages/admin/data-integrity` + BE `internal/data-integrity`. Arista débil **018[BE] <≈> 018[FE]** (la UI consume el endpoint). No declarada como dep formal, pero deben ir coordinados.

- **013 be-csv-export** cambia endpoints de export **XLSX -> CSV** y borra excel. Esto crea una **dependencia de consumo cross-repo no declarada**: cualquier página FE que descargue/parseé el export (típicamente en **014** master-data y en **015** dashboard) asume el nuevo formato. Arista débil:
  - `013[BE] -> 014[FE]` (débil, cross-repo — revisar consumo del export)
  - `013[BE] -> 015[FE]` (débil/incierta — si el dashboard exporta)
  - Si se mergea 013 sin actualizar el FE consumidor, las descargas quedan rotas en runtime aunque todo compile.

- **008 identity-tenant-context** habilita el switcher de tenants. Débil **008 -> 010/014/...**: las pantallas multi-tenant (projects, master-data) cobran sentido recién con el contexto de tenant del 008. No es dep de build, es de experiencia. Marcar **incierta -> verificar** si alguna página de 010/014 lee el `TenantContext` de 008.

### 4.2 Inciertas (archivos compartidos / a confirmar a nivel símbolo)

Estas nacen de los **archivos compartidos/peligrosos** conocidos del repo. Son aristas "de archivo", no "de feature": dos features tocan el mismo archivo, así que su **orden de aplicación de hunks** importa aunque no haya dependencia lógica directa.

Archivos MEZCLADOS y features que los tocan:

- **`wire/wire.go`, `wire/wire_gen.go`, `cmd/api/main.go`, `cmd/api/http_server.go`** → núcleo de **023**, pero reciben hunks de **007** (`actor_providers`) y **012** (`companion_providers`), y potencialmente de 008/009/010.
  - Implica: `007 -> 023`, `012 -> 023` ya son fuertes (declaradas). Lo **incierto** es que el *mismo archivo* recibe pedazos de varias features → al portar con `git restore -p 777e5f6a -- wire/...` hay que traer los hunks de cada módulo **junto a ese módulo**, no en bloque. Conflicto de aplicación, no de lógica.
- **`cmd/config/loadconfig.go`** → núcleo de **005**; recibe hunk de **012** (config de companion/axis). `005 -> 012` (ya fuerte).
- **`go.mod` / `go.sum`** → tocados por **021** (build/deploy). **OJO:** los bumps de `go-jose/v4` y `x/net` ya están **DONE (#124)** → al portar 021 hay que **separar** esos hunks (ya en `develop`) del resto. Arista de archivo incierta: cualquier feature que agregue un import nuevo (p.ej. 012 con su cliente HTTP, 013 con lib CSV) toca `go.mod/go.sum`. Resolver por orden de merge, no por dep lógica.
- **`Makefile`** → núcleo de **019**; puede recibir targets de 013 (csv), 005 (config), 020 (ci). Incierta.
- **`internal/shared/handlers/**`, `internal/shared/models/base.go`, `internal/shared/repository/**`** → base transversal. **001** (tenancy.Scope) y **002/009** (lifecycle/archive) los reescriben. Cualquier feature de dominio (007,010,...) que herede de `base.go`/shared repo **depende implícitamente** de que 001/002/009 ya hayan tocado esos shared. Refuerza las aristas `001 -> *` y `002 -> 009 -> *`.

> Regla operativa para archivos MEZCLADOS: **siempre `git restore -p` desde `777e5f6a`**, seleccionando solo los hunks del módulo que estás portando, y mergeando ese archivo en el PR de su feature dueña (023 para wire/cmd-api, 005 para config, 019 para Makefile). Nunca traer el archivo entero de una.

---

## 5. Quién bloquea a quién (resumen invertido)

Lista de **bloqueadores** ordenada por impacto (cuántos features esperan por cada uno):

| Bloqueador | Bloquea directamente a | Alcance |
|---|---|---|
| **001** be-platform-tenancy-refactor | 003, 007, 023, 025, 027 (+ implícito todos los dominios BE vía shared) | **Máximo** — base de todo el BE |
| **006** fe-design-system | 007(FE), 014, 015, 016, 017, 026 (+ implícito todo el FE) | **Máximo FE** — base de todo el FE |
| **007** actor-system | 008, 010, 014, 023 | Alto (full-stack) |
| **002** be-crudar-lifecycle-framework | 009, 007, 025 | Alto |
| **009** crudar-archive-surface | 010, 014, 023, 025 | Alto |
| **005** be-config-modularization | 012, 023 | Medio |
| **003** be-multitenant-db-hardening | 007 | Medio (gate de datos) |
| **004** shared-text-propername | 007 | Bajo (util, pero gatea 007) |
| **012** ai-companion-integration | 023 | Bajo |
| **008** identity-tenant-context | 023 (+ débiles UX) | Bajo |

**Hojas (no bloquean a nadie):** 010, 011, 013, 015, 016, 017, 018, 019, 020, 021, 022, 024, 025, 026, 027. Son las candidatas naturales a mergear al final o en paralelo (sujeto a sus propios prereqs).

**Convergencia notable:** **023 be-wire-di** es el sumidero del BE: depende de 001,005,007,008,009,012. Es lo **último** del backend que se integra, porque cablea todo lo anterior. Su `deps` declaradas coinciden con el conjunto de módulos cuyos providers cablea.

---

## 6. Orden de merge sugerido (topológico)

Respetando todas las aristas fuertes. Las olas dentro de un mismo nivel pueden ir en paralelo.

### BE (este repo)
```
Ola 0 (raíces, paralelo):   001 · 002 · 004 · 005 · 013 · 019 · 022 · 024
Ola 1:                      003 (←001) · 009 (←002) · 027 (←001) · 012 (←005)
Ola 2:                      007 (←001,002,003,004 [+006 FE])
Ola 3:                      008 (←007) · 010 (←007,009) · 011[BE] (coordinado)
Ola 4:                      023 (←001,005,007,008,009,012)
Follow-up:                  025 (←001,002,009) · 018[BE] (coordinado) · 020[BE] · 021[BE]
```

### FE (repo frontend) — cross-repo
```
Ola 0:                      006
Ola 1 (←006, paralelo):     015 · 016 · 017 · 026
Ola 2 (←006 + contratos BE): 007[FE] (tras 007[BE]) · 014 (←006,007,009 BE)
Coordinados con BE:         008[FE] (tras 008[BE]) · 010[FE] (tras 010[BE]) · 011[FE] (con 011[BE]) · 018[FE] (con 018[BE])
Follow-up:                  013-consumers (revisar export CSV en 014/015) · 020[FE] · 021[FE]
```

**Sincronización full-stack (BE-first salvo 011/018 coordinados):**
- 007: BE primero, luego FE.
- 008: BE primero, luego FE.
- 010: BE primero, luego FE.
- 012: BE-first (FE de ai puede seguir).
- 011: **coordinado** (mergear FE+BE juntos — shape change).
- 018: **coordinado** (UI + endpoint juntos).

---

## 7. Exclusiones (DONE — ya en `develop`)

No entran en ningún paquete; aparecen aquí solo para que no se re-porteen y como prereqs ya satisfechos:

- **table-select-filters** — FE #104.
- **reports-dark-mode** — FE #105. *(Nota: la limpieza de json-tags del dominio BE NO está porteada → eso va en **027**.)*
- **lot-metrics / total_tons** — FE+BE #117/#121/#124.
- **tentative-prices** — FE+BE #121/#124 → **excluir de 018**.
- **dependency-bumps** (`go-jose`, `x/net`) — BE #124 → **excluir de 021** (separar esos hunks de `go.mod/go.sum`).
- **lots / workorders** master-data — parcialmente DONE (#104/#117) → al portar **014**, descontar lo ya hecho de esas dos entidades.

---

## 8. Grafo compacto (todas las aristas, una línea por arista)

Formato `prereq -> feature [fuerza]`. `develop` y `develop(FE)` son las raíces.

```
develop -> 001 [fuerte]
develop -> 002 [fuerte]
develop -> 004 [fuerte]
develop -> 005 [fuerte]
develop -> 011 [fuerte]        # bugfix sin deps de build
develop -> 013 [fuerte]
develop -> 018 [fuerte]
develop -> 019 [fuerte]
develop -> 020 [fuerte]
develop -> 021 [fuerte]
develop -> 022 [fuerte]
develop -> 024 [fuerte]
develop(FE) -> 006 [fuerte]

001 -> 003 [fuerte]
001 -> 007 [fuerte]
001 -> 023 [fuerte]
001 -> 025 [fuerte]
001 -> 027 [fuerte]

002 -> 007 [fuerte]
002 -> 009 [fuerte]
002 -> 025 [fuerte]

003 -> 007 [fuerte]
004 -> 007 [fuerte]
006 -> 007 [fuerte, cross-repo]
006 -> 014 [fuerte]
006 -> 015 [fuerte]
006 -> 016 [fuerte]
006 -> 017 [fuerte]
006 -> 026 [fuerte]

007 -> 008 [fuerte]
007 -> 010 [fuerte]
007 -> 014 [fuerte, cross-repo]
007 -> 023 [fuerte]

008 -> 023 [fuerte]

009 -> 010 [fuerte]
009 -> 014 [fuerte, cross-repo]
009 -> 023 [fuerte]
009 -> 025 [fuerte]

005 -> 012 [fuerte]
005 -> 023 [fuerte]
012 -> 023 [fuerte]

# debiles / coordinados (producto, no build)
011[BE] <≈> 011[FE] [debil, coordinado]
018[BE] <≈> 018[FE] [debil, coordinado]
010 -> 011 [incierta, verificar shape project_id]
013 -> 014 [debil, cross-repo, consumo export CSV]
013 -> 015 [incierta, cross-repo, si dashboard exporta]
008 -> 010 [debil, UX TenantContext]
008 -> 014 [incierta, UX TenantContext]

# inciertas de archivo compartido (orden de hunks, no logica)
001 -> {dominios BE} [incierta-implicita, shared handlers/models/repository]
002 -> {dominios BE} [incierta-implicita, shared lifecycle]
005 -> 012 [archivo, loadconfig.go]            # ya fuerte
007 -> 023 [archivo, wire/cmd-api]             # ya fuerte
012 -> 023 [archivo, wire/cmd-api/config]      # ya fuerte
019 <≈> {005,013,020} [incierta, Makefile]
021 <≈> {012,013,...} [incierta, go.mod/go.sum imports]
```

---

## 9. Riesgos de ordenamiento (top)

1. **Saltarse 001 antes de cualquier dominio BE** → no compila (símbolo `tenancy.Scope` ausente). 001 es no-negociable como primer merge BE.
2. **003 con datos stale** → migr 224/225 hace backfill y luego pone constraints; si quedan filas sin tenant, el constraint falla. Gate de datos antes de 007.
3. **014 sin 007/009 en BE** → las pantallas master-data llaman endpoints (`/api/v1/actors`, `archive/hard/archived`) que no existen aún → 404 en runtime. Cross-repo, no lo detecta el build del FE.
4. **013 sin actualizar consumidores FE** → exports CSV vs FE esperando XLSX → descarga rota en runtime.
5. **011 desincronizado** → dropdown de campañas vacío. Mergear FE+BE juntos.
6. **Portar wire/cmd-api/go.mod en bloque** desde `777e5f6a` → arrastra hunks de features que aún no se mergearon → build roto o features fantasma. Usar `git restore -p` por hunk, por feature.
7. **Re-portear DONE** (tentative-prices, dep-bumps, lots/workorders) → conflictos o duplicación. Ver sección 7.
```
