//go:build wireinject
// +build wireinject

package wire

import (
	"github.com/google/wire"

	gorm "github.com/alphacodinggroup/ponti-backend/pkg/databases/sql/gorm"
	mdw "github.com/alphacodinggroup/ponti-backend/pkg/http/middlewares/gin"
	ginsrv "github.com/alphacodinggroup/ponti-backend/pkg/http/servers/gin"
	config "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/cmd/config"

	crop "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/crop"
	customer "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/customer"
	field "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/field"
	investor "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/investor"
	lot "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/lot"
	manager "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/manager"
	project "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/project"
)

type Dependencies struct {
	AppConfig      *config.ConfigSet
	GinServer      *ginsrv.Server
	GormRepository *gorm.Repository

	Middlewares *mdw.Middlewares

	CropHandler     *crop.Handler
	CustomerHandler *customer.Handler
	ManagerHandler  *manager.Handler
	FieldHandler    *field.Handler
	InvestorHandler *investor.Handler
	LotHandler      *lot.Handler
	ProjectHandler  *project.Handler

	CropUseCases     crop.UseCases
	CustomerUseCases customer.UseCases
	FieldUseCases    field.UseCases
	InvestorUseCases investor.UseCases
	LotUseCases      lot.UseCases
	ProjectUseCases  project.UseCases
}

func Initialize(cfgSet *config.ConfigSet) (*Dependencies, error) {
	wire.Build(
		// Inyectamos la configuración cargada en main.go
		wire.Value(cfgSet),

		// Servidor HTTP y repositorio
		GinSet,
		GormSet,

		// Middlewares
		GlobalMiddlewareSet,
		ValidationMiddlewareSet,
		AuthMiddlewareSet,

		// Crop
		ProvideCropRepository,
		ProvideCropUseCases,
		ProvideCropHandler,

		// Customer
		ProvideCustomerRepository,
		ProvideCustomerUseCases,
		ProvideCustomerHandler,

		// Manager
		ProvideManagerRepository,
		ProvideManagerUseCases,
		ProvideManagerHandler,

		// Field
		ProvideFieldRepository,
		ProvideFieldUseCases,
		ProvideFieldHandler,

		// Investor
		ProvideInvestorRepository,
		ProvideInvestorUseCases,
		ProvideInvestorHandler,

		// Lot
		ProvideLotRepository,
		ProvideLotUseCases,
		ProvideLotHandler,

		// Project
		ProvideProjectRepository,
		ProvideProjectUseCases,
		ProvideProjectHandler,

		// Construye el struct completo
		wire.Struct(new(Dependencies), "*"),
	)
	return &Dependencies{}, nil
}
