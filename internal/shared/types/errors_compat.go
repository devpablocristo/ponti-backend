package pkgtypes

import (
	"github.com/devpablocristo/core/errors/go/domainerr"
)

type ErrorType string

const (
	ErrInternal     ErrorType = "INTERNAL"
	ErrValidation   ErrorType = "VALIDATION_ERROR"
	ErrInvalidInput ErrorType = "VALIDATION_ERROR"
	ErrBadRequest   ErrorType = "VALIDATION_ERROR"
	ErrConflict     ErrorType = "CONFLICT"
	ErrNotFound     ErrorType = "NOT_FOUND"
)

func NewError(kind ErrorType, message string, _ error) error {
	switch kind {
	case ErrValidation:
		return domainerr.Validation(message)
	case ErrConflict:
		return domainerr.Conflict(message)
	case ErrNotFound:
		return domainerr.NotFound(message)
	default:
		return domainerr.Internal(message)
	}
}

func NewInvalidIDError(message string, _ error) error {
	return domainerr.Validation(message)
}

func GetErrorType(err error) (ErrorType, bool) {
	switch {
	case domainerr.IsKind(err, domainerr.KindValidation):
		return ErrValidation, true
	case domainerr.IsKind(err, domainerr.KindConflict):
		return ErrConflict, true
	case domainerr.IsKind(err, domainerr.KindNotFound):
		return ErrNotFound, true
	case domainerr.IsKind(err, domainerr.KindInternal):
		return ErrInternal, true
	default:
		return "", false
	}
}
