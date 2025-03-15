package browserEvent

import (
	"context"
	"net/http"
	"time"

	domain "github.com/devpablocristo/monorepo/projects/qh/internal/browser-events/usecases/domain"
)

type UseCases interface {
	BrowserEvent(ctx context.Context, browserEvent *domain.BrowserEvent) error
}

type Cache interface {
	StoreRefreshToken(context.Context, string, string, time.Time) error
	RetrieveRefreshToken(context.Context, string) (string, error)
	Close()
}

// WebSocket define la interfaz del adaptador para conexiones WebSocket.
type WebSocket interface {
	Ping(http.ResponseWriter, *http.Request)
	BrowserEvent(http.ResponseWriter, *http.Request)
}

type Repository interface {
	SaveBrowserEvent(context.Context, *domain.BrowserEvent) error
	GetBrowserEventsByCandidateID(ctx context.Context, candidateID string) ([]*domain.BrowserEvent, error)
	GetBrowserEventsByAsssementID(context.Context, string) ([]*domain.BrowserEvent, error)
}
