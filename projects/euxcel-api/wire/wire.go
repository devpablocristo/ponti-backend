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

	authe "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/authe"
	category "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/category"
	config "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/config"
	item "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/item"
	macrocategory "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/macrocategory"
	manager "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/manager"
	notification "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/notification"
	person "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/person"
	supplier "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/supplier"
	user "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/user"

	// Nuevas entidades
	crop "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/crop"
	customer "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/customer"
	field "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/field"
	investor "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/investor"
	lot "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/lot"
	project "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/project"
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

	PersonHandler        *person.Handler
	UserHandler          *user.Handler
	AutheHandler         *authe.Handler
	NotificationHandler  *notification.Handler
	ItemHandler          *item.Handler
	CategoryHandler      *category.Handler
	MacroCategoryHandler *macrocategory.Handler
	SupplierHandler      *supplier.Handler
	CropHandler          *crop.Handler
	CustomerHandler      *customer.Handler
	ManagerHandler       *manager.Handler
	FieldHandler         *field.Handler
	InvestorHandler      *investor.Handler
	LotHandler           *lot.Handler
	ProjectHandler       *project.Handler

	// Para pruebas (use cases existentes y de nuevas entidades si se requieren)
	PersonUseCases   person.UseCases
	UserUseCases     user.UseCases
	ItemUseCases     item.UseCases
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

		// Authe
		ProvideAutheCache,
		ProvideAutheHttpClient,
		ProvideAutheJwtService,
		ProvideAutheUseCases,
		ProvideAutheHandler,

		// Item
		ProvideItemRepository,
		ProvideItemUseCases,
		ProvideItemHandler,

		// Category
		ProvideCategoryRepository,
		ProvideCategoryUseCases,
		ProvideCategoryHandler,

		// MacroCategory
		ProvideMacroCategoryRepository,
		ProvideMacroCategoryUseCases,
		ProvideMacroCategoryHandler,

		// Supplier
		ProvideSupplierRepository,
		ProvideSupplierUseCases,
		ProvideSupplierHandler,

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
