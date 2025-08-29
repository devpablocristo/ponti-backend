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
🌱 SEED DASHBOARD - Herramienta para poblar la base de datos con datos mínimos para el dashboard

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
  Carga automáticamente todos los scripts SQL del directorio sql/ en orden alfabético:
  
  - 01_dashboard_minimal.sql: Datos básicos (Soja)
  - 02_add_maiz_crop.sql: Agregar Maíz
  - 03_add_trigo_crop.sql: Agregar Trigo
  - ... (cualquier script adicional que agregues)

RESULTADO:
  El dashboard mostrará todos los cultivos configurados en los scripts SQL
  - Soja: 10.5 hectáreas
  - Maíz: 8.5 hectáreas  
  - Trigo: 12.0 hectáreas
  - Total: 31.0 hectáreas

NOTA:
  Los scripts se ejecutan en orden alfabético, por eso usamos prefijos numéricos
  (01_, 02_, 03_, etc.) para controlar el orden de ejecución.
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
		"lots",
		"fields",
		"projects",
		"customers",
		"campaigns",
		"crops",
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

	// Leer todos los archivos .sql del directorio
	files, err := filepath.Glob(filepath.Join(sqlDir, "*.sql"))
	if err != nil {
		log.Fatalf("❌ Error leyendo directorio SQL: %v", err)
	}

	if len(files) == 0 {
		log.Fatalf("❌ No se encontraron archivos SQL en %s", sqlDir)
	}

	// Ordenar archivos alfabéticamente (para mantener el orden 01_, 02_, 03_, etc.)
	sort.Strings(files)

	fmt.Printf("📁 Encontrados %d scripts SQL en %s:\n", len(files), sqlDir)
	for i, file := range files {
		fmt.Printf("   %d. %s\n", i+1, filepath.Base(file))
	}
	fmt.Println()

	// Ejecutar cada script en orden
	for i, file := range files {
		fmt.Printf("🔄 Ejecutando script %d/%d: %s\n", i+1, len(files), filepath.Base(file))

		if err := executeSQLFile(db, file); err != nil {
			log.Fatalf("❌ Error ejecutando script %s: %v", filepath.Base(file), err)
		}

		fmt.Printf("✅ Script %s ejecutado exitosamente\n", filepath.Base(file))
		fmt.Println()
	}

	fmt.Println("🎉 Todos los scripts SQL ejecutados exitosamente")
	fmt.Println("🎯 Ahora puedes probar el endpoint: GET /api/v1/dashboard")
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
