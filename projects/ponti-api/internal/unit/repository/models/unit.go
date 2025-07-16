package models

import (
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/unit/usecases/domain"
)

type Unit struct {
	ID   int64  `gorm:"primaryKey;autoIncrement;column:id"`
	Name string `gorm:"type:varchar(30);unique;not null"`
}

func (m *Unit) ToDomain() *domain.Unit {
	return &domain.Unit{
		ID:   m.ID,
		Name: m.Name,
	}
}

func FromDomain(d *domain.Unit) *Unit {
	return &Unit{
		ID:   d.ID,
		Name: d.Name,
	}
}
