package domain

import (
	shareddomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/domain"
)

type Category struct {
	ID   int64  // unique id
	Name string // category name (e.g., "Herbicides", "Fertilizers", etc.)

	shareddomain.Base
}
