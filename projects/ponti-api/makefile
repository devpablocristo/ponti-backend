SHELL := /bin/bash

# Variables
ROOT_DIR := $(shell pwd)
BUILD_DIR := $(ROOT_DIR)/bin
COMPOSE_DEV := $(ROOT_DIR)/docker-compose.dev.yml
COMPOSE_STG := $(ROOT_DIR)/docker-compose.stg.yml
DB_URL=postgres://admin:admin@localhost:5432/ponti_api_db?sslmode=disable
MIGRATIONS_DIR := ./migrations
MIGRATE := migrate -path $(MIGRATIONS_DIR) -database $(DB_URL)
MIGRATIONS_NAME=$(name)

.PHONY: all build run test clean lint \
	dev-build dev-up dev-down dev-reset dev-logs \
	stg-build stg-up stg-down stg-logs \
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

# Default target
all: build run

# --- Go Core ---
build:
	@echo "🔨 Build..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/main -gcflags "all=-N -l" -ldflags "-X main.Version=$(VERSION)" ./cmd/

run:
	@echo "🚀 Run..."
	go run ./cmd/

test:
	@echo "🧪 Test..."
	go test ./...

clean:
	@echo "🧹 Clean..."
	rm -f $(BUILD_DIR)/main

lint:
	@echo "🔍 Lint..."
	golangci-lint run --config .golangci.yml --verbose

run-api:
	@echo "🚦 Run API..."
	go run ./cmd/api/

# --- Docker Compose Helpers ---

define compose_cmd
	@echo "$(2) services ($(1))..."
	docker compose -f $($(1)) $(3)
endef

# --- Dev ---
dev-build:
	$(call compose_cmd,COMPOSE_DEV,Building,up --build -d)
	@$(MAKE) dev-logs

dev-up:
	$(call compose_cmd,COMPOSE_DEV,Starting,up -d)
	@$(MAKE) dev-logs

dev-down:
	$(call compose_cmd,COMPOSE_DEV,Stopping,down --remove-orphans)

dev-reset:
	@echo "⚠️  Limpiando volúmenes y recreando (esto borra la base de datos)..."
	docker compose -f $(COMPOSE_DEV) down -v --remove-orphans
	docker compose -f $(COMPOSE_DEV) up --build -d
	@$(MAKE) dev-logs

dev-logs:
	$(call compose_cmd,COMPOSE_DEV,Logs,logs -f)

# --- Stg/Test ---
stg-build:
	$(call compose_cmd,COMPOSE_STG,Building,up --build -d)
	@$(MAKE) stg-logs

stg-up:
	$(call compose_cmd,COMPOSE_STG,Starting,up -d)
	@$(MAKE) stg-logs

stg-down:
	$(call compose_cmd,COMPOSE_STG,Stopping,down --remove-orphans)

stg-logs:
	$(call compose_cmd,COMPOSE_STG,Logs,logs -f)
