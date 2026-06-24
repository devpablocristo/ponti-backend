package dto

import (
	"strings"

	domain "github.com/devpablocristo/ponti-backend/internal/actors/usecases/domain"
)

// ResolveActorRequest es el body de POST /actors (resolve-or-create).
type ResolveActorRequest struct {
	Name           string  `json:"name" binding:"required,min=1"`
	TaxID          *string `json:"tax_id"`
	Role           string  `json:"role" binding:"required"`
	AllowCreate    bool    `json:"allow_create"`
	RejectExisting bool    `json:"reject_existing"`
}

func (r ResolveActorRequest) ToDomain() domain.ResolveInput {
	return domain.ResolveInput{
		Name:           r.Name,
		TaxID:          r.TaxID,
		Role:           r.Role,
		AllowCreate:    r.AllowCreate,
		RejectExisting: r.RejectExisting,
	}
}

// UpdateActorRequest edita los campos de display de un actor.
type UpdateActorRequest struct {
	DisplayName string `json:"display_name" binding:"required,min=1"`
	PartyType   string `json:"party_type" binding:"omitempty,oneof=org person unknown"`
}

// SetRolesRequest reemplaza el conjunto de roles del actor.
type SetRolesRequest struct {
	Roles []string `json:"roles" binding:"required"`
}

// SetTaxIDRequest corrige la clave fiscal (CUIT/DNI) del actor.
type SetTaxIDRequest struct {
	TaxID string `json:"tax_id" binding:"required,min=1"`
}

func (r UpdateActorRequest) ToDomain(id int64) *domain.Actor {
	pt := strings.ToLower(strings.TrimSpace(r.PartyType))
	if pt == "" {
		pt = "unknown"
	}
	return &domain.Actor{
		ID:          id,
		DisplayName: strings.TrimSpace(r.DisplayName),
		PartyType:   pt,
	}
}
