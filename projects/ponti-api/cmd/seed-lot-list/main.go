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
		allSeeds = flag.Bool("all", false, "Ejecutar todos los seeds para LIST LOT")
		reset    = flag.Bool("reset", false, "Limpiar todos los datos antes de sembrar")
		specific = flag.String("seed", "", "Ejecutar seed específico")
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
🌱 SEED TOOL 2 - Generador de datos de prueba para LIST LOT

USO:
  go run cmd/seed-tool2/main.go [OPCIONES]

OPCIONES:
  -all              Ejecutar todos los seeds para LIST LOT en orden
  -reset            Limpiar todos los datos antes de sembrar
  -seed <nombre>    Ejecutar seed específico
  -help             Mostrar esta ayuda

EJEMPLOS:
  # Ejecutar todos los seeds para LIST LOT
  go run cmd/seed-tool2/main.go -all

  # Limpiar y ejecutar todos los seeds
  go run cmd/seed-tool2/main.go -reset -all

  # Ejecutar seed específico
  go run cmd/seed-tool2/main.go -seed users
  go run cmd/seed-tool2/main.go -seed fields
  go run cmd/seed-tool2/main.go -seed lots

SEEDS DISPONIBLES PARA LIST LOT:
  - reset                    Limpiar base de datos
  - users                    Usuario demo (ID: 123)
  - types                    Tipos base del sistema
  - lease-types              Tipos de arriendo
  - managers                 Managers del proyecto
  - investors                Inversores
  - units                    Unidades de medida
  - customers                Clientes
  - campaigns                Campañas
  - crops                    Cultivos (Soja, Maíz)
  - projects                 Proyecto principal
  - project-managers         Asociación managers
  - project-investors        Asociación inversores
  - fields                   Campos (A y B)
  - lots                     Lotes (3 parcelas)
  - supplies                 Insumos
  - labors                   Labores (categorías corregidas)
  - workorders               Órdenes de trabajo
  - tons                     Toneladas de los lotes
  - crop-commercializations  Precios de comercialización
  - lot-dates                Fechas de siembra/cosecha
  - verification             Verificación final

ORDEN DE EJECUCIÓN:
  1. reset (opcional, para limpiar)
  2. users, types, lease-types, managers, investors, units
  3. customers, campaigns, crops
  4. projects, project-managers, project-investors
  5. fields, lots, supplies, labors
  6. workorders, tons, crop-commercializations, lot-dates
  7. verification (para verificar todo)

DATOS GENERADOS:
  - Proyecto: "Construcción Torre Norte" (ID: 1)
  - Admin Cost: $15,000
  - Campo A: Arriendo fijo $100/ha
  - Campo B: Arriendo 15% de ingresos
  - 3 Lotes con datos diferenciados
  - Labores con categorías correctas (9=Siembra, 13=Cosecha)
  - Workorders con fechas correctas por estación
  - Precios realistas de comercialización

RESULTADO ESPERADO:
  - 3 lots con datos diferenciados
  - Cálculos coherentes en lot_table_view
  - Valores realistas para testing de LIST LOT
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
	fmt.Println("🌱 Ejecutando todos los seeds para LIST LOT...")

	// Leer directorio SQL y ordenar por nombre
	files, err := filepath.Glob("sql/*.sql")
	if err != nil {
		log.Fatalf("Error leyendo directorio sql/: %v", err)
	}

	if len(files) == 0 {
		log.Fatal("No se encontraron archivos SQL en el directorio sql/")
	}

	// Ordenar archivos para ejecutar en orden correcto
	sort.Strings(files)

	fmt.Println("📋 Archivos SQL encontrados:")
	for i, file := range files {
		fmt.Printf("   %d. %s\n", i+1, filepath.Base(file))
	}

	fmt.Println("\n🚀 Ejecutando seeds en orden...")

	for i, file := range files {
		fmt.Printf("\n   🌱 [%d/%d] Ejecutando %s...\n", i+1, len(files), filepath.Base(file))
		if err := executeSQLFile(db, file); err != nil {
			log.Fatalf("Error ejecutando %s: %v", file, err)
		}
		fmt.Printf("   ✅ [%d/%d] %s completado\n", i+1, len(files), filepath.Base(file))
	}

	fmt.Println("\n🎉 Todos los seeds para LIST LOT ejecutados exitosamente!")
	fmt.Println("📊 Ahora puedes probar el endpoint list lot con datos coherentes")
	fmt.Println("🔗 URL: http://localhost:8080/api/v1/lots?project_id=1")
}

