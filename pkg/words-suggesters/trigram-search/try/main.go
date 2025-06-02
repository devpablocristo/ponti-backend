// File: ./try/main.go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	pkggorm "github.com/alphacodinggroup/ponti-backend/pkg/databases/sql/gorm"
	pgksuggester "github.com/alphacodinggroup/ponti-backend/pkg/words-suggesters/trigram-search"
)

func main() {
	// Leer configuraciones de entorno
	table := os.Getenv("SUGGESTER_TABLE")
	column := os.Getenv("SUGGESTER_COLUMN")
	limit := 10 // resultados por página
	offset := 0 // desde el primer resultado

	// Inicializar repositorio GORM con pkggorm (usa variables DB_*)
	repo, err := pkggorm.Bootstrap("", "", "", "", "", "", 0)
	if err != nil {
		log.Fatalf("failed to initialize repository: %v", err)
	}

	// Crear adapter para el sugeridor
	adapter := pgksuggester.NewPkggormAdapter(repo)

	// Inicializar WordsSuggester
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

	// Ejecutar sugerencias con paginado
	ctx := context.Background()
	results, total, err := suggester.Suggest(ctx, table, column, "pab", limit, offset)
	if err != nil {
		log.Fatalf("suggest error: %v", err)
	}

	fmt.Printf("Total resultados posibles: %d\n", total)
	for _, r := range results {
		fmt.Printf("ID:%d Text:%s\n", r.ID, r.Text)
	}
}
