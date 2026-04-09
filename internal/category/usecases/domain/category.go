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