func runSpecificSeed(db *sql.DB, seedName string) {
	fmt.Printf("🎯 Ejecutando seed específico: %s\n", seedName)

	// Mapear nombres de seed a archivos (específico para LIST LOT)
	seedFiles := map[string]string{
		"reset":                   "00_reset.sql",
		"users":                   "01_users.sql",
		"types":                   "02_types.sql",
		"lease-types":             "03_lease_types.sql",
		"managers":                "04_managers.sql",
		"investors":               "05_investors.sql",
		"categories":              "06_categories.sql",
		"customers":               "07_customers.sql",
		"campaigns":               "08_campaigns.sql",
		"crops":                   "09_crops.sql",
		"projects":                "10_projects.sql",
		"project-managers":        "11_project_managers.sql",
		"project-investors":       "12_project_investors.sql",
		"fields":                  "13_fields.sql",
		"lots":                    "14_lots.sql",
		"supplies":                "15_supplies.sql",
		"labors":                  "16_labors.sql",
		"workorders":              "17_workorders.sql",
		"tons":                    "18_tons.sql",
		"crop-commercializations": "19_crop_commercializations.sql",
		"lot-dates":               "20_lot_dates.sql",
		"verification":            "21_verification.sql",
	}

	fileName, exists := seedFiles[seedName]
	if !exists {
		log.Fatalf("Seed '%s' no encontrado.\n\nSeeds disponibles para LIST LOT:\n%s",
			seedName, getAvailableSeedsList(seedFiles))
	}

	filePath := filepath.Join("sql", fileName)
	fmt.Printf("📁 Archivo: %s\n", fileName)

	if err := executeSQLFile(db, filePath); err != nil {
		log.Fatalf("Error ejecutando %s: %v", fileName, err)
	}

	fmt.Printf("✅ Seed %s ejecutado exitosamente\n", seedName)
}

func getAvailableSeedsList(seedFiles map[string]string) string {
	var seeds []string
	for seed := range seedFiles {
		seeds = append(seeds, seed)
	}
	sort.Strings(seeds)

	var result strings.Builder
	for i, seed := range seeds {
		if i > 0 && i%3 == 0 {
			result.WriteString("\n")
		}
		result.WriteString(fmt.Sprintf("  - %-25s", seed))
	}
	return result.String()
}

func executeSQLFile(db *sql.DB, filePath string) error {
	// Leer contenido SQL
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error leyendo archivo %s: %w", filePath, err)
	}

	// Dividir en statements
	contentStr := string(content)

	// Remover comentarios y líneas vacías
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

	// Filtrar statements vacíos
	var validStatements []string
	for _, statement := range statements {
		statement = strings.TrimSpace(statement)
		if statement != "" {
			validStatements = append(validStatements, statement)
		}
	}

	fmt.Printf("   📝 Archivo %s: %d statements válidos encontrados\n",
		filepath.Base(filePath), len(validStatements))

	statementCount := 0
	for _, statement := range validStatements {
		statementCount++

		// Mostrar preview del statement
		preview := statement
		if len(preview) > 60 {
			preview = preview[:60] + "..."
		}
		fmt.Printf("   🔄 [%d/%d] Ejecutando: %s\n",
			statementCount, len(validStatements), preview)

		// Ejecutar statement
		result, err := db.Exec(statement)
		if err != nil {
			return fmt.Errorf("error ejecutando statement %d: %w\nStatement: %s",
				statementCount, err, statement)
		}

		// Verificar filas afectadas
		rowsAffected, _ := result.RowsAffected()
		fmt.Printf("   ✅ [%d/%d] Completado. Filas afectadas: %d\n",
			statementCount, len(validStatements), rowsAffected)
	}

	return nil
}
