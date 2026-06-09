SHELL := /bin/bash

# Variables base
ROOT_DIR           := $(shell pwd)
VERSION            := 1.0
BUILD_DIR          := $(ROOT_DIR)/bin
DOCKER_COMPOSE_YML := $(ROOT_DIR)/docker-compose.yml
GO_MODULES_TOKEN   ?=

MIGRATIONS_DIR     := ./migrations_v4
MIGRATIONS_NAME    := $(NAME)  # pasar NAME=nombre al crear

.PHONY: all bin-build run test bin-clean lint \
        build up down logs reset rebuild clean docker-cleanup dev dev-logs \
        run-api up-ponti-local down-ponti-local reset-local-db-from-prod e2e-changes \
        smoke-axis smoke-axis-chat smoke-axis-governance smoke-axis-all \
        migrate-create openapi \
        db-reset db-migrate-up db-validate db-schema-snapshot db-schema-diff db-verify db-adopt-baseline actors-backfill-sync

# --------------------------------------------------
# OpenAPI / codegen
# --------------------------------------------------
# Genera docs/openapi/swagger.{yaml,json} desde anotaciones @Summary/@Router/@Success en handlers.
# Requiere ~/go/bin/swag (instalar con: go install github.com/swaggo/swag/cmd/swag@latest).
# El FE consume el yaml para generar tipos TS (ver ../web/ui/package.json: codegen:openapi).
openapi:
	@echo "Generating OpenAPI spec via swag..."
	@~/go/bin/swag init -g cmd/api/main.go -o docs/openapi --parseDependency --parseInternal --outputTypes json,yaml
	@echo "OK: docs/openapi/swagger.yaml updated. Now run 'yarn codegen:openapi' in web/ui."

define compose_cmd
GO_MODULES_TOKEN="$(GO_MODULES_TOKEN)" docker compose -f $(DOCKER_COMPOSE_YML)
endef

# Crea una nueva migración usando la variable NAME
migrate-create:
	@echo "Creating migration $(MIGRATIONS_NAME)..."
	@migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(MIGRATIONS_NAME)

# --------------------------------------------------
# Compilación y ejecución
# --------------------------------------------------
all: build run

bin-build:
	@echo "Building the project..."
	@mkdir -p $(BUILD_DIR)
	SERVICE_NAME=ponti-api go build -gcflags "all=-N -l" \
		-o $(BUILD_DIR)/ponti-api \
		-ldflags "-X main.Version=$(VERSION)" \
		$(ROOT_DIR)/cmd/api

run:
	@echo "Running the project (local)..."
	@go run ./cmd/api

test:
	@echo "Running tests..."
	@go test ./...

bin-clean:
	@echo "Cleaning up build artifacts..."
	@rm -rf $(BUILD_DIR)

lint:
	@echo "Linting the project..."
	@go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.11.4 run --timeout=5m

run-api:
	@echo "Starting API server..."
	@go run ./cmd/api/

up-ponti-local:
	@echo "Starting full local stack (core + web + axis)..."
	@bash ./scripts/run_ponti_local.sh

down-ponti-local:
	@echo "Stopping full local stack (core + web + axis)..."
	@bash ./scripts/down_ponti_local.sh

# --------------------------------------------------
# Base de datos (reset local + migraciones + dump data-only PROD → local)
# --------------------------------------------------
reset-local-db-from-prod:
	@echo "Resetting local DB from PROD data-only dump..."
	@echo "Asegurando que la DB local esté levantada..."
	@echo "Tip: si no seteás SRC_PASS, el script intenta leer db-password-dev desde Secret Manager (requiere gcloud auth)."
	@docker compose -f $(DOCKER_COMPOSE_YML) up -d ponti-db 2>/dev/null || true
	@bash ./scripts/db/reset-local-db-from-prod.sh
	@echo "Local DB reset + migrations + PROD data-only restore completed."

actors-backfill-sync:
	@echo "Re-ejecutando backfill/sync de actors sobre datos locales..."
	@set -a && source .env && set +a && \
	PGPASSWORD="$$DB_PASSWORD" psql -h "$$DB_HOST" -p "$$DB_PORT" -U "$$DB_USER" -d "$$DB_NAME" -v ON_ERROR_STOP=1 -f scripts/db/actors_backfill_sync.sql

