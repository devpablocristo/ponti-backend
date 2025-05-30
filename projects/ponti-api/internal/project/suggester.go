package project

import (
	"context"

	pgs "github.com/alphacodinggroup/ponti-backend/pkg/words-suggesters/pg_trgm-gin"
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/project/usecases/domain"
)

// SuggesterEnginePort es la interfaz del motor externo de sugerencias.
type SuggesterEnginePort interface {
	Suggest(ctx context.Context, table, column, prefix string) ([]pgs.Suggestion, error)
	Close() error
	Health(ctx context.Context) error
}

// Suggester adapta SuggesterEnginePort al puerto de dominio.
type Suggester struct {
	eng SuggesterEnginePort
}

// NewSuggesterAdapter recibe el motor externo más la tabla/columna a usar.
func NewSuggester(eng SuggesterEnginePort) *Suggester {
	return &Suggester{
		eng: eng,
	}
}

// Suggest llama internamente a eng.Suggest con table y column "inyectados",
// y mapea el resultado a domain.ListedProject.
func (a *Suggester) Suggest(ctx context.Context, prefix string) ([]domain.ListedProject, error) {
	ext, err := a.eng.Suggest(ctx, "projects", "name", prefix)
	if err != nil {
		return nil, err
	}
	out := make([]domain.ListedProject, len(ext))
	for i, s := range ext {
		out[i] = domain.ListedProject{
			ID:   int64(s.ID),
			Name: s.Text,
		}
	}
	return out, nil
}

// Close delega el cierre de recursos al motor externo.
func (a *Suggester) Close() error {
	return a.eng.Close()
}

// Health delega el chequeo de salud al motor externo.
func (a *Suggester) Health(ctx context.Context) error {
	return a.eng.Health(ctx)
}
