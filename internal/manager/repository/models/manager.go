package models

import (
	domain "github.com/devpablocristo/ponti-backend/internal/manager/usecases/domain"
	projectmod "github.com/devpablocristo/ponti-backend/internal/project/repository/models"
	shareddomain "github.com/devpablocristo/ponti-backend/internal/shared/domain"
	sharedmodels "github.com/devpablocristo/ponti-backend/internal/shared/models"
)

type Manager struct {
	ID       int64                `gorm:"primaryKey;autoIncrement"`
	Name     string               `gorm:"type:varchar(255);not null;unique"`
	Projects []projectmod.Project `gorm:"many2many:project_managers;"`
	sharedmodels.Base
}

func (m Manager) ToDomain() *domain.Manager {
	d := &domain.Manager{
		ID:   m.ID,
		Name: m.Name,
		Base: shareddomain.Base{
			CreatedAt: m.CreatedAt,
			UpdatedAt: m.UpdatedAt,
			CreatedBy: m.CreatedBy,
			UpdatedBy: m.UpdatedBy,
		},
	}
	if m.DeletedAt.Valid {
		t := m.DeletedAt.Time
		d.ArchivedAt = &t
	}
	return d
}

func FromDomain(d *domain.Manager) *Manager {
	m := &Manager{
		Name: d.Name,
		Base: sharedmodels.Base{
			CreatedBy: d.CreatedBy,
			UpdatedBy: d.UpdatedBy,
		},
	}
	if d.ID > 0 {
		m.ID = d.ID
	}
	return m
}
