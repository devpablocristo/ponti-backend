package domain

import (
	"time"

	leasetypedom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/leasetype/usecases/domain"
	lotdom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/lot/usecases/domain"
)

type Field struct {
	ID               int64
	ProjectID        int64
	Name             string
	LeaseType        *leasetypedom.LeaseType
	LeaseTypePercent *float64
	LeaseTypeValue   *float64
	Lots             []lotdom.Lot
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
