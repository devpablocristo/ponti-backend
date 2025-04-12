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
	mongo "github.com/alphacodinggroup/euxcel-backend/pkg/databases/nosql/mongodb/mongo-driver"
	pg "github.com/alphacodinggroup/euxcel-backend/pkg/databases/sql/postgresql/pgxpool"
	resty "github.com/alphacodinggroup/euxcel-backend/pkg/http/clients/resty"
	smtp "github.com/alphacodinggroup/euxcel-backend/pkg/notification/smtp"

	authe "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/authe"

	category "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/category"
	config "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/config"

	item "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/item"
	macrocategory "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/macrocategory"
	notification "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/notification"
	person "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/person"
	supplier "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/supplier"
	user "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/user"
)

// Dependencies reúne todas las dependencias de la aplicación que se inyectan con Wire.
type Dependencies struct {
	ConfigLoader       config.Loader
	GinServer          ginsrv.Server
	GormRepository     gorm.Repository
	MongoRepository    mongo.Repository
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

	// Para pruebas
	PersonUseCases person.UseCases
	UserUseCases   user.UseCases
	ItemUseCases   item.UseCases
}

// Initialize se encarga de inyectar todas las dependencias usando Wire.
func Initialize() (*Dependencies, error) {
	wire.Build(
		// Proveedores bootstrap
		ProvideConfigLoader,
		ProvideGinServer,
		ProvideGormRepository,
		ProvideMongoDbRepository,
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

		wire.Struct(new(Dependencies), "*"),
	)
	return &Dependencies{}, nil
}
