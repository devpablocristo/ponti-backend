#!/usr/bin/env sh
set -eu

# Script de migración para docker-compose
rm -f /migrations/000000_baseline_schema.up.sql /migrations/000000_baseline_schema.down.sql

# ponti-db escucha en 5432 dentro del contenedor; DB_PORT es el puerto del host
exec migrate -path /migrations \
  -database "postgres://${DB_USER}:${DB_PASSWORD}@ponti-db:5432/${DB_NAME}?sslmode=${DB_SSL_MODE}" \
  up
