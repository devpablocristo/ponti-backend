package sharedhandlers

import (
	"strconv"

	"github.com/devpablocristo/core/errors/go/domainerr"
)

// ParseParamID mantiene compatibilidad con handlers viejos que aún reciben el valor ya extraído.
func ParseParamID(raw string, name string) (int64, error) {
	if raw == "" {
		return 0, domainerr.Validation(name + " is required")
	}
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, domainerr.Validation("invalid " + name)
	}
	return id, nil
}
