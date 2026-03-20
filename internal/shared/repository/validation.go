package sharedrepo

import (
	"fmt"

	types "github.com/devpablocristo/ponti-backend/pkg/types"
)

// ValidateEntity valida que el payload no sea nil.
func ValidateEntity(entity any, name string) error {
	if entity == nil {
		return types.NewError(types.ErrValidation, name+" is nil", nil)
	}
	return nil
}

// ValidateID valida que un ID sea mayor que cero.
func ValidateID(id int64, name string) error {
	if id <= 0 {
		return types.NewInvalidIDError(fmt.Sprintf("invalid %s id: %d", name, id), nil)
	}
	return nil
}
