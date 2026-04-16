package dto

import domain "github.com/devpablocristo/ponti-backend/internal/class-type/usecases/domain"

type CreateClassTypeRequest struct {
	Name string `json:"name" binding:"required"`
}

func (r *CreateClassTypeRequest) ToDomain() *domain.ClassType {
	return &domain.ClassType{Name: r.Name}
}

type UpdateClassTypeRequest struct {
	Name string `json:"name" binding:"required"`
}

func (r *UpdateClassTypeRequest) ToDomain(id int64) *domain.ClassType {
	return &domain.ClassType{ID: id, Name: r.Name}
}
