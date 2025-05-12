package models

import (
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/manager/usecases/domain"
)

// Manager representa el modelo GORM para un manager.
type Manager struct {
	ID   int64  `gorm:"primaryKey"`
	Name string `gorm:"type:varchar(100);not null"`
	Type string `gorm:"type:varchar(100);not null"`
}

// ToDomain convierte el modelo Manager a la entidad de dominio.
func (c Manager) ToDomain() *domain.Manager {
	return &domain.Manager{
		ID:   c.ID,
		Name: c.Name,
		Type: c.Type,
	}
}

// FromDomainManager convierte una entidad de dominio a su modelo GORM.
func FromDomainManager(d *domain.Manager) *Manager {
	return &Manager{
		ID:   d.ID,
		Name: d.Name,
		Type: d.Type,
	}
}
