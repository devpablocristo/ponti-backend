package domain

import (
	shareddomain "github.com/devpablocristo/ponti-backend/internal/shared/domain"
)

type Campaign struct {
	ID        int64
	Name      string
	ProjectID int64

	shareddomain.Base
}
