.PHONY: help all clean

# Variables
BINARY_NAME=seed-dashboard
BUILD_DIR=build

help: ## Mostrar ayuda
	@echo "🌱 SEED DASHBOARD - Comandos disponibles:"
	@echo ""
	@echo "  make all        - Cargar todos los scripts SQL del dashboard"
	@echo "  make clean      - Limpiar archivos compilados"
	@echo "  make help       - Mostrar esta ayuda"
	@echo ""

all: ## Cargar todos los scripts SQL del dashboard
	@echo "🌱 Cargando todos los scripts SQL del dashboard..."
	@go run main.go -reset
	@echo ""
	@echo "✅ Dashboard completo creado con todos los cultivos configurados"
	@echo "🎯 Prueba el endpoint: GET /api/v1/dashboard"

clean: ## Limpiar archivos compilados
	@echo "🧹 Limpiando archivos compilados..."
	@rm -rf $(BUILD_DIR)
	@echo "✅ Limpieza completada"

# Comando por defecto
.DEFAULT_GOAL := all
