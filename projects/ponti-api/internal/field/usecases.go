package field

import (
	"context"
	"fmt"

	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/field/usecases/domain"
	lot "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/lot"
)

type useCases struct {
	repo Repository
	lot  lot.UseCases
}

// NewUseCases wires up Field and Lot repositories.
func NewUseCases(fRepo Repository, lRepo lot.Repository) UseCases {
	return &useCases{repo: fRepo, lot: lRepo}
}

// CreateField creates the Field and its Lots.
func (u *useCases) CreateField(ctx context.Context, f *domain.Field) (int64, error) {
	fieldID, err := u.repo.CreateField(ctx, f)
	if err != nil {
		return 0, fmt.Errorf("create field %q: %w", f.Name, err)
	}
	for _, l := range f.Lots {
		if _, err := u.lot.CreateLot(ctx, &l); err != nil {
			return 0, fmt.Errorf("create lot %q for field %q: %w", l.Name, f.Name, err)
		}
	}
	return fieldID, nil
}

// ListFields returns all fields.
func (u *useCases) ListFields(ctx context.Context) ([]domain.Field, error) {
	return u.repo.ListFields(ctx)
}

// GetField retrieves a field by ID.
func (u *useCases) GetField(ctx context.Context, id int64) (*domain.Field, error) {
	return u.repo.GetField(ctx, id)
}

// UpdateField updates an existing field.
func (u *useCases) UpdateField(ctx context.Context, f *domain.Field) error {
	return u.repo.UpdateField(ctx, f)
}

// DeleteField deletes a field by ID.
func (u *useCases) DeleteField(ctx context.Context, id int64) error {
	return u.repo.DeleteField(ctx, id)
}
