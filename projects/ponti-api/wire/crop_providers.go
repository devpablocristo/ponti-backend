package wire

import (
	"github.com/google/wire"

	pgorm "github.com/alphacodinggroup/ponti-backend/pkg/databases/sql/gorm"
	mwr "github.com/alphacodinggroup/ponti-backend/pkg/http/middlewares/gin"
	pgin "github.com/alphacodinggroup/ponti-backend/pkg/http/servers/gin"
	config "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/cmd/config"

	crop "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/crop"
)

// ProvideCropRepository crea la implementación concreta de crop.Repository.
func ProvideCropRepository(repo crop.GormEnginePort) *crop.Repository {
	return crop.NewRepository(repo)
}

// ProvideCropRepositoryPort adapta *crop.Repository a la interfaz crop.RepositoryPort.
func ProvideCropRepositoryPort(r *crop.Repository) crop.RepositoryPort {
	return r
}

// ProvideCropUseCases agrupa repositorios en crop.UseCases.
func ProvideCropUseCases(
	rep crop.RepositoryPort,
) *crop.UseCases {
	return crop.NewUseCases(rep)
}

// ProvideCropUseCasesPort adapta *crop.UseCases a la interfaz crop.UseCasesPort.
func ProvideCropUseCasesPort(uc *crop.UseCases) crop.UseCasesPort {
	return uc
}

// ProvideCropHandler construye el handler HTTP para Crop.
func ProvideCropHandler(
	server crop.GinServerPort,
	useCases crop.UseCasesPort,
	cfg crop.ConfigAPIPort,
	middlewares crop.MiddlewaresPort,
) *crop.Handler {
	return crop.NewHandler(useCases, server, cfg, middlewares)
}

// ProvideCropAPIConfig extrae la configuración específica de API para Crop.
func ProvideCropAPIConfig(cfg *config.ConfigSet) crop.ConfigAPIPort {
	return &cfg.API
}

// ProvideCropGormEnginePort adapta *pgorm.Repository a crop.GormEnginePort.
func ProvideCropGormEnginePort(repo *pgorm.Repository) crop.GormEnginePort {
	return repo
}

// ProvideCropGinServerPort adapta *pgin.Server a crop.GinServerPort.
func ProvideCropGinServerPort(srv *pgin.Server) crop.GinServerPort {
	return srv
}

// ProvideCropMiddlewaresPort adapta *mwr.Middlewares a crop.MiddlewaresPort.
func ProvideCropMiddlewaresPort(m *mwr.Middlewares) crop.MiddlewaresPort {
	return m
}

// CropSet expone todos los providers necesarios para Crop.
var CropSet = wire.NewSet(
	ProvideCropRepository,
	ProvideCropRepositoryPort,
	ProvideCropUseCases,
	ProvideCropUseCasesPort,
	ProvideCropHandler,
	ProvideCropAPIConfig,
	ProvideCropGormEnginePort,
	ProvideCropGinServerPort,
	ProvideCropMiddlewaresPort,
)
