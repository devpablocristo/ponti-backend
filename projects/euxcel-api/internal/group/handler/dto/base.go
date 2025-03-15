package dto

import (
	"time"

	"github.com/devpablocristo/monorepo/projects/qh/internal/group/usecases/domain"
)

type Group struct {
	Name               string   `json:"name" binding:"required,min=3,max=50"`
	Description        string   `json:"description" binding:"omitempty,max=250"`
	Category           string   `json:"category" binding:"omitempty,oneof=music sports tech art"`
	Privacy            string   `json:"privacy" binding:"required,oneof=public private"`
	MaxMembers         int      `json:"max_members" binding:"omitempty,gte=1"`
	ImageURL           string   `json:"image_url" binding:"omitempty,url"`
	OrganizerID        string   `json:"organizer_id" binding:"required"`
	IsOrganizerCompany bool     `json:"is_organizer_company"`
	Members            []string `json:"members" binding:"required,min=1,dive,gt=0"`
}

// ToDomain convierte un Group a un Group del dominio
func (d *Group) ToDomain() *domain.Group {
	// Convertir los miembros del DTO a GroupMember en el dominio
	groupMembers := make([]domain.GroupMember, len(d.Members))
	for i, memberID := range d.Members {
		role := "member"
		if memberID == d.OrganizerID {
			role = "admin" // El organizador es automáticamente un administrador
		}

		groupMembers[i] = domain.GroupMember{
			GroupID:  "", // Se asignará al guardar en la base de datos
			UserID:   memberID,
			Role:     role,
			JoinedAt: time.Now(), // Fecha de unión inicial
		}
	}

	// Manejar MaxMembers como puntero si está configurado
	var maxMembers *int
	if d.MaxMembers > 0 {
		maxMembers = &d.MaxMembers
	}

	// Construir el grupo del dominio
	return &domain.Group{
		Name:               d.Name,
		Description:        d.Description,
		Category:           d.Category,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
		Privacy:            d.Privacy,
		MaxMembers:         maxMembers,
		ImageURL:           d.ImageURL,
		OrganizerID:        d.OrganizerID,
		IsOrganizerCompany: d.IsOrganizerCompany,
		GroupMembers:       groupMembers,
		AssociatedEventIDs: []string{}, // Inicialmente sin eventos asociados
		Status:             "active",   // Estado inicial del grupo
	}
}
