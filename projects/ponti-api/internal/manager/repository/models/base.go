package models

import (
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/base"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/manager/usecases/domain"
)

type Manager struct {
	ID   int64  `gorm:"primaryKey;autoIncrement"`
	Name string `gorm:"type:varchar(255);not null;unique"`
	base.BaseModel
}

func (c Manager) ToDomain() *domain.Manager {
	return &domain.Manager{
		ID:   c.ID,
		Name: c.Name,
	}
}

// Solo pasa ID si es mayor a 0 (evita null value en insert)
func FromDomain(d *domain.Manager) *Manager {
	m := &Manager{
		Name: d.Name,
	}
	if d.ID > 0 {
		m.ID = d.ID
	}
	return m
}
