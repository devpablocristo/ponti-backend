# risks.md — feature-019 · be-local-tooling-db-scripts

## Funcionales

- **Tooling "verde-pero-inútil":** los `.sql` (golden masters, auditorías) y
  `lint-tenant-leaks.sh` corren sin error de sintaxis aunque las features de las que
  dependen (007 actors, 003 tenancy, 001 platform) no estén. El resultado puede ser
  vacío o engañoso. *Mitigación:* documentar el pre-requisito en `scripts/README.md` y
  no presentar el lint como "verde" hasta que 001 esté en `develop`.
- **`actors_backfill_sync.sql` hace `TRUNCATE`** sobre `project_responsibles`,
  `project_investor_allocations`, `project_admin_cost_allocations`, `field_lease_participants`
  y reinserta. Es idempotente por diseño, pero destructivo sobre tablas de actors.
  *Mitigación:* solo correr en local tras restore; nunca apuntar a una DB con datos productivos editados a mano.

## Técnicos

- **Makefile compartido (alto):** extraerlo entero arrastra `openapi:` (024),
  rename `core/*`→`platform/*` (001), `lint golangci-lint v2.11.4` (020/022) y la baja
  de `seed`/`select-ponti-*`. *Mitigación:* `git restore -p` y aceptar solo los hunks de tooling-DB.
- **`smoke-companion/main.go` no compila sin `internal/axis` (012)** → rompe
  `go build ./scripts/...` y cualquier CI que compile todo el árbol. *Mitigación:* condicionar
  su inclusión a que 012 esté en `develop`; si no, postergarlo a follow-up.
- **Bit de ejecución:** `reset-local-db-from-prod.sh` cambió a `100755`. Si `git checkout`
  no preserva el modo, el `make` que lo invoca con `bash ...` igual funciona, pero la ejecución
  directa `./script.sh` falla. *Mitigación:* `chmod +x` + `git update-index --chmod=+x`.

## Integración / cross-repo

- **Cross-repo: ninguno** (solo-BE). Riesgo de mergear "solo este repo" = ninguno, no hay
  contraparte FE que deba ir en sync.
- `run_ponti_local.sh`/`down_ponti_local.sh` asumen `web/` en `$ROOT_DIR/web` (o `$PONTI_WEB_DIR`)
  y axis corriendo aparte. Si el dev no tiene esa topología, el script avisa con WARN (no rompe core).

## Datos / migración

- **`reset-local-db-from-prod.sh` es el script más peligroso:** origen PROD (read-only) +
  destino local destructivo (reset + truncate + restore). Guardas: `DB_HOST` debe ser
  `localhost`/`127.0.0.1` o aborta; `DRY_RUN=1` soportado; lee `SRC_PASS` de Secret Manager
  si no se pasa. *Mitigación:* validar la guarda tras extraer (`grep -n "destino bloqueado" scripts/db/reset-local-db-from-prod.sh`),
  correr siempre con `DRY_RUN=1` la primera vez.
- **No hay migraciones** en esta feature: los `.sql` no alteran el schema versionado.

## Archivos compartidos

- `Makefile`: ver "Técnicos". Único archivo compartido. Riesgo de doble-merge con 001/020/022/024.

## Extracción parcial

- Riesgo de **olvidar borrar** `repair_stocks_investor_granularity.sql` y `schema.snapshot.sql`
  (están `D` en SOURCE). Si quedan, reaparece un artefacto generado de 6589 líneas en el repo.
  *Mitigación:* `git rm` explícito + verificar `.gitignore`.
- Riesgo de **traer hunks de más** del Makefile. *Mitigación:* `git diff develop..HEAD -- Makefile`
  y revisar que no aparezcan `openapi:`, `swag`, `golangci-lint v2.11.4`, `platform/*` warnings.

## Riesgo de mergear solo este repo / solo el otro

- **Solo BE:** seguro. No hay FE asociado; nada queda desincronizado.
- **Solo el otro (FE):** no aplica — el FE no tiene cambios para esta feature.
