package models

import (
	domain "github.com/devpablocristo/ponti-backend/internal/lease-type/usecases/domain"
	shareddomain "github.com/devpablocristo/ponti-backend/internal/shared/domain"
	sharedmodels "github.com/devpablocristo/ponti-backend/internal/shared/models"
)

type LeaseType struct {
	ID   int64  `gorm:"primaryKey;autoIncrement;column:id"`
	Name string `gorm:"size:100;not null;unique;column:name"`

	sharedmodels.Base
}

// Mapeo Model → Domain
func (m *LeaseType) ToDomain() *domain.LeaseType {
	return &domain.LeaseType{
		ID:   m.ID,
		Name: m.Name,
		Base: shareddomain.Base{
			CreatedAt: m.CreatedAt,
			UpdatedAt: m.UpdatedAt,
			CreatedBy: m.CreatedBy,
			UpdatedBy: m.UpdatedBy,
		},
	}
}

// Mapeo Domain → Model
func FromDomainLeaseType(d *domain.LeaseType) *LeaseType {
	return &LeaseType{
		ID:   d.ID,
		Name: d.Name,
		Base: sharedmodels.Base{
			CreatedBy: d.CreatedBy,
			UpdatedBy: d.UpdatedBy,
		},
	}
}
