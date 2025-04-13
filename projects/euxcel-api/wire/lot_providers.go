package wire

import (
	"errors"

	gorm "github.com/alphacodinggroup/euxcel-backend/pkg/databases/sql/gorm"
	mdw "github.com/alphacodinggroup/euxcel-backend/pkg/http/middlewares/gin"
	ginsrv "github.com/alphacodinggroup/euxcel-backend/pkg/http/servers/gin"

	lot "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/lot"
)

func ProvideLotRepository(repo gorm.Repository) (lot.Repository, error) {
	if repo == nil {
		return nil, errors.New("gorm repository cannot be nil")
	}
	return lot.NewRepository(repo), nil
}

func ProvideLotUseCases(repo lot.Repository) lot.UseCases {
	return lot.NewUseCases(repo)
}

func ProvideLotHandler(server ginsrv.Server, usecases lot.UseCases, middlewares *mdw.Middlewares) *lot.Handler {
	return lot.NewHandler(server, usecases, middlewares)
}
