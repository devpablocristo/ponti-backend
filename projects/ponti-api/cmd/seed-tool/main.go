package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "github.com/lib/pq"
)

func main() {
	var (
		allSeeds = flag.Bool("all", false, "Ejecutar todos los seeds")
		reset    = flag.Bool("reset", false, "Limpiar todos los datos")
		specific = flag.String("seed", "", "Ejecutar seed específico (ej: users, customers)")
		help     = flag.Bool("help", false, "Mostrar ayuda")
	)

	flag.Parse()

	if *help {
		showHelp()
		return
	}

	// Conectar a la base de datos
	db, err := connectDB()
	if err != nil {
		log.Fatalf("Error conectando a la BD: %v", err)
	}
	defer db.Close()

	if *reset {
		resetDatabase(db)
		return
	}

	if *allSeeds {
		runAllSeeds(db)
		return
	}

	if *specific != "" {
		runSpecificSeed(db, *specific)
		return
	}

	showHelp()
}

func showHelp() {
	fmt.Println(`
🌱 SEED TOOL - Herramienta para poblar la base de datos con datos de prueba

USO:
  go run cmd/seed-tool/main.go [OPCIONES]

OPCIONES:
  -all              Ejecutar todos los seeds en orden
  -reset            Limpiar todos los datos antes de sembrar
  -seed <nombre>    Ejecutar seed específico (ej: users, customers, crops)
  -help             Mostrar esta ayuda

EJEMPLOS:
  # Ejecutar todos los seeds
  go run cmd/seed-tool/main.go -all

  # Limpiar y ejecutar todos los seeds
  go run cmd/seed-tool/main.go -reset -all

  # Ejecutar seed específico
  go run cmd/seed-tool/main.go -seed users
  go run cmd/seed-tool/main.go -seed customers
  go run cmd/seed-tool/main.go -seed crops

SEEDS DISPONIBLES:
  - users           Usuarios básicos
  - types           Tipos del sistema
  - categories      Categorías del sistema
  - lease-types     Tipos de arriendo
  - managers        Managers
  - investors       Inversores (con datos financieros)
  - units           Unidades de medida
  - customers       Clientes
  - campaigns       Campañas
  - crops           Cultivos
  - projects        Proyectos (con versionado)
  - lots            Lotes (incluye campos, variety, sowing_date, version)
  - lot-dates       Fechas de lotes (sowing, harvest)
  - labor-types     Tipos de labor
  - labor-categories Categorías de labor
  - labors          Labores del proyecto
  - project-investors Relaciones proyecto-inversores
  - project-managers Relaciones proyecto-managers
  - supplies        Suministros (con unidades y categorías)
  - providers       Proveedores
  - stocks          Stock de suministros
  - supply-movements Movimientos de suministros
  - workorders      Órdenes de trabajo
  - workorder-items Items de órdenes de trabajo
  - invoices        Facturas
  - project-dollar-values Valores del dólar por proyecto
  - crop-commercializations Comercialización de cultivos
`)
}

func connectDB() (*sql.DB, error) {
	// Configuración de la base de datos
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_USER", "admin"),
		getEnv("DB_PASSWORD", "admin"),
		getEnv("DB_NAME", "ponti_api_db"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_SSL_MODE", "disable"),
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("error abriendo conexión: %w", err)
	}

	// Verificar conexión
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error verificando conexión: %w", err)
	}

	return db, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func resetDatabase(db *sql.DB) {
	fmt.Println("🧹 Limpiando todos los datos...")

	// Usar archivo de reset
	resetFile := "sql/00_reset.sql"
	if err := executeSQLFile(db, resetFile); err != nil {
		log.Fatalf("Error ejecutando reset: %v", err)
	}

	fmt.Println("✅ Todos los datos limpiados exitosamente")
}

