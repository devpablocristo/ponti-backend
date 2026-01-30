package wire

import (
	"github.com/google/wire"

	config "github.com/alphacodinggroup/ponti-backend/cmd/config"
	provider "github.com/alphacodinggroup/ponti-backend/internal/provider"
	pgorm "github.com/alphacodinggroup/ponti-backend/pkg/databases/sql/gorm"
	mwr "github.com/alphacodinggroup/ponti-backend/pkg/http/middlewares/gin"
	pgin "github.com/alphacodinggroup/ponti-backend/pkg/http/servers/gin"
)

// ProvideProviderRepository crea la implementación concreta de provider.Repository.
func ProvideProviderRepository(db provider.GormEnginePort) *provider.Repository {
	return provider.NewRepository(db)
}

// ProvideProviderRepositoryPort adapta *provider.Repository a provider.RepositoryPort.
func ProvideProviderRepositoryPort(r *provider.Repository) provider.RepositoryPort {
	return r
}

// ProvideProviderUseCases agrupa repositorios en provider.UseCases.
func ProvideProviderUseCases(repo provider.RepositoryPort) *provider.UseCases {
	return provider.NewUseCases(repo)
}

// ProvideProviderUseCasesPort adapta *provider.UseCases a provider.UseCasesPort.
func ProvideProviderUseCasesPort(uc *provider.UseCases) provider.UseCasesPort {
	return uc
}

// ProvideProviderHandler construye el handler HTTP para Provider.
func ProvideProviderHandler(
	server provider.GinEnginePort,
	useCases provider.UseCasesPort,
	cfg provider.ConfigAPIPort,
	middlewares provider.MiddlewaresEnginePort,
) *provider.Handler {
	return provider.NewHandler(useCases, server, cfg, middlewares)
}

// ProvideProviderConfigAPI extrae la configuración específica de API para Provider.
func ProvideProviderConfigAPI(cfg *config.Config) provider.ConfigAPIPort {
	return &cfg.API
}

// ProvideProviderGormEnginePort adapta *pgorm.Repository a provider.GormEnginePort.
func ProvideProviderGormEnginePort(r *pgorm.Repository) provider.GormEnginePort {
	return r
}

// ProvideProviderGinEnginePort adapta *pgin.Server a provider.GinEnginePort.
func ProvideProviderGinEnginePort(s *pgin.Server) provider.GinEnginePort {
	return s
}

// ProvideProviderMiddlewaresEnginePort adapta *mwr.Middlewares a provider.MiddlewaresEnginePort.
func ProvideProviderMiddlewaresEnginePort(m *mwr.Middlewares) provider.MiddlewaresEnginePort {
	return m
}

// ProviderSet expone todos los providers necesarios para Provider.
var ProviderSet = wire.NewSet(
	ProvideProviderRepository,
	ProvideProviderRepositoryPort,
	ProvideProviderUseCases,
	ProvideProviderUseCasesPort,
	ProvideProviderHandler,
	ProvideProviderConfigAPI,
	ProvideProviderGormEnginePort,
	ProvideProviderGinEnginePort,
	ProvideProviderMiddlewaresEnginePort,
)