# Smoke tests de release (incluye divisor de aportes). Uso: make e2e-changes [BASE_URL=http://...]
e2e-changes:
	@bash ./scripts/smoke_release.sh $(BASE_URL)

smoke-axis:
	@bash ./scripts/axis/smoke-ponti-axis-readonly.sh

smoke-axis-chat:
	@bash ./scripts/axis/smoke-ponti-axis-chat.sh

smoke-axis-governance:
	@bash ./scripts/axis/smoke-ponti-axis-draft-actions.sh
	@bash ./scripts/axis/smoke-ponti-axis-draft-previews.sh
	@bash ./scripts/axis/smoke-ponti-axis-nexus-approved-draft.sh

smoke-axis-all:
	@bash ./scripts/axis/smoke-ponti-axis-all.sh

# --------------------------------------------------
# Base de datos (verificación local v4)
# --------------------------------------------------
db-reset:
	@bash ./scripts/db/db_reset.sh

db-migrate-up:
	@bash ./scripts/db/db_migrate_up.sh

db-validate:
	@bash ./scripts/db/db_validate.sh

db-schema-snapshot:
	@bash ./scripts/db/db_schema_snapshot.sh

db-schema-diff:
	@bash ./scripts/db/db_schema_diff.sh

db-verify: db-reset db-migrate-up db-validate db-schema-snapshot db-schema-diff
	@echo "DB verify completed."

db-adopt-baseline:
	@echo "Uso: make db-adopt-baseline DB_HOST=... DB_NAME=... [DB_USER=...] [DB_PORT=...] [DB_SSL_MODE=...]"
	@bash ./scripts/db/db_adopt_baseline.sh $(DB_HOST) $(DB_NAME) $(DB_USER) $(DB_PORT) $(DB_SSL_MODE)

# --------------------------------------------------
# Desarrollo con hot reload (Air)
# --------------------------------------------------
dev:
	@echo "Starting dev environment with hot reload (Air)..."
	@if [ -z "$(GO_MODULES_TOKEN)" ]; then echo "WARN: GO_MODULES_TOKEN vacio. Si el cache de Go esta frio, ponti-api no podra bajar github.com/devpablocristo/platform/*."; fi
	$(compose_cmd) up --build -d ponti-db
	@echo "Waiting for DB to be healthy..."
	@until $(compose_cmd) exec ponti-db pg_isready -U $${DB_USER:-admin} -q 2>/dev/null; do sleep 1; done
	$(compose_cmd) up --build ponti-api

dev-logs:
	$(compose_cmd) logs -f ponti-api

# --------------------------------------------------
# Docker Compose
# --------------------------------------------------
up:
	@echo "Starting services in background (compose up -d)..."
	@if [ -z "$(GO_MODULES_TOKEN)" ]; then echo "WARN: GO_MODULES_TOKEN vacio. Si el cache de Go esta frio, ponti-api no podra bajar github.com/devpablocristo/platform/*."; fi
	$(compose_cmd) up -d

down:
	@echo "Stopping services (compose down)..."
	$(compose_cmd) down --remove-orphans

reset: down up
	@echo "Reset done (down & up)."

build:
	@echo "Building services images (compose build)..."
	$(compose_cmd) build

clean:
	@echo "Cleaning: stopping services, removing containers, volumes, networks and build artifacts..."
	@docker compose -f $(DOCKER_COMPOSE_YML) down -v --remove-orphans
	@docker system prune -f --volumes
	@rm -rf $(BUILD_DIR)
	@echo "Clean completed."

rebuild: clean build
	@echo "Rebuild completed (clean + build)."

logs:
	@echo "Tailing logs (compose logs -f)..."
	$(compose_cmd) logs -f

# Limpieza total de Docker (interactivo, requiere confirmación). Usa scripts/docker_cleanup.sh.
docker-cleanup:
	@bash ./scripts/docker_cleanup.sh
