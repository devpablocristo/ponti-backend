// Package domain define modelos de dominio para managers.
package domain

import (
	"time"

	shareddomain "github.com/devpablocristo/ponti-backend/internal/shared/domain"
)

type Manager struct {
	ID         int64
	Name       string
	Type       string
	ArchivedAt *time.Time
	shareddomain.Base
}
