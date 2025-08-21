# 🌱 SEED TOOL MAKEFILE
# Herramienta para poblar la base de datos con datos de prueba

.PHONY: help all reset users types lease-types managers investors units customers campaigns crops projects lots labor-types labor-categories project-investors project-managers invoices project-dollar-values clean build run

# Variables
TOOL_DIR := $(shell pwd)
MAIN_FILE := main.go

# Comandos principales
help: ## Mostrar ayuda
	@echo "🌱 SEED TOOL - Comandos disponibles:"
	@echo ""
	@echo "📋 COMANDOS PRINCIPALES:"
	@echo "  make all          - Ejecutar todos los seeds"
	@echo "  make reset        - Limpiar todos los datos"
	@echo "  make clean        - Limpiar y ejecutar todos los seeds"
	@echo ""
	@echo "🎯 SEEDS INDIVIDUALES:"
	@echo "  make users        - Solo usuarios"
	@echo "  make types        - Solo tipos"
	@echo "  make lease-types  - Solo tipos de arriendo"
	@echo "  make managers     - Solo managers"
	@echo "  make investors    - Solo inversores"
	@echo "  make units        - Solo unidades"
	@echo "  make customers    - Solo clientes"
	@echo "  make campaigns    - Solo campañas"
	@echo "  make crops        - Solo cultivos"
	@echo "  make projects     - Solo proyectos"
	@echo "  make lots         - Solo lotes (incluye campos)"
	@echo "  make labor-types  - Solo tipos de labor"
	@echo "  make labor-categories - Solo categorías de labor"
	@echo "  make project-investors - Solo relaciones proyecto-inversores"
	@echo "  make project-managers - Solo relaciones proyecto-managers"
	@echo "  make invoices     - Solo facturas"
	@echo "  make project-dollar-values - Solo valores del dólar por proyecto"
	@echo ""
	@echo "🔧 UTILIDADES:"
	@echo "  make build        - Compilar el tool"
	@echo "  make run          - Ejecutar con go run"
	@echo "  make tidy         - Limpiar dependencias Go"

# Comandos principales
all: ## Ejecutar todos los seeds
	@echo "🌱 Ejecutando todos los seeds..."
	go run $(MAIN_FILE) -all

reset: ## Limpiar todos los datos
	@echo "🧹 Limpiando todos los datos..."
	go run $(MAIN_FILE) -reset

clean: ## Limpiar y ejecutar todos los seeds
	@echo "🧹 Limpiando y ejecutando todos los seeds..."
	go run $(MAIN_FILE) -reset -all

# Seeds individuales
users: ## Sembrar solo usuarios
	@echo "👥 Sembrando usuarios..."
	go run $(MAIN_FILE) -seed users

types: ## Sembrar solo tipos
	@echo "📝 Sembrando tipos..."
	go run $(MAIN_FILE) -seed types

lease-types: ## Sembrar solo tipos de arriendo
	@echo "📋 Sembrando tipos de arriendo..."
	go run $(MAIN_FILE) -seed lease-types

managers: ## Sembrar solo managers
	@echo "👨‍💼 Sembrando managers..."
	go run $(MAIN_FILE) -seed managers

investors: ## Sembrar solo inversores
	@echo "💰 Sembrando inversores..."
	go run $(MAIN_FILE) -seed investors



units: ## Sembrar solo unidades
	@echo "📏 Sembrando unidades..."
	go run $(MAIN_FILE) -seed units

customers: ## Sembrar solo clientes
	@echo "🏢 Sembrando clientes..."
	go run $(MAIN_FILE) -seed customers

campaigns: ## Sembrar solo campañas
	@echo "📅 Sembrando campañas..."
	go run $(MAIN_FILE) -seed campaigns

crops: ## Sembrar solo cultivos
	@echo "🌾 Sembrando cultivos..."
	go run $(MAIN_FILE) -seed crops

projects: ## Sembrar solo proyectos
	@echo "📁 Sembrando proyectos..."
	go run $(MAIN_FILE) -seed projects

lots: ## Sembrar solo lotes
	@echo "🏞️ Sembrando lotes..."
	go run $(MAIN_FILE) -seed lots

labor-types: ## Sembrar solo tipos de labor
	@echo "🔧 Sembrando tipos de labor..."
	go run $(MAIN_FILE) -seed labor-types

labor-categories: ## Sembrar solo categorías de labor
	@echo "📋 Sembrando categorías de labor..."
	go run $(MAIN_FILE) -seed labor-categories

project-investors: ## Sembrar solo relaciones proyecto-inversores
	@echo "💰 Sembrando relaciones proyecto-inversores..."
	go run $(MAIN_FILE) -seed project-investors

project-managers: ## Sembrar solo relaciones proyecto-managers
	@echo "👨‍💼 Sembrando relaciones proyecto-managers..."
	go run $(MAIN_FILE) -seed project-managers

invoices: ## Sembrar solo facturas
	@echo "🧾 Sembrando facturas..."
	go run $(MAIN_FILE) -seed invoices

project-dollar-values: ## Sembrar solo valores del dólar por proyecto
	@echo "💵 Sembrando valores del dólar por proyecto..."
	go run $(MAIN_FILE) -seed project-dollar-values

# Utilidades
build: ## Compilar el tool
	@echo "🔨 Compilando seed-tool..."
	go build -o seed-tool $(MAIN_FILE)
	@echo "✅ Compilado exitosamente como 'seed-tool'"

run: ## Ejecutar con go run
	@echo "🚀 Ejecutando seed-tool..."
	go run $(MAIN_FILE) -help

tidy: ## Limpiar dependencias Go
	@echo "🧹 Limpiando dependencias Go..."
	go mod tidy
	@echo "✅ Dependencias limpiadas"

# Comandos de desarrollo
dev-setup: ## Configurar entorno de desarrollo
	@echo "⚙️ Configurando entorno de desarrollo..."
	@echo "📝 Asegúrate de tener estas variables de entorno:"
	@echo "   DB_HOST=localhost"
	@echo "   DB_USER=admin"
	@echo "   DB_PASSWORD=admin"
	@echo "   DB_NAME=ponti_api_db"
	@echo "   DB_PORT=5432"
	@echo "   DB_SSL_MODE=disable"
	@echo ""
	@echo "🔧 O ejecuta: source ../../.env"

test-run: ## Ejecutar test básico
	@echo "🧪 Ejecutando test básico..."
	go run $(MAIN_FILE) -users -types -units
	@echo "✅ Test básico completado"

# Comandos de ejemplo
example-basic: ## Ejecutar seeds básicos para testing
	@echo "🎯 Ejecutando seeds básicos para testing..."
	go run $(MAIN_FILE) -seed users -seed types -seed units -seed customers -seed crops -seed projects

example-full: ## Ejecutar seeds completos para desarrollo
	@echo "🚀 Ejecutando seeds completos para desarrollo..."
	go run $(MAIN_FILE) -all

example-dashboard: ## Ejecutar seeds necesarios para dashboard
	@echo "📊 Ejecutando seeds necesarios para dashboard..."
	go run $(MAIN_FILE) -all 