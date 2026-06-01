# Descartado / Cuestionable — develop-problematico (BE `ponti-backend`)

> **Repo:** `ponti-backend` — `/home/pablocristo/Proyectos/pablo/ponti/core`
> **Fecha del análisis:** 2026-05-30
> **Rango fuente:** `0972e565..777e5f6a`
> **SOURCE de extracción:** `develop-problematico~1` = **SHA `777e5f6a`** (el "pico"), NUNCA el tip.
> **Destino:** `develop` (tip actual `003a9b8f`, PR #124).
> **Por qué `~1`:** el último commit de `develop-problematico` es `66f2b602 restore: app a estado pre-new-cns3 + mantener tooling local actual`, que **vacía** la rama. El árbol real con las features vive en su padre `777e5f6a`.

Este documento cubre, para **este repo BE**:
1. Lo que se debe **DESCARTAR** (no portear): basura local, docs de handoff/scratch, dead code.
2. Lo **CUESTIONABLE / baja confianza**: requiere decisión humana antes de mover.
3. Lo **YA PORTEADO (DONE)** con su PR#, para no re-portearlo.
4. Archivos compartidos/peligrosos y **preguntas abiertas**.
5. **Riesgos** de traer esto a `develop`.

Todos los comandos `git` en este doc son **sugerencias** (no se ejecutaron). Cero cambios a código.

---

## 1. DESCARTAR (no portear)

### 1.1 `.claude/settings.json` — config local del agente, además con rutas de OTRA máquina

- **Estado:** `A` (added) en el rango.
- **Por qué descartar:** es configuración personal de la herramienta Claude Code, no del producto. Encima los paths del allowlist apuntan a `/home/pablo/Projects/Pablo/...` (otra máquina/usuario), no a `/home/pablocristo/Proyectos/pablo/...` de este entorno. Trae permisos hacia repos **fuera** de ponti-backend (`saas-core`, `nexus`, `core/ai/python`, cloud-sql-proxy, gcloud, psql con credenciales hardcodeadas en patrones de comando).
- **Riesgo si se trae:** ruido en el repo, falsa sensación de allowlist, fuga de topología de repos ajenos. Debería estar en `.gitignore`.
- **Acción sugerida:** NO portear. Si se quiere versionar config del agente, hacerlo en un commit aparte y saneando paths.

### 1.2 `docs/INVESTIGACION_ECOSISTEMA_IA_HANDOFF_CLAUDE.md` — documento de handoff a otro asistente

- **Estado:** `A`.
- **Qué es:** un "Investigación: ecosistema Pablo, IA (Ponti vs Pymes), reutilización (`core`/`modules`)" cuyo **propósito declarado** es "transferencia de contexto a otro asistente (p. ej. Claude)". Describe `~/Projects/Pablo`, Pymes, Nexus, ToolLab, Medmory: repos y productos **ajenos a ponti-backend**.
- **Por qué descartar:** no es documentación del producto; es scratch/handoff transversal con rutas absolutas de otra máquina e inferencias sobre otros repos. No aporta a `develop` de ponti-backend.
- **Acción sugerida:** NO portear (o moverlo fuera del repo, p. ej. a notas personales).

### 1.3 Docs de auditoría / scratch de proceso (decisión humana — ver §3 de feature-024)

Los siguientes `.md` añadidos son **auditorías puntuales fechadas** o planes de proceso, no contrato ni doc de referencia estable. Quedan a criterio de quien arme feature-024 (openapi-and-docs); por defecto **baja prioridad / candidatos a descartar**:

| Archivo | Qué es | Recomendación |
|---|---|---|
| `docs/BACKEND_CLEANUP_AUDIT.md` (Fecha 2026-05-12) | Auditoría de limpieza puntual | Scratch — descartar o archivar |
| `docs/projects-archive-audit.md` (Fecha 2026-05-26) | Auditoría arquitectónica de proyectos/archivado | Scratch — descartar o archivar |
| `docs/audit-custom-errors.md` | "read-only sweep" de `fmt.Errorf`/`errors.New` con propuestas de migración a `domainerr` | Scratch — descartar; es una TODO-list de un refactor |
| `CRUDAR_PLAN.md` | Plan de homogeneización FE "CRUDAR universal", estado a 2026-05-19 sobre rama `new-cns3` | Plan de proceso — descartar o mover a feature-024 |

> Nota: docs de **contrato/referencia** estables (`docs/openapi/*`, `docs/ERROR_CATALOG.md`, `docs/MULTI_TENANT_100_EVIDENCE.md`, `docs/crudar-lifecycle.md`, `docs/DATA_INTEGRITY_CONTRACT.md`, `docs/ARCHITECTURE.md`, `docs/OBSERVABILITY.md`, `CLAUDE.md`) SÍ van en feature-024; no se descartan aquí.

> **PDFs / binarios:** se buscaron (`*.pdf`, `.DS_Store`, `.idea`, `.vscode`, `*.bak/.orig/.log`) y **no existen** en el árbol de `777e5f6a`. El candidato genérico "PDFs" del brief **no aplica en este repo BE**.

---

## 2. CUESTIONABLE / baja confianza (requiere decisión humana)

### 2.1 `internal/axis/nexus_client.go` (+ `nexus_types.go` y su provider) — **dead-wired, explícitamente descartado**

- **Estado:** `A`. Parte del paquete `internal/axis/` que pertenece a **feature-012 (ai-companion-integration)**.
- **Hallazgo clave (verificado):** en `wire/wire_gen.go:300` el cliente de Nexus está **cableado a la basura**:
  ```go
  _, err = ProvideNexusClient(nexusCfg) // Nexus opcional, descartado hasta ola 2
  ```
  El resultado se asigna a `_`. El comentario del propio código dice "descartado hasta ola 2". `wire/companion_providers.go` lo describe como "Nexus opcional".
- **Interpretación:** el **cliente Companion** de `axis` está plenamente usado (`internal/ai/companion_adapter.go`), pero la **mitad Nexus** (`nexus_client.go`, `nexus_types.go`, `ProvideNexusClient`, `config.Nexus`, `defaultNexusScopes`) es funcionalidad **inacabada/aparcada** que solo compila pero no se ejerce.
- **Decisión humana:** al portear feature-012, ¿se trae la mitad Nexus o se deja afuera?
  - **Recomendado:** portear feature-012 **sin** la parte Nexus (excluir `internal/axis/nexus_client.go`, `internal/axis/nexus_types.go`, `ProvideNexusClient` y la config `Nexus`), para no meter dead code en `develop`. Companion sí.
  - Si se decide traerlo "por completitud", marcarlo explícitamente como dormido y dejar el comentario "descartado hasta ola 2".

### 2.2 `internal/reviewproxy/` — **YA ESTÁ EN `develop`** (ver también §3)

- `internal/reviewproxy/client.go` y `client_test.go` **ya existen en `develop`** (verificado con `git ls-tree -r develop -- internal/reviewproxy`).
- Es un **alias delgado** sobre `platform/kernels/governance/go/governanceclient` (dependencia ya presente en `go.mod` de develop: `github.com/devpablocristo/platform/kernels/governance/go v0.1.0`).
- **Cuestionable de cara a la extracción:** si algún paquete (p. ej. feature-023 wire / `cmd/api/http_server.go`) se trae con `restore`, NO re-introducir `internal/reviewproxy`. `cmd/api/http_server.go:21,141` lo importa y lo usa (`reviewproxy.NewClient(...)`); cuidar que el merge no duplique ni revierta lo que ya está en develop.
- **Acción:** tratar `reviewproxy` como **ya porteado**; no incluirlo en ningún paquete de extracción.

### 2.3 `internal/actor/legacy_sync.go` — puente de compatibilidad (feature-007)

- **Estado:** `A`. Sincroniza entidades legacy (`customers/investors/managers/providers/workorders.contractor/invoices.company/labors.contractor_name`) hacia el nuevo `actor`.
- **Cuestionable:** es código de **migración/transición**. ¿Es permanente o se retira una vez migrados los datos? Va atado a feature-007 y a las migraciones 223/226/231/234. Decisión: ¿se mantiene el sync legacy en `develop` o se considera scaffolding temporal?
- **Acción:** portear con feature-007, pero dejar nota de que es puente legacy (revisar si debe sobrevivir post-migración).

### 2.4 `cmd/archive-cleanup/main.go` — comando operativo (no servidor)

- **Estado:** `A`. CLI de remediación de archivado (`--dry-run`/`--apply`/`--tenant-id`/`--output`). Documentado en `scripts/data-audit/README.md`.
- **Cuestionable:** muta datos en `--apply`. Pertenece conceptualmente a **feature-002/009 (lifecycle/archive)** + tooling (feature-019). No es servicio, es herramienta operativa.
- **Acción:** portear acompañando a `internal/shared/lifecycle/` (depende de `lifecycle.*` y del gorm repo). Confirmar con humano que el `--apply` está protegido por aprobación (el README lo indica) antes de exponerlo.

### 2.5 `internal/shared/authz/authz.go` — capa authz nueva

- **Estado:** `A`. Usado por `legacy_sync.go` (`internal/shared/authz`). Compartido/transversal.
- **Cuestionable:** vive en `internal/shared/**` (zona peligrosa, ver §4). Verificar a qué feature pertenece su "dueño" (probablemente 007/008) y que no choque con authz preexistente en develop.

---

## 3. YA PORTEADO (DONE) — NO re-portear

Verificado contra `develop` (tip `003a9b8f`, PR #124) donde indicado.

| Item | Dónde / PR# | Evidencia verificada | Acción |
|---|---|---|---|
| **lot-metrics / `total_tons`** | FE+BE **#117 / #121 / #124** | `internal/lot/handler/dto/lot_metrics.go`, `lot_metrics_test.go`, `internal/lot/repository.go`, `internal/lot/usecases/domain/lot.go` presentes en `develop` | Excluir de cualquier paquete de lot |
| **tentative-prices** | FE+BE **#121 / #124** | `internal/data-integrity/handler/dto/tentative_prices.go`, `handler.go`, `usecases.go`, `usecases/domain/types.go`, `usecases_mock_test.go`, `internal/supply/repository.go`, `internal/supply/usecases/domain/supply.go` presentes en `develop` | **Excluir de feature-018** (data-integrity-admin) |
| **dependency-bumps (`go-jose/v4`, `golang.org/x/net`)** | BE **#124** | `go.mod` de `develop` y de `777e5f6a` coinciden: `go-jose/go-jose/v4 v4.1.4` y `golang.org/x/net v0.55.0` | **Excluir de feature-021** (build-and-deploy-config) |
| **`internal/reviewproxy/`** (client + test) | ya en `develop` | `git ls-tree -r develop -- internal/reviewproxy` devuelve ambos archivos | No incluir en feature-023 ni en ningún paquete |
| **table-select-filters** | FE **#104** | (FE, fuera de este repo) | n/a en BE |
| **reports-dark-mode** | FE **#105** | (FE) — OJO: la limpieza de json-tags del dominio BE NO está en #105 | la limpieza BE va en **feature-027** |

> Para el resto de DONE puramente FE (no en este repo) ver el doc de la rama FE.

---

## 4. Archivos compartidos / peligrosos (MEZCLADOS) — manejo

Estos archivos concentran cambios de **varias features** a la vez. No portear "el archivo entero"; traer **por hunks** con `git restore -p <SHA> -- <archivo>` junto al módulo dueño de cada hunk.

- `wire/wire.go`, `wire/wire_gen.go`, `cmd/api/main.go`, `cmd/api/http_server.go` → **feature-023**. OJO: `http_server.go` referencia `reviewproxy` (ya en develop) y wire referencia `actor_providers` (007), `companion_providers` (012, **con Nexus dead-wired → §2.1**).
- `cmd/config/loadconfig.go` (+ `companion.go`, `reporting.go`, `security.go` añadidos) → feature-005/012.
- `go.mod`, `go.sum` → feature-021, **menos** los bumps go-jose/x-net (DONE #124, §3).
- `Makefile` → feature-019.
- `internal/shared/handlers/**`, `internal/shared/models/base.go`, `internal/shared/repository/**`, `internal/shared/authz/**` → transversales; identificar feature-dueña por hunk.

---

## 5. Preguntas abiertas (para el humano)

1. **Nexus (axis):** ¿se trae la mitad Nexus de `internal/axis` o se deja afuera por estar "descartado hasta ola 2" (`wire_gen.go:300`)? *(Recomendación: dejar afuera; portear solo Companion.)*
2. **`legacy_sync.go`:** ¿el sync de actores legacy es permanente o scaffolding a retirar tras backfill (migr 223/226/231/234)?
3. **`cmd/archive-cleanup --apply`:** ¿se expone como comando versionado en `develop`? Muta datos; el README exige aprobación manual — confirmar guardrails.
4. **Docs de auditoría/scratch (§1.3):** ¿se descartan o se archivan en algún `docs/history/`? No son contrato.
5. **`.claude/settings.json`:** ¿se versiona config del agente? Si sí, sanear paths (`/home/pablo/` → `/home/pablocristo/`) y quitar permisos a repos ajenos; mejor `.gitignore`.
6. **feature-027 (cleanup):** la limpieza de json-tags del dominio BE NO entró con reports-dark-mode (#105). Confirmar que `internal/report/...` json-tag removal + `staticcheck` + borrado de jwt utils legacy + remoción de `core/governance` se canalizan por 027. Nota: `internal/reviewproxy/client.go:2` aún comenta "core/governance"; pero el import real es `platform/kernels/governance` — el comentario es legacy, revisar en 027.

---

## 6. Riesgos de traer esto a `develop`

- **Re-introducir lo ya porteado (DONE):** lot-metrics, tentative-prices, reviewproxy y dep bumps ya están en develop. Un `restore` ciego de archivos mezclados (wire, go.mod, http_server) puede **revertir** #124/#121/#117 o duplicar símbolos. Mitigar con `restore -p` por hunks.
- **Dead code:** portear la mitad Nexus de `axis` mete código inerte y config (`config.Nexus`, scopes) que solo añade superficie sin uso.
- **Migraciones con riesgo de datos:** feature-003 (224/225 backfill→constraints) y 002 (227/228/232/233) pueden fallar con datos stale; `cmd/archive-cleanup --apply` muta. Orquestar BE-first con dry-run y aprobación.
- **Zona `internal/shared/**`:** cambios transversales (authz, base models, handlers, repository) tocan ~todos los dominios; un merge grueso puede arrastrar features no deseadas.
- **Ruido/seguridad:** `.claude/settings.json` y el doc de handoff exponen topología de otros repos y rutas de otra máquina; no aportan a develop.
- **`develop-problematico` tip vacío:** recordar SIEMPRE extraer desde `777e5f6a` (`develop-problematico~1`), nunca desde el tip `66f2b602` (restore que vacía la rama).
