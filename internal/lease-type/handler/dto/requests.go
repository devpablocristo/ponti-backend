package dto

import domain "github.com/devpablocristo/ponti-backend/internal/lease-type/usecases/domain"

type CreateLeaseTypeRequest struct {
	Name string `json:"name" binding:"required"`
}

func (r *CreateLeaseTypeRequest) ToDomain() *domain.LeaseType {
	return &domain.LeaseType{Name: r.Name}
}

type UpdateLeaseTypeRequest struct {
	Name string `json:"name" binding:"required"`
}

func (r *UpdateLeaseTypeRequest) ToDomain(id int64) *domain.LeaseType {
	return &domain.LeaseType{ID: id, Name: r.Name}
}
