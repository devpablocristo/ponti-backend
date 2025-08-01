SHELL := /bin/bash

# Variables
ROOT_DIR := $(shell pwd)
VERSION := 1.0
BUILD_DIR := $(ROOT_DIR)/bin
DOCKER_COMPOSE_MIX_DEV := $(ROOT_DIR)/docker-compose.mix.dev.yml
DOCKER_COMPOSE_DOCKER_DEV := $(ROOT_DIR)/docker-compose.docker.dev.yml
DOCKER_COMPOSE_DOCKER_STG := $(ROOT_DIR)/docker-compose.docker.stg.yml
DB_URL=postgres://admin:admin@localhost:5432/ponti_api_db?sslmode=disable
MIGRATIONS_DIR := ./migrations
MIGRATE := migrate -path $(MIGRATIONS_DIR) -database $(DB_URL)
MIGRATIONS_NAME=$(name)

.PHONY: all build run test clean lint \
	mix-dev-build mix-dev-up mix-dev-down mix-dev-logs \
	docker-dev-build docker-dev-up docker-dev-down docker-dev-logs \
	docker-stg-build docker-stg-up docker-stg-down docker-stg-logs \
	run-api seed \
	migrate-up migrate-down migrate-force migrate-version migrate-create

# Migrations
migrate-up: 
	@echo "Running migrations..."
	@$(MIGRATE) up

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

all: build run

build:
	@echo "Building the project..."
	@mkdir -p $(BUILD_DIR)
	APP_NAME=ponti-api go build -gcflags "all=-N -l" -o $(BUILD_DIR)/ponti-api -ldflags "-X main.Version=$(VERSION)" $(ROOT_DIR)/cmd/

run:
	@echo "Running the project..."
	@go run $(ROOT_DIR)/cmd/

test:
	@echo "Running tests..."
	@go test ./...

clean:
	@echo "Cleaning up..."
	@rm -rf $(BUILD_DIR)

lint:
	@echo "Linting the project..."
	@golangci-lint run --config .golangci.yml --verbose

run-api:
	@echo "Starting API..."
	go run ./cmd/api/

seed:
	@echo "Seeding database..."
	go run ./cmd/seed/main.go

# Mix Dev Docker commands
mix-dev-build:
	@echo "Building services in mix.dev mode..."
	docker compose -f $(DOCKER_COMPOSE_MIX_DEV) up --build -d
	@$(MAKE) mix-dev-logs

mix-dev-build:
	@echo "Building services in mix.dev mode..."
	docker compose -f $(DOCKER_COMPOSE_MIX_DEV) up --build -d
	@$(MAKE) mix-dev-logs

mix-dev-up:
	@echo "Starting services in mix.dev mode..."
	docker compose -f $(DOCKER_COMPOSE_MIX_DEV) up -d
	@$(MAKE) mix-dev-logs

mix-dev-down:
	@echo "Stopping services in mix.dev mode..."
	docker compose -f $(DOCKER_COMPOSE_MIX_DEV) down --remove-orphans

mix-dev-logs:
	@echo "Fetching logs for mix.dev services..."
	docker compose -f $(DOCKER_COMPOSE_MIX_DEV) logs -f

# Docker Dev commands
docker-dev-build:
	@echo "Building services in docker.dev mode..."
	docker compose -f $(DOCKER_COMPOSE_DOCKER_DEV) up --build -d
	@$(MAKE) docker-dev-logs

docker-dev-up:
	@echo "Starting services in docker.dev mode..."
	docker compose -f $(DOCKER_COMPOSE_DOCKER_DEV) up -d
	@$(MAKE) docker-dev-logs

docker-dev-down:
	@echo "Stopping services in docker.dev mode..."
	docker compose -f $(DOCKER_COMPOSE_DOCKER_DEV) down --remove-orphans

docker-dev-logs:
	@echo "Fetching logs for docker.dev services..."
	docker compose -f $(DOCKER_COMPOSE_DOCKER_DEV) logs -f

# Docker Stg commands
docker-stg-build:
	@echo "Building services in docker.stg mode..."
	docker compose -f $(DOCKER_COMPOSE_DOCKER_STG) up --build -d
	@$(MAKE) docker-stg-logs

docker-stg-up:
	@echo "Starting services in docker.stg mode..."
	docker compose -f $(DOCKER_COMPOSE_DOCKER_STG) up -d
	@$(MAKE) docker-stg-logs

docker-stg-down:
	@echo "Stopping services in docker.stg mode..."
	docker compose -f $(DOCKER_COMPOSE_DOCKER_STG) down --remove-orphans

docker-stg-logs:
	@echo "Fetching logs for docker.stg services..."
	docker compose -f $(DOCKER_COMPOSE_DOCKER_STG) logs -f
