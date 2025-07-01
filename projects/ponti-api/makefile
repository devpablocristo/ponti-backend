SHELL := /bin/bash

# Variables
ROOT_DIR := $(shell pwd)
VERSION := 1.0
BUILD_DIR := $(ROOT_DIR)/bin
DOCKER_COMPOSE_DEV := $(ROOT_DIR)/docker-compose.dev.yml
DOCKER_COMPOSE_TEST := $(ROOT_DIR)/docker-compose.test.yml
DB_URL=postgres://admin:admin@localhost:5432/ponti_api_db?sslmode=disable
MIGRATIONS_DIR := ./migrations
MIGRATE := migrate -path $(MIGRATIONS_DIR) -database $(DB_URL)
MIGRATIONS_NAME=$(name)

# Phony targets
.PHONY: all build run test clean lint \
	docker-dev-build docker-dev-up docker-dev-down docker-dev-logs \
	docker-stg-build docker-stg-up docker-stg-down docker-stg-logs \
	migrate-up migrate-down migrate-force migrate-version

# Migrations
# migrate -path ./migrations -database "postgres://admin:admin@localhost:5432/ponti_api_db?sslmode=disable" up
migrate-up: 
	@echo "Running migrations..."
	@$(MIGRATE) up

# migrate -path ./migrations -database "postgres://admin:admin@localhost:5432/ponti_api_db?sslmode=disable" down 1
migrate-down: 
	@echo "Running migrations down..."
	@$(MIGRATE) down

migrate-create:
	@echo "Creating migration..."
	@migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(MIGRATIONS_NAME)

migrate-force:
	@echo "Forcing migration..."
	@$(MIGRATE) force -1

migrate-version:
	@echo "Getting migration version..."
	@$(MIGRATE) version

# Default target
all: build run

# Core commands
build:
	@echo "Building the project..."
	@mkdir -p $(BUILD_DIR)
	APP_NAME=main go build -gcflags "all=-N -l" -o $(BUILD_DIR)/main -ldflags "-X main.Version=$(VERSION)" $(ROOT_DIR)/cmd/

run:
	@echo "Running the project..."
	@go run $(ROOT_DIR)/cmd/

test:
	@echo "Running tests..."
	@go test ./...

clean:
	@echo "Cleaning up..."
	@rm -f $(BUILD_DIR)/main

lint:
	@echo "Linting the project..."
	@golangci-lint run --config .golangci.yml --verbose

run-api:
	@echo "Starting API..."
	go run ./cmd/api/

# ----------- Docker Compose Commands -----------

# Development
dev-build:
	@echo "Building services in dev mode..."
	docker compose -f $(DOCKER_COMPOSE_DEV) up --build -d
	@$(MAKE) dev-logs

dev-up:
	@echo "Starting services in dev mode..."
	docker compose -f $(DOCKER_COMPOSE_DEV) up -d
	@$(MAKE) dev-logs

dev-down:
	@echo "Stopping services in dev mode..."
	docker compose -f $(DOCKER_COMPOSE_DEV) down --remove-orphans

dev-logs:
	@echo "Fetching logs for dev services..."
	docker compose -f $(DOCKER_COMPOSE_DEV) logs -f

# Test
test-build:
	@echo "Building services in test mode..."
	docker compose -f $(DOCKER_COMPOSE_TEST) up --build -d
	@$(MAKE) test-logs

test-up:
	@echo "Starting services in test mode..."
	docker compose -f $(DOCKER_COMPOSE_TEST) up -d
	@$(MAKE) test-logs

test-down:
	@echo "Stopping services in test mode..."
	docker compose -f $(DOCKER_COMPOSE_TEST) down --remove-orphans

test-logs:
	@echo "Fetching logs for test services..."
	docker compose -f $(DOCKER_COMPOSE_TEST) logs -f