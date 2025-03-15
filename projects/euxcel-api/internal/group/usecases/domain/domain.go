package domain

import (
	"time"
)

type Group struct {
	ID                 string
	Name               string
	Description        string
	Category           string
	CreatedAt          time.Time
	UpdatedAt          time.Time
	Privacy            string // "public" o "private"
	MaxMembers         *int   // Nil indica sin límite.
	ImageURL           string
	OrganizerID        string        // ID del organizador, puede ser persona o empresa.
	IsOrganizerCompany bool          // Indica si el organizador es una empresa.
	GroupMembers       []GroupMember // Lista de miembros.
	AssociatedEventIDs []string      // Lista de IDs de eventos relacionados.
	Status             string        // "active", "inactive", etc.
}

// GroupMember represents the many-to-many relationship between Users and Groups
type GroupMember struct {
	GroupID  string
	UserID   string
	Role     string // e.g., "admin", "moderator", "member"
	JoinedAt time.Time
}
