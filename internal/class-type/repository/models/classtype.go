package models

import (
	domain "github.com/devpablocristo/ponti-backend/internal/class-type/usecases/domain"
	shareddomain "github.com/devpablocristo/ponti-backend/internal/shared/domain"
	sharedmodels "github.com/devpablocristo/ponti-backend/internal/shared/models"
)

// ClassType ("supply types": fertilizer, seed, etc.) is a GLOBAL catalog
// shared across tenants — the DB table has no `tenant_id` column. The
// previous tenanted shape (the now-removed `TenantID uuid.UUID` field)
// caused `SELECT ... FROM types WHERE tenant_id = ?` to fail with
// "column does not exist", producing a 500 on every `/api/v1/types` call.
type ClassType struct {
	ID   int64  `gorm:"primaryKey;autoIncrement;column:id"`
	Name string `gorm:"type:varchar(50);not null"`

	sharedmodels.Base
}

func (ClassType) TableName() string {
	return "types"
}

func (m *ClassType) ToDomain() *domain.ClassType {
	return &domain.ClassType{
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

func FromDomain(d *domain.ClassType) *ClassType {
	return &ClassType{
		ID:   d.ID,
		Name: d.Name,
		Base: sharedmodels.Base{
			CreatedBy: d.CreatedBy,
			UpdatedBy: d.UpdatedBy,
		},
	}
}
