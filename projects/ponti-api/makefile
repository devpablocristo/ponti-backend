SHELL := /bin/bash

# Variables
ROOT_DIR := $(shell pwd)
VERSION := 1.0
BUILD_DIR := $(ROOT_DIR)/bin
DOCKER_COMPOSE_DEV := $(ROOT_DIR)/docker-compose.dev.yml
DOCKER_COMPOSE_STG := $(ROOT_DIR)/docker-compose.stg.yml
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

# Development Docker commands (sin profiles)
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

# Staging Docker commands (sin profiles)
stg-build:
	@echo "Building services in staging mode..."
	docker compose -f $(DOCKER_COMPOSE_STG) up --build -d
	@$(MAKE) stg-logs

stg-up:
	@echo "Starting services in staging mode..."
	docker compose -f $(DOCKER_COMPOSE_STG) up -d
	@$(MAKE) stg-logs

stg-down:
	@echo "Stopping services in staging mode..."
	docker compose -f $(DOCKER_COMPOSE_STG) down --remove-orphans

stg-logs:
	@echo "Fetching logs for staging services..."
	docker compose -f $(DOCKER_COMPOSE_STG) logs -f


# SHELL := /bin/bash

# # Variables
# ROOT_DIR := $(shell pwd)
# VERSION := 1.0
# BUILD_DIR := $(ROOT_DIR)/bin
# DOCKER_COMPOSE_DEV := $(ROOT_DIR)/docker-compose.dev.yml
# DOCKER_COMPOSE_STG := $(ROOT_DIR)/docker-compose.stg.yml

# # Phony targets declaration
# .PHONY: all build run test clean lint \
# 	docker-dev-build docker-dev-up docker-dev-down docker-dev-logs \
# 	docker-stg-build docker-stg-up docker-stg-down docker-stg-logs \
# 	dev-build dev-up dev-down dev-logs \
# 	stg-build stg-up stg-down stg-logs
	
# # Default target
# all: build run

# # Core commands
# build:
# 	@echo "Building the project..."
# 	@mkdir -p ${BUILD_DIR}
# 	APP_NAME=main go build -gcflags "all=-N -l" -o ${BUILD_DIR}/${APP_NAME} -ldflags "-X main.Version=${VERSION}" ${ROOT_DIR}/cmd/

# run:
# 	@echo "Running the project..."
# 	@go run ${ROOT_DIR}/cmd/

# test:
# 	@echo "Running tests..."
# 	@go test ./...

# clean:
# 	@echo "Cleaning up..."
# 	@rm -f ${BUILD_DIR}/main

# lint:
# 	@echo "Linting the project..."
# 	@golangci-lint run --config .golangci.yml --verbose

# # Development Docker commands
# docker-dev-build:
# 	@echo "Building $(PROFILE) services in dev mode..."
# 	docker compose -f $(DOCKER_COMPOSE_DEV) --profile $(PROFILE) up --build -d
# 	@$(MAKE) docker-dev-logs PROFILE=$(PROFILE)

# docker-dev-up:
# 	@echo "Starting $(PROFILE) services in dev mode..."
# 	docker compose -f $(DOCKER_COMPOSE_DEV) --profile $(PROFILE) up -d
# 	@$(MAKE) docker-dev-logs PROFILE=$(PROFILE)

# docker-dev-down:
# 	@echo "Stopping $(PROFILE) services in dev mode..."
# 	docker compose -f $(DOCKER_COMPOSE_DEV) --profile $(PROFILE) down --remove-orphans

# docker-dev-logs:
# 	@echo "Fetching logs for $(PROFILE) services in dev..."
# 	docker compose -f $(DOCKER_COMPOSE_DEV) --profile $(PROFILE) logs -f

# # Backend profile targets for dev
# dev-build:
# 	@$(MAKE) docker-dev-build PROFILE=ponti-api

# dev-up:
# 	@$(MAKE) docker-dev-up PROFILE=ponti-api

# dev-down:
# 	@$(MAKE) docker-dev-down PROFILE=ponti-api

# dev-logs:
# 	@$(MAKE) docker-dev-logs PROFILE=ponti-api

# # ------------------------------------------------------------------------
# # Staging Docker commands
# # ------------------------------------------------------------------------

# docker-stg-build:
# 	@echo "Building $(PROFILE) services in staging mode..."
# 	docker compose -f $(DOCKER_COMPOSE_STG) --profile $(PROFILE) up --build -d
# 	@$(MAKE) docker-stg-logs PROFILE=$(PROFILE)

# docker-stg-up:
# 	@echo "Starting $(PROFILE) services in staging mode..."
# 	docker compose -f $(DOCKER_COMPOSE_STG) --profile $(PROFILE) up -d
# 	@$(MAKE) docker-stg-logs PROFILE=$(PROFILE)

# docker-stg-down:
# 	@echo "Stopping $(PROFILE) services in staging mode..."
# 	docker compose -f $(DOCKER_COMPOSE_STG) --profile $(PROFILE) down --remove-orphans

# docker-stg-logs:
# 	@echo "Fetching logs for $(PROFILE) services in staging..."
# 	docker compose -f $(DOCKER_COMPOSE_STG) --profile $(PROFILE) logs -f

# # Backend profile targets for staging
# stg-build:
# 	@$(MAKE) docker-stg-build PROFILE=ponti-api

# stg-up:
# 	@$(MAKE) docker-stg-up PROFILE=ponti-api

# stg-down:
# 	@$(MAKE) docker-stg-down PROFILE=ponti-api

# stg-logs:
# 	@$(MAKE) docker-stg-logs PROFILE=ponti-api
