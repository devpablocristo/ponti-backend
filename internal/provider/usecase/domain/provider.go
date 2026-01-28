package domain

import (
	shareddomain "github.com/alphacodinggroup/ponti-backend/internal/shared/domain"
)

type Provider struct {
	ID   int64
	Name string
	shareddomain.Base
}
