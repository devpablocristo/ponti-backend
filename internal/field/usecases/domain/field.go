package domain

import (
	"github.com/shopspring/decimal"

	invdom "github.com/alphacodinggroup/ponti-backend/internal/investor/usecases/domain"
	leasetypedom "github.com/alphacodinggroup/ponti-backend/internal/lease-type/usecases/domain"
	lotdom "github.com/alphacodinggroup/ponti-backend/internal/lot/usecases/domain"
	shareddomain "github.com/alphacodinggroup/ponti-backend/internal/shared/domain"
)

type Field struct {
	ID                int64
	ProjectID         int64
	Name              string
	LeaseType         *leasetypedom.LeaseType
	LeaseTypePercent  *decimal.Decimal
	LeaseTypeValue    *decimal.Decimal
	Investors         []invdom.Investor
	Lots              []lotdom.Lot
	shareddomain.Base // Incluye CreatedAt, UpdatedAt, etc
}
