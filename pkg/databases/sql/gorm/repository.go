package pkggorm

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"cloud.google.com/go/cloudsqlconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// ConfigPort es la interfaz para manejar configuraciones del cliente GORM
type ConfigPort interface {
	GetDBType() DBType
	GetHost() string
	GetUser() string
	GetSSLMode() string
	GetPassword() string
	GetDBName() string
	GetPort() int
	GetSQLitePath() string
	GetSchema() string // Schema de PostgreSQL
	Validate() error
}

// Repository es la implementación de Repository
type Repository struct {
	client  *gorm.DB
	address string
	config  ConfigPort
	sqlDB   *sql.DB
}

// NewRepository inicializa un nuevo repositorio sin usar singleton
func newRepository(c ConfigPort) (*Repository, error) {
	repo := &Repository{
		config: c,
	}
	if err := repo.Connect(c); err != nil {
		return nil, fmt.Errorf("failed to initialize Repository: %w", err)
	}
	return repo, nil
}

func (r *Repository) Connect(config ConfigPort) error {
	// Primero crear la base si no existe
	if err := r.createDatabaseIfNotExists(config); err != nil {
		return fmt.Errorf("failed to create database: %w", err)
	}

	dialector, err := getDialector(config)
	if err != nil {
		return err
	}

	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		return err
	}

	r.client = db
	r.address = config.GetHost()

	if config.GetDBType() != SQLite {
		sqlDB, err := db.DB()
		if err != nil {
			return fmt.Errorf("failed to get sql.DB from gorm.DB: %w", err)
		}
		if err := sqlDB.Ping(); err != nil {
			return fmt.Errorf("failed to ping database: %w", err)
		}
		r.sqlDB = sqlDB

		// Inicializar schema para PostgreSQL
		if config.GetDBType() == Postgres {
			schema := config.GetSchema()
			if err := r.initializeSchema(context.Background(), sqlDB, schema); err != nil {
				return fmt.Errorf("failed to initialize schema: %w", err)
			}
			// search_path ya está configurado en el DSN, se aplica automáticamente a todas las conexiones
		}
	}

	log.Printf("Gorm successfully connected to %s database: %s (schema: %s)", config.GetDBType(), config.GetDBName(), config.GetSchema())
	return nil
}

func getDialector(config ConfigPort) (gorm.Dialector, error) {
	// if os.Getenv("K_SERVICE") != "" {
	//	return connectWithConnectorIAMAuthN(config)
	//}

	var dialector gorm.Dialector
	switch config.GetDBType() {
	case Postgres:
		schema := config.GetSchema()
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
			config.GetHost(), config.GetUser(), config.GetPassword(), config.GetDBName(), config.GetPort(), config.GetSSLMode())
		
		// Agregar search_path al DSN usando options=-c (se aplica a TODAS las conexiones del pool)
		// PostgreSQL aplica estos parámetros automáticamente a cada conexión nueva
		// Esto funciona con lib/pq y pgx, independientemente del driver usado por GORM
		if schema != "" && schema != "public" {
			// Validar schema name antes de agregarlo al DSN
			if err := validateSchemaName(schema); err != nil {
				return nil, fmt.Errorf("invalid schema name: %w", err)
			}
			// Escapar el schema name para el DSN (reemplazar comillas y espacios)
			escapedSchema := strings.ReplaceAll(schema, "'", "''")
			dsn += fmt.Sprintf(" options=-csearch_path='%s',public", escapedSchema)
		}
		
		dialector = postgres.Open(dsn)

	case MySQL:
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			config.GetUser(), config.GetPassword(), config.GetHost(), config.GetPort(), config.GetDBName())
		dialector = mysql.Open(dsn)

	case SQLite:
		dsn := config.GetSQLitePath()
		dialector = sqlite.Open(dsn)

	default:
		return dialector, fmt.Errorf("unsupported database type: %s", config.GetDBType())
	}

	return dialector, nil
}

func connectWithConnectorIAMAuthN(config ConfigPort) (gorm.Dialector, error) {
	mustGetenv := func(k string) string {
		v := os.Getenv(k)
		if v == "" {
			log.Fatalf("Warning: %s environment variable not set.", k)
		}
		return v
	}

	var (
		dbUser                 = config.GetUser()
		dbName                 = config.GetDBName()
		instanceConnectionName = mustGetenv("INSTANCE_CONNECTION_NAME")
		usePrivate             = os.Getenv("PRIVATE_IP")
	)

	d, err := cloudsqlconn.NewDialer(
		context.Background(),
		cloudsqlconn.WithIAMAuthN(),
		cloudsqlconn.WithLazyRefresh(),
	)
	if err != nil {
		return nil, fmt.Errorf("cloudsqlconn.NewDialer: %w", err)
	}
	var opts []cloudsqlconn.DialOption
	if usePrivate != "" {
		opts = append(opts, cloudsqlconn.WithPrivateIP())
	}

	dsn := fmt.Sprintf("user=%s database=%s", dbUser, dbName)
	dbConfigPort, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	dbConfigPort.DialFunc = func(ctx context.Context, network, instance string) (net.Conn, error) {
		return d.Dial(ctx, instanceConnectionName, opts...)
	}
	dbURI := stdlib.RegisterConnConfig(dbConfigPort)
	sqlDB, err := sql.Open("pgx", dbURI)
	if err != nil {
		return nil, fmt.Errorf("sql.Open: %w", err)
	}

	return postgres.New(postgres.Config{Conn: sqlDB}), nil
}

