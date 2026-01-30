#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
ROOT_DIR="$(cd "$BACKEND_DIR/.." && pwd)"
AUTH_DIR="$ROOT_DIR/ponti-auth"
FRONTEND_DIR="$ROOT_DIR/ponti-frontend"

require_dir() {
  local dir="$1"
  local label="$2"
  if [[ ! -d "$dir" ]]; then
    echo "ERROR: no existe $label en $dir" >&2
    exit 1
  fi
}

require_dir "$BACKEND_DIR" "backend"
require_dir "$AUTH_DIR" "auth"
require_dir "$FRONTEND_DIR" "frontend"

http_ok() {
  local url="$1"
  if command -v curl >/dev/null 2>&1; then
    curl -fsS "$url" >/dev/null 2>&1
    return $?
  fi
  python - <<PY
import sys, urllib.request
try:
    urllib.request.urlopen("$url", timeout=2)
    sys.exit(0)
except Exception:
    sys.exit(1)
PY
}

ensure_env_file() {
  local dir="$1"
  if [[ -f "$dir/.env" ]]; then
    return 0
  fi
  if [[ -f "$dir/.env.example" ]]; then
    cp "$dir/.env.example" "$dir/.env"
    return 0
  fi
  echo "ERROR: falta $dir/.env y $dir/.env.example" >&2
  exit 1
}

echo "Levantando backend (DB + migraciones) con Docker..."
docker compose -f "$BACKEND_DIR/docker-compose.yml" up -d

echo "Levantando backend API..."
if http_ok "http://localhost:8080/ping"; then
  echo "Backend API ya está levantado en :8080"
else
  make -C "$BACKEND_DIR" run-api &
fi

echo "Levantando auth (DB) con Docker..."
docker compose -f "$AUTH_DIR/docker-compose.yml" up -d

ensure_env_file "$AUTH_DIR"
set -a
source "$AUTH_DIR/.env"
set +a

echo "Aplicando migraciones de auth..."
if [[ -f "$AUTH_DIR/migrations/create_users_table.sql" ]]; then
  if docker exec -i postgres psql -U "$DB_USER" -d "$DB_NAME" -tAc "SELECT to_regclass('public.users');" | grep -q "users"; then
    echo "Auth: tabla users ya existe, se omite migración"
  else
    docker exec -i postgres psql -U "$DB_USER" -d "$DB_NAME" < "$AUTH_DIR/migrations/create_users_table.sql"
  fi
else
  echo "WARN: no existe migrations/create_users_table.sql en auth"
fi

echo "Levantando auth API..."
if http_ok "http://localhost:8081/api/v1/auth/ping"; then
  echo "Auth API ya está levantado en :8081"
else
  (cd "$AUTH_DIR" && GO_ENVIRONMENT=production PORT=8081 go run ./cmd/api) &
fi

# Evitar que el PORT del auth contamine frontend (conflicto 8081)
unset PORT

if [[ ! -f "$FRONTEND_DIR/ui/package.json" ]]; then
  echo "ERROR: falta $FRONTEND_DIR/ui/package.json" >&2
  exit 1
fi
if [[ ! -f "$FRONTEND_DIR/api/package.json" ]]; then
  echo "ERROR: falta $FRONTEND_DIR/api/package.json" >&2
  exit 1
fi

frontend_cmd() {
  local dir="$1"
  local label="$2"
  local script="${3:-dev}"
  local env_vars="${4:-}"
  local pm="yarn"
  if ! command -v yarn >/dev/null 2>&1; then
    pm="npm"
  fi
  if [[ ! -d "$dir/node_modules" ]]; then
    echo "Instalando dependencias ($label) con $pm..."
    (cd "$dir" && $pm install)
  fi
  echo "Levantando $label con $pm..."
  if [[ -n "$env_vars" ]]; then
    (cd "$dir" && env $env_vars $pm run "$script") &
  else
    (cd "$dir" && $pm run "$script") &
  fi
}

echo "Levantando frontend UI..."
frontend_cmd "$FRONTEND_DIR/ui" "frontend UI" "dev"

echo "Levantando frontend API..."
frontend_cmd "$FRONTEND_DIR/api" "frontend API" "local"

echo "Todos los servicios fueron lanzados. Logs en esta terminal."
wait
