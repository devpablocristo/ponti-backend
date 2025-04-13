package wire

import (
	"errors"

	gorm "github.com/alphacodinggroup/euxcel-backend/pkg/databases/sql/gorm"
	mdw "github.com/alphacodinggroup/euxcel-backend/pkg/http/middlewares/gin"
	ginsrv "github.com/alphacodinggroup/euxcel-backend/pkg/http/servers/gin"

	crop "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/crop"
)

func ProvideCropRepository(repo gorm.Repository) (crop.Repository, error) {
	if repo == nil {
		return nil, errors.New("gorm repository cannot be nil")
	}
	return crop.NewRepository(repo), nil
}

func ProvideCropUseCases(repo crop.Repository) crop.UseCases {
	return crop.NewUseCases(repo)
}

func ProvideCropHandler(server ginsrv.Server, usecases crop.UseCases, middlewares *mdw.Middlewares) *crop.Handler {
	return crop.NewHandler(server, usecases, middlewares)
}
