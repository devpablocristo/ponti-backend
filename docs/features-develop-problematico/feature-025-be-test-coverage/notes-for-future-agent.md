# notes-for-future-agent.md — feature-025 · BE test coverage sweep

## Resumen corto

45 archivos de test del backend (44 nuevos + 1 modificado), todos bajo `internal/<modulo>/`. Son la red de
seguridad de los refactors de tenancy (001), lifecycle/crudar (002) y archive surface (009). Es una feature
**derivada**: no produce runtime, solo prueba el de otras. "Sigue a su módulo" → va de follow-up, nunca antes
que sus features productoras.

## Qué está en FE y en BE

- **FE:** nada. Sin cambios. (Anotar en cross-repo-map del FE: "feature-025: sin carpeta FE".)
- **BE:** las tres familias —
  - `repository_tenant_test.go` (23) → aislamiento multi-tenant + strict mode (valida 001/003).
  - `repository_archived_refs_test.go` (10, incl. `repository_movement_archived_refs_test.go` en supply) → integridad vs entidades archivadas (valida 009).
  - `handler_test.go` (13 nuevos + 1 modificado) → rutas, actor, lifecycle archive/restore/hard (valida 002/009).

## Archivos esenciales / peligrosos / mezclados

- **Esencial / con más señal:** `internal/work-order/handler_test.go` (el único `M`): renombra el stub
  `DeleteWorkOrderByID` → `HardDeleteWorkOrder`, agrega `ListArchivedWorkOrders(...domain.ArchivedWorkOrderFilter)`
  al stub, y agrega `TestWorkOrderActionRoutesCallExplicitUseCases` que prueba `POST .../archive`,
  `POST .../restore`, `DELETE .../hard` (espera 204 + `actionCall` correcto). Traerlo ENTERO desde 777e5f6a.
- **Peligroso:** ninguno destruye datos; el peligro es de ORDEN (mergear antes de 001/002/009 rompe CI).
- **Mezclados (partial-hunks):** NINGUNO. No toca `wire/*`, `cmd/api/*`, `go.mod`, `go.sum`, `Makefile`,
  `internal/shared/**`. 100% archivos de test aislados.

## Decisiones ya tomadas

- Extracción = **whole-file para los 45** (ninguno es parcial).
- Rama sugerida: `pr/feature-025-be-test-coverage-be`.
- Recomendación: partir en 3 sub-PRs enganchados a 001/009/002, o un PR único SOLO si los cuatro
  (001/002/003/009) ya están en develop.

## Dudas abiertas (para el humano)

1. ¿1 PR o 3 sub-PRs? La nota de feature ("pueden ir como follow-up", "sigue a su módulo") sugiere 3.
2. ¿Las firmas de producción mergeadas en develop (001/002/009) coinciden con las del SOURCE 777e5f6a?
   Si no, hay que ajustar el TEST (nunca la producción).

## Qué comandos mirar primero

```
cat /tmp/flists/be-025.txt
git -C <core> grep -ln "func assertCustomerReferencesActive" develop -- internal/customer/   # ¿está 009?
git -C <core> grep -ln "func.*HardDeleteWorkOrder" develop -- internal/work-order/            # ¿está 002?
git -C <core> grep -n  "contextkeys.OrgID" develop -- internal/                               # ¿está 001?
git -C <core> show 777e5f6a:internal/work-order/handler_test.go | head -120                   # el M
```

## Errores a evitar

- NO mergear este PR antes que 001/002/003/009 (rompe build de CI: `undefined: ...`).
- NO editar código de producción para "hacer pasar" un test. Si no compila, falta una feature o difiere una
  firma → ajustar el test.
- NO usar `develop-problematico` (tip = restore/vacío). SOURCE = `develop-problematico~1` = 777e5f6a.
- NO traer producción: lot-metrics/total_tons y tentative-prices YA están en develop (#117/#121/#124); no
  re-portar nada de eso desde aquí.
- NO tocar `go.mod`/`go.sum`: las deps de test (sqlite, uuid, gin, contextkeys, domainerr) ya están.

## Camino más seguro

1. Confirmar 001/002/003/009 en develop. 2. `git checkout -b pr/feature-025-be-test-coverage-be`.
3. `git checkout 777e5f6a -- <los 45 paths>`. 4. `go build ./...` + `go test ./internal/...`.
5. Si algo no compila → es orden/firma, no producción. 6. Abrir PR cuando todo esté verde.

## Qué PR del otro repo va antes/después

Ninguno. Solo-BE; sin acoplamiento cross-repo. El único orden a respetar es intra-BE: 025 va DESPUÉS de
001, 002, 003 y 009.
