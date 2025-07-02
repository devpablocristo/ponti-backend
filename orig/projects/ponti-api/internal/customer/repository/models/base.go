package models

import (
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/base"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/customer/usecases/domain"
)

type Customer struct {
	ID   int64  `gorm:"primaryKey;autoIncrement"`
	Name string `gorm:"type:varchar(255);not null;unique"`
	base.BaseModel
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
