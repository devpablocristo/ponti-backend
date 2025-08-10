package domain

import (
	classdomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/classtype/usecases/domain"
	shareddomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/domain"
)

type Supply struct {
	ID         int64
	ProjectID  int64
	Name       string
	UnitID     int64
	UnitName   string
	Price      float64
	CategoryID int64
	CategoryName string
	Type       classdomain.ClassType

	shareddomain.Base // Audit fields
}
