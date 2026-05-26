package domain

import (
	shareddomain "github.com/devpablocristo/ponti-backend/internal/shared/domain"
)

type Category struct {
	ID     int64
	Name   string
	TypeID int64

	shareddomain.Base
}

// ListFilters contiene los filtros opcionales para `ListCategories`.
// Conviene mantener la convención struct (alineada con `actor.ListFilters`)
// para que futuros filtros se agreguen sin tocar signatures.
type ListFilters struct {
	TypeID *int64
}
