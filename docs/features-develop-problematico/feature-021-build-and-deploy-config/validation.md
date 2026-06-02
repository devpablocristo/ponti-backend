# validation.md — feature-021 (BE)

## Checklist pre-PR

- [ ] El PR toca SÓLO `docker-compose.yml` y `.gitignore` (y, si aplica más adelante, `Dockerfile`). `git diff --stat` lo confirma.
- [ ] El PR **NO** toca `go.mod` ni `go.sum`.
- [ ] `git grep -n "CORE_REPO_DIR\|/home/pablo/Projects/Pablo/core" docker-compose.yml` → vacío.
- [ ] `git grep -n "GOWORK=off" docker-compose.yml` → vacío.
- [ ] `git grep -n "go.work" .gitignore` → SIGUE presente (no se arrastró el revert).
- [ ] `.gitignore` ignora `scripts/db/schema.snapshot.sql`.
- [ ] No se reintroducen los bumps go-jose/x/net (no debería, al no tocar go.mod/go.sum).
- [ ] `git diff --check` sin warnings de whitespace.

## Tests sugeridos (BE)

```bash
go build ./...
go test ./...                 # deben quedar verdes (son de otras features, no de 021)
go vet ./...
docker compose config         # valida el YAML resultante
docker compose build          # imagen buildea con prefetch platform/*
docker compose up -d && docker compose logs ponti-be   # NOTA: en este entorno usar docker logs, NO docker exec stdout
```

Cierre futuro de módulos (sólo DESPUÉS de mergear 001/002/005/013/023):
```bash
go mod tidy
go mod verify
git diff go.mod go.sum   # acá deben aparecer observability/persistence/prometheus y desaparecer excelize si 013 así lo decidió
```

## Manual / runtime

- `docker compose up` levanta el backend sin requerir un `core/` local montado.
- Sin `replace` a paths locales en `go.mod` (verificar `grep -n "replace" go.mod` → vacío o sólo replaces intencionales).
- Si un dev quiere iterar contra `platform/` local: la NOTA del compose indica agregar `replace` + mount `PLATFORM_REPO_DIR`.

## Casos borde

- Build en CI sin `GO_MODULES_TOKEN`: el `Dockerfile` usa `required=false` para el secret; debe seguir funcionando.
- Dev con `go.work` viejo apuntando a `core`: ahora `go.work` sigue gitignorado; quitar el `GOWORK=off` del compose hace que el contenedor respete cualquier `go.work` montado vía `.:/app` → si el dev tiene un `go.work` roto en el repo, podría afectar el build dentro del contenedor. Validar que no haya `go.work` en el working tree al buildear.

## Qué revisar en UI / API / DB / env

- UI: N/A (BE).
- API: N/A. (El endpoint `/observability/metrics` es de 005/020, no de 021.)
- DB: N/A; sólo el ignore de `schema.snapshot.sql`.
- ENV: `CORE_REPO_DIR` deja de usarse; documentar `PLATFORM_REPO_DIR` como opcional.

## Qué validar en el otro repo (FE-021)

- Que el FE buildee con su propia config (vite/tailwind/eslint/knip/tsconfig) y que lockfiles + generated client estén consistentes. Sin dependencia con este PR.

## Señales de incompletitud / incompatibilidad

- `go build` falla por símbolos de `observability`/`persistence` → el código de 001/005 no está; correcto que 021 no haya tocado go.mod.
- `go build` falla por `excelize` no encontrado → alguien quitó excelize de go.mod antes de 013; revertir.
- `docker compose up` pide un mount a `/home/pablo/Projects/Pablo/core` → el hunk del compose no se aplicó.
