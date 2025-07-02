package pkgsuggester

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type WordsSuggester struct {
	db        DB
	limit     int
	threshold float64
	logger    Logger
}

func newSuggester(cfg *Config) *WordsSuggester {
	// Habilitar extensiones necesarias (pg_trgm, unaccent) al inicializar.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Instalar pg_trgm
	errTrgm := cfg.DB.WithContext(ctx).
		Exec("CREATE EXTENSION IF NOT EXISTS pg_trgm").
		Error()
	if errTrgm != nil {
		cfg.logger.Error("failed to create pg_trgm extension", errTrgm)
	}

	// Instalar unaccent
	errUnaccent := cfg.DB.WithContext(ctx).
		Exec("CREATE EXTENSION IF NOT EXISTS unaccent").
		Error()
	if errUnaccent != nil {
		cfg.logger.Error("failed to create unaccent extension", errUnaccent)
	}

	// Ajustar umbral global de similitud
	_ = cfg.DB.WithContext(ctx).
		Exec(fmt.Sprintf("SET pg_trgm.similarity_threshold = %f", cfg.Threshold)).
		Error()

	return &WordsSuggester{
		db:        cfg.DB,
		limit:     cfg.Limit,
		threshold: cfg.Threshold,
		logger:    cfg.logger,
	}
}

func (s *WordsSuggester) Suggest(
	ctx context.Context,
	table, column, prefix string,
	limit, offset int,
) ([]Suggestion, int64, error) {
	if !validIdentifier.MatchString(table) {
		return nil, 0, fmt.Errorf("invalid table name: %s", table)
	}
	if !validIdentifier.MatchString(column) {
		return nil, 0, fmt.Errorf("invalid column name: %s", column)
	}
	q := strings.TrimSpace(prefix)
	if q == "" {
		return nil, 0, nil
	}

	// 1. Query paginada
	sqlStmt := fmt.Sprintf(
		`SELECT id, %s AS text
		 FROM %s
		 WHERE %s ILIKE ?
		 ORDER BY similarity(%s, ?) DESC
		 LIMIT ? OFFSET ?`,
		column, table, column, column,
	)
	var out []Suggestion
	start := time.Now()
	res := s.db.WithContext(ctx).
		Raw(sqlStmt, q+"%", q, limit, offset).
		Scan(&out)
	if err := res.Error(); err != nil {
		s.logger.Error("suggest query failed", err)
		return nil, 0, fmt.Errorf("suggest: %w", err)
	}

	// 2. Query de conteo total (sin limit ni offset)
	sqlCount := fmt.Sprintf(
		`SELECT COUNT(*) FROM %s WHERE %s ILIKE ?`,
		table, column,
	)
	var total int64
	countRes := s.db.WithContext(ctx).
		Raw(sqlCount, q+"%").
		Scan(&total)
	if err := countRes.Error(); err != nil {
		s.logger.Error("count query failed", err)
		return nil, 0, fmt.Errorf("count: %w", err)
	}

	s.logger.Debug(fmt.Sprintf("latency: %s", time.Since(start)))
	return out, total, nil
}

func (s *WordsSuggester) Close() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (s *WordsSuggester) Health(ctx context.Context) error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}
