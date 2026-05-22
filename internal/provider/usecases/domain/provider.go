// Package domain define modelos de dominio para proveedores.
package domain

import (
	shareddomain "github.com/devpablocristo/ponti-backend/internal/shared/domain"
)

type Provider struct {
	ID      int64
	Name    string
	ActorID *int64
	shareddomain.Base
}
