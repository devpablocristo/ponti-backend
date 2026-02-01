SHELL := /bin/bash

# Variables base
ROOT_DIR           := $(shell pwd)
VERSION            := 1.0
BUILD_DIR          := $(ROOT_DIR)/bin
DOCKER_COMPOSE_YML := $(ROOT_DIR)/docker-compose.yml

# Recomiendo usar variables de entorno para la base de datos
DB_URL             := postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSL_MODE}
MIGRATIONS_DIR     := ./migrations_v4
MIGRATIONS_NAME    := $(NAME)  # pasar NAME=nombre al crear

.PHONY: all bin-build run test bin-clean lint \
        build up down logs reset rebuild clean \
        run-api run-ponti-local seed seed-dashboard download-gcp-db \
        migrate-up migrate-down migrate-force migrate-force-dc migrate-version migrate-create \
        db-reset db-migrate-up db-validate db-schema-snapshot db-schema-diff db-verify db-adopt-baseline

# --------------------------------------------------
# Migraciones
# --------------------------------------------------
migrate-up:
	@echo "Running migrations..."
	@bash ./scripts/db/db_migrate_up.sh

migrate-down:
	@echo "Running migrations down..."
	@echo "No hay migrate-down v4 implementado en scripts. Usá el binario migrate manualmente."

# Crea una nueva migración usando la variable NAME
migrate-create:
	@echo "Creating migration $(MIGRATIONS_NAME)..."
	@migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(MIGRATIONS_NAME)

migrate-force:
	@echo "Forcing migration to -1..."
	@$(MIGRATE) force -1

migrate-version:
	@echo "Current migration version:"
	@$(MIGRATE) version

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
		$(ROOT_DIR)/cmd/

run:
	@echo "Running the project (local)..."
	@go run $(ROOT_DIR)/cmd/

test:
	@echo "Running tests..."
	@go test ./...

bin-clean:
	@echo "Cleaning up build artifacts..."
	@rm -rf $(BUILD_DIR)

lint:
	@echo "Linting the project..."
	@golangci-lint run --config .golangci.yml --verbose

run-api:
	@echo "Starting API server..."
	@go run ./cmd/api/

run-ponti-local:
	@echo "Running full local stack (backend + auth + frontend + ai)..."
	@bash ./scripts/run-ponti-local.sh

seed:
	@echo "Seeding database..."
	@go run ./cmd/seed/main.go

seed-dashboard:
	@echo "Seeding dashboard test data..."
	@go run ./cmd/api/main.go seed

# --------------------------------------------------
# Base de datos (descarga GCP DEV)
# --------------------------------------------------
download-gcp-db:
	@echo "Downloading GCP DB and applying business_parameters rename..."
	@set -a && source docs/GCP_DB_CREDS.md && set +a && ./scripts/download-gcp-db.sh

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
# Docker Compose
# --------------------------------------------------
up:
	@echo "Starting services (compose up)..."
	docker compose -f $(DOCKER_COMPOSE_YML) up

down:
	@echo "Stopping services (compose down)..."
	docker compose -f $(DOCKER_COMPOSE_YML) down --remove-orphans

reset: down up
	@echo "Reset done (down & up)."

build:
	@echo "Building services (compose up --build)..."
	docker compose -f $(DOCKER_COMPOSE_YML) up --build

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
	docker compose -f $(DOCKER_COMPOSE_YML) logs -f