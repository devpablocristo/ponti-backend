package sharedrepo

import (
	"fmt"

	"github.com/devpablocristo/platform/errors/go/domainerr"
	gormdb "github.com/devpablocristo/platform/databases/postgres/go"
)

// HandleGormError centraliza el manejo de ErrRecordNotFound y errores internos.
func HandleGormError(err error, entity string, id int64) error {
	if gormdb.IsNotFound(err) {
		return domainerr.NotFound(fmt.Sprintf("%s %d not found", entity, id))
	}
	return domainerr.Internal(fmt.Sprintf("failed to get %s", entity))
}
