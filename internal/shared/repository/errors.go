package sharedrepo

import (
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/devpablocristo/saas-core/shared/domainerr"
)

// HandleGormError centraliza el manejo de ErrRecordNotFound y errores internos.
func HandleGormError(err error, entity string, id int64) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domainerr.NotFound(fmt.Sprintf("%s %d not found", entity, id))
	}
	return domainerr.Internal(fmt.Sprintf("failed to get %s", entity))
}
