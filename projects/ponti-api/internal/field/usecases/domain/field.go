package domain

import (
	leasetypedom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/leasetype/usecases/domain"
	lotdom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/lot/usecases/domain"
	shareddomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/domain"
)

type Field struct {
	ID                int64
	ProjectID         int64
	Name              string
	LeaseType         *leasetypedom.LeaseType
	LeaseTypePercent  *float64
	LeaseTypeValue    *float64
	Lots              []lotdom.Lot
	shareddomain.Base // Incluye CreatedAt, UpdatedAt, etc
}
