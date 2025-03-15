SHELL := /bin/bash

# Variables
ROOT_DIR := $(shell pwd)
VERSION := 1.0
BUILD_DIR := $(ROOT_DIR)/bin
DOCKER_COMPOSE_DEV := $(ROOT_DIR)/docker-compose.dev.yml
DOCKER_COMPOSE_STG := $(ROOT_DIR)/docker-compose.stg.yml

# Phony targets declaration
.PHONY: all build run test clean lint \
	docker-dev-build docker-dev-up docker-dev-down docker-dev-logs \
	docker-stg-build docker-stg-up docker-stg-down docker-stg-logs \
	euxcel-api-dev-build euxcel-api-dev-up euxcel-api-dev-down euxcel-api-dev-logs \
	euxcel-api-stg-build euxcel-api-stg-up euxcel-api-stg-down euxcel-api-stg-logs
	
# Default target
all: build run

# Core commands
build:
	@echo "Building the project..."
	@mkdir -p ${BUILD_DIR}
	APP_NAME=main go build -gcflags "all=-N -l" -o ${BUILD_DIR}/${APP_NAME} -ldflags "-X main.Version=${VERSION}" ${ROOT_DIR}/cmd/

run:
	@echo "Running the project..."
	@go run ${ROOT_DIR}/cmd/

test:
	@echo "Running tests..."
	@go test ./...

clean:
	@echo "Cleaning up..."
	@rm -f ${BUILD_DIR}/main

lint:
	@echo "Linting the project..."
	@golangci-lint run --config .golangci.yml --verbose

# Development Docker commands
docker-dev-build:
	@echo "Building $(PROFILE) services in dev mode..."
	docker compose -f $(DOCKER_COMPOSE_DEV) --profile $(PROFILE) up --build -d
	@$(MAKE) docker-dev-logs PROFILE=$(PROFILE)

docker-dev-up:
	@echo "Starting $(PROFILE) services in dev mode..."
	docker compose -f $(DOCKER_COMPOSE_DEV) --profile $(PROFILE) up -d
	@$(MAKE) docker-dev-logs PROFILE=$(PROFILE)

docker-dev-down:
	@echo "Stopping $(PROFILE) services in dev mode..."
	docker compose -f $(DOCKER_COMPOSE_DEV) --profile $(PROFILE) down --remove-orphans

docker-dev-logs:
	@echo "Fetching logs for $(PROFILE) services in dev..."
	docker compose -f $(DOCKER_COMPOSE_DEV) --profile $(PROFILE) logs -f

# Backend profile targets for dev
euxcel-api-dev-build:
	@$(MAKE) docker-dev-build PROFILE=euxcel-api

euxcel-api-dev-up:
	@$(MAKE) docker-dev-up PROFILE=euxcel-api

euxcel-api-dev-down:
	@$(MAKE) docker-dev-down PROFILE=euxcel-api

euxcel-api-dev-logs:
	@$(MAKE) docker-dev-logs PROFILE=euxcel-api

# ------------------------------------------------------------------------
# Staging Docker commands
# ------------------------------------------------------------------------

docker-stg-build:
	@echo "Building $(PROFILE) services in staging mode..."
	docker compose -f $(DOCKER_COMPOSE_STG) --profile $(PROFILE) up --build -d
	@$(MAKE) docker-stg-logs PROFILE=$(PROFILE)

docker-stg-up:
	@echo "Starting $(PROFILE) services in staging mode..."
	docker compose -f $(DOCKER_COMPOSE_STG) --profile $(PROFILE) up -d
	@$(MAKE) docker-stg-logs PROFILE=$(PROFILE)

docker-stg-down:
	@echo "Stopping $(PROFILE) services in staging mode..."
	docker compose -f $(DOCKER_COMPOSE_STG) --profile $(PROFILE) down --remove-orphans

docker-stg-logs:
	@echo "Fetching logs for $(PROFILE) services in staging..."
	docker compose -f $(DOCKER_COMPOSE_STG) --profile $(PROFILE) logs -f

# Backend profile targets for staging
euxcel-api-stg-build:
	@$(MAKE) docker-stg-build PROFILE=euxcel-api

euxcel-api-stg-up:
	@$(MAKE) docker-stg-up PROFILE=euxcel-api

euxcel-api-stg-down:
	@$(MAKE) docker-stg-down PROFILE=euxcel-api

euxcel-api-stg-logs:
	@$(MAKE) docker-stg-logs PROFILE=euxcel-api
