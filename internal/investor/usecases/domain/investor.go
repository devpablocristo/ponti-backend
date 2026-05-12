package domain

import (
	"time"

	shareddomain "github.com/devpablocristo/ponti-backend/internal/shared/domain"
)

type Investor struct {
	ID         int64
	Name       string
	Percentage int
	ActorID    *int64
	ArchivedAt *time.Time
	shareddomain.Base
}
