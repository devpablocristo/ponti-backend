SHELL := /bin/bash

# Variables base
ROOT_DIR           := $(shell pwd)
VERSION            := 1.0
BUILD_DIR          := $(ROOT_DIR)/bin
DOCKER_COMPOSE_YML := $(ROOT_DIR)/docker-compose.yml
GO_MODULES_TOKEN   ?=

# Recomiendo usar variables de entorno para la base de datos
DB_URL             := postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSL_MODE}
MIGRATIONS_DIR     := ./migrations_v4
MIGRATIONS_NAME    := $(NAME)  # pasar NAME=nombre al crear

.PHONY: all bin-build run test bin-clean lint \
        build up down logs reset rebuild clean docker-cleanup dev dev-logs \
        run-api up-ponti-local down-ponti-local seed seed-dashboard db-staging-to-local db-reset-from-staging staging-db-2-dev-db e2e-changes \
        migrate-create \
        db-reset db-migrate-up db-validate db-schema-snapshot db-schema-diff db-verify db-adopt-baseline db-force-reset-gcp db-gcp-reset-and-load-local

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

up-ponti-local:
	@echo "Starting full local stack (backend + frontend + ai)..."
	@if [ -f ./scripts/run-ponti-local.sh ]; then bash ./scripts/run-ponti-local.sh; else bash ./scripts/run_ponti_local.sh; fi

down-ponti-local:
	@echo "Stopping full local stack (backend + frontend + ai)..."
	@bash ./scripts/down_ponti_local.sh

seed:
	@echo "Seeding database..."
	@go run ./cmd/seed/main.go

seed-dashboard:
	@echo "Seeding dashboard test data..."
	@go run ./cmd/api/main.go seed

# --------------------------------------------------
# Base de datos (descarga GCP STAGING → local, data-only)
# --------------------------------------------------
db-staging-to-local:
	@echo "Downloading GCP STAGING and restoring data-only to local..."
	@echo "Asegurando que la DB local esté levantada..."
	@echo "Tip: si no seteás SRC_PASS, el script intenta leer db-password-dev desde Secret Manager (requiere gcloud auth)."
	@docker compose -f $(DOCKER_COMPOSE_YML) up -d ponti-db 2>/dev/null || true
	@bash ./scripts/db/db_staging_to_local.sh

# Reset local DB + migraciones + carga data-only desde STAGING
db-reset-from-staging: db-reset db-migrate-up db-staging-to-local
	@echo "Local DB reset + migrate + staging data-only restore completed."

# Copia datos GCP STAGING → GCP DEV (data-only). Requiere scripts/staging_db_2_dev_db.env.
staging-db-2-dev-db:
	@set -a && [ -f scripts/staging_db_2_dev_db.env ] && source scripts/staging_db_2_dev_db.env; set +a && \
	bash ./scripts/staging_db_2_dev_db.sh

# Smoke tests de release (incluye divisor de aportes). Uso: make e2e-changes [BASE_URL=http://...]
e2e-changes:
	@bash ./scripts/smoke_release.sh $(BASE_URL)

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

# Fuerza reset de la DB en GCP (DROP schema public + migraciones). Requiere scripts/db/db_force_reset_gcp.env.
db-force-reset-gcp:
	@bash ./scripts/db/db_force_reset_gcp.sh

# Después del merge: reset GCP + migraciones + cargar datos desde DB local. Requiere scripts/db/db_gcp_reset_and_load_local.env.
db-gcp-reset-and-load-local:
	@bash ./scripts/db/db_gcp_reset_and_load_local.sh

# --------------------------------------------------
# Desarrollo con hot reload (Air)
# --------------------------------------------------
dev:
	@echo "Starting dev environment with hot reload (Air)..."
	@if [ -z "$(GO_MODULES_TOKEN)" ]; then echo "WARN: GO_MODULES_TOKEN vacio. Si el cache de Go esta frio, ponti-api no podra bajar github.com/devpablocristo/core/*."; fi
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
	@if [ -z "$(GO_MODULES_TOKEN)" ]; then echo "WARN: GO_MODULES_TOKEN vacio. Si el cache de Go esta frio, ponti-api no podra bajar github.com/devpablocristo/core/*."; fi
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
