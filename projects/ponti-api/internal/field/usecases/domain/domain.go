package domain

import (
	"time"

	lotdom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/lot/usecases/domain"
)

type Field struct {
	ID               int64
	ProjectID        int64
	Name             string
	LeaseTypeID      int64
	LeaseTypePercent *float64
	LeaseTypeValue   *float64
	Lots             []lotdom.Lot
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
