//go:build wireinject
// +build wireinject

package wire

import (
	mwr "github.com/devpablocristo/ponti-backend/internal/platform/http/middlewares/gin"
	gin "github.com/devpablocristo/ponti-backend/internal/platform/http/servers/gin"
	gorm "github.com/devpablocristo/ponti-backend/internal/platform/persistence/gorm"
	sug "github.com/devpablocristo/ponti-backend/internal/platform/words-suggesters/trigram-search"

	config "github.com/devpablocristo/ponti-backend/cmd/config"
	admin "github.com/devpablocristo/ponti-backend/internal/admin"
	ai "github.com/devpablocristo/ponti-backend/internal/ai"
	bparams "github.com/devpablocristo/ponti-backend/internal/business-parameters"
	campaign "github.com/devpablocristo/ponti-backend/internal/campaign"
	category "github.com/devpablocristo/ponti-backend/internal/category"
	classtype "github.com/devpablocristo/ponti-backend/internal/class-type"
	commercialization "github.com/devpablocristo/ponti-backend/internal/commercialization"
	crop "github.com/devpablocristo/ponti-backend/internal/crop"
	customer "github.com/devpablocristo/ponti-backend/internal/customer"
	dashboard "github.com/devpablocristo/ponti-backend/internal/dashboard"
	dataintegrity "github.com/devpablocristo/ponti-backend/internal/data-integrity"
	dollar "github.com/devpablocristo/ponti-backend/internal/dollar"
	field "github.com/devpablocristo/ponti-backend/internal/field"
	investor "github.com/devpablocristo/ponti-backend/internal/investor"
	invoice "github.com/devpablocristo/ponti-backend/internal/invoice"
	labor "github.com/devpablocristo/ponti-backend/internal/labor"
	leasetype "github.com/devpablocristo/ponti-backend/internal/lease-type"
	lot "github.com/devpablocristo/ponti-backend/internal/lot"
	manager "github.com/devpablocristo/ponti-backend/internal/manager"
	project "github.com/devpablocristo/ponti-backend/internal/project"
	provider "github.com/devpablocristo/ponti-backend/internal/provider"
	report "github.com/devpablocristo/ponti-backend/internal/report"
	"github.com/devpablocristo/ponti-backend/internal/stock"
	supply "github.com/devpablocristo/ponti-backend/internal/supply"
	workorder "github.com/devpablocristo/ponti-backend/internal/work-order"
	"github.com/google/wire"
)

type Dependencies struct {
	Config                    *config.Config
	GinEngine                 *gin.Server
	GormRepo                  *gorm.Repository
	Middlewares               *mwr.Middlewares
	WordsSuggester            *sug.WordsSuggester
	CustomerHandler           *customer.Handler
	CampaignHandler           *campaign.Handler
	DashboardHandler          *dashboard.Handler
	DataIntegrityHandler      *dataintegrity.Handler
	InvestorHandler           *investor.Handler
	CropHandler               *crop.Handler
	LotHandler                *lot.Handler
	FieldHandler              *field.Handler
	ManagerHandler            *manager.Handler
	ProjectHandler            *project.Handler
	ProviderHandler           *provider.Handler
	ReportHandler             *report.ReportHandler
	LeaseTypeHandler          *leasetype.Handler
	SupplyHandler             *supply.Handler
	CategoryHandler           *category.Handler
	BusinessParametersHandler *bparams.Handler
	ClassTypeHandler          *classtype.Handler
	DollarHandler             *dollar.Handler
	WorkOrderHandler          *workorder.Handler
	LaborHandler              *labor.Handler
	InvoiceHandler            *invoice.Handler
	CommercializationHandler  *commercialization.Handler
	StockHandler              *stock.Handler
	StockUseCases             *stock.UseCases
	AIHandler                 *ai.Handler
	AdminHandler              *admin.Handler
}

func Initialize() (*Dependencies, error) {
	wire.Build(
		ConfigSet,
		GormSet,
		GinSet,
		MiddlewareSet,
		SuggesterSet,
		CustomerSet,
		CampaignSet,
		DashboardSet,
		DataIntegritySet,
		AISet,
		InvestorSet,
		CropSet,
		CommercializationSet,
		LotSet,
		FieldSet,
		ManagerSet,
		ProviderSet,
		ProjectSet,
		ReportSet,
		LeaseTypeSet,
		SupplySet,
		CategorySet,
		BusinessParametersSet,
		ClassTypeSet,
		DollarSet,
		WorkOrderSet,
		LaborSet,
		StockSet,
		InvoiceSet,
		AdminSet,
		wire.Struct(new(Dependencies), "*"),
	)
	return &Dependencies{}, nil
}
