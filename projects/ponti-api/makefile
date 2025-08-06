SHELL := /bin/bash

# Variables base (ahora más parametrizables)
ROOT_DIR           := $(shell pwd)
VERSION            := 1.0
BUILD_DIR          := $(ROOT_DIR)/bin
DOCKER_COMPOSE_DEV := $(ROOT_DIR)/docker-compose.mix.dev.yml
DOCKER_COMPOSE_STG := $(ROOT_DIR)/docker-compose.mix.stg.yml   # renombrado para consistencia
# Recomiendo usar variables de entorno en vez de hardcodear “localhost”
DB_URL             := postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSL_MODE}
MIGRATIONS_DIR     := ./migrations
# MIGRATIONS_NAME estaba usando $(name) que no viene definido
MIGRATIONS_NAME    := $(NAME)  # deberías pasar NAME=name_al_migrar o cambiar aquí

.PHONY: all build run test clean lint \
        mix-dev-build mix-dev-up mix-dev-down mix-dev-logs \
        mix-stg-build mix-stg-up mix-stg-down mix-stg-logs \
        run-api seed \
        migrate-up migrate-down migrate-force migrate-version migrate-create

# --------------------------------------------------
# Migrations
# --------------------------------------------------
migrate-up:
	@echo "Running migrations..."
	@$(MIGRATE) up

migrate-down:
	@echo "Running migrations down..."
	@$(MIGRATE) down

# Usamos la var NAME en lugar de name
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
# Build & Run
# --------------------------------------------------
all: build run

build:
	@echo "Building the project..."
	@mkdir -p $(BUILD_DIR)
	APP_NAME=ponti-api go build -gcflags "all=-N -l" \
		-o $(BUILD_DIR)/ponti-api \
		-ldflags "-X main.Version=$(VERSION)" \
		$(ROOT_DIR)/cmd/

run:
	@echo "Running the project (local)..."
	@go run $(ROOT_DIR)/cmd/

test:
	@echo "Running tests..."
	@go test ./...

clean:
	@echo "Cleaning up build artifacts..."
	@rm -rf $(BUILD_DIR)

lint:
	@echo "Linting the project..."
	@golangci-lint run --config .golangci.yml --verbose

run-api:
	@echo "Starting API server..."
	@go run ./cmd/api/

seed:
	@echo "Seeding database..."
	@go run ./cmd/seed/main.go

# --------------------------------------------------
# Docker Compose: Mix Dev
# --------------------------------------------------
mix-dev-build:
	@echo "Building services in mix.dev mode..."
	docker compose -f $(DOCKER_COMPOSE_DEV) up --build -d
	@$(MAKE) mix-dev-logs

mix-dev-up:
	@echo "Starting services in mix.dev mode..."
	docker compose -f $(DOCKER_COMPOSE_DEV) up -d
	@$(MAKE) mix-dev-logs

mix-dev-down:
	@echo "Stopping services in mix.dev mode..."
	docker compose -f $(DOCKER_COMPOSE_DEV) down --remove-orphans

mix-dev-logs:
	@echo "Tailing logs for mix.dev services..."
	docker compose -f $(DOCKER_COMPOSE_DEV) logs -f

# --------------------------------------------------
# Docker Compose: Staging (Mix Stg)
# --------------------------------------------------
mix-stg-build:
	@echo "Building services in mix-stg mode..."
	docker compose -f $(DOCKER_COMPOSE_STG) up --build -d
	@$(MAKE) mix-stg-logs

mix-stg-up:
	@echo "Starting services in mix-stg mode..."
	docker compose -f $(DOCKER_COMPOSE_STG) up -d
	@$(MAKE) mix-stg-logs

mix-stg-down:
	@echo "Stopping services in mix-stg mode..."
	docker compose -f $(DOCKER_COMPOSE_STG) down --remove-orphans

mix-stg-logs:
	@echo "Tailing logs for mix-stg services..."
	docker compose -f $(DOCKER_COMPOSE_STG) logs -f
