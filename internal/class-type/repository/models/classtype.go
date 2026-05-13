package models

import (
	domain "github.com/devpablocristo/ponti-backend/internal/class-type/usecases/domain"
	shareddomain "github.com/devpablocristo/ponti-backend/internal/shared/domain"
	sharedmodels "github.com/devpablocristo/ponti-backend/internal/shared/models"
	"github.com/google/uuid"
)

type ClassType struct {
	ID       int64     `gorm:"primaryKey;autoIncrement;column:id"`
	TenantID uuid.UUID `gorm:"column:tenant_id;type:uuid;index"`
	Name     string    `gorm:"type:varchar(50);not null"`

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
