package models

import (
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/manager/usecases/domain"
	projectmod "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/project/repository/models"
	sharedmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/models"
)

type Manager struct {
	ID       int64                `gorm:"primaryKey;autoIncrement"`
	Name     string               `gorm:"type:varchar(255);not null;unique"`
	Projects []projectmod.Project `gorm:"many2many:project_managers;"` // <<--- AGREGAR ESTA LINEA
	sharedmodels.Base
}

func (c Manager) ToDomain() *domain.Manager {
	return &domain.Manager{
		ID:   c.ID,
		Name: c.Name,
	}
}

func FromDomain(d *domain.Manager) *Manager {
	m := &Manager{
		Name: d.Name,
	}
	if d.ID > 0 {
		m.ID = d.ID
	}
	return m
}
