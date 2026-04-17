package domain

import (
	shareddomain "github.com/alphacodinggroup/ponti-backend/internal/shared/domain"
)

type Category struct {
	ID     int64
	Name   string
	TypeID int64

	shareddomain.Base
}