func (r *Repository) Client() *gorm.DB {
	return r.client
}

func (r *Repository) GetSQLDB() *sql.DB {
	return r.sqlDB
}

func (r *Repository) Address() string {
	return r.address
}

func (r *Repository) AutoMigrate(models ...any) error {
	return r.client.AutoMigrate(models...)
}

func (r *Repository) createDatabaseIfNotExists(config ConfigPort) error {
	if os.Getenv("K_SERVICE") != "" {
		return nil
	}

	switch config.GetDBType() {
	case Postgres:
		dsn := fmt.Sprintf(
			"host=%s user=%s password=%s dbname=postgres port=%d sslmode=%s",
			config.GetHost(), config.GetUser(), config.GetPassword(),
			config.GetPort(), config.GetSSLMode(),
		)
		db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
			DisableForeignKeyConstraintWhenMigrating: true,
		})
		if err != nil {
			return fmt.Errorf("failed to open sql.DB: %w", err)
		}

		sqlDB, err := db.DB()
		if err != nil {
			return fmt.Errorf("failed to get sql.DB: %w", err)
		}
		defer sqlDB.Close()

		var exists bool
		checkQuery := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = '%s')", config.GetDBName())
		if err := db.Raw(checkQuery).Scan(&exists).Error; err != nil {
			return fmt.Errorf("failed to check if database exists: %w", err)
		}

		if !exists {
			if err := db.Exec(fmt.Sprintf("CREATE DATABASE \"%s\"", config.GetDBName())).Error; err != nil {
				return fmt.Errorf("failed to create database: %w", err)
			}
		}
	case MySQL:
		dsn := fmt.Sprintf(
			"%s:%s@tcp(%s:%d)/?charset=utf8mb4&parseTime=True&loc=Local",
			config.GetUser(), config.GetPassword(),
			config.GetHost(), config.GetPort(),
		)
		db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
			DisableForeignKeyConstraintWhenMigrating: true,
		})
		if err != nil {
			return fmt.Errorf("failed to connect to MySQL server: %w", err)
		}

		sqlDB, err := db.DB()
		if err != nil {
			return fmt.Errorf("failed to get sql.DB: %w", err)
		}
		defer sqlDB.Close()

		createDBSQL := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", config.GetDBName())
		if err := db.Exec(createDBSQL).Error; err != nil {
			return fmt.Errorf("failed to create database: %w", err)
		}
	case SQLite:
		if _, err := os.Stat(config.GetSQLitePath()); os.IsNotExist(err) {
			fmt.Println("Automatically created by SQLite")
		}
		return nil

	default:
		return fmt.Errorf("unsupported database type: %s", config.GetDBType())
	}

	return nil
}

// initializeSchema crea el schema si no existe
// NO setea search_path aquí porque se configura en el DSN usando options=-c
func (r *Repository) initializeSchema(ctx context.Context, sqlDB *sql.DB, schema string) error {
	if schema == "" {
		schema = "public"
	}

	// Validar nombre de schema (seguridad básica)
	if err := validateSchemaName(schema); err != nil {
		return fmt.Errorf("invalid schema name: %w", err)
	}

	log.Printf("Initializing schema: %s", schema)

	// Crear schema si no existe
	createSchemaSQL := fmt.Sprintf(`CREATE SCHEMA IF NOT EXISTS %s`, quoteIdentifier(schema))
	if _, err := sqlDB.ExecContext(ctx, createSchemaSQL); err != nil {
		return fmt.Errorf("failed to create schema %s: %w", schema, err)
	}

	log.Printf("Schema %s created (search_path configured via DSN)", schema)
	return nil
}

// validateSchemaName valida que el nombre del schema sea seguro
func validateSchemaName(schema string) error {
	if schema == "" {
		return fmt.Errorf("schema name cannot be empty")
	}

	// Nombres reservados de PostgreSQL
	reserved := []string{"pg_catalog", "pg_toast", "information_schema", "pg_temp", "pg_toast_temp"}
	for _, r := range reserved {
		if strings.EqualFold(schema, r) {
			return fmt.Errorf("schema name '%s' is reserved", schema)
		}
	}

	// Validar caracteres (solo alfanuméricos, guiones bajos y guiones)
	for _, r := range schema {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-') {
			return fmt.Errorf("schema name contains invalid character: %c", r)
		}
	}

	// No puede empezar con número
	if len(schema) > 0 && schema[0] >= '0' && schema[0] <= '9' {
		return fmt.Errorf("schema name cannot start with a number")
	}

	return nil
}

// quoteIdentifier escapa un identificador de PostgreSQL de forma segura
func quoteIdentifier(name string) string {
	// Reemplazar comillas dobles escapándolas
	escaped := strings.ReplaceAll(name, `"`, `""`)
	return fmt.Sprintf(`"%s"`, escaped)
}
