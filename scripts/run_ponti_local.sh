#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
ROOT_DIR="$(cd "$BACKEND_DIR/.." && pwd)"
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
  # Chequear si el puerto que vamos a usar para la DB ya está ocupado.
  local port="${DB_PORT:-5432}"
  if ss -tlnp 2>/dev/null | grep -qE "(:${port}|\\.${port})\\s"; then
    echo "WARN: Puerto ${port} ya en uso. Docker DB podría fallar al iniciar."
  fi
}

ensure_env_file "$BACKEND_DIR"
ensure_env_file "$FRONTEND_DIR/api"
ensure_env_file "$AI_DIR"

# Importante: este script NO setea defaults de infraestructura ni modifica .env.
# La configuración debe vivir en los archivos `.env` de cada servicio (local)
# o en las variables de entorno del ambiente (dev/staging/prod).
set -a
source "$BACKEND_DIR/.env"
set +a

# Si DB_PORT=5432 pero el 5432 está ocupado, fallar con instrucción clara.
if [[ "${DB_PORT:-5432}" == "5432" ]]; then
  if ss -tlnp 2>/dev/null | grep -qE '(:5432|\.5432)\s'; then
    echo "ERROR: 5432 está ocupado. Para correr el stack local, setea DB_PORT=5433 en $BACKEND_DIR/.env" >&2
    exit 1
  fi
fi

# Validación mínima de coherencia local:
# - Si el backend corre con AUTH_ENABLED=false, el BFF debería usar LOCAL_DEV_AUTH=1.
if grep -qE '^AUTH_ENABLED=false' "$BACKEND_DIR/.env" 2>/dev/null; then
  if ! grep -qE '^LOCAL_DEV_AUTH=1' "$FRONTEND_DIR/api/.env" 2>/dev/null; then
    echo "WARN: $BACKEND_DIR/.env tiene AUTH_ENABLED=false pero $FRONTEND_DIR/api/.env no tiene LOCAL_DEV_AUTH=1. Login local puede fallar."
  fi
fi

echo "Bajando contenedores antes de levantar..."
docker compose -f "$BACKEND_DIR/docker-compose.yml" down --remove-orphans
docker compose -f "$AI_DIR/docker-compose.yml" down --remove-orphans -v
if [[ -f "$FRONTEND_DIR/docker-compose.yml" ]]; then
  # A veces Vite/Yarn se quedan colgados y el stop falla; hacer down "best effort".
  docker compose -f "$FRONTEND_DIR/docker-compose.yml" down --remove-orphans --timeout 1 || true
  docker compose -f "$FRONTEND_DIR/docker-compose.yml" kill || true
  docker compose -f "$FRONTEND_DIR/docker-compose.yml" down --remove-orphans --timeout 1 || true
fi

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
  # Detach: si salís del follow de logs (Ctrl+C), no debería matar la API.
  nohup env DB_PORT="${DB_PORT:-5432}" make -C "$BACKEND_DIR" run-api >"$BACKEND_DIR/.run-api.local.log" 2>&1 &
  echo "Backend API iniciada (detached). Log: $BACKEND_DIR/.run-api.local.log"
fi

echo "Levantando AI (DB + API) con Docker..."
# Levantar solo lo necesario (evitar ai-test en local).
# Ollama solo si el .env indica LLM_PROVIDER=ollama.
ai_services=(ai-db ai-migrate ponti-ai)
llm_provider="$(grep -E '^LLM_PROVIDER=' "$AI_DIR/.env" 2>/dev/null | tail -1 | cut -d= -f2- | tr -d '\r' | tr '[:upper:]' '[:lower:]' || true)"
if [[ "$llm_provider" == "ollama" ]]; then
  ai_services+=(ollama)
fi
docker compose -f "$AI_DIR/docker-compose.yml" up -d "${ai_services[@]}"

echo "Levantando frontend con Docker Compose..."
if [[ -f "$FRONTEND_DIR/docker-compose.yml" ]]; then
  docker compose -f "$FRONTEND_DIR/docker-compose.yml" up -d
else
  echo "ERROR: falta $FRONTEND_DIR/docker-compose.yml (el FE ahora usa docker-compose)" >&2
  exit 1
fi

echo "Todos los servicios fueron lanzados. Mostrando logs (Ctrl+C para salir)..."
docker compose -f "$BACKEND_DIR/docker-compose.yml" logs -f &
docker compose -f "$AI_DIR/docker-compose.yml" logs -f &
docker compose -f "$FRONTEND_DIR/docker-compose.yml" logs -f &
wait
