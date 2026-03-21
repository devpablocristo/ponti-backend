package dollar

import (
	"context"

	"github.com/devpablocristo/core/backend/go/domainerr"
	"github.com/devpablocristo/ponti-backend/internal/dollar/usecases/domain"
)

type RepositoryPort interface {
	ListByProject(context.Context, int64) ([]domain.DollarAverage, error)
	Create(context.Context, *domain.DollarAverage) (int64, error)
	Update(context.Context, *domain.DollarAverage) error
	GetByComposite(ctx context.Context, projectID, year int64, month string) (*domain.DollarAverage, error)
}

type UseCases struct {
	repo RepositoryPort
}

func NewUseCases(repo RepositoryPort) *UseCases {
	return &UseCases{repo: repo}
}

func (u *UseCases) ListByProject(ctx context.Context, projectID int64) ([]domain.DollarAverage, error) {
	if projectID == 0 {
		return nil, domainerr.Validation("projectID is required")
	}
	return u.repo.ListByProject(ctx, projectID)
}

func (u *UseCases) CreateOrUpdateBulk(ctx context.Context, items []domain.DollarAverage) error {
	if len(items) == 0 {
		return domainerr.Internal("no values provided")
	}

	base := items[0]
	seen := make(map[string]bool)

	for _, d := range items {
		if d.ProjectID != base.ProjectID || d.Year != base.Year {
			return domainerr.Validation("all items must share projectID and year")
		}
		if seen[d.Month] {
			return domainerr.Validation("duplicate month")
		}
		seen[d.Month] = true

		existing, err := u.repo.GetByComposite(ctx, d.ProjectID, d.Year, d.Month)
		if err != nil {
			return domainerr.Internal("error checking existence for month")
		}

		if existing == nil {
			id, err := u.repo.Create(ctx, &d)
			if err != nil {
				return domainerr.Internal("error creating month")
			}
			d.ID = id
		} else {
			d.ID = existing.ID
			if err := u.repo.Update(ctx, &d); err != nil {
				return domainerr.Internal("error updating month")
			}
		}
	}

	return nil
}
