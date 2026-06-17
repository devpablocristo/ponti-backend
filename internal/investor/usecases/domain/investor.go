package domain

import (
	"time"

	shareddomain "github.com/devpablocristo/ponti-backend/internal/shared/domain"
)

type Investor struct {
	ID         int64
	Name       string
	Percentage int
	ArchivedAt *time.Time
	TaxID      *string `json:"-"` // transitorio: CUIT para el Identity Gate
	shareddomain.Base
}
