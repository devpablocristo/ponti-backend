# validation.md — feature-022 · lefthook-git-hooks (BE)

## Checklist pre-PR
- [ ] `lefthook.yml` existe en la raíz del repo BE tras la extracción.
- [ ] Es byte-idéntico al SOURCE:
  ```bash
  diff <(git -C /home/pablocristo/Proyectos/pablo/ponti/core show HEAD:lefthook.yml) \
       <(git -C /home/pablocristo/Proyectos/pablo/ponti/core show 777e5f6a:lefthook.yml)
  # salida vacía = OK
  ```
- [ ] `git diff --cached --stat` lista SOLO `lefthook.yml` (ningún otro archivo colado).
- [ ] `git diff --check` no reporta whitespace ni conflict markers.
- [ ] YAML válido: `yamllint lefthook.yml` o `lefthook validate` (si Lefthook está instalado).
- [ ] No se modificó `Makefile`, `go.mod`, `go.sum`, `.gitignore`, `.github/workflows/**`.

## Validación manual (con Lefthook instalado)
```bash
# instalar (una vez)
brew install lefthook        # macOS
# o: sudo apt install lefthook  (Debian/Ubuntu)
lefthook install             # registra .git/hooks/

# correr el set pre-commit a mano sin commitear
lefthook run pre-commit
```
- [ ] `lefthook run pre-commit` ejecuta gofmt / go-vet / golangci-lint sin error en un árbol limpio.
- [ ] Stagear un `.go` con formato roto → el comando `gofmt` bloquea el commit y lista el archivo.
- [ ] `git commit --no-verify` saltea los hooks (bypass de emergencia).
- [ ] `git push` dispara `go test ./... -short`; `git push --no-verify` lo saltea.

## Casos borde a verificar
- **golangci-lint v2 de sistema + `--fast`:** confirmar si `golangci-lint run --fast` falla con la versión instalada del dev. Si falla, anotar para el follow-up (alinear con `go run ...@v2.11.4`).
- **Commit sin archivos `.go` staged:** los comandos con `glob: "*.go"` deben skippearse; `go vet ./...` no tiene glob específico de archivos staged (corre a nivel paquete) — verificar que no corra innecesariamente en commits sin Go.
- **Dev sin Lefthook instalado:** confirmar que el commit/push funciona normal (archivo inerte).
- **Repo sin tests / tests lentos:** `pre-push` con `-short` debe terminar rápido.

## Tests sugeridos
- **BE:** no aplica suite específica para esta feature. El hook *invoca* `go test ./... -count=1 -short`; opcionalmente correrlo a mano para confirmar que la suite pasa antes de habilitar el hook al equipo.
- No hay tests unitarios para un archivo de config.

## Qué revisar en UI / API / DB / env
- **UI:** N/A (BE).
- **API:** N/A — no toca endpoints.
- **DB:** N/A — no toca migraciones ni datos.
- **Env / herramientas:** confirmar que el equipo tiene (o instalará) `lefthook`, `gofmt` (viene con Go), `go` (1.26.3), y `golangci-lint` v2 alineado al Makefile.

## Qué validar en el otro repo (FE)
- Que el `lefthook.yml` del FE (feature-022) se mergee de forma consistente (mismos verbos: instalación + bypass documentados).
- No hay validación técnica cruzada: son archivos independientes.

## Señales de incompletitud / incompatibilidad
- `lefthook.yml` ausente o con menos de los 4 comandos (gofmt, go-vet, golangci-lint, tests) → extracción incompleta.
- Aparición de cambios en `Makefile`/`go.mod`/CI en el mismo commit → contaminación de otras features (019/020/021/024); revertir esos hunks.
- `lefthook install` falla al parsear el YAML → revisar identación copiada.
