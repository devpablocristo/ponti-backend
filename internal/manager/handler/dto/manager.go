package dto

import (
	"time"

	"github.com/alphacodinggroup/ponti-backend/internal/manager/usecases/domain"
	shareddomain "github.com/alphacodinggroup/ponti-backend/internal/shared/domain"
)

// Manager es el DTO para un manager específico.
type Manager struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"Type"`
	CreatedAt time.Time `json:"CreatedAt"`
	UpdatedAt time.Time `json:"UpdatedAt"`
	CreatedBy *int64    `json:"CreatedBy"`
	UpdatedBy *int64    `json:"UpdatedBy"`
}

// ToDomain convierte el DTO Manager a la entidad de dominio.
func (c Manager) ToDomain() *domain.Manager {
	return &domain.Manager{
		ID:   c.ID,
		Name: c.Name,
		Type: c.Type,
		Base: shareddomain.Base{
			CreatedAt: c.CreatedAt,
			UpdatedAt: c.UpdatedAt,
			CreatedBy: c.CreatedBy,
			UpdatedBy: c.UpdatedBy,
		},
	}
}

// FromDomain convierte una entidad de dominio a su DTO.
func FromDomain(d domain.Manager) *Manager {
	return &Manager{
		ID:        d.ID,
		Name:      d.Name,
		Type:      d.Type,
		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
		CreatedBy: d.CreatedBy,
		UpdatedBy: d.UpdatedBy,
	}
}
