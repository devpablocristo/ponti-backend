package models

import (
	"time"

	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/field/usecases/domain"
)

type Field struct {
	ID          int64     `gorm:"primaryKey;autoIncrement;column:id"`
	ProjectID   int64     `gorm:"index;column:project_id"`
	Name        string    `gorm:"size:100;not null;column:name"`
	LeaseTypeID int64     `gorm:"not null;index;column:lease_type_id"`
	CreatedAt   time.Time `gorm:"autoCreateTime;column:created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime;column:updated_at"`
}

func (m Field) ToDomain() *domain.Field {
	d := &domain.Field{
		ID:          m.ID,
		ProjectID:   m.ProjectID,
		Name:        m.Name,
		LeaseTypeID: m.LeaseTypeID,
	}
	return d
}

func FromDomain(d *domain.Field) *Field {
	m := &Field{
		ID:          d.ID,
		ProjectID:   d.ProjectID,
		Name:        d.Name,
		LeaseTypeID: d.LeaseTypeID,
	}
	return m
}
