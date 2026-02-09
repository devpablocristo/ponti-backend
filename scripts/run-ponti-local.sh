#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
ROOT_DIR="$(cd "$BACKEND_DIR/.." && pwd)"
AUTH_DIR="$ROOT_DIR/ponti-auth"
FRONTEND_DIR="$ROOT_DIR/ponti-frontend"
AI_DIR="$ROOT_DIR/ponti-ai"

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
require_dir "$AI_DIR" "ai"

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

stop_system_postgres() {
  # Verificar conflicto solo si usamos 5432 (DB_PORT por defecto)
  local port="${DB_PORT:-5432}"
  if [[ "$port" = "5432" ]] && ss -tlnp 2>/dev/null | grep -qE '(:5432|\.5432)\s'; then
    echo "ERROR: PostgreSQL del sistema detectado en puerto 5432."
    echo "       Docker no puede usar el mismo puerto."
    echo ""
    echo "Ejecuta primero: sudo systemctl stop postgresql"
    echo "O configura DB_PORT=5433 en .env para usar otro puerto."
    exit 1
  fi
}

stop_frontend_ports() {
  # Detener procesos del FE (UI/API) por puertos conocidos
  local ports=("5173" "5174" "3000")
  local port pids
  if command -v lsof >/dev/null 2>&1; then
    for port in "${ports[@]}"; do
      pids="$(lsof -ti :"$port" || true)"
      if [[ -n "$pids" ]]; then
        echo "Deteniendo FE en puerto ${port}..."
        kill $pids || true
      fi
    done
    return 0
  fi
  if command -v fuser >/dev/null 2>&1; then
    for port in "${ports[@]}"; do
      if fuser -s "${port}/tcp"; then
        echo "Deteniendo FE en puerto ${port}..."
        fuser -k "${port}/tcp" || true
      fi
    done
  fi
}

ensure_env_file "$BACKEND_DIR"
set -a
source "$BACKEND_DIR/.env"
set +a

echo "Bajando contenedores antes de levantar..."
docker compose -f "$BACKEND_DIR/docker-compose.yml" down --remove-orphans
docker compose -f "$AUTH_DIR/docker-compose.yml" down --remove-orphans
docker compose -f "$AI_DIR/docker-compose.yml" down --remove-orphans

echo "Deteniendo frontend antes de levantar..."
stop_frontend_ports

echo "Verificando conflictos de puerto PostgreSQL..."
stop_system_postgres

echo "Levantando backend (DB + migraciones) con Docker..."
docker compose -f "$BACKEND_DIR/docker-compose.yml" up -d

echo "Levantando backend API..."
if http_ok "http://localhost:8080/ping"; then
  echo "Backend API ya está levantado en :8080"
else
  if [[ -z "${AI_SERVICE_URL:-}" || -z "${AI_SERVICE_KEY:-}" ]]; then
    echo "WARN: AI_SERVICE_URL / AI_SERVICE_KEY no configurados. Endpoints AI no funcionarán."
  fi
  make -C "$BACKEND_DIR" run-api &
fi

echo "Levantando auth (DB) con Docker..."
docker compose -f "$AUTH_DIR/docker-compose.yml" up -d

echo "Levantando AI (DB + API) con Docker..."
docker compose -f "$AI_DIR/docker-compose.yml" up -d

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
