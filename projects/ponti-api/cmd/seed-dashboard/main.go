package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/lib/pq"
)

func main() {
	var (
		reset = flag.Bool("reset", false, "Limpiar todos los datos antes de sembrar")
		help  = flag.Bool("help", false, "Mostrar ayuda")
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
		runAllSQLScripts(db)
		return
	}

	// Por defecto, crear datos básicos del dashboard
	runAllSQLScripts(db)
}

func showHelp() {
	fmt.Println(`
🌱 SEED DASHBOARD - Herramienta para poblar la base de datos con datos del dashboard

USO:
  go run cmd/seed-dashboard/main.go [OPCIONES]

OPCIONES:
  -reset            Limpiar todos los datos antes de sembrar
  -help             Mostrar esta ayuda

EJEMPLOS:
  # Crear datos del dashboard (ejecuta todos los scripts SQL en orden)
  go run cmd/seed-dashboard/main.go

  # Limpiar y crear datos del dashboard
  go run cmd/seed-dashboard/main.go -reset

DESCRIPCIÓN:
  Carga automáticamente todos los scripts SQL del directorio sql/ en orden específico:
  
  - 00_base_data.sql: Datos base (types, categories, labor_types, providers, users)
  - 01_basic_entities.sql: Entidades básicas (customers, campaigns, projects, fields, crops, lots)
  - 99_complete_dashboard_data.sql: Datos completos del dashboard (supplies, labors, workorders, investors, stock, invoices)

RESULTADO:
  El dashboard mostrará datos reales y ricos:
  - 20 workorders con fechas reales
  - 20 supplies con precios reales
  - 20 labors con contratistas reales
  - 5 inversores con porcentajes
  - Stock valorado en inventario
  - Facturas por montos reales
  - Métricas completas de siembra, cosecha, costos e ingresos

NOTA:
  Los scripts se ejecutan en orden específico para respetar las dependencias de foreign keys
`)
}

func connectDB() (*sql.DB, error) {
	// Configuración de la base de datos - por defecto usa Docker Compose
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_USER", "admin"),
		getEnv("DB_PASSWORD", "admin"),
		getEnv("DB_NAME", "ponti_api_db"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_SSL_MODE", "disable"),
	)

	fmt.Printf("🔌 Conectando a: %s:%s/%s (usuario: %s)\n",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_NAME", "ponti_api_db"),
		getEnv("DB_USER", "admin"))

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("error abriendo conexión: %w", err)
	}

	// Verificar conexión
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error conectando a la BD: %w", err)
	}

	fmt.Println("✅ Conexión exitosa a la base de datos")
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

	// Limpiar en orden correcto (por foreign keys)
	tables := []string{
		"invoices",
		"stocks",
		"workorders",
		"project_investors",
		"labors",
		"supplies",
		"lots",
		"fields",
		"projects",
		"customers",
		"campaigns",
		"crops",
		"investors",
		"users",
		"providers",
		"labor_categories",
		"labor_types",
		"categories",
		"types",
		"lease_types",
	}

	for _, table := range tables {
		fmt.Printf("   🗑️  Limpiando tabla %s...\n", table)
		_, err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table))
		if err != nil {
			log.Printf("⚠️  Advertencia al limpiar %s: %v", table, err)
		} else {
			fmt.Printf("   ✅ Tabla %s limpiada\n", table)
		}
	}

	fmt.Println("✅ Todos los datos limpiados exitosamente")
}

func runAllSQLScripts(db *sql.DB) {
	fmt.Println("🌱 Cargando todos los scripts SQL del dashboard...")

	// Buscar el directorio sql en el directorio actual o en cmd/seed-dashboard/sql
	sqlDir := "sql"
	if _, err := os.Stat(sqlDir); os.IsNotExist(err) {
		sqlDir = "cmd/seed-dashboard/sql"
	}

	// Verificar que el directorio existe
	if _, err := os.Stat(sqlDir); os.IsNotExist(err) {
		log.Fatalf("❌ Directorio SQL no encontrado: %s", sqlDir)
	}

	// Scripts en orden específico para respetar dependencias
	scriptOrder := []string{
		"00_base_data.sql",
		"01_basic_entities.sql",
		"99_complete_dashboard_data.sql",
	}

	fmt.Printf("📁 Ejecutando %d scripts SQL en %s:\n", len(scriptOrder), sqlDir)
	for i, script := range scriptOrder {
		fmt.Printf("   %d. %s\n", i+1, script)
	}
	fmt.Println()

	// Ejecutar cada script en orden específico
	for i, scriptName := range scriptOrder {
		scriptPath := filepath.Join(sqlDir, scriptName)

		// Verificar que el archivo existe
		if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
			log.Fatalf("❌ Script no encontrado: %s", scriptPath)
		}

		fmt.Printf("🔄 Ejecutando script %d/%d: %s\n", i+1, len(scriptOrder), scriptName)

		if err := executeSQLFile(db, scriptPath); err != nil {
			log.Fatalf("❌ Error ejecutando script %s: %v", scriptName, err)
		}

		fmt.Printf("✅ Script %s ejecutado exitosamente\n", scriptName)
		fmt.Println()
	}

	fmt.Println("🎉 Todos los scripts SQL ejecutados exitosamente")
	fmt.Println("🎯 Ahora puedes probar el endpoint: GET /api/v1/dashboard")
	fmt.Println("📈 El dashboard mostrará datos reales y ricos en lugar de solo 0s")
}

func executeSQLFile(db *sql.DB, filePath string) error {
	// Leer contenido SQL
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error leyendo archivo %s: %w", filePath, err)
	}

	// Dividir en statements
	contentStr := string(content)
	statements := strings.Split(contentStr, ";")

	fmt.Printf("   📝 Archivo %s: %d statements encontrados\n", filepath.Base(filePath), len(statements))

	statementCount := 0
	for _, statement := range statements {
		statement = strings.TrimSpace(statement)
		if statement == "" {
			continue
		}

		statementCount++
		fmt.Printf("   🔄 Ejecutando statement %d: %s\n", statementCount, strings.TrimSpace(statement[:min(len(statement), 50)])+"...")

		// Ejecutar statement
		result, err := db.Exec(statement)
		if err != nil {
			return fmt.Errorf("error ejecutando statement %d: %w\nStatement: %s", statementCount, err, statement)
		}

		// Verificar filas afectadas
		rowsAffected, _ := result.RowsAffected()
		fmt.Printf("   ✅ Statement %d ejecutado. Filas afectadas: %d\n", statementCount, rowsAffected)
	}

	fmt.Printf("   📊 Total de statements ejecutados: %d\n", statementCount)
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
