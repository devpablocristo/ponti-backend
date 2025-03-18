package wire

import (
	"errors"

	gorm "github.com/alphacodinggroup/euxcel-backend/pkg/databases/sql/gorm"
	mdw "github.com/alphacodinggroup/euxcel-backend/pkg/http/middlewares/gin"
	ginsrv "github.com/alphacodinggroup/euxcel-backend/pkg/http/servers/gin"

	"github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/supplier"
)

func ProvideSupplierRepository(repo gorm.Repository) (supplier.Repository, error) {
	if repo == nil {
		return nil, errors.New("gorm repository cannot be nil")
	}
	return supplier.NewRepository(repo), nil
}

func ProvideSupplierUseCases(
	repo supplier.Repository,
) supplier.UseCases {
	return supplier.NewUseCases(repo)
}

func ProvideSupplierHandler(
	server ginsrv.Server,
	usecases supplier.UseCases,
	middlewares *mdw.Middlewares,
) *supplier.Handler {
	return supplier.NewHandler(server, usecases, middlewares)
}
