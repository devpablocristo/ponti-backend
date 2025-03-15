package patient

import (
	"context"

	types "github.com/devpablocristo/monorepo/pkg/types"

	domain "github.com/devpablocristo/monorepo/projects/qh/internal/patient/usecases/domain"
)

// useCases es la implementación de la interfaz UseCases.
type useCases struct {
	repo Repository
}

func NewUseCases(r Repository) UseCases {
	return &useCases{
		repo: r,
	}
}

func (uc *useCases) GetAllPatients(ctx context.Context) ([]domain.Patient, error) {
	patients, err := uc.repo.GetAllPatients(ctx)
	if err != nil {
		return nil, err
	}
	return patients, nil
}

func (uc *useCases) GetPatientByID(ctx context.Context, id string) (*domain.Patient, error) {
	patient, err := uc.repo.GetPatientByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return patient, nil
}

func (uc *useCases) CreatePatient(ctx context.Context, patient *domain.Patient) error {
	if patient == nil {
		return types.NewError(
			types.ErrInvalidInput,
			"patient cannot be nil",
			nil,
		)
	}
	if err := uc.repo.CreatePatient(ctx, patient); err != nil {
		return err
	}
	return nil
}
