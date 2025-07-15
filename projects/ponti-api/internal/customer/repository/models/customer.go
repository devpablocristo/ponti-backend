package models

import (
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/customer/usecases/domain"

	sharedmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/models"
)

type Customer struct {
	ID   int64  `gorm:"primaryKey;autoIncrement"`
	Name string `gorm:"type:varchar(255);not null;unique"`
	sharedmodels.Base
}

func (c Customer) ToDomain() *domain.Customer {
	return &domain.Customer{
		ID:   c.ID,
		Name: c.Name,
	}
}

func FromDomain(d *domain.Customer) *Customer {
	m := &Customer{
		Name: d.Name,
	}
	if d.ID > 0 {
		m.ID = d.ID
	}
	return m
}
