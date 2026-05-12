// Package domain define modelos de dominio para proveedores.
package domain

import (
	shareddomain "github.com/devpablocristo/ponti-backend/internal/shared/domain"
)

type Provider struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	ActorID *int64 `json:"actor_id,omitempty"`
	shareddomain.Base
}
