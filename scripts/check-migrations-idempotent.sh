#!/usr/bin/env bash
# Guard de idempotencia de migraciones (sin DB, solo grep).
#
# Falla si una migración *.up.sql contiene formas NO idempotentes:
#   - CREATE TABLE            sin  IF NOT EXISTS
#   - CREATE [UNIQUE] INDEX   sin  IF NOT EXISTS
#   - CREATE TRIGGER          sin un DROP TRIGGER IF EXISTS en el mismo archivo
#
# golang-migrate corre file-per-version y marca la versión "dirty" ante un fallo
# parcial, sin auto-force: una sentencia no idempotente impide reintentar/re-aplicar.
#
# Por defecto solo audita los .up.sql NUEVOS o MODIFICADOS respecto de origin/main
# (la baseline histórica usa formas viejas y queda fuera de alcance). Se le pueden
# pasar archivos explícitos como argumentos.
set -euo pipefail

MIG_DIR="$(cd "$(dirname "$0")/.." && pwd)/migrations_v4"

files=()
if [ "$#" -gt 0 ]; then
  files=("$@")
elif git rev-parse --verify origin/main >/dev/null 2>&1; then
  while IFS= read -r f; do
    [ -n "$f" ] && [ -f "$f" ] && files+=("$f")
  done < <(git diff --name-only --diff-filter=AM origin/main...HEAD -- 'migrations_v4/*.up.sql' 2>/dev/null || true)
else
  while IFS= read -r f; do files+=("$f"); done < <(find "$MIG_DIR" -name '*.up.sql' | sort)
fi

if [ "${#files[@]}" -eq 0 ]; then
  echo "check-migrations-idempotent: nada para auditar."
  exit 0
fi

violations=0
report() { echo "  ✗ $1"; violations=$((violations + 1)); }

for f in "${files[@]}"; do
  [ -f "$f" ] || continue

  # CREATE TABLE sin IF NOT EXISTS
  if grep -nEi '^[[:space:]]*CREATE[[:space:]]+TABLE[[:space:]]+' "$f" \
      | grep -viE 'IF[[:space:]]+NOT[[:space:]]+EXISTS' >/dev/null; then
    while IFS= read -r line; do
      report "$f: CREATE TABLE sin IF NOT EXISTS -> $line"
    done < <(grep -nEi '^[[:space:]]*CREATE[[:space:]]+TABLE[[:space:]]+' "$f" | grep -viE 'IF[[:space:]]+NOT[[:space:]]+EXISTS')
  fi

  # CREATE [UNIQUE] INDEX sin IF NOT EXISTS
  if grep -nEi '^[[:space:]]*CREATE[[:space:]]+(UNIQUE[[:space:]]+)?INDEX[[:space:]]+' "$f" \
      | grep -viE 'IF[[:space:]]+NOT[[:space:]]+EXISTS' >/dev/null; then
    while IFS= read -r line; do
      report "$f: CREATE INDEX sin IF NOT EXISTS -> $line"
    done < <(grep -nEi '^[[:space:]]*CREATE[[:space:]]+(UNIQUE[[:space:]]+)?INDEX[[:space:]]+' "$f" | grep -viE 'IF[[:space:]]+NOT[[:space:]]+EXISTS')
  fi

  # CREATE TRIGGER <name>: CADA trigger requiere su PROPIO 'DROP TRIGGER IF EXISTS <name>'
  # (chequeo por nombre, no a nivel de archivo: un DROP de otro trigger no alcanza).
  while IFS= read -r tname; do
    [ -z "$tname" ] && continue
    if ! grep -qEi "DROP[[:space:]]+TRIGGER[[:space:]]+IF[[:space:]]+EXISTS[[:space:]]+${tname}([[:space:]]|;|\$)" "$f"; then
      report "$f: CREATE TRIGGER ${tname} sin 'DROP TRIGGER IF EXISTS ${tname}' previo"
    fi
  done < <(grep -oiE 'CREATE[[:space:]]+TRIGGER[[:space:]]+[a-zA-Z_][a-zA-Z0-9_]*' "$f" | sed -E 's/.*[[:space:]]//' | sort -u)
done

if [ "$violations" -gt 0 ]; then
  echo ""
  echo "check-migrations-idempotent: $violations violación(es). Usá IF NOT EXISTS / DROP ... IF EXISTS."
  exit 1
fi
echo "check-migrations-idempotent: OK (${#files[@]} archivo(s))."
