SHELL := /bin/bash

# Variables base
ROOT_DIR           := $(shell pwd)
VERSION            := 1.0
BUILD_DIR          := $(ROOT_DIR)/bin
DOCKER_COMPOSE_YML := $(ROOT_DIR)/docker-compose.yml
GO_MODULES_TOKEN   ?=
PONTI_ENV_FILE     ?=
PONTI_LOCAL_ENV    ?=

# Recomiendo usar variables de entorno para la base de datos
DB_URL             := postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSL_MODE}
MIGRATIONS_DIR     := ./migrations_v4
MIGRATIONS_NAME    := $(NAME)  # pasar NAME=nombre al crear

.PHONY: all bin-build run test bin-clean lint \
        build up down logs reset rebuild clean docker-cleanup dev dev-logs \
        dev-main dev-develop run-api up-ponti-local up-ponti-stg-local up-ponti-dev-local down-ponti-local seed seed-dashboard db-staging-to-local db-staging-to-main-local db-reset-from-staging db-main-reset-from-staging db-reset-main db-reset-develop db-migrate-up-main db-migrate-up-develop db-copy-current-dev-to-develop-local staging-db-2-dev-db e2e-changes \
        migrate-create \
        db-reset db-migrate-up db-validate db-schema-snapshot db-schema-diff db-verify db-adopt-baseline db-force-reset-gcp db-gcp-reset-and-load-local

define compose_cmd
PONTI_ENV_FILE="$(PONTI_ENV_FILE)" PONTI_LOCAL_ENV="$(PONTI_LOCAL_ENV)" GO_MODULES_TOKEN="$(GO_MODULES_TOKEN)" $(ROOT_DIR)/scripts/compose_with_env.sh
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
	@if [ -f ./scripts/run-ponti-local.sh ]; then PONTI_ENV_FILE="$(PONTI_ENV_FILE)" PONTI_LOCAL_ENV="$(PONTI_LOCAL_ENV)" bash ./scripts/run-ponti-local.sh; else PONTI_ENV_FILE="$(PONTI_ENV_FILE)" PONTI_LOCAL_ENV="$(PONTI_LOCAL_ENV)" bash ./scripts/run_ponti_local.sh; fi

up-ponti-stg-local:
	@PONTI_LOCAL_ENV=stg $(MAKE) up-ponti-local

up-ponti-dev-local:
	@PONTI_LOCAL_ENV=dev $(MAKE) up-ponti-local

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
	@PONTI_ENV_FILE="$(PONTI_ENV_FILE)" $(ROOT_DIR)/scripts/compose_with_env.sh up -d ponti-db 2>/dev/null || true
	@bash ./scripts/db/db_staging_to_local.sh

db-staging-to-main-local:
	@PONTI_ENV_FILE=.env.main.local $(MAKE) db-staging-to-local

# Reset local DB + migraciones + carga data-only desde STAGING
db-reset-from-staging: db-reset db-migrate-up db-staging-to-local
	@echo "Local DB reset + migrate + staging data-only restore completed."

db-main-reset-from-staging:
	@PONTI_ENV_FILE=.env.main.local $(MAKE) db-reset
	@PONTI_ENV_FILE=.env.main.local $(MAKE) db-migrate-up
	@PONTI_ENV_FILE=.env.main.local $(MAKE) db-staging-to-local
	@echo "Main/stg local DB reset + migrate + staging data-only restore completed."

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
	@PONTI_ENV_FILE="$(PONTI_ENV_FILE)" bash ./scripts/db/db_ensure_exists.sh
	$(compose_cmd) up --build ponti-api

dev-main:
	@PONTI_ENV_FILE=.env.main.local $(MAKE) dev

dev-develop:
	@PONTI_ENV_FILE=.env.develop.local $(MAKE) dev

dev-logs:
	$(compose_cmd) logs -f ponti-api

db-reset-main:
	@PONTI_ENV_FILE=.env.main.local $(MAKE) db-reset

db-reset-develop:
	@PONTI_ENV_FILE=.env.develop.local $(MAKE) db-reset

db-migrate-up-main:
	@PONTI_ENV_FILE=.env.main.local $(MAKE) db-migrate-up

db-migrate-up-develop:
	@PONTI_ENV_FILE=.env.develop.local $(MAKE) db-migrate-up

db-copy-current-dev-to-develop-local:
	@bash ./scripts/db/db_copy_local.sh new_ponti_db_dev new_ponti_db_develop_local

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
