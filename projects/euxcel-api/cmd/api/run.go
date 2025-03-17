package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	cass "github.com/alphacodinggroup/euxcel-backend/pkg/databases/nosql/cassandra/gocql"
	gorm "github.com/alphacodinggroup/euxcel-backend/pkg/databases/sql/gorm"

	assessmentmodels "github.com/alphacodinggroup/euxcel-backend/internal/assessment/repository/models"
	candidatemodels "github.com/alphacodinggroup/euxcel-backend/internal/candidate/repository/models"
	categorymodels "github.com/alphacodinggroup/euxcel-backend/internal/category/repository/models"
	groupmodels "github.com/alphacodinggroup/euxcel-backend/internal/group/repository/models"
	itemmodels "github.com/alphacodinggroup/euxcel-backend/internal/item/repository/models"
	macrocategorymodels "github.com/alphacodinggroup/euxcel-backend/internal/macrocategory/repository/models"
	personmodels "github.com/alphacodinggroup/euxcel-backend/internal/person/repository/models"
	suppliermodels "github.com/alphacodinggroup/euxcel-backend/internal/supplier/repository/models"
	usermodels "github.com/alphacodinggroup/euxcel-backend/internal/user/repository/models"

	wire "github.com/alphacodinggroup/euxcel-backend/wire"
)

// RunHttpServer registers routes in the Gin router and starts the HTTP server.
func RunHttpServer(ctx context.Context, deps *wire.Dependencies) error {
	if deps == nil {
		return errors.New("dependencies cannot be nil")
	}

	log.Println("Registering HTTP routes...")

	// Configure global middlewares if any.
	if len(deps.Middlewares.Global) > 0 {
		deps.GinServer.GetRouter().Use(deps.Middlewares.Global...)
	}

	// Register all application routes.
	log.Println("Starting HTTP Server...")
	registerHttpRoutes(deps)

	// Start the HTTP server (e.g., on port 8080).
	return deps.GinServer.RunServer(ctx)
}

// registerHttpRoutes registers all application routes in the Gin router.
func registerHttpRoutes(deps *wire.Dependencies) {
	deps.EventHandler.Routes()
	deps.GroupHandler.Routes()
	deps.PersonHandler.Routes()
	deps.AssessmentHandler.Routes()
	deps.CandidateHandler.Routes()
	deps.UserHandler.Routes()
	deps.AutheHandler.Routes()
	deps.NotificationHandler.Routes()
	deps.TweetHandler.Routes()
	deps.BrowserEventsHandler.Routes()
	deps.ItemHandler.Routes()
	deps.CategoryHandler.Routes()
	deps.MacroCategoryHandler.Routes()
	deps.SupplierHandler.Routes()
}

// RunGormMigrations runs SQL migrations using GORM.
func RunGormMigrations(ctx context.Context, repo gorm.Repository) error {
	log.Println("Starting GORM migrations...")

	// Obtain the underlying database connection.
	sqlDB, err := repo.Client().DB()
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}
	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("database connection failed: %w", err)
	}

	// List of models to migrate.
	modelsToMigrate := []any{
		&groupmodels.Group{},
		&groupmodels.GroupMember{},
		&assessmentmodels.Assessment{},
		&assessmentmodels.Problem{},
		&assessmentmodels.SkillConfig{},
		&assessmentmodels.UnitTest{},
		&candidatemodels.Candidate{},
		&personmodels.Person{},
		&assessmentmodels.Link{},
		&usermodels.User{},
		&usermodels.Follow{},
		&itemmodels.Item{},
		&categorymodels.Category{},
		&macrocategorymodels.MacroCategory{},
		&suppliermodels.Supplier{},
	}

	start := time.Now()
	if err := repo.AutoMigrate(modelsToMigrate...); err != nil {
		return fmt.Errorf("failed to migrate database models: %w", err)
	}
	duration := time.Since(start)
	log.Printf("GORM migrations completed successfully in %s.", duration)
	return nil
}

// RunCassandraMigrations runs Cassandra migrations.
func RunCassandraMigrations(ctx context.Context, repo cass.Repository) error {
	log.Println("Starting Cassandra migrations...")
	session := repo.GetSession()

	// Create keyspace if it doesn't exist.
	createKeyspaceCQL := `
		CREATE KEYSPACE IF NOT EXISTS mi_keyspace 
		WITH REPLICATION = { 'class': 'SimpleStrategy', 'replication_factor': 1 }`
	if err := session.Query(createKeyspaceCQL).WithContext(ctx).Exec(); err != nil {
		return fmt.Errorf("failed to create keyspace: %w", err)
	}
	log.Println("Keyspace 'mi_keyspace' created or already exists.")

	// Create table "tweets".
	createTweetsTableCQL := `
		CREATE TABLE IF NOT EXISTS tweets (
			id uuid PRIMARY KEY,
			user_id text,
			content text,
			created_at timestamp
		)`
	if err := session.Query(createTweetsTableCQL).WithContext(ctx).Exec(); err != nil {
		return fmt.Errorf("failed to create table 'tweets': %w", err)
	}
	log.Println("Table 'tweets' created or already exists.")

	// Create denormalized table "timeline_by_user".
	createTimelineTableCQL := `
		CREATE TABLE IF NOT EXISTS timeline_by_user (
			user_id text,
			created_at timestamp,
			tweet_id text,
			content text,
			PRIMARY KEY (user_id, created_at, tweet_id)
		) WITH CLUSTERING ORDER BY (created_at DESC)
	`
	if err := session.Query(createTimelineTableCQL).WithContext(ctx).Exec(); err != nil {
		return fmt.Errorf("failed to create table 'timeline_by_user': %w", err)
	}
	log.Println("Table 'timeline_by_user' created or already exists.")

	log.Println("Cassandra migrations completed successfully.")
	return nil
}
