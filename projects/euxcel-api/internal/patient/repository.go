package patient

import (
	"context"

	"github.com/google/uuid"

	gorm "github.com/devpablocristo/monorepo/pkg/databases/sql/gorm"

	models "github.com/devpablocristo/monorepo/projects/qh/internal/patient/repository/models"
	domain "github.com/devpablocristo/monorepo/projects/qh/internal/patient/usecases/domain"
)

type repository struct {
	db gorm.Repository
}

// NewRepository crea una nueva instancia del repositorio de Patient.
func NewRepository(db gorm.Repository) Repository {
	return &repository{
		db: db,
	}
}

// CreatePatient crea un registro de paciente en la base de datos.
func (r *repository) CreatePatient(ctx context.Context, patient *domain.Patient) error {
	pModel := &models.CreatePatient{}
	// Transfiere los datos del dominio al modelo de persistencia.
	pModel.FromDomain(patient)
	// Asigna un identificador único.
	pModel.ID = uuid.New().String()
	return r.db.Client().WithContext(ctx).Create(pModel).Error
}

// GetPatientByID obtiene un paciente a partir de su ID.
func (r *repository) GetPatientByID(ctx context.Context, id string) (*domain.Patient, error) {
	var pModel models.CreatePatient
	// Se busca por el ID en el modelo.
	if err := r.db.Client().WithContext(ctx).Where("id = ?", id).First(&pModel).Error; err != nil {
		return nil, err
	}

	// Se convierte el modelo de persistencia al dominio.
	patient, err := pModel.ToDomain()
	if err != nil {
		return nil, err
	}
	return patient, nil
}

// UpdatePatient actualiza un paciente en la base de datos.
func (r *repository) UpdatePatient(ctx context.Context, patient *domain.Patient) error {
	pModel := &models.CreatePatient{}
	pModel.FromDomain(patient)
	return r.db.Client().WithContext(ctx).Save(pModel).Error
}

// DeletePatient elimina un paciente a partir de su ID.
func (r *repository) DeletePatient(ctx context.Context, id string) error {
	return r.db.Client().WithContext(ctx).Delete(&models.CreatePatient{}, "id = ?", id).Error
}

// ListPatients retorna la lista de pacientes almacenados.
func (r *repository) GetAllPatients(ctx context.Context) ([]domain.Patient, error) {
	var pModels []models.CreatePatient
	if err := r.db.Client().WithContext(ctx).Find(&pModels).Error; err != nil {
		return nil, err
	}

	patients := make([]domain.Patient, 0, len(pModels))
	for _, m := range pModels {
		patient, err := m.ToDomain()
		if err != nil {
			return nil, err
		}
		patients = append(patients, *patient)
	}
	return patients, nil
}
