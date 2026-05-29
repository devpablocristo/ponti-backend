package pkggorm

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"

	"cloud.google.com/go/cloudsqlconn"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	gormdb "github.com/devpablocristo/core/databases/postgres/go"
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

	// Usar core/databases/postgres/go para la conexión GORM.
	dsn := buildDSN(config)
	driver := mapDriver(config.GetDBType())
	gormCfg := gormdb.DefaultGormConfig()
	gormCfg.Driver = driver

	db, err := gormdb.OpenGorm(dsn, gormCfg)
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
	}

	log.Printf("Gorm successfully connected to %s database: %s", config.GetDBType(), config.GetDBName())
	return nil
}

func buildDSN(config ConfigPort) string {
	switch config.GetDBType() {
	case Postgres:
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d",
			config.GetHost(), config.GetUser(), config.GetPassword(), config.GetDBName(), config.GetPort())
		if ssl := config.GetSSLMode(); ssl != "" {
			dsn += " sslmode=" + ssl
		}
		return dsn
	case MySQL:
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			config.GetUser(), config.GetPassword(), config.GetHost(), config.GetPort(), config.GetDBName())
	case SQLite:
		return config.GetSQLitePath()
	default:
		return ""
	}
}

func mapDriver(t DBType) gormdb.DriverType {
	switch t {
	case MySQL:
		return gormdb.DriverMySQL
	case SQLite:
		return gormdb.DriverSQLite
	default:
		return gormdb.DriverPostgres
	}
}

// connectWithConnectorIAMAuthN conecta a Cloud SQL con IAM auth (usado en Cloud Run).
//
//nolint:unused
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

	dbConfigPort.DialFunc = func(ctx context.Context, network, addr string) (net.Conn, error) {
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
		gormCfg := gormdb.DefaultGormConfig()
		gormCfg.Driver = gormdb.DriverPostgres
		db, err := gormdb.OpenGorm(dsn, gormCfg)
		if err != nil {
			return fmt.Errorf("failed to open sql.DB: %w", err)
		}
		sqlDB, err := db.DB()
		if err != nil {
			return fmt.Errorf("failed to get sql.DB: %w", err)
		}
		defer func() { _ = sqlDB.Close() }()

		var exists bool
		if err := db.Raw("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = ?)", config.GetDBName()).Scan(&exists).Error; err != nil {
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
		gormCfg := gormdb.DefaultGormConfig()
		gormCfg.Driver = gormdb.DriverMySQL
		db, err := gormdb.OpenGorm(dsn, gormCfg)
		if err != nil {
			return fmt.Errorf("failed to connect to MySQL server: %w", err)
		}
		sqlDB, err := db.DB()
		if err != nil {
			return fmt.Errorf("failed to get sql.DB: %w", err)
		}
		defer func() { _ = sqlDB.Close() }()

		createDBSQL := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", config.GetDBName())
		if err := db.Exec(createDBSQL).Error; err != nil {
			return fmt.Errorf("failed to create database: %w", err)
		}
	case SQLite:
		if _, err := os.Stat(config.GetSQLitePath()); os.IsNotExist(err) {
			f, err := os.Create(config.GetSQLitePath())
			if err != nil {
				return fmt.Errorf("failed to create SQLite file: %w", err)
			}
			_ = f.Close()
		}
	}

	return nil
}
