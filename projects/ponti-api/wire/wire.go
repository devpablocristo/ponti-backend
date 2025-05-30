//go:build wireinject
// +build wireinject

package wire

import (
	"github.com/google/wire"

	gorm "github.com/alphacodinggroup/ponti-backend/pkg/databases/sql/gorm"
	mwr "github.com/alphacodinggroup/ponti-backend/pkg/http/middlewares/gin"
	gin "github.com/alphacodinggroup/ponti-backend/pkg/http/servers/gin"
	sug "github.com/alphacodinggroup/ponti-backend/pkg/words-suggesters/pg_trgm-gin"

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
	Config          *config.AllConfigs
	GinEngine       *gin.Server
	GormRepo        *gorm.Repository
	Middlewares     *mwr.Middlewares
	Suggester       *sug.Suggester
	CustomerHandler *customer.Handler
	InvestorHandler *investor.Handler
	CropHandler     *crop.Handler
	LotHandler      *lot.Handler
	FieldHandler    *field.Handler
	ManagerHandler  *manager.Handler
	ProjectHandler  *project.Handler
}

func Initialize() (*Dependencies, error) {
	wire.Build(
		ConfigSet,
		GormSet,
		GinSet,
		MiddlewareSet,
		SuggesterSet,
		CustomerSet,
		InvestorSet,
		CropSet,
		LotSet,
		FieldSet,
		ManagerSet,
		ProjectSet,
		wire.Struct(new(Dependencies), "*"),
	)
	return &Dependencies{}, nil
}