func runAllSeeds(db *sql.DB) {
	fmt.Println("🌱 Ejecutando todos los seeds...")

	// Leer directorio SQL
	files, err := filepath.Glob("sql/*.sql")
	if err != nil {
		log.Fatalf("Error leyendo directorio sql/: %v", err)
	}

	if len(files) == 0 {
		log.Fatal("No se encontraron archivos SQL en el directorio sql/")
	}

	sort.Strings(files)

	for _, file := range files {
		fmt.Printf("   🌱 Ejecutando %s...\n", filepath.Base(file))
		if err := executeSQLFile(db, file); err != nil {
			log.Fatalf("Error ejecutando %s: %v", file, err)
		}
		fmt.Printf("   ✅ %s completado\n", filepath.Base(file))
	}

	fmt.Println("✅ Todos los seeds ejecutados exitosamente")
}

func runSpecificSeed(db *sql.DB, seedName string) {
	fmt.Printf("🎯 Ejecutando seed específico: %s\n", seedName)

	// Mapear nombres de seed a archivos
	seedFiles := map[string]string{
		"users":                   "01_users.sql",
		"types":                   "02_types.sql",
		"categories":              "03_categories.sql",
		"lease-types":             "04_lease_types.sql",
		"managers":                "05_managers.sql",
		"investors":               "06_investors.sql",
		"units":                   "07_units.sql",
		"customers":               "08_customers.sql",
		"campaigns":               "09_campaigns.sql",
		"crops":                   "10_crops.sql",
		"projects":                "11_projects.sql",
		"lots":                    "12_lots.sql",
		"lot-dates":               "13_lot_dates.sql",
		"labor-types":             "14_labor_types.sql",
		"labor-categories":        "15_labor_categories.sql",
		"labors":                  "16_labors.sql",
		"project-investors":       "17_project_investors.sql",
		"project-managers":        "18_project_managers.sql",
		"supplies":                "19_supplies.sql",
		"providers":               "20_providers.sql",
		"stocks":                  "21_stocks.sql",
		"supply-movements":        "22_supply_movements.sql",
		"workorders":              "23_workorders.sql",
		"workorder-items":         "24_workorder_items.sql",
		"invoices":                "25_invoices.sql",
		"project-dollar-values":   "26_project_dollar_values.sql",
		"crop-commercializations": "27_crop_commercializations.sql",
	}

	fileName, exists := seedFiles[seedName]
	if !exists {
		log.Fatalf("Seed '%s' no encontrado. Seeds disponibles: %v", seedName, getAvailableSeeds(seedFiles))
	}

	filePath := filepath.Join("sql", fileName)
	if err := executeSQLFile(db, filePath); err != nil {
		log.Fatalf("Error ejecutando %s: %v", fileName, err)
	}

	fmt.Printf("✅ Seed %s ejecutado exitosamente\n", seedName)
}

func getAvailableSeeds(seedFiles map[string]string) []string {
	seeds := make([]string, 0, len(seedFiles))
	for seed := range seedFiles {
		seeds = append(seeds, seed)
	}
	return seeds
}

func executeSQLFile(db *sql.DB, filePath string) error {
	// Leer contenido SQL
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error leyendo archivo %s: %w", filePath, err)
	}

	// Dividir en statements
	contentStr := string(content)

	// Remover comentarios
	lines := strings.Split(contentStr, "\n")
	var cleanLines []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "--") {
			cleanLines = append(cleanLines, line)
		}
	}

	// Unir líneas
	cleanContent := strings.Join(cleanLines, " ")
	statements := strings.Split(cleanContent, ";")

	fmt.Printf("   📝 Archivo %s: %d statements encontrados\n", filepath.Base(filePath), len(statements))

	statementCount := 0
	for _, statement := range statements {
		statement = strings.TrimSpace(statement)
		if statement == "" {
			continue
		}

		statementCount++
		fmt.Printf("   🔄 Ejecutando statement %d: %s\n", statementCount, strings.TrimSpace(statement[:min(len(statement), 50)])+"...")

		// Ejecutar statements
		result, err := db.Exec(statement)
		if err != nil {
			return fmt.Errorf("error ejecutando statement %d: %w\nStatement: %s", statementCount, err, statement)
		}

		// Verificar filas afectadas
		rowsAffected, _ := result.RowsAffected()
		fmt.Printf("   ✅ Statement %d ejecutado. Filas afectadas: %d\n", statementCount, rowsAffected)
	}

	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
