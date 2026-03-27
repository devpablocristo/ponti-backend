// Package shareddomain contiene tipos de dominio compartidos.
package shareddomain

import (
	"time"
)

type Base struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	CreatedBy *string
	UpdatedBy *string
}
