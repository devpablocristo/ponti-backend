#!/usr/bin/env bash
# Levanta el stack local de Ponti (backend + frontend).
#
# Importante: el servicio AI vive en `axis/companion` (repo paralelo en
# `/home/<user>/Proyectos/pablo/axis/`). Este script asume que el stack axis
# ya está corriendo (`docker compose up -d` desde axis/). Si no está, se avisa
# con WARN — los endpoints `/api/v1/ai/*` retornarán error hasta que esté.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
ROOT_DIR="$(cd "$BACKEND_DIR/.." && pwd)"
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
require_dir "$FRONTEND_DIR" "frontend"

http_ok() {
  local url="$1"
  if command -v curl >/dev/null 2>&1; then
    curl --connect-timeout 2 --max-time 2 -fsS "$url" >/dev/null 2>&1
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
  local port="${DB_PORT:-5432}"
  if ss -tlnp 2>/dev/null | grep -qE "(:${port}|\\.${port})\\s"; then
    echo "WARN: Puerto ${port} ya en uso. Docker DB podría fallar al iniciar."
  fi
}

ensure_env_file "$BACKEND_DIR"
ensure_env_file "$FRONTEND_DIR/api"

# Importante: este script NO setea defaults de infraestructura ni modifica .env.
# La configuración debe vivir en los archivos `.env` de cada servicio (local)
# o en las variables de entorno del ambiente (dev/staging/prod).
set -a
source "$BACKEND_DIR/.env"
set +a

if [[ "${DB_PORT:-5432}" == "5432" ]]; then
  if ss -tlnp 2>/dev/null | grep -qE '(:5432|\.5432)\s'; then
    echo "ERROR: 5432 está ocupado. Para correr el stack local, setea DB_PORT=5433 en $BACKEND_DIR/.env" >&2
    exit 1
  fi
fi

# Validación mínima de coherencia local:
# - El BFF usa siempre Identity Platform (sin dev-mode de auth).
if ! grep -qE '^IDENTITY_PLATFORM_API_KEY=' "$FRONTEND_DIR/api/.env" 2>/dev/null; then
  echo "WARN: falta IDENTITY_PLATFORM_API_KEY en $FRONTEND_DIR/api/.env. Login local puede fallar."
fi
if ! grep -qE '^IDENTITY_PLATFORM_PROJECT_ID=' "$FRONTEND_DIR/api/.env" 2>/dev/null; then
  echo "WARN: falta IDENTITY_PLATFORM_PROJECT_ID en $FRONTEND_DIR/api/.env. Login local puede fallar."
fi

# Check de axis (companion + nexus) — viven en repo paralelo. Si no están UP,
# advertir; los endpoints /api/v1/ai/* van a fallar hasta que se levante axis.
if [[ -n "${COMPANION_BASE_URL:-}" ]]; then
  if ! http_ok "${COMPANION_BASE_URL}/readyz"; then
    echo "WARN: Companion (${COMPANION_BASE_URL}) no responde. Levantá axis: cd ../../axis && docker compose up -d"
  fi
fi
if [[ -n "${NEXUS_BASE_URL:-}" ]]; then
  if ! http_ok "${NEXUS_BASE_URL}/readyz"; then
    echo "WARN: Nexus (${NEXUS_BASE_URL}) no responde. Opcional para MVP solo-chat."
  fi
fi

echo "Bajando contenedores antes de levantar..."
docker compose --progress quiet -f "$BACKEND_DIR/docker-compose.yml" down --remove-orphans
if [[ -f "$FRONTEND_DIR/docker-compose.yml" ]]; then
  docker compose --progress quiet -f "$FRONTEND_DIR/docker-compose.yml" down --remove-orphans --timeout 1 || true
  docker compose --progress quiet -f "$FRONTEND_DIR/docker-compose.yml" kill || true
  docker compose --progress quiet -f "$FRONTEND_DIR/docker-compose.yml" down --remove-orphans --timeout 1 || true
fi

echo "Verificando conflictos de puerto PostgreSQL..."
stop_system_postgres

echo "Levantando backend (DB + migraciones) con Docker..."
docker compose --progress quiet -f "$BACKEND_DIR/docker-compose.yml" up -d ponti-db

echo "Levantando backend API (docker)..."
docker compose --progress quiet -f "$BACKEND_DIR/docker-compose.yml" up -d --build --quiet-pull ponti-api

if ! http_ok "http://localhost:8080/ping"; then
  echo "WARN: backend API aún no responde en :8080 (puede tardar por build/migrate inicial)." >&2
fi

echo "Levantando frontend con Docker Compose..."
if [[ -f "$FRONTEND_DIR/docker-compose.yml" ]]; then
  docker compose --progress quiet -f "$FRONTEND_DIR/docker-compose.yml" up -d --quiet-pull
else
  echo "ERROR: falta $FRONTEND_DIR/docker-compose.yml (el FE ahora usa docker-compose)" >&2
  exit 1
fi

echo "Todos los servicios fueron lanzados. Mostrando logs (Ctrl+C para salir)..."
docker compose -f "$BACKEND_DIR/docker-compose.yml" logs -f &
docker compose -f "$FRONTEND_DIR/docker-compose.yml" logs -f &
wait
