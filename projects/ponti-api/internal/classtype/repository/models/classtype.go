package models

import (
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/classtype/usecases/domain"
)

type ClassType struct {
	ID   int64  `gorm:"primaryKey;autoIncrement;column:id"`
	Name string `gorm:"type:varchar(50);unique;not null"`
}

func (m *ClassType) ToDomain() *domain.ClassType {
	return &domain.ClassType{
		ID:   m.ID,
		Name: m.Name,
	}
}
func FromDomain(d *domain.ClassType) *ClassType {
	return &ClassType{
		ID:   d.ID,
		Name: d.Name,
	}
}
