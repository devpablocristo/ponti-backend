//go:build wireinject
// +build wireinject

package wire

import (
	"github.com/google/wire"

	jwt "github.com/devpablocristo/monorepo/pkg/authe/jwt/v5"
	rabbit "github.com/devpablocristo/monorepo/pkg/brokers/rabbitmq/amqp091/producer"
	redis "github.com/devpablocristo/monorepo/pkg/databases/cache/redis/v8"
	cass "github.com/devpablocristo/monorepo/pkg/databases/nosql/cassandra/gocql"
	mongo "github.com/devpablocristo/monorepo/pkg/databases/nosql/mongodb/mongo-driver"
	gorm "github.com/devpablocristo/monorepo/pkg/databases/sql/gorm"
	pg "github.com/devpablocristo/monorepo/pkg/databases/sql/postgresql/pgxpool"
	resty "github.com/devpablocristo/monorepo/pkg/http/clients/resty"
	mdw "github.com/devpablocristo/monorepo/pkg/http/middlewares/gin"
	gin "github.com/devpablocristo/monorepo/pkg/http/servers/gin"
	smtp "github.com/devpablocristo/monorepo/pkg/notification/smtp"
	ws "github.com/devpablocristo/monorepo/pkg/websocket/gorilla"

	assessment "github.com/devpablocristo/monorepo/projects/qh/internal/assessment"
	authe "github.com/devpablocristo/monorepo/projects/qh/internal/authe"
	browserevent "github.com/devpablocristo/monorepo/projects/qh/internal/browser-events"
	candidate "github.com/devpablocristo/monorepo/projects/qh/internal/candidate"
	config "github.com/devpablocristo/monorepo/projects/qh/internal/config"
	event "github.com/devpablocristo/monorepo/projects/qh/internal/event"
	group "github.com/devpablocristo/monorepo/projects/qh/internal/group"
	notification "github.com/devpablocristo/monorepo/projects/qh/internal/notification"
	person "github.com/devpablocristo/monorepo/projects/qh/internal/person"
	tweet "github.com/devpablocristo/monorepo/projects/qh/internal/tweet"
	user "github.com/devpablocristo/monorepo/projects/qh/internal/user"
)

// Dependencies reúne todas las dependencias de la aplicación que se
// inyectan con Wire.
type Dependencies struct {
	ConfigLoader        config.Loader
	GinServer           gin.Server
	GormRepository      gorm.Repository
	MongoRepository     mongo.Repository
	PostgresRepository  pg.Repository
	RedisCache          redis.Cache
	JwtService          jwt.Service
	RestyClient         resty.Client
	SmtpService         smtp.Service
	RabbitProducer      rabbit.Producer
	CassandraRepository cass.Repository
	WebSocket           ws.Upgrader

	Middlewares *mdw.Middlewares

	PersonHandler          *person.Handler
	GroupHandler           *group.Handler
	EventHandler           *event.Handler
	UserHandler            *user.Handler
	AssessmentHandler      *assessment.Handler
	CandidateHandler       *candidate.Handler
	BrowserEventsHandler   *browserevent.Handler
	BrowserEventsWebSocket browserevent.WebSocket
	AutheHandler           *authe.Handler
	NotificationHandler    *notification.Handler
	TweetHandler           *tweet.Handler

	// para pruebas
	PersonUseCases person.UseCases
	UserUseCases   user.UseCases
	TweetUseCases  tweet.UseCases
}

// Initialize se encarga de inyectar todas las dependencias usando Wire.
func Initialize() (*Dependencies, error) {
	wire.Build(
		// Proveedores boostrap
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
		ProvideRabbitProducer,
		ProvideCassandraRepository,
		ProvideWebSocketUpgrader,

		// Person
		ProvidePersonRepository,
		ProvidePersonUseCases,
		ProvidePersonHandler,

		// Group
		ProvideGroupRepository,
		ProvideGroupUseCases,
		ProvideGroupHandler,

		// Event
		ProvideEventRepository,
		ProvideEventUseCases,
		ProvideEventHandler,

		// User
		ProvideUserRepository,
		ProvideUserUseCases,
		ProvideUserHandler,

		// Assessment
		ProvideAssessmentRepository,
		ProvideAssessmentUseCases,
		ProvideAssessmentHandler,

		// Candidate
		ProvideCandidateRepository,
		ProvideCandidateUseCases,
		ProvideCandidateHandler,

		// Browser Events
		ProvideBrowserEventsRepository,
		ProvideBrowserEventsUseCases,
		ProvideBrowserEventsWebsocket,
		ProvideBrowserEventsHandler,

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

		// Tweet
		ProvideTweetBroker,
		ProvideTweetCache,
		ProvideTweetRepository,
		ProvideTweetUseCases,
		ProvideTweetHandler,

		wire.Struct(new(Dependencies), "*"),
	)
	return &Dependencies{}, nil
}
