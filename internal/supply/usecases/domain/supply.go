package domain

import (
	"time"

	classdomain "github.com/alphacodinggroup/ponti-backend/internal/class-type/usecases/domain"
	shareddomain "github.com/alphacodinggroup/ponti-backend/internal/shared/domain"
	"github.com/shopspring/decimal"
)

func ZeroPrice() decimal.Decimal {
	return decimal.Zero
}

type SupplyOrigin struct {
	Type            string
	SourceProjectID *int64
	SourceProject   string
	MovementID      *int64
	ReferenceNumber string
	ProviderName    string
	MovementDate    *time.Time
}

type Supply struct {
	ID        int64
	ProjectID int64
	Name      string
	UnitID    int64
	UnitName  string
	Price     decimal.Decimal
	IsPartialPrice bool
	IsPending      bool
	CategoryID     int64
	CategoryName   string
	Type           classdomain.ClassType
	Origin         *SupplyOrigin

	shareddomain.Base // Audit fields
}
