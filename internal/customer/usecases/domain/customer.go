package domain

import (
	"time"

	shareddomain "github.com/devpablocristo/ponti-backend/internal/shared/domain"
)

type Customer struct {
	ID         int64
	Name       string
	ArchivedAt *time.Time

	// TaxID es transitorio (no se persiste en customers): se pasa al Identity Gate
	// para resolver/crear el actor por CUIT. La CUIT vive en actor_keys.
	TaxID *string `json:"-"`

	shareddomain.Base
}

type ListedCustomer struct {
	ID   int64
	Name string
}
