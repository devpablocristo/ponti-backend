# notes-for-future-agent.md — feature-022 · lefthook-git-hooks (BE)

## Resumen corto
Feature de **un solo archivo**: `lefthook.yml` (nuevo, raíz del repo BE). Define git hooks opt-in con Lefthook: pre-commit (gofmt + go vet + golangci-lint) y pre-push (go test -short). Es tooling local; no toca código, deps, migraciones ni CI. **Extraer entero, riesgo de merge ~nulo.**

## Qué está en FE y en BE
- **BE (este paquete):** `lefthook.yml` con comandos Go.
- **FE (repo web, mismo feature-022):** `lefthook.yml` propio con comandos JS/TS. Independiente.
- Son espejos por consistencia de DX; **sin acoplamiento técnico**. Orden de merge indistinto.

## Archivos esenciales
- `lefthook.yml` — la feature completa. Único archivo de `/tmp/flists/be-022.txt`.

## Archivos peligrosos / mezclados
- **Ninguno.** Esta feature NO toca archivos compartidos. NO traer `Makefile`, `go.mod`, `go.sum`, `.gitignore`, `wire/*`, `cmd/api/*`, `.github/workflows/**` — pertenecen a otras features (019/020/021/024/025).

## Decisiones ya tomadas
- Extraer `lefthook.yml` **whole-file** desde `develop-problematico~1` (SHA 777e5f6a).
- NO crear `.golangci.yml` aquí (no existe en el SOURCE).
- La corrección del comando golangci-lint queda como **follow-up no bloqueante**.

## Dudas abiertas (para humano)
1. ¿golangci-lint del sistema o `go run` pineado v2.11.4 (como el Makefile)? El hook usa el primero con `--fast`; en golangci-lint **v2** `--fast` fue eliminado → posible fallo. Decidir y alinear hook+Makefile+CI.
2. ¿Mergear BE y FE de feature-022 juntos? (cosmético).
3. ¿Agregar bootstrap (`make hooks` → `lefthook install`) para reducir fricción?

## Comandos para mirar primero
```bash
cat /tmp/flists/be-022.txt
git -C /home/pablocristo/Proyectos/pablo/ponti/core show 777e5f6a:lefthook.yml
git -C /home/pablocristo/Proyectos/pablo/ponti/core cat-file -e develop:lefthook.yml   # confirma que NO está en develop
git -C /home/pablocristo/Proyectos/pablo/ponti/core grep -l -i lefthook 777e5f6a       # único match: el propio archivo
git -C /home/pablocristo/Proyectos/pablo/ponti/core show 777e5f6a:Makefile | grep -n golangci   # ver versión pineada v2.11.4
```

## Errores a evitar
- NO usar `develop-problematico` (tip vacío/restore). Usar `develop-problematico~1` / `777e5f6a`.
- NO arrastrar hunks de otros archivos en el commit (revisar `git diff --cached --stat`).
- NO "arreglar" el comando golangci-lint dentro de esta extracción si el objetivo es portar fiel; si se decide arreglarlo, hacerlo como cambio explícito y anotado, no silencioso.
- NO inventar un `.golangci.yml`.

## Camino más seguro
1. `git checkout develop && git checkout -b pr/feature-022-lefthook-git-hooks-be`
2. `git checkout 777e5f6a -- lefthook.yml`
3. `git diff --cached --stat` (solo lefthook.yml) + `git diff --check`
4. commit + PR. Listo.

## PR del otro repo: antes/después
- **Indistinto.** El PR de feature-022 en FE no bloquea ni es bloqueado. Recomendación: anunciar ambos juntos al equipo para que instalen Lefthook en los dos repos a la vez.
