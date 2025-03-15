package browserEvent

import (
	"context"
	"fmt"

	"github.com/alphacodinggroup/euxcel-backend/internal/browser-events/usecases/domain"
)

type useCases struct {
	repository Repository
}

func NewUseCases(rp Repository) UseCases {
	return &useCases{
		repository: rp,
	}
}

func (u *useCases) BrowserEvent(ctx context.Context, event *domain.BrowserEvent) error {
	fmt.Println("BrowserEvent")

	if err := u.repository.SaveBrowserEvent(ctx, event); err != nil {
		return fmt.Errorf("error saving browser event: %w", err)
	}

	return nil
}
