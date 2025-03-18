package dto

import (
	"github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/supplier/usecases/domain"
)

// Supplier is the DTO for a supplier.
type Supplier struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// ToDomain converts the DTO to the domain entity.
func (s Supplier) ToDomain() *domain.Supplier {
	return &domain.Supplier{
		ID:   s.ID,
		Name: s.Name,
	}
}

// FromDomain converts a domain Supplier to the DTO.
func FromDomain(d domain.Supplier) *Supplier {
	return &Supplier{
		ID:   d.ID,
		Name: d.Name,
	}
}
