package monitoring

import (
	"context"
)

// Repository define la interfaz para interactuar con la base de datos.
type Repository interface {
	CheckDbConn(ctx context.Context) error
}

// UseCases define la interfaz para los casos de uso relacionados con el monitoreo.
type UseCases interface {
	CheckDbConn(ctx context.Context) error
}
