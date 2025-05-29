//go:build wireinject
// +build wireinject

package wire

import (
	"github.com/google/wire"

	gorm "github.com/alphacodinggroup/ponti-backend/pkg/databases/sql/gorm"
	mwr "github.com/alphacodinggroup/ponti-backend/pkg/http/middlewares/gin"
	gin "github.com/alphacodinggroup/ponti-backend/pkg/http/servers/gin"

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
	Config      *config.ConfigSet
	GinServer   *gin.Server
	GormRepo    *gorm.Repository
	Middlewares *mwr.Middlewares

	// Los Handlers que tu main va a montar en las rutas:
	CustomerHandler *customer.Handler
	InvestorHandler *investor.Handler
	CropHandler     *crop.Handler
	LotHandler      *lot.Handler
	FieldHandler    *field.Handler
	ManagerHandler  *manager.Handler
	ProjectHandler  *project.Handler
}

func Initialize(cfgSet *config.ConfigSet) (*Dependencies, error) {
	wire.Build(
		// Boostraps
		GormSet,
		GinSet,
		MiddlewareSet,

		// dominios
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
