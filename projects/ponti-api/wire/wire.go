//go:build wireinject
// +build wireinject

package wire

import (
	gorm "github.com/alphacodinggroup/ponti-backend/pkg/databases/sql/gorm"
	mwr "github.com/alphacodinggroup/ponti-backend/pkg/http/middlewares/gin"
	gin "github.com/alphacodinggroup/ponti-backend/pkg/http/servers/gin"
	sug "github.com/alphacodinggroup/ponti-backend/pkg/words-suggesters/trigram-search"
	config "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/cmd/config"
	campaign "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/campaign"
	category "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/category"
	classtype "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/classtype"
	commercialization "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/commercialization"
	crop "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/crop"
	customer "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/customer"
	dollar "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/dollar"
	field "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/field"
	investor "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/investor"
	labor "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/labor"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/leasetype"
	lot "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/lot"
	manager "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/manager"
	project "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/project"
	supply "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/supply"
	unit "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/unit"
	"github.com/google/wire"
)

type Dependencies struct {
	Config                   *config.Config
	GinEngine                *gin.Server
	GormRepo                 *gorm.Repository
	Middlewares              *mwr.Middlewares
	WordsSuggester           *sug.WordsSuggester
	CustomerHandler          *customer.Handler
	CampaignHandler          *campaign.Handler
	InvestorHandler          *investor.Handler
	CropHandler              *crop.Handler
	CommercializationHandler *commercialization.Handler
	LotHandler               *lot.Handler
	FieldHandler             *field.Handler
	ManagerHandler           *manager.Handler
	ProjectHandler           *project.Handler
	LeaseTypeHandler         *leasetype.Handler
	SupplyHandler            *supply.Handler
	CategoryHandler          *category.Handler
	UnitHandler              *unit.Handler
	ClassTypeHandler         *classtype.Handler
	DollarHandler            *dollar.Handler
	LaborHandler             *labor.Handler
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
		InvestorSet,
		CropSet,
		CommercializationSet,
		LotSet,
		FieldSet,
		ManagerSet,
		ProjectSet,
		LeaseTypeSet,
		SupplySet,
		CategorySet,
		UnitSet,
		ClassTypeSet,
		DollarSet,
		LaborSet,
		wire.Struct(new(Dependencies), "*"),
	)
	return &Dependencies{}, nil
}
