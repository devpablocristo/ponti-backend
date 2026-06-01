# notes-for-future-agent.md — feature-002 · be-crudar-lifecycle-framework

## Resumen corto

Esta feature crea SOLO el paquete compartido `internal/shared/lifecycle` (5 .go
de producción + 4 de test) y 4 pares de migraciones (227/228/232/233). Es el
"motor" CRUDAR (archive/restore con cascadas, invariante "archived = no existe",
métrica). Es 100% BE. NO trae a sus consumidores: eso es feature 009.

## Qué está en FE y en BE

- **FE**: NADA. Sin carpeta, sin cambios. (En el cross-repo-map del FE: "feature-
  002 sin cambios FE".)
- **BE**: el paquete + migraciones (los 17 paths de `/tmp/flists/be-002.txt`).

## Archivos esenciales

- `internal/shared/lifecycle/policy.go` — el mapa `Policies` es la fuente de
  verdad de la jerarquía y de las políticas por entidad. Leerlo primero para
  entender todo.
- `internal/shared/lifecycle/cascade.go` — `RunCascadeArchive`/`RunCascadeRestore`
  (recursivos sobre `Policies`), `WouldOrphanActiveChildren`.
- `internal/shared/lifecycle/lifecycle.go` — `Cause`, `ArchiveBatch`,
  `RequireActive`/`RequireAllActive` (la barrera Go).

## Archivos peligrosos

- `migrations_v4/000227_*` — toca 14 tablas + cambia índice único de customers +
  migra `actors.archived_at → deleted_at`. El down NO recompone `archived_at`.
- `migrations_v4/000233_*` — triggers sin guarda de existencia de tabla; requiere
  data limpia (avisa de `scripts/data-audit/archived_invariants.sql`).
- `go.mod` / `go.sum` — NO son míos pero el paquete no compila sin 3 deps nuevas
  (prometheus + observability/go + persistence/gorm/go). NO arrastres el diff
  entero (mete otras features + colisiona con #124 go-jose/x/net YA porteado).

## Archivos mezclados (compartidos / partial-hunks)

- `go.mod`, `go.sum`: tomar SOLO los hunks de las 3 deps.
- `cmd/api/main.go`: NO traer (compartido, feature 023/005). Llama
  `lifecycle.RegisterMetrics`.
- `internal/*/repository.go` (20), `cmd/archive-cleanup/main.go`: NO traer
  (feature 009).

## Decisiones ya tomadas

- DECISIÓN: extraer el paquete como whole-file, PERO arreglar antes (1) deps,
  (2) orden de migraciones. No es "extraer tal cual".
- Tests no necesitan Postgres (sqlite in-memory) → corren en CI tal cual.

## Dudas abiertas (para humano)

1. ¿El runner de migraciones es estricto por número? Si sí, 227/228 (< 229/230
   ya en develop) hay que renumerar a > 234 manteniendo orden relativo.
2. ¿Las 3 deps Go vienen de feature 021 o se agregan aquí?
3. ¿Se mergea 002 solo (paquete inerte) o junto/antes de 009 (consumidores +
   data-audit)? Recomendado: 002 primero (con deps), 009 inmediatamente después.

## Qué comandos mirar primero

```bash
cat /tmp/flists/be-002.txt
git -C /home/pablocristo/Proyectos/pablo/ponti/core show 777e5f6a:internal/shared/lifecycle/policy.go | head -120
git -C .../core ls-tree 003a9b8f -- internal/shared/lifecycle/          # vacío = no existe en develop
git -C .../core ls-tree 003a9b8f --name-only -- migrations_v4/ | grep -E '00022[789]|00023[0-4]'  # ver 229/230 ya presentes
git -C .../core show 003a9b8f:go.mod | grep -iE 'prometheus|observability/go|persistence/gorm'    # vacío = faltan deps
git -C .../core grep -l "shared/lifecycle" 777e5f6a -- 'internal/**/*.go' | grep -v 'shared/lifecycle/'  # los 20 consumidores (feature 009)
```

## Errores a evitar

- NO usar `develop-problematico` (tip vacío); usar `develop-problematico~1`
  (777e5f6a).
- NO arrastrar el diff completo de go.mod (otras features + go-jose/x/net de #124).
- NO traer `internal/*/repository.go` ni `cmd/*` (rompe el build por símbolos de
  otras features y mezcla alcance con 009).
- NO renumerar migraciones a ciegas sin confirmar el runner y sin chequear
  referencias en feature 009.
- NO aplicar 233 sobre data sucia sin correr antes el data-audit/cleanup.

## Camino más seguro

1. Confirmar/agregar deps en go.mod (coordinar con 021).
2. Confirmar comportamiento del runner de migraciones; renumerar 227/228/232/233
   solo si es estricto.
3. Traer los 9 .go + 8 .sql whole-file en una rama desde develop.
4. `go build ./... && go test ./internal/shared/lifecycle/...`.
5. Validar migraciones up/down en staging (especialmente 232/233).
6. PR BE-first. Luego feature 009.

## Qué PR del otro repo debe ir antes/después

- Ninguno del FE (Solo-BE).
- Mismo repo BE: feature 021 (deps) ANTES o junto; feature 009 (consumidores)
  DESPUÉS.
