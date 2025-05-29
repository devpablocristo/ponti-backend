// File: suggester_adapter.go
package project

import (
	"context"

	pgs "github.com/alphacodinggroup/ponti-backend/pkg/words-suggesters/pg_trgm-gin"

	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/project/usecases/domain"
)

// pkg WordsSuggesters
type SuggesterEnginePort interface {
	Suggest(ctx context.Context, prefix string) ([]pgs.Suggestion, error)
	Close() error
	Health(ctx context.Context) error
}

// Suggester wraps a SuggesterEnginePort and adapts it to the project's needs.
type Suggester struct {
	sugEng SuggesterEnginePort
}

// NewSuggesterAdapter returns a SuggesterEnginePort implementation that converts
// external suggestions into domain.ListedProject entries.
func NewSuggester(se SuggesterEnginePort) *Suggester {
	return &Suggester{sugEng: se}
}

// Suggest delegates to the external sugEng and maps results to domain.ListedProject.
func (a *Suggester) Suggest(ctx context.Context, prefix string) ([]domain.ListedProject, error) {
	suggestions, err := a.sugEng.Suggest(ctx, prefix)
	if err != nil {
		return nil, err
	}
	projects := make([]domain.ListedProject, len(suggestions))
	for i, s := range suggestions {
		projects[i] = domain.ListedProject{ID: int64(s.ID), Name: s.Text}
	}
	return projects, nil
}

// Close cleans up resources in the external sugEng.
func (a *Suggester) Close() error {
	return a.sugEng.Close()
}

// Health checks the readiness of the external sugEng.
func (a *Suggester) Health(ctx context.Context) error {
	return a.sugEng.Health(ctx)
}
