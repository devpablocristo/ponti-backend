//go:build wireinject
// +build wireinject

package wire

import (
	gorm "github.com/alphacodinggroup/ponti-backend/pkg/databases/sql/gorm"
	mwr "github.com/alphacodinggroup/ponti-backend/pkg/http/middlewares/gin"
	gin "github.com/alphacodinggroup/ponti-backend/pkg/http/servers/gin"
	sug "github.com/alphacodinggroup/ponti-backend/pkg/words-suggesters/trigram-search"

	config "github.com/alphacodinggroup/ponti-backend/cmd/config"
	bparams "github.com/alphacodinggroup/ponti-backend/internal/business-parameters"
	campaign "github.com/alphacodinggroup/ponti-backend/internal/campaign"
	category "github.com/alphacodinggroup/ponti-backend/internal/category"
	classtype "github.com/alphacodinggroup/ponti-backend/internal/class-type"
	commercialization "github.com/alphacodinggroup/ponti-backend/internal/commercialization"
	crop "github.com/alphacodinggroup/ponti-backend/internal/crop"
	customer "github.com/alphacodinggroup/ponti-backend/internal/customer"
	dashboard "github.com/alphacodinggroup/ponti-backend/internal/dashboard"
	data_integrity "github.com/alphacodinggroup/ponti-backend/internal/data-integrity"
	dollar "github.com/alphacodinggroup/ponti-backend/internal/dollar"
	field "github.com/alphacodinggroup/ponti-backend/internal/field"
	investor "github.com/alphacodinggroup/ponti-backend/internal/investor"
	invoice "github.com/alphacodinggroup/ponti-backend/internal/invoice"
	labor "github.com/alphacodinggroup/ponti-backend/internal/labor"
	leasetype "github.com/alphacodinggroup/ponti-backend/internal/lease-type"
	lot "github.com/alphacodinggroup/ponti-backend/internal/lot"
	manager "github.com/alphacodinggroup/ponti-backend/internal/manager"
	provider "github.com/alphacodinggroup/ponti-backend/internal/provider"
	project "github.com/alphacodinggroup/ponti-backend/internal/project"
	report "github.com/alphacodinggroup/ponti-backend/internal/report"
	"github.com/alphacodinggroup/ponti-backend/internal/stock"
	supply "github.com/alphacodinggroup/ponti-backend/internal/supply"
	workorder "github.com/alphacodinggroup/ponti-backend/internal/work-order"
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
	DataIntegrityHandler      *data_integrity.Handler
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
		wire.Struct(new(Dependencies), "*"),
	)
	return &Dependencies{}, nil
}
