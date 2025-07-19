package models

import (
	"time"

	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/dollar/usecases/domain"
)

type ProjectDollarValue struct {
	ID           int64     `gorm:"primaryKey;autoIncrement"`
	ProjectID    int64     `gorm:"not null;index"`
	Year         int64     `gorm:"not null;index"`
	Month        string    `gorm:"size:20;not null;index"`
	StartValue   float64   `gorm:"type:numeric(12,2);not null"`
	EndValue     float64   `gorm:"type:numeric(12,2);not null"`
	AverageValue float64   `gorm:"type:numeric(12,2);not null"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime"`
}

func FromDomain(d *domain.DollarAverage) *ProjectDollarValue {
	return &ProjectDollarValue{
		ID:           d.ID,
		ProjectID:    d.ProjectID,
		Year:         d.Year,
		Month:        d.Month,
		StartValue:   d.StartValue,
		EndValue:     d.EndValue,
		AverageValue: d.AvgValue,
		CreatedAt:    d.CreatedAt,
		UpdatedAt:    d.UpdatedAt,
	}
}

func (m *ProjectDollarValue) ToDomain() *domain.DollarAverage {
	return &domain.DollarAverage{
		ID:         m.ID,
		ProjectID:  m.ProjectID,
		Year:       m.Year,
		Month:      m.Month,
		StartValue: m.StartValue,
		EndValue:   m.EndValue,
		AvgValue:   m.AverageValue,
		CreatedAt:  m.CreatedAt,
		UpdatedAt:  m.UpdatedAt,
	}
}
