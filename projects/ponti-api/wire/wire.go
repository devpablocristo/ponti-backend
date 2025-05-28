//go:build wireinject
// +build wireinject

package wire

import (
	"github.com/google/wire"

	gorm "github.com/alphacodinggroup/ponti-backend/pkg/databases/sql/gorm"
	mwr "github.com/alphacodinggroup/ponti-backend/pkg/http/middlewares/gin"
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
	Config      *config.ConfigSet
	GinServer   *ginsrv.Server
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
		// 1) Inyectar la config que ya cargaste en main.go
		wire.Value(cfgSet),

		// 2) Infraestructuras compartidas
		GormSet,
		GinSet,

		// 3) Middlewares
		GlobalMiddlewareSet,
		ValidationMiddlewareSet,
		AuthMiddlewareSet,

		// 4) Todos los dominios
		CustomerSet,
		InvestorSet,
		CropSet,
		LotSet,
		FieldSet,
		ManagerSet,
		ProjectSet,

		// 5) Ensamblar el struct final
		wire.Struct(new(Dependencies), "*"),
	)
	return &Dependencies{}, nil
}
