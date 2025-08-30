.PHONY: help all clean

# Variables
BINARY_NAME=seed-dashboard
BUILD_DIR=build

help: ## Mostrar ayuda
	@echo "🌱 SEED DASHBOARD - Comandos disponibles:"
	@echo ""
	@echo "  make all        - Cargar todos los scripts SQL del dashboard usando main.go"
	@echo "  make clean      - Limpiar archivos compilados"
	@echo "  make help       - Mostrar esta ayuda"
	@echo ""

all: ## Cargar todos los scripts SQL del dashboard usando main.go
	@echo "🌱 Cargando scripts SQL del dashboard usando main.go..."
	@go run main.go
	@echo ""
	@echo "✅ Dashboard completo creado con datos ricos y reales"
	@echo "🎯 Prueba el endpoint: GET /api/v1/dashboard"
	@echo "📈 Ahora tendrás:"
	@echo "   - 20 workorders con fechas reales"
	@echo "   - 20 supplies con precios reales"
	@echo "   - 20 labors con contratistas reales"
	@echo "   - 5 inversores con porcentajes"
	@echo "   - Stock valorado en inventario"
	@echo "   - Facturas por montos reales"

clean: ## Limpiar archivos compilados
	@echo "🧹 Limpiando archivos compilados..."
	@rm -rf $(BUILD_DIR)
	@echo "✅ Limpieza completada"

# Comando por defecto
.DEFAULT_GOAL := all
