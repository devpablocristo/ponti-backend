package patient

import (
	"context"

	domain "github.com/devpablocristo/monorepo/projects/qh/internal/patient/usecases/domain"
)

// Repository define los métodos que debe implementar el repositorio para interactuar con Patient.
// Repository define los métodos que debe implementar el repositorio para Patient.
type Repository interface {
	CreatePatient(ctx context.Context, patient *domain.Patient) error
	GetPatientByID(ctx context.Context, id string) (*domain.Patient, error)
	UpdatePatient(ctx context.Context, patient *domain.Patient) error
	DeletePatient(ctx context.Context, id string) error
	GetAllPatients(ctx context.Context) ([]domain.Patient, error)
}

// UseCases define la interfaz para los casos de uso de Patient.
type UseCases interface {
	GetAllPatients(ctx context.Context) ([]domain.Patient, error)
	GetPatientByID(ctx context.Context, id string) (*domain.Patient, error)
	CreatePatient(ctx context.Context, patient *domain.Patient) error
}
