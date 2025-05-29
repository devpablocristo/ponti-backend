// File: ./try/main.go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	pkggorm "github.com/alphacodinggroup/ponti-backend/pkg/databases/sql/gorm"
	pgksuggester "github.com/alphacodinggroup/ponti-backend/pkg/words-suggesters/pg_trgm-gin"
)

func main() {
	// Leer configuraciones de entorno
	table := os.Getenv("SUGGESTER_TABLE")
	column := os.Getenv("SUGGESTER_COLUMN")

	// Inicializar repositorio GORM con pkggorm (usa variables GORM_*)
	repo, err := pkggorm.Bootstrap("", "", "", "", "", "", 0)
	if err != nil {
		log.Fatalf("failed to initialize repository: %v", err)
	}

	// Crear adapter para el sugeridor
	adapter := pgksuggester.NewPkggormAdapter(repo)

	// Inicializar Suggester
	suggester, err := pgksuggester.Bootstrap(
		pgksuggester.WithDB(adapter),
	)
	if err != nil {
		log.Fatalf("bootstrap error: %v", err)
	}
	defer func() {
		if cerr := suggester.Close(); cerr != nil {
			log.Printf("error closing suggester: %v", cerr)
		}
	}()

	// Ejecutar sugerencias
	ctx := context.Background()
	results, err := suggester.Suggest(ctx, table, column, "pab")
	if err != nil {
		log.Fatalf("suggest error: %v", err)
	}

	for _, r := range results {
		fmt.Printf("ID:%d Text:%s\n", r.ID, r.Text)
	}
}
