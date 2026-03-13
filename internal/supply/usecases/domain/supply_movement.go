package domain

import (
	"time"

	investordomain "github.com/alphacodinggroup/ponti-backend/internal/investor/usecases/domain"
	providerdomain "github.com/alphacodinggroup/ponti-backend/internal/provider/usecases/domain"
	shareddomain "github.com/alphacodinggroup/ponti-backend/internal/shared/domain"
	"github.com/shopspring/decimal"
)

type SupplyMovement struct {
	ID                   int64
	StockId              int64
	Quantity             decimal.Decimal
	MovementType         string
	MovementDate         *time.Time
	ReferenceNumber      string
	ProjectId            int64
	ProjectDestinationId int64
	Supply               *Supply
	Investor             *investordomain.Investor
	Provider             *providerdomain.Provider
	OriginProjectID      *int64
	OriginProjectName    *string
	DestinationProjectID *int64
	DestinationProject   *string
	DestinationCustomer  *string
	DestinationCampaign  *string
	IsEntry              bool
	shareddomain.Base
}
