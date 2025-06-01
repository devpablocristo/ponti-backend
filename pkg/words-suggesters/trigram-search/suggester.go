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
	// Ajustar umbral global una vez
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// CORREGIDO: Usar fmt.Sprintf para embedir el valor
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

// Suggest recibe tabla, columna y prefijo
func (s *WordsSuggester) Suggest(ctx context.Context, table, column, prefix string) ([]Suggestion, error) {
	// Validar
	if !validIdentifier.MatchString(table) {
		return nil, fmt.Errorf("invalid table name: %s", table)
	}
	if !validIdentifier.MatchString(column) {
		return nil, fmt.Errorf("invalid column name: %s", column)
	}
	q := strings.TrimSpace(prefix)
	if q == "" {
		return nil, nil
	}
	sqlStmt := fmt.Sprintf(
		"SELECT id, %s AS text FROM %s WHERE %s ILIKE ? ORDER BY similarity(%s, ?) DESC LIMIT %d",
		column, table, column, column, s.limit,
	)
	var out []Suggestion
	start := time.Now()
	res := s.db.WithContext(ctx).
		Raw(sqlStmt, q+"%", q).
		Scan(&out)
	if err := res.Error(); err != nil {
		s.logger.Error("suggest query failed", err)
		return nil, fmt.Errorf("suggest: %w", err)
	}
	s.logger.Debug(fmt.Sprintf("latency: %s", time.Since(start)))
	return out, nil
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
