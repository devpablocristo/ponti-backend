package wire

import (
	"errors"

	gorm "github.com/alphacodinggroup/euxcel-backend/pkg/databases/sql/gorm"
	mdw "github.com/alphacodinggroup/euxcel-backend/pkg/http/middlewares/gin"
	ginsrv "github.com/alphacodinggroup/euxcel-backend/pkg/http/servers/gin"

	investor "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/investor"
)

func ProvideInvestorRepository(repo gorm.Repository) (investor.Repository, error) {
	if repo == nil {
		return nil, errors.New("gorm repository cannot be nil")
	}
	return investor.NewRepository(repo), nil
}

func ProvideInvestorUseCases(repo investor.Repository) investor.UseCases {
	return investor.NewUseCases(repo)
}

func ProvideInvestorHandler(server ginsrv.Server, usecases investor.UseCases, middlewares *mdw.Middlewares) *investor.Handler {
	return investor.NewHandler(server, usecases, middlewares)
}
