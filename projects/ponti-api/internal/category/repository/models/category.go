package models

import (
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/category/usecases/domain"
)

type Category struct {
	ID   int64  `gorm:"primaryKey;autoIncrement;column:id"`
	Name string `gorm:"type:varchar(50);unique;not null"`
}

func (m *Category) ToDomain() *domain.Category {
	return &domain.Category{
		ID:   m.ID,
		Name: m.Name,
	}
}

func FromDomain(d *domain.Category) *Category {
	return &Category{
		ID:   d.ID,
		Name: d.Name,
	}
}
