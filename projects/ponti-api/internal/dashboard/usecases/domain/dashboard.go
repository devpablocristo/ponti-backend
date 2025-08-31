package domain

import (
	shareddomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/domain"
)

type Dashboard struct {
	ID int64

	shareddomain.Base
}
