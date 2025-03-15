package monitoring

import (
	"context"
)

type useCases struct {
	repository Repository
}

// NewUseCases crea una nueva instancia de casos de uso de monitoreo.
func NewUseCases(r Repository) UseCases {
	return &useCases{
		repository: r,
	}
}

// CheckDbConn verifica la conexión a la base de datos usando el repositorio.
func (u *useCases) CheckDbConn(ctx context.Context) error {
	return u.repository.CheckDbConn(ctx)
}
