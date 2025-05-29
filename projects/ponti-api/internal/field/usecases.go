package field

import (
	"context"
	"fmt"

	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/field/usecases/domain"
	lotdom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/lot/usecases/domain"
)

type LotUseCasesPort interface {
	CreateLot(context.Context, *lotdom.Lot) (int64, error)
	ListLots(context.Context, int64) ([]lotdom.Lot, error)
	GetLot(context.Context, int64) (*lotdom.Lot, error)
	UpdateLot(context.Context, *lotdom.Lot) error
	DeleteLot(context.Context, int64) error
}

type RepositoryPort interface {
	CreateField(ctx context.Context, f *domain.Field) (int64, error)
	ListFields(ctx context.Context) ([]domain.Field, error)
	GetField(ctx context.Context, id int64) (*domain.Field, error)
	UpdateField(ctx context.Context, f *domain.Field) error
	DeleteField(ctx context.Context, id int64) error
}

type UseCases struct {
	repo RepositoryPort
	lot  LotUseCasesPort
}

func NewUseCases(repo RepositoryPort, lot LotUseCasesPort) *UseCases {
	return &UseCases{
		repo: repo,
		lot:  lot,
	}
}

func (u *UseCases) CreateField(ctx context.Context, f *domain.Field) (int64, error) {
	// 1) Crear el Field y obtener su ID
	fieldID, err := u.repo.CreateField(ctx, f)
	if err != nil {
		return 0, fmt.Errorf("create field %q: %w", f.Name, err)
	}

	for _, l := range f.Lots {
		l.FieldID = fieldID
		if _, err := u.lot.CreateLot(ctx, &l); err != nil {
			if delErr := u.repo.DeleteField(ctx, fieldID); delErr != nil {
				return 0, fmt.Errorf(
					"rollback delete field %d failed: %v (original error: %w)",
					fieldID, delErr, err,
				)
			}
			return 0, fmt.Errorf("create lot %q for field %q: %w", l.Name, f.Name, err)
		}
	}

	return fieldID, nil
}

func (u *UseCases) ListFields(ctx context.Context) ([]domain.Field, error) {
	fields, err := u.repo.ListFields(ctx)
	if err != nil {
		return nil, err
	}
	for i := range fields {
		if err := u.enrichField(ctx, &fields[i]); err != nil {
			return nil, err
		}
	}
	return fields, nil
}

func (u *UseCases) GetField(ctx context.Context, id int64) (*domain.Field, error) {
	f, err := u.repo.GetField(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := u.enrichField(ctx, f); err != nil {
		return nil, err
	}
	return f, nil
}

func (u *UseCases) UpdateField(ctx context.Context, f *domain.Field) error {
	return u.repo.UpdateField(ctx, f)
}

func (u *UseCases) DeleteField(ctx context.Context, id int64) error {
	return u.repo.DeleteField(ctx, id)
}

func (u *UseCases) enrichField(ctx context.Context, f *domain.Field) error {
	allLots, err := u.lot.ListLots(ctx, f.ID)
	if err != nil {
		return fmt.Errorf("listar lots: %w", err)
	}

	var related []lotdom.Lot
	for _, l := range allLots {
		if l.FieldID == f.ID {
			related = append(related, l)
		}
	}
	f.Lots = related
	return nil
}
