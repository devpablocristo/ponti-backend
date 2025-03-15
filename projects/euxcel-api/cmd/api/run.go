package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	cass "github.com/devpablocristo/monorepo/pkg/databases/nosql/cassandra/gocql"
	gorm "github.com/devpablocristo/monorepo/pkg/databases/sql/gorm"

	assessmentmodels "github.com/devpablocristo/monorepo/projects/qh/internal/assessment/repository/models"
	candidatemodels "github.com/devpablocristo/monorepo/projects/qh/internal/candidate/repository/models"
	groupmodels "github.com/devpablocristo/monorepo/projects/qh/internal/group/repository/models"
	personmodels "github.com/devpablocristo/monorepo/projects/qh/internal/person/repository/models"
	usermodels "github.com/devpablocristo/monorepo/projects/qh/internal/user/repository/models"

	wire "github.com/devpablocristo/monorepo/projects/qh/wire"
)

// RunWebServer registra las rutas en el router de Gin y arranca el servidor HTTP.
func RunWebServer(ctx context.Context, deps *wire.Dependencies) error {
	if deps == nil {
		return errors.New("dependencies cannot be nil")
	}

	log.Println("Registering HTTP routes...")

	// Se configuran middlewares globales en caso de tenerlos.
	if len(deps.Middlewares.Global) > 0 {
		deps.GinServer.GetRouter().Use(deps.Middlewares.Global...)
	}

	// Registrar todas las rutas de la aplicación.
	log.Println("Starting HTTP Server...")
	registerHttpRoutes(deps)

	// Arranca el servidor HTTP (por ejemplo, en el puerto 8080).
	return deps.GinServer.RunServer(ctx)
}

// registerRoutes registra todas las rutas de la aplicación en el router de Gin.
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
}

// RunGormMigrations ejecuta las migraciones de la base de datos SQL utilizando GORM.
func RunGormMigrations(ctx context.Context, repo gorm.Repository) error {
	log.Println("Starting GORM migrations...")

	// Se obtiene la conexión subyacente.
	sqlDB, err := repo.Client().DB()
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}
	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("database connection failed: %w", err)
	}

	// Lista de modelos a migrar.
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
	}

	start := time.Now()
	if err := repo.AutoMigrate(modelsToMigrate...); err != nil {
		return fmt.Errorf("failed to migrate database models: %w", err)
	}
	duration := time.Since(start)
	log.Printf("GORM migrations completed successfully in %s.", duration)
	return nil
}

// RunCassandraMigrations ejecuta las migraciones para Cassandra.
func RunCassandraMigrations(ctx context.Context, repo cass.Repository) error {
	log.Println("Starting Cassandra migrations...")
	session := repo.GetSession()

	// Crear keyspace si no existe.
	createKeyspaceCQL := `
		CREATE KEYSPACE IF NOT EXISTS mi_keyspace 
		WITH REPLICATION = { 'class': 'SimpleStrategy', 'replication_factor': 1 }`
	if err := session.Query(createKeyspaceCQL).WithContext(ctx).Exec(); err != nil {
		return fmt.Errorf("failed to create keyspace: %w", err)
	}
	log.Println("Keyspace 'mi_keyspace' created or already exists.")

	// Crear tabla "tweets".
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

	// Crear tabla desnormalizada "timeline_by_user".
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
