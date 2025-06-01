package project

import (
	"context"

	pgs "github.com/alphacodinggroup/ponti-backend/pkg/words-suggesters/trigram-search"
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/project/usecases/domain"
)

// SuggesterEnginePort es la interfaz del motor externo de sugerencias.
type SuggesterEnginePort interface {
	Suggest(context.Context, string, string, string) ([]pgs.Suggestion, error)
	Close() error
	Health(context.Context) error
}

// WordsSuggester adapta SuggesterEnginePort al puerto de dominio.
type WordsSuggester struct {
	eng SuggesterEnginePort
}

// NewSuggesterAdapter recibe el motor externo más la tabla/columna a usar.
func NewSuggester(eng SuggesterEnginePort) *WordsSuggester {
	return &WordsSuggester{
		eng: eng,
	}
}

// Suggest llama internamente a eng.Suggest con table y column "inyectados",
// y mapea el resultado a domain.ListedProject.
func (a *WordsSuggester) Suggest(ctx context.Context, prefix string) ([]domain.ListedProject, error) {
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
func (a *WordsSuggester) Close() error {
	return a.eng.Close()
}

// Health delega el chequeo de salud al motor externo.
func (a *WordsSuggester) Health(ctx context.Context) error {
	return a.eng.Health(ctx)
}
