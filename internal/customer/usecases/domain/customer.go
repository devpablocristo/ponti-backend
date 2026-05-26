package domain

import (
	"time"

	shareddomain "github.com/devpablocristo/ponti-backend/internal/shared/domain"
)

type Customer struct {
	ID         int64
	Name       string
	ActorID    *int64
	ArchivedAt *time.Time

	shareddomain.Base
}

type ListedCustomer struct {
	ID      int64
	Name    string
	ActorID *int64
}
