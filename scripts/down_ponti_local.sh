#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CORE_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
ROOT_DIR="$(cd "$CORE_DIR/.." && pwd)"
WEB_DIR="${PONTI_WEB_DIR:-$ROOT_DIR/web}"

echo "Bajando web..."
if [[ -f "$WEB_DIR/docker-compose.yml" ]]; then
  docker compose --progress quiet -f "$WEB_DIR/docker-compose.yml" down --remove-orphans --timeout 5 || true
fi

echo "Bajando core..."
docker compose --progress quiet -f "$CORE_DIR/docker-compose.yml" down --remove-orphans

echo "Stack local detenido. Nota: axis (companion + nexus) sigue corriendo aparte (gestionar desde /home/pablocristo/Proyectos/pablo/axis/)."
