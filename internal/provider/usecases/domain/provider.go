// Package domain define modelos de dominio para proveedores.
package domain

import (
	shareddomain "github.com/alphacodinggroup/ponti-backend/internal/shared/domain"
)

type Provider struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	shareddomain.Base
}
