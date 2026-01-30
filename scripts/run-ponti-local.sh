#!/usr/bin/env bash
set -euo pipefail

BACKEND_DIR="/home/pablo/Projects/Pablo/ponti-backend"
AUTH_DIR="/home/pablo/Projects/Pablo/ponti-auth"
FRONTEND_DIR="/home/pablo/Projects/Pablo/ponti-frontend"

echo "Levantando backend (DB + migraciones) con Docker..."
cd "$BACKEND_DIR"
docker compose up -d

echo "Levantando auth (DB) con Docker..."
cd "$AUTH_DIR"
docker compose up -d

echo "Aplicando migraciones de auth..."
go run ./cmd/migration/main.go UP

echo "Levantando backend API..."
cd "$BACKEND_DIR"
go run ./cmd/api &

echo "Levantando auth API..."
cd "$AUTH_DIR"
go run ./cmd/api &

echo "Levantando frontend (api + ui)..."
cd "$FRONTEND_DIR"
yarn dev &

echo "Todos los servicios fueron lanzados. Logs en esta terminal."
wait
