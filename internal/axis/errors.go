package axis

import (
	"errors"
	"fmt"

	domainerr "github.com/devpablocristo/platform/errors/go/domainerr"
)

// ErrNotConfigured se retorna cuando el cliente fue construido sin baseURL.
// Los usecases deben tratarlo como "AI no disponible" y caer a un fallback
// dummy (mismo patrón que el cliente actual de ponti-ai).
var ErrNotConfigured = errors.New("axis: companion not configured")

// mapHTTPError traduce status de Companion a errores del catálogo de Ponti.
// El body se incluye en el mensaje solo cuando no es JSON estructurado (no
// queremos leakear stacks ni IDs internos en el message).
func mapHTTPError(status int, body []byte) error {
	preview := string(body)
	if len(preview) > 200 {
		preview = preview[:200] + "..."
	}
	switch {
	case status == 401:
		return domainerr.Unauthorized("companion: authentication failed")
	case status == 403:
		return domainerr.Forbidden("companion: forbidden")
	case status == 404:
		return domainerr.NotFound("companion: resource not found")
	case status == 409:
		return domainerr.Conflict(fmt.Sprintf("companion: conflict: %s", preview))
	case status >= 400 && status < 500:
		return domainerr.Validation(fmt.Sprintf("companion: bad request: %s", preview))
	case status >= 500:
		return domainerr.Unavailable(fmt.Sprintf("companion: upstream error (%d)", status))
	default:
		return domainerr.Internal(fmt.Sprintf("companion: unexpected status %d: %s", status, preview))
	}
}
