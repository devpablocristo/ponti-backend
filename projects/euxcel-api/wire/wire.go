//go:build wireinject
// +build wireinject

package wire

import (
	"github.com/google/wire"

	jwt "github.com/alphacodinggroup/euxcel-backend/pkg/authe/jwt/v5"
	rabbit "github.com/alphacodinggroup/euxcel-backend/pkg/brokers/rabbitmq/amqp091/producer"
	redis "github.com/alphacodinggroup/euxcel-backend/pkg/databases/cache/redis/v8"
	cass "github.com/alphacodinggroup/euxcel-backend/pkg/databases/nosql/cassandra/gocql"
	mongo "github.com/alphacodinggroup/euxcel-backend/pkg/databases/nosql/mongodb/mongo-driver"
	gorm "github.com/alphacodinggroup/euxcel-backend/pkg/databases/sql/gorm"
	pg "github.com/alphacodinggroup/euxcel-backend/pkg/databases/sql/postgresql/pgxpool"
	resty "github.com/alphacodinggroup/euxcel-backend/pkg/http/clients/resty"
	mdw "github.com/alphacodinggroup/euxcel-backend/pkg/http/middlewares/gin"
	gin "github.com/alphacodinggroup/euxcel-backend/pkg/http/servers/gin"
	smtp "github.com/alphacodinggroup/euxcel-backend/pkg/notification/smtp"
	ws "github.com/alphacodinggroup/euxcel-backend/pkg/websocket/gorilla"

	assessment "github.com/alphacodinggroup/euxcel-backend/internal/assessment"
	authe "github.com/alphacodinggroup/euxcel-backend/internal/authe"
	browserevent "github.com/alphacodinggroup/euxcel-backend/internal/browser-events"
	candidate "github.com/alphacodinggroup/euxcel-backend/internal/candidate"
	config "github.com/alphacodinggroup/euxcel-backend/internal/config"
	event "github.com/alphacodinggroup/euxcel-backend/internal/event"
	group "github.com/alphacodinggroup/euxcel-backend/internal/group"
	notification "github.com/alphacodinggroup/euxcel-backend/internal/notification"
	person "github.com/alphacodinggroup/euxcel-backend/internal/person"
	tweet "github.com/alphacodinggroup/euxcel-backend/internal/tweet"
	user "github.com/alphacodinggroup/euxcel-backend/internal/user"
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
