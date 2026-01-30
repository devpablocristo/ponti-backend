package sharedrepo

import (
	"errors"
	"fmt"

	"gorm.io/gorm"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
)

// HandleGormError centraliza el manejo de ErrRecordNotFound y errores internos.
func HandleGormError(err error, entity string, id int64) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return types.NewError(types.ErrNotFound, fmt.Sprintf("%s %d not found", entity, id), err)
	}
	return types.NewError(types.ErrInternal, fmt.Sprintf("failed to get %s", entity), err)
}
