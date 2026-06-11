// Package domain define modelos de dominio para proveedores.
package domain

import (
	"time"

	shareddomain "github.com/devpablocristo/ponti-backend/internal/shared/domain"
)

type Provider struct {
	ID         int64      `json:"id"`
	Name       string     `json:"name"`
	ArchivedAt *time.Time `json:"archived_at,omitempty"`
	TaxID      *string    `json:"-"` // transitorio: CUIT para el Identity Gate
	shareddomain.Base
}
