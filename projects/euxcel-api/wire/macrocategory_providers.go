package wire

import (
	"errors"

	gorm "github.com/alphacodinggroup/euxcel-backend/pkg/databases/sql/gorm"
	mdw "github.com/alphacodinggroup/euxcel-backend/pkg/http/middlewares/gin"
	ginsrv "github.com/alphacodinggroup/euxcel-backend/pkg/http/servers/gin"

	"github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/macrocategory"
)

func ProvideMacroCategoryRepository(repo gorm.Repository) (macrocategory.Repository, error) {
	if repo == nil {
		return nil, errors.New("gorm repository cannot be nil")
	}
	return macrocategory.NewRepository(repo), nil
}

func ProvideMacroCategoryUseCases(
	repo macrocategory.Repository,
) macrocategory.UseCases {
	return macrocategory.NewUseCases(repo)
}

func ProvideMacroCategoryHandler(
	server ginsrv.Server,
	usecases macrocategory.UseCases,
	middlewares *mdw.Middlewares,
) *macrocategory.Handler {
	return macrocategory.NewHandler(server, usecases, middlewares)
}
