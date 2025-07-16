package domain

import (
	shareddomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/domain"
)

type Unit struct {
	ID   int64  // unique id
	Name string // unit name (e.g., "Lts", "Kg", "Ha")

	shareddomain.Base // Campos de auditoría (CreatedAt, UpdatedAt, etc)
}
