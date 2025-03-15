package models

import (
	"time"

	domain "github.com/devpablocristo/monorepo/projects/qh/internal/group/usecases/domain"
)

// Group representa un grupo en el sistema, con configuraciones para GORM.
type Group struct {
	ID                 string        `gorm:"type:uuid;primaryKey"`
	Name               string        `gorm:"size:255;not null"`
	Description        string        `gorm:"type:text"`
	Category           string        `gorm:"size:100"`
	CreatedAt          time.Time     `gorm:"autoCreateTime"`
	UpdatedAt          time.Time     `gorm:"autoUpdateTime"`
	Privacy            string        `gorm:"size:50;not null"` // "public" o "private"
	MaxMembers         *int          `gorm:"default:null"`
	ImageURL           string        `gorm:"size:255"`
	OrganizerID        string        `gorm:"not null"`
	IsOrganizerCompany bool          `gorm:"not null"`
	Status             string        `gorm:"size:50;not null;default:'active'"`               // "active", "inactive", etc.
	GroupMembers       []GroupMember `gorm:"foreignKey:GroupID;constraint:OnDelete:CASCADE;"` // Relación con los miembros
	AssociatedEventIDs []string      `gorm:"-"`                                               // IDs de eventos asociados (sin carga automática)
}

// GroupMember representa la relación muchos-a-muchos entre Usuarios y Grupos.
type GroupMember struct {
	GroupID  string    `gorm:"primaryKey"`
	UserID   string    `gorm:"primaryKey"`
	Role     string    `gorm:"size:50;not null"` // Ejemplo: "admin", "moderator", "member"
	JoinedAt time.Time `gorm:"autoCreateTime"`
}

func (g *Group) TableName() string {
	return "groups"
}

// FromDomain mapea un `domain.Group` a un `Group`.
func (g *Group) FromDomain(group *domain.Group) {
	g.ID = group.ID
	g.Name = group.Name
	g.Description = group.Description
	g.Category = group.Category
	g.CreatedAt = group.CreatedAt
	g.UpdatedAt = group.UpdatedAt
	g.Privacy = group.Privacy
	g.MaxMembers = group.MaxMembers
	g.ImageURL = group.ImageURL
	g.OrganizerID = group.OrganizerID
	g.IsOrganizerCompany = group.IsOrganizerCompany
	g.Status = group.Status

	// Mapear miembros del grupo
	g.GroupMembers = make([]GroupMember, len(group.GroupMembers))
	for i, member := range group.GroupMembers {
		g.GroupMembers[i] = GroupMember{
			GroupID:  member.GroupID,
			UserID:   member.UserID,
			Role:     member.Role,
			JoinedAt: member.JoinedAt,
		}
	}

	// Asociar solo los IDs de los eventos
	g.AssociatedEventIDs = group.AssociatedEventIDs
}

// ToDomain convierte un `Group` a un `domain.Group`.
func (g *Group) ToDomain() *domain.Group {
	groupMembers := make([]domain.GroupMember, len(g.GroupMembers))
	for i, member := range g.GroupMembers {
		groupMembers[i] = domain.GroupMember{
			GroupID:  member.GroupID,
			UserID:   member.UserID,
			Role:     member.Role,
			JoinedAt: member.JoinedAt,
		}
	}

	return &domain.Group{
		ID:                 g.ID,
		Name:               g.Name,
		Description:        g.Description,
		Category:           g.Category,
		CreatedAt:          g.CreatedAt,
		UpdatedAt:          g.UpdatedAt,
		Privacy:            g.Privacy,
		MaxMembers:         g.MaxMembers,
		ImageURL:           g.ImageURL,
		OrganizerID:        g.OrganizerID,
		IsOrganizerCompany: g.IsOrganizerCompany,
		Status:             g.Status,
		GroupMembers:       groupMembers,
		AssociatedEventIDs: g.AssociatedEventIDs,
	}
}

func (gm *GroupMember) TableName() string {
	return "group_members"
}
