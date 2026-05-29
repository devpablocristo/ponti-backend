#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
ROOT_DIR="$(cd "$BACKEND_DIR/.." && pwd)"
FRONTEND_DIR="$ROOT_DIR/ponti-frontend"
AI_DIR="$ROOT_DIR/ponti-ai"

echo "Bajando frontend..."
if [[ -f "$FRONTEND_DIR/docker-compose.yml" ]]; then
  docker compose -f "$FRONTEND_DIR/docker-compose.yml" down --remove-orphans --timeout 5 || true
fi

echo "Bajando AI..."
if [[ -f "$AI_DIR/docker-compose.yml" ]]; then
  docker compose -f "$AI_DIR/docker-compose.yml" down --remove-orphans || true
fi

echo "Bajando backend..."
docker compose -f "$BACKEND_DIR/docker-compose.yml" down --remove-orphans

echo "Stack local detenido."
