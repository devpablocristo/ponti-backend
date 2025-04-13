package wire

import (
	"errors"

	gorm "github.com/alphacodinggroup/euxcel-backend/pkg/databases/sql/gorm"
	mdw "github.com/alphacodinggroup/euxcel-backend/pkg/http/middlewares/gin"
	ginsrv "github.com/alphacodinggroup/euxcel-backend/pkg/http/servers/gin"

	field "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/field"
)

func ProvideFieldRepository(repo gorm.Repository) (field.Repository, error) {
	if repo == nil {
		return nil, errors.New("gorm repository cannot be nil")
	}
	return field.NewRepository(repo), nil
}

func ProvideFieldUseCases(repo field.Repository) field.UseCases {
	return field.NewUseCases(repo)
}

func ProvideFieldHandler(server ginsrv.Server, usecases field.UseCases, middlewares *mdw.Middlewares) *field.Handler {
	return field.NewHandler(server, usecases, middlewares)
}
