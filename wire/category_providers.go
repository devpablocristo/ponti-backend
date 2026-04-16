package wire

import (
	"github.com/google/wire"

	cfg "github.com/devpablocristo/ponti-backend/cmd/config"
	category "github.com/devpablocristo/ponti-backend/internal/category" // Corrected import for category
	mwr "github.com/devpablocristo/ponti-backend/internal/platform/http/middlewares/gin"
	pgin "github.com/devpablocristo/ponti-backend/internal/platform/http/servers/gin"
	gormpkg "github.com/devpablocristo/ponti-backend/internal/platform/persistence/gorm"
)

// ProvideCategoryGormEnginePort ...
func ProvideCategoryGormEnginePort(r *gormpkg.Repository) category.GormEnginePort {
	return r
}
func ProvideCategoryRepository(r category.GormEnginePort) *category.Repository {
	return category.NewRepository(r)
}
func ProvideCategoryRepositoryPort(repo *category.Repository) category.RepositoryPort {
	return repo
}

// ProvideCategoryConfigAPI ...
func ProvideCategoryConfigAPI(c *cfg.Config) category.ConfigAPIPort {
	return &c.API
}

// ProvideCategoryGinEnginePort ...
func ProvideCategoryGinEnginePort(s *pgin.Server) category.GinEnginePort {
	return s
}
func ProvideCategoryMiddlewaresEnginePort(m *mwr.Middlewares) category.MiddlewaresEnginePort {
	return m
}

// ProvideCategoryUseCases ...
func ProvideCategoryUseCases(rep category.RepositoryPort) *category.UseCases {
	return category.NewUseCases(rep)
}

func ProvideCategoryUseCasesPort(u *category.UseCases) category.UseCasesPort {
	return u
}

// ProvideCategoryHandler ...
func ProvideCategoryHandler(
	server category.GinEnginePort,
	ucs category.UseCasesPort,
	cfg category.ConfigAPIPort,
	mws category.MiddlewaresEnginePort,
) *category.Handler {
	return category.NewHandler(ucs, server, cfg, mws)
}

// CategorySet ...
var CategorySet = wire.NewSet(
	ProvideCategoryGormEnginePort,
	ProvideCategoryRepository,
	ProvideCategoryRepositoryPort,
	ProvideCategoryConfigAPI,
	ProvideCategoryGinEnginePort,
	ProvideCategoryMiddlewaresEnginePort,

	ProvideCategoryUseCases,
	ProvideCategoryUseCasesPort,
	ProvideCategoryHandler,
)
