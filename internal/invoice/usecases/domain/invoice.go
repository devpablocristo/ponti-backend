package domain

import (
	"time"

	shareddomain "github.com/devpablocristo/ponti-backend/internal/shared/domain"
)

type Invoice struct {
	ID          int64
	WorkOrderID int64
	InvestorID  int64
	Number      string
	Company     string
	Date        time.Time
	Status      string

	shareddomain.Base
}
