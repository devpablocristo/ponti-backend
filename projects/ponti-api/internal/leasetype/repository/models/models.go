package models

import (
	"time"

	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/leasetype/usecases/domain"
)

type LeaseType struct {
	ID        int64     `gorm:"primaryKey;autoIncrement;column:id"`
	Name      string    `gorm:"size:100;not null;unique;column:name"`
	CreatedAt time.Time `gorm:"autoCreateTime;column:created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime;column:updated_at"`
}

func (m *LeaseType) ToDomain() *domain.LeaseType {
	return &domain.LeaseType{
		ID:   m.ID,
		Name: m.Name,
	}
}

func FromDomainLeaseType(d *domain.LeaseType) *LeaseType {
	return &LeaseType{
		ID:   d.ID,
		Name: d.Name,
	}
}
