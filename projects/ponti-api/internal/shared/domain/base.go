package shareddomain

import (
	"time"
)

type Base struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	CreatedBy *int64
	UpdatedBy *int64
	Version   int64
}

// IncrementVersion incrementa la versión del modelo
func (b *Base) IncrementVersion() {
	b.Version++
}

// GetVersion retorna la versión actual
func (b *Base) GetVersion() int64 {
	return b.Version
}
