package pkgutils

import (
	"fmt"
	"strconv"
	"strings"

	pkgtypes "github.com/alphacodinggroup/ponti-backend/pkg/types"
)

// ValidateStringID converts a string to uint and checks if it's positive.
func ValidateStringID(idParam string) (uint, error) {
	id, err := strconv.Atoi(strings.TrimSpace(idParam))
	if err != nil || id <= 0 {
		return 0, pkgtypes.NewInvalidIDError("invalid ID parameter", err)
	}
	return uint(id), nil
}

// ValidateNumericID ensures a numeric ID is > 0.
func ValidateNumericID(id int64) error {
	if id <= 0 {
		return fmt.Errorf("id must be greater than 0")
	}
	return nil
}
