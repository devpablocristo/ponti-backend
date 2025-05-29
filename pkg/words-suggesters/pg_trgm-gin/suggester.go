package pkgsuggester

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type Suggester struct {
	db     DB
	table  string
	column string
	limit  int
	logger Logger
}

func newSuggester(cfg *Config) *Suggester {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = cfg.DB.WithContext(ctx).
		Exec("SET pg_trgm.similarity_threshold = ?", cfg.Threshold).
		Error() // ignore failure: logged internally
	return &Suggester{
		db:     cfg.DB,
		table:  cfg.Table,
		column: cfg.Column,
		limit:  cfg.Limit,
		logger: cfg.logger,
	}
}

func (s *Suggester) Suggest(ctx context.Context, prefix string) ([]Suggestion, error) {
	q := strings.TrimSpace(prefix)
	if q == "" {
		return nil, nil
	}
	sqlStmt := fmt.Sprintf(
		"SELECT id, %s AS text FROM %s WHERE %s ILIKE ? ORDER BY similarity(%s, ?) DESC LIMIT %d",
		s.column, s.table, s.column, s.column, s.limit,
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

func (s *Suggester) Close() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (s *Suggester) Health(ctx context.Context) error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}
