package sharedrepo

import (
	"fmt"

	"github.com/devpablocristo/core/saas/go/shared/domainerr"
)

// ValidateEntity valida que el payload no sea nil.
func ValidateEntity(entity any, name string) error {
	if entity == nil {
		return domainerr.Validation(name + " is nil")
	}
	return nil
}

// ValidateID valida que un ID sea mayor que cero.
func ValidateID(id int64, name string) error {
	if id <= 0 {
		return domainerr.Validation(fmt.Sprintf("invalid %s id: %d", name, id))
	}
	return nil
}
