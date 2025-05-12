//go:build wireinject
// +build wireinject

package wire

import (
	gorm "github.com/alphacodinggroup/euxcel-backend/pkg/databases/sql/gorm"
	mdw "github.com/alphacodinggroup/euxcel-backend/pkg/http/middlewares/gin"
	ginsrv "github.com/alphacodinggroup/euxcel-backend/pkg/http/servers/gin"
	"github.com/google/wire"

	jwt "github.com/alphacodinggroup/euxcel-backend/pkg/authe/jwt/v5"
	redis "github.com/alphacodinggroup/euxcel-backend/pkg/databases/cache/redis/v8"
	pg "github.com/alphacodinggroup/euxcel-backend/pkg/databases/sql/postgresql/pgxpool"
	resty "github.com/alphacodinggroup/euxcel-backend/pkg/http/clients/resty"
	smtp "github.com/alphacodinggroup/euxcel-backend/pkg/notification/smtp"

	config "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/cmd/config"

	crop "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/crop"
	customer "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/customer"
	field "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/field"
	investor "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/investor"
	lot "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/lot"
	manager "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/manager"
	notification "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/notification"
	person "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/person"
	project "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/project"
	user "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/user"
)

// Dependencies reúne todas las dependencias de la aplicación que se inyectan con Wire.
type Dependencies struct {
	ConfigLoader       config.Loader
	GinServer          ginsrv.Server
	GormRepository     gorm.Repository
	PostgresRepository pg.Repository
	RedisCache         redis.Cache
	JwtService         jwt.Service
	RestyClient        resty.Client
	SmtpService        smtp.Service

	Middlewares *mdw.Middlewares

	PersonHandler       *person.Handler
	UserHandler         *user.Handler
	NotificationHandler *notification.Handler
	CropHandler         *crop.Handler
	CustomerHandler     *customer.Handler
	ManagerHandler      *manager.Handler
	FieldHandler        *field.Handler
	InvestorHandler     *investor.Handler
	LotHandler          *lot.Handler
	ProjectHandler      *project.Handler

	PersonUseCases   person.UseCases
	UserUseCases     user.UseCases
	CropUseCases     crop.UseCases
	CustomerUseCases customer.UseCases
	FieldUseCases    field.UseCases
	InvestorUseCases investor.UseCases
	LotUseCases      lot.UseCases
	ProjectUseCases  project.UseCases
}

// Initialize se encarga de inyectar todas las dependencias usando Wire.
func Initialize() (*Dependencies, error) {
	wire.Build(
		// Proveedores bootstrap
		ProvideConfigLoader,
		ProvideGinServer,
		ProvideGormRepository,
		ProvidePostgresRepository,
		ProvideJwtMiddleware,
		ProvideMiddlewares,
		ProvideRedisCache,
		ProvideJwtService,
		ProvideHttpClient,
		ProvideSmtpService,

		// Person
		ProvidePersonRepository,
		ProvidePersonUseCases,
		ProvidePersonHandler,

		// User
		ProvideUserRepository,
		ProvideUserUseCases,
		ProvideUserHandler,

		// Notification
		ProvideNotificationSmtpService,
		ProvideNotificationUseCases,
		ProvideNotificationHandler,

		// Nuevas entidades
		ProvideCropRepository,
		ProvideCropUseCases,
		ProvideCropHandler,

		ProvideCustomerRepository,
		ProvideCustomerUseCases,
		ProvideCustomerHandler,

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
