package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	pgksuggester "github.com/alphacodinggroup/ponti-backend/pkg/words-suggestors/pg_trgm-gin"
)

func main() {
	dsn := os.Getenv("DATABASE_DSN")
	table := os.Getenv("SUGGESTER_TABLE")
	column := os.Getenv("SUGGESTER_COLUMN")

	// Open underlying GORM DB
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("db open error: %v", err)
	}

	// Wrap in our adapter
	adapter := pgksuggester.NewGormAdapter(db)

	// Bootstrap using the adapter
	suggester, err := pgksuggester.Bootstrap(
		pgksuggester.WithDB(adapter),
		pgksuggester.WithTable(table),
		pgksuggester.WithColumn(column),
	)
	if err != nil {
		log.Fatalf("bootstrap error: %v", err)
	}
	defer suggester.Close()

	results, err := suggester.Suggest(context.Background(), "pab")
	if err != nil {
		log.Fatalf("suggest error: %v", err)
	}
	for _, r := range results {
		fmt.Printf("ID:%d Text:%s\n", r.ID, r.Text)
	}
}
