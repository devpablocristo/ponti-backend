package dto

import (
	domain "github.com/devpablocristo/ponti-backend/internal/actors/usecases/domain"
)

// ResolveActorRequest es el body de POST /actors (resolve-or-create).
type ResolveActorRequest struct {
	Name        string  `json:"name" binding:"required,min=1"`
	TaxID       *string `json:"tax_id"`
	Role        string  `json:"role" binding:"required"`
	AllowCreate bool    `json:"allow_create"`
}

func (r ResolveActorRequest) ToDomain() domain.ResolveInput {
	return domain.ResolveInput{
		Name:        r.Name,
		TaxID:       r.TaxID,
		Role:        r.Role,
		AllowCreate: r.AllowCreate,
	}
}
