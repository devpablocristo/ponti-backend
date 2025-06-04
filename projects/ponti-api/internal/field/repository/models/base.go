package models

import (
	"time"

	fielddom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/field/usecases/domain"
)

type Field struct {
	ID          int64     `gorm:"primaryKey;autoIncrement;column:id"`
	Name        string    `gorm:"size:100;not null;column:name"`
	ProjectID   int64     `gorm:"not null;index;column:project_id"`
	LeaseTypeID int64     `gorm:"not null;column:lease_type_id"`
	CreatedAt   time.Time `gorm:"autoCreateTime;column:created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime;column:updated_at"`
}

// FROM DOMAIN (para INSERT: solo escalares)
func FromDomain(d *fielddom.Field) *Field {
	return &Field{
		ID:          d.ID,
		Name:        d.Name,
		ProjectID:   d.ProjectID,
		LeaseTypeID: d.LeaseTypeID,
	}
}

// TO DOMAIN (sin preload de lots)
func (m *Field) ToDomain() *fielddom.Field {
	return &fielddom.Field{
		ID:          m.ID,
		Name:        m.Name,
		ProjectID:   m.ProjectID,
		LeaseTypeID: m.LeaseTypeID,
		// El campo Lots se debe enriquecer en el usecase si lo necesitas.
	}
}
