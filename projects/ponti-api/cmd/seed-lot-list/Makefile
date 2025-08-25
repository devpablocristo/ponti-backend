# =======================
# SEED TOOL 2 - MAKEFILE
# =======================
# Makefile específico para generar datos de prueba para LIST LOT

.PHONY: help build run clean reset all users types lease-types managers investors categories customers campaigns crops projects project-managers project-investors fields lots supplies labors workorders tons crop-commercializations lot-dates verification

# Variables
BINARY_NAME=seed-tool2
BUILD_DIR=build
SQL_DIR=sql

# Colores para output
GREEN=\033[0;32m
YELLOW=\033[1;33m
BLUE=\033[0;34m
RED=\033[0;31m
NC=\033[0m # No Color

# Ayuda
help:
	@echo "$(BLUE)🌱 SEED TOOL 2 - Generador de datos de prueba para LIST LOT$(NC)"
	@echo ""
	@echo "$(YELLOW)COMANDOS DISPONIBLES:$(NC)"
	@echo "  $(GREEN)make help$(NC)                    - Mostrar esta ayuda"
	@echo "  $(GREEN)make build$(NC)                   - Compilar la aplicación"
	@echo "  $(GREEN)make clean$(NC)                   - Limpiar archivos compilados"
	@echo ""
	@echo "$(YELLOW)EJECUCIÓN COMPLETA:$(NC)"
	@echo "  $(GREEN)make all$(NC)                     - Ejecutar todos los seeds para LIST LOT"
	@echo "  $(GREEN)make reset$(NC)                   - Limpiar base de datos"
	@echo "  $(GREEN)make reset-all$(NC)               - Limpiar y ejecutar todos los seeds"
	@echo ""
	@echo "$(YELLOW)SEEDS INDIVIDUALES:$(NC)"
	@echo "  $(GREEN)make users$(NC)                   - Crear usuario demo"
	@echo "  $(GREEN)make types$(NC)                   - Crear tipos base"
	@echo "  $(GREEN)make lease-types$(NC)             - Crear tipos de arriendo"
	@echo "  $(GREEN)make managers$(NC)                - Crear managers"
	@echo "  $(GREEN)make investors$(NC)               - Crear investors"
	@echo "  $(GREEN)make categories$(NC)              - Crear categorías"

	@echo "  $(GREEN)make customers$(NC)               - Crear customers"
	@echo "  $(GREEN)make campaigns$(NC)               - Crear campaigns"
	@echo "  $(GREEN)make crops$(NC)                   - Crear crops"
	@echo "  $(GREEN)make projects$(NC)                - Crear proyecto principal"
	@echo "  $(GREEN)make project-managers$(NC)        - Asociar managers"
	@echo "  $(GREEN)make project-investors$(NC)       - Asociar investors"
	@echo "  $(GREEN)make fields$(NC)                  - Crear campos"
	@echo "  $(GREEN)make lots$(NC)                    - Crear lotes"
	@echo "  $(GREEN)make supplies$(NC)                - Crear supplies"
	@echo "  $(GREEN)make labors$(NC)                  - Crear labores"
	@echo "  $(GREEN)make workorders$(NC)              - Crear workorders"
	@echo "  $(GREEN)make tons$(NC)                    - Actualizar toneladas"
	@echo "  $(GREEN)make crop-commercializations$(NC) - Crear precios"
	@echo "  $(GREEN)make lot-dates$(NC)               - Crear fechas"
	@echo "  $(GREEN)make verification$(NC)            - Verificar datos"
	@echo ""
	@echo "$(YELLOW)ORDEN RECOMENDADO:$(NC)"
	@echo "  1. make reset (opcional)"
	@echo "  2. make users types lease-types managers investors categories"
	@echo "  3. make customers campaigns crops"
	@echo "  4. make projects project-managers project-investors"
	@echo "  5. make fields lots supplies labors"
	@echo "  6. make workorders tons crop-commercializations lot-dates"
	@echo "  7. make verification"
	@echo ""
	@echo "$(YELLOW)RESULTADO ESPERADO:$(NC)"
	@echo "  - 3 lots con datos diferenciados"
	@echo "  - Cálculos coherentes en lot_table_view"
	@echo "  - Valores realistas para testing de LIST LOT"

# Compilar
build:
	@echo "$(BLUE)🔨 Compilando SEED TOOL 2...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) main.go
	@echo "$(GREEN)✅ Compilación completada: $(BUILD_DIR)/$(BINARY_NAME)$(NC)"

# Limpiar
clean:
	@echo "$(YELLOW)🧹 Limpiando archivos compilados...$(NC)"
	@rm -rf $(BUILD_DIR)
	@echo "$(GREEN)✅ Limpieza completada$(NC)"

# Ejecutar todos los seeds
all:
	@echo "$(BLUE)🌱 Ejecutando todos los seeds para LIST LOT...$(NC)"
	@go run main.go -all

# Limpiar base de datos
reset:
	@echo "$(YELLOW)🧹 Limpiando base de datos...$(NC)"
	@go run main.go -reset

# Limpiar y ejecutar todo
reset-all:
	@echo "$(BLUE)🔄 Limpiando y ejecutando todos los seeds...$(NC)"
	@go run main.go -reset -all

# Seeds individuales
users:
	@echo "$(BLUE)👤 Creando usuario demo...$(NC)"
	@go run main.go -seed users

types:
	@echo "$(BLUE)🏷️ Creando tipos base...$(NC)"
	@go run main.go -seed types

lease-types:
	@echo "$(BLUE)🏠 Creando tipos de arriendo...$(NC)"
	@go run main.go -seed lease-types

managers:
	@echo "$(BLUE)👨‍💼 Creando managers...$(NC)"
	@go run main.go -seed managers

investors:
	@echo "$(BLUE)💰 Creando investors...$(NC)"
	@go run main.go -seed investors

categories:
	@echo "$(BLUE)🏷️ Creando categorías...$(NC)"
	@go run main.go -seed categories

customers:
	@echo "$(BLUE)👥 Creando customers...$(NC)"
	@go run main.go -seed customers

campaigns:
	@echo "$(BLUE)📢 Creando campaigns...$(NC)"
	@go run main.go -seed campaigns

crops:
	@echo "$(BLUE)🌾 Creando crops...$(NC)"
	@go run main.go -seed crops

projects:
	@echo "$(BLUE)🏗️ Creando proyecto principal...$(NC)"
	@go run main.go -seed projects

project-managers:
	@echo "$(BLUE)🔗 Asociando managers...$(NC)"
	@go run main.go -seed project-managers

project-investors:
	@echo "$(BLUE)🔗 Asociando investors...$(NC)"
	@go run main.go -seed project-investors

fields:
	@echo "$(BLUE)🌍 Creando campos...$(NC)"
	@go run main.go -seed fields

lots:
	@echo "$(BLUE)📦 Creando lotes...$(NC)"
	@go run main.go -seed lots

supplies:
	@echo "$(BLUE)📦 Creando supplies...$(NC)"
	@go run main.go -seed supplies

labors:
	@echo "$(BLUE)🔧 Creando labores...$(NC)"
	@go run main.go -seed labors

workorders:
	@echo "$(BLUE)📋 Creando workorders...$(NC)"
	@go run main.go -seed workorders

tons:
	@echo "$(BLUE)⚖️ Actualizando toneladas...$(NC)"
	@go run main.go -seed tons

crop-commercializations:
	@echo "$(BLUE)💰 Creando precios de comercialización...$(NC)"
	@go run main.go -seed crop-commercializations

lot-dates:
	@echo "$(BLUE)📅 Creando fechas de lotes...$(NC)"
	@go run main.go -seed lot-dates

verification:
	@echo "$(BLUE)✅ Verificando datos...$(NC)"
	@go run main.go -seed verification

# Comandos de desarrollo
dev-setup:
	@echo "$(BLUE)🚀 Configurando entorno de desarrollo...$(NC)"
	@go mod tidy
	@go mod download
	@echo "$(GREEN)✅ Entorno configurado$(NC)"

test:
	@echo "$(BLUE)🧪 Ejecutando pruebas...$(NC)"
	@go test ./...
	@echo "$(GREEN)✅ Pruebas completadas$(NC)"

# Comandos de verificación
check-sql:
	@echo "$(BLUE)🔍 Verificando archivos SQL...$(NC)"
	@ls -la $(SQL_DIR)/*.sql
	@echo "$(GREEN)✅ Verificación completada$(NC)"

list-seeds:
	@echo "$(BLUE)📋 Listando seeds disponibles...$(NC)"
	@go run main.go -help

# Comandos de utilidad
status:
	@echo "$(BLUE)📊 Estado del proyecto...$(NC)"
	@echo "Archivos SQL: $(shell ls $(SQL_DIR)/*.sql | wc -l)"
	@echo "Tamaño total: $(shell du -sh . | cut -f1)"
	@echo "Última modificación: $(shell date -r main.go)"

# Comando para ejecutar secuencia completa en orden
sequence:
	@echo "$(BLUE)🔄 Ejecutando secuencia completa para LIST LOT...$(NC)"
	@echo "$(YELLOW)1. Limpiando base de datos...$(NC)"
	@go run main.go -reset
	@echo "$(YELLOW)2. Creando datos base...$(NC)"
	@go run main.go -seed users
	@go run main.go -seed types
	@go run main.go -seed lease-types
	@go run main.go -seed managers
	@go run main.go -seed investors
	@go run main.go -seed categories

	@echo "$(YELLOW)3. Creando entidades principales...$(NC)"
	@go run main.go -seed customers
	@go run main.go -seed campaigns
	@go run main.go -seed crops
	@go run main.go -seed projects
	@go run main.go -seed project-managers
	@go run main.go -seed project-investors
	@echo "$(YELLOW)4. Creando estructura agrícola...$(NC)"
	@go run main.go -seed fields
	@go run main.go -seed lots
	@go run main.go -seed supplies
	@go run main.go -seed labors
	@echo "$(YELLOW)5. Creando datos operativos...$(NC)"
	@go run main.go -seed workorders
	@go run main.go -seed tons
	@go run main.go -seed crop-commercializations
	@go run main.go -seed lot-dates
	@echo "$(YELLOW)6. Verificando datos...$(NC)"
	@go run main.go -seed verification
	@echo "$(GREEN)🎉 Secuencia completada exitosamente!$(NC)"
	@echo "$(BLUE)🔗 Ahora puedes probar: curl -H 'X-API-KEY: abc123secreta' -H 'X-USER-ID: 123' 'http://localhost:8080/api/v1/lots?project_id=1'$(NC)"
