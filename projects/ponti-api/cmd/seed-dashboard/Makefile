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
	@echo "🌱 Cargando scripts SQL del dashboard..."
	@echo "📊 Ejecutando 00_base_data.sql..."
	@docker compose exec ponti-db psql -U admin -d ponti_api_db -f /tmp/00_base_data.sql
	@echo "📊 Ejecutando 99_complete_dashboard_data.sql..."
	@docker compose exec ponti-db psql -U admin -d ponti_api_db -f /tmp/99_complete_dashboard_data.sql
	@echo ""
	@echo "✅ Dashboard completo creado con datos ricos y reales"
	@echo "🎯 Prueba el endpoint: GET /api/v1/dashboard"
	@echo "📈 Ahora tendrás:"
	@echo "   - 20 workorders con fechas reales"
	@echo "   - 20 supplies con precios reales"
	@echo "   - 20 labors con contratistas reales"
	@echo "   - 5 inversores con porcentajes"
	@echo "   - Stock valorado en ~$700,000"
	@echo "   - Facturas por $173,000"

clean: ## Limpiar archivos compilados
	@echo "🧹 Limpiando archivos compilados..."
	@rm -rf $(BUILD_DIR)
	@echo "✅ Limpieza completada"

# Comando por defecto
.DEFAULT_GOAL := all
