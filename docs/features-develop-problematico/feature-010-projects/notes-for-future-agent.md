# notes-for-future-agent.md — feature-010 projects (BE)

## Resumen corto
feature-010 BE migra `internal/project` a multi-tenant + actor-system + lifecycle/archive. Son 9 archivos, todos en `internal/project/`. El cambio grande es `repository.go` (2075 líneas de diff). Es FULL-STACK: el mismo `feature-010` existe en el repo FE (`pages/admin/projects` + BFF `projects.ts`). **BE-first.**

## Qué está en FE y qué en BE
- **BE (este paquete):** tenancy scope, actor-sync de customer/manager/investor, cascada de archive/restore vía lifecycle, hard-delete con bloqueo (ruta `/hard`), `assertProjectReferencesActive`, hidratación de `actor_id` legacy, `GetProjectByNameCustomerAndCampaignID`, `GetRawAdminCostTotal`.
- **FE:** consume el contrato — path `DELETE /:id/hard`, campo `actor_id`, manejo de 409s. No documentado aquí.

## Archivos esenciales
- `internal/project/repository.go` — el corazón. Helpers `ensureCustomer/Campaign/Manager/Investor/Crop` + variantes `*ForUpdate`, cascada de archive/restore/hard-delete, `assertProjectReferencesActive`, `hydrateLegacyActorIDs`.
- `internal/project/handler.go` — renombre de ruta y de método del port.
- `internal/project/handler/dto/project.go` — `ActorID` + `CanonicalizeName`.

## Archivos peligrosos / mezclados
- `repository.go`: mezcla tenancy + dominio + actor-sync + lifecycle. No se puede trocear sin riesgo; traer entero del SOURCE.
- `GetRawAdminCostTotal` dentro de `repository.go` pertenece conceptualmente a 018 (data-integrity la consume). No la dupliques en 018.

## Decisiones ya tomadas
- **Extracción whole-file** de los 9 archivos (no partial-hunks): el módulo es un snapshot coherente a `777e5f6a`.
- **No** incluir migraciones ni `go.mod`/`go.sum`/`wire`/`cmd` en este PR (los aportan las dependencias).
- **Orden de merge:** tenancy-bump → 004 → 008 → 007 → 009 → **010** → (011, 018) → FE-010.

## Dudas abiertas (para humano)
- ¿`develop` ya tiene 004/007/008/009 + bump tenancy? (hoy: NO; verificado MISSING).
- ¿018 ya está mergeado con un stub de `GetRawAdminCostTotal`?
- ¿FE-010 ya espera `/hard` o aún `DELETE /:id`?

## Comandos a mirar primero
```sh
cat /tmp/flists/be-010.txt
git -C ~/Proyectos/pablo/ponti/core diff 0972e565..777e5f6a -- internal/project/repository.go | less
git -C ~/Proyectos/pablo/ponti/core diff 0972e565..777e5f6a -- internal/project/handler.go
# verificar deps en develop:
for p in internal/actor internal/shared/lifecycle internal/shared/authz internal/shared/text; do \
  git -C ~/Proyectos/pablo/ponti/core show develop:$p >/dev/null 2>&1 && echo "$p OK" || echo "$p MISSING"; done
git -C ~/Proyectos/pablo/ponti/core show develop:go.mod | grep persistence/gorm
```

## Errores a evitar
- NO usar `develop-problematico` (tip) como fuente — usar SIEMPRE `develop-problematico~1` (777e5f6a).
- NO extraer 010 antes que sus dependencias: el build de develop se rompe.
- NO trocear `repository.go` por hunks.
- NO crear migraciones nuevas aquí.
- NO mergear FE-010 antes que BE-010.
- NO reintroducir lot-metrics/tentative-prices (ya DONE en otros archivos del módulo, no en estos 9).

## Camino más seguro
1. Confirmar deps en develop (paso 0).
2. Rama desde develop; `git checkout develop-problematico~1 -- <los 9 paths>`.
3. `git diff --check`; `go build ./...`; `go test ./internal/project/...`.
4. PR BE-first.
5. Coordinar FE-010.

## PR del otro repo: antes/después
- **Antes (BE):** este PR (feature-010 BE).
- **Después (FE):** FE feature-010 (path `/hard`, `actor_id`, 409s). El FE debe ir DESPUÉS del merge de este BE.
