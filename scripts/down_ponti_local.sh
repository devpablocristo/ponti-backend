#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
ROOT_DIR="$(cd "$BACKEND_DIR/.." && pwd)"
FRONTEND_DIR="$ROOT_DIR/ponti-frontend"

echo "Bajando frontend..."
if [[ -f "$FRONTEND_DIR/docker-compose.yml" ]]; then
  docker compose --progress quiet -f "$FRONTEND_DIR/docker-compose.yml" down --remove-orphans --timeout 5 || true
fi

echo "Bajando backend..."
docker compose --progress quiet -f "$BACKEND_DIR/docker-compose.yml" down --remove-orphans

echo "Stack local detenido. Nota: axis (companion + nexus) sigue corriendo aparte (gestionar desde /home/pablocristo/Proyectos/pablo/axis/)."
