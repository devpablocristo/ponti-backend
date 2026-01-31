#!/usr/bin/env sh
set -eu

# Script de migración para docker-compose
rm -f /migrations/000000_baseline_schema.up.sql /migrations/000000_baseline_schema.down.sql

exec migrate -path /migrations \
  -database "postgres://${DB_USER}:${DB_PASSWORD}@ponti-db:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSL_MODE}" \
  up
