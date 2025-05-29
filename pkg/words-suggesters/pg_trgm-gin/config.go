package pkgsuggester

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
)

const (
	defaultLimit     = 10
	defaultThreshold = 0.3
)

var validIdentifier = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

type DB interface {
	WithContext(ctx context.Context) *miniGorm
	Exec(query string, args ...any) *miniGorm
	Raw(query string, args ...any) *miniGorm
	Scan(dest any) *miniGorm
	DB() (*sql.DB, error)
	Error() error
}

type Option func(*Config)

type Config struct {
	DB        DB
	Table     string
	Column    string
	Limit     int
	Threshold float64
	logger    Logger
}

func WithDB(db DB) Option            { return func(c *Config) { c.DB = db } }
func WithTable(name string) Option   { return func(c *Config) { c.Table = name } }
func WithColumn(name string) Option  { return func(c *Config) { c.Column = name } }
func WithLimit(n int) Option         { return func(c *Config) { c.Limit = n } }
func WithThreshold(t float64) Option { return func(c *Config) { c.Threshold = t } }
func WithLogger(l Logger) Option     { return func(c *Config) { c.logger = l } }

func newConfig(opts ...Option) (*Config, error) {
	cfg := &Config{
		DB:        nil,
		Table:     "",
		Column:    "",
		Limit:     defaultLimit,
		Threshold: defaultThreshold,
		logger:    noopLogger{},
	}
	for _, o := range opts {
		o(cfg)
	}
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) validate() error {
	if c.DB == nil {
		return fmt.Errorf("must provide a DB via WithDB")
	}
	if !validIdentifier.MatchString(c.Table) {
		return fmt.Errorf("invalid table name: %s", c.Table)
	}
	if !validIdentifier.MatchString(c.Column) {
		return fmt.Errorf("invalid column name: %s", c.Column)
	}
	if c.Limit <= 0 {
		return fmt.Errorf("limit must be > 0, got %d", c.Limit)
	}
	if c.Threshold < 0 || c.Threshold > 1 {
		return fmt.Errorf("threshold must be between 0 and 1, got %f", c.Threshold)
	}
	return nil
}
