package domain

import (
	campdom "github.com/devpablocristo/ponti-backend/internal/campaign/usecases/domain"
	customerdom "github.com/devpablocristo/ponti-backend/internal/customer/usecases/domain"
	fieldom "github.com/devpablocristo/ponti-backend/internal/field/usecases/domain"
	investordom "github.com/devpablocristo/ponti-backend/internal/investor/usecases/domain"
	managerdom "github.com/devpablocristo/ponti-backend/internal/manager/usecases/domain"
	shareddomain "github.com/devpablocristo/ponti-backend/internal/shared/domain"
	"github.com/shopspring/decimal"
)

type Project struct {
	ID                 int64
	Name               string
	AdminCost          decimal.Decimal
	PlannedCost        decimal.Decimal
	Customer           customerdom.Customer
	Campaign           campdom.Campaign
	Managers           []managerdom.Manager
	Investors          []investordom.Investor
	AdminCostInvestors []investordom.Investor
	Fields             []fieldom.Field

	shareddomain.Base
}

type ListedProject struct {
	ID   int64
	Name string
}
