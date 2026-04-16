package domain

import (
	shareddomain "github.com/devpablocristo/ponti-backend/internal/shared/domain"
)

type Crop struct {
	ID   int64
	Name string

	shareddomain.Base
}
