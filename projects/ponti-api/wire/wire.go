//go:build wireinject
// +build wireinject

package wire

import (
	gorm "github.com/alphacodinggroup/ponti-backend/pkg/databases/sql/gorm"
	mdw "github.com/alphacodinggroup/ponti-backend/pkg/http/middlewares/gin"
	ginsrv "github.com/alphacodinggroup/ponti-backend/pkg/http/servers/gin"
	"github.com/google/wire"

	config "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/cmd/config"

	campaign "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/campaign"
	crop "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/crop"
	customer "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/customer"
	field "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/field"
	investor "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/investor"
	lot "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/lot"
	manager "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/manager"
	project "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/project"
)

type Dependencies struct {
	ConfigLoader   config.Loader
	GinServer      ginsrv.Server
	GormRepository gorm.Repository

	Middlewares *mdw.Middlewares

	CropHandler     *crop.Handler
	CustomerHandler *customer.Handler
	ManagerHandler  *manager.Handler
	FieldHandler    *field.Handler
	InvestorHandler *investor.Handler
	LotHandler      *lot.Handler
	ProjectHandler  *project.Handler
	CampaignHandler *campaign.Handler

	CropUseCases     crop.UseCases
	CustomerUseCases customer.UseCases
	CampaignUseCases campaign.UseCases

	FieldUseCases    field.UseCases
	InvestorUseCases investor.UseCases
	LotUseCases      lot.UseCases
	ProjectUseCases  project.UseCases
}

func Initialize() (*Dependencies, error) {
	wire.Build(
		ProvideConfigLoader,
		ProvideGinServer,
		ProvideGormRepository,
		ProvideJwtMiddleware,
		ProvideMiddlewares,

		ProvideCropRepository,
		ProvideCropUseCases,
		ProvideCropHandler,

		ProvideCustomerRepository,
		ProvideCustomerUseCases,
		ProvideCustomerHandler,

		ProvideCampaignRepository,
		ProvideCampaignUseCases,
		ProvideCampaigHandler,

		ProvideManagerRepository,
		ProvideManagerUseCases,
		ProvideManagerHandler,

		ProvideFieldRepository,
		ProvideFieldUseCases,
		ProvideFieldHandler,

		ProvideInvestorRepository,
		ProvideInvestorUseCases,
		ProvideInvestorHandler,

		ProvideLotRepository,
		ProvideLotUseCases,
		ProvideLotHandler,

		ProvideProjectRepository,
		ProvideProjectUseCases,
		ProvideProjectHandler,

		wire.Struct(new(Dependencies), "*"),
	)
	return &Dependencies{}, nil
}
