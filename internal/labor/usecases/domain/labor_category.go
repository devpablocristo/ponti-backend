package domain

import sharedmodels "github.com/devpablocristo/ponti-backend/internal/shared/models"

type LaborCategory struct {
	ID          int64
	Name        string
	LaborTypeId int64
	LaborType   LaborType

	sharedmodels.Base
}

type LaborType struct {
	ID   int64
	Name string

	sharedmodels.Base
}
