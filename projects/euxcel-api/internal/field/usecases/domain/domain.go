// File: internal/field/usecases/domain/field.go
package domain

import (
	lotdom "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/lot/usecases/domain"
)

type Field struct {
	ID          int64
	Name        string
	LeaseTypeID int64
	Lots        []lotdom.Lot
}
