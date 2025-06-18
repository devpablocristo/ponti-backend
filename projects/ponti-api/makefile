SHELL := /bin/bash

# Variables
ROOT_DIR := $(shell pwd)
VERSION := 1.0
BUILD_DIR := $(ROOT_DIR)/bin
DOCKER_COMPOSE_DEV := $(ROOT_DIR)/docker-compose.dev.yml
DOCKER_COMPOSE_TEST := $(ROOT_DIR)/docker-compose.test.yml

# Phony targets
.PHONY: all build run test clean lint run-api \
	dev-build dev-up dev-down dev-logs \
	test-build test-up test-down test-logs

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


# SHELL := /bin/bash

# # Variables
# ROOT_DIR := $(shell pwd)
# VERSION := 1.0
# BUILD_DIR := $(ROOT_DIR)/bin
# DOCKER_COMPOSE_DEV := $(ROOT_DIR)/docker-compose.dev.yml
# DOCKER_COMPOSE_test := $(ROOT_DIR)/docker-compose.test.yml

# # Phony targets
# .PHONY: all build run test clean lint \
# 	docker-dev-build docker-dev-up docker-dev-down docker-dev-logs \
# 	docker-test-build docker-test-up docker-test-down docker-test-logs

# # Default target
# all: build run

# # Core commands
# build:
# 	@echo "Building the project..."
# 	@mkdir -p $(BUILD_DIR)
# 	APP_NAME=main go build -gcflags "all=-N -l" -o $(BUILD_DIR)/main -ldflags "-X main.Version=$(VERSION)" $(ROOT_DIR)/cmd/

# run:
# 	@echo "Running the project..."
# 	@go run $(ROOT_DIR)/cmd/

# test:
# 	@echo "Running tests..."
# 	@go test ./...

# clean:
# 	@echo "Cleaning up..."
# 	@rm -f $(BUILD_DIR)/main

# lint:
# 	@echo "Linting the project..."
# 	@golangci-lint run --config .golangci.yml --verbose

# run-api:
# 	@echo "Starting API..."
# 	go run ./cmd/api/

# # Development Docker commands (sin profiles)
# dev-build:
# 	@echo "Building services in dev mode..."
# 	docker compose -f $(DOCKER_COMPOSE_DEV) up --build -d
# 	@$(MAKE) dev-logs

# dev-up:
# 	@echo "Starting services in dev mode..."
# 	docker compose -f $(DOCKER_COMPOSE_DEV) up -d
# 	@$(MAKE) dev-logs

# dev-down:
# 	@echo "Stopping services in dev mode..."
# 	docker compose -f $(DOCKER_COMPOSE_DEV) down --remove-orphans

# dev-logs:
# 	@echo "Fetching logs for dev services..."
# 	docker compose -f $(DOCKER_COMPOSE_DEV) logs -f

# # test Docker commands (sin profiles)
# test-build:
# 	@echo "Building services in test mode..."
# 	docker compose -f $(DOCKER_COMPOSE_test) up --build -d
# 	@$(MAKE) test-logs

# test-up:
# 	@echo "Starting services in test mode..."
# 	docker compose -f $(DOCKER_COMPOSE_test) up -d
# 	@$(MAKE) test-logs

# test-down:
# 	@echo "Stopping services in test mode..."
# 	docker compose -f $(DOCKER_COMPOSE_test) down --remove-orphans

# test-logs:
# 	@echo "Fetching logs for test services..."
# 	docker compose -f $(DOCKER_COMPOSE_test) logs -f