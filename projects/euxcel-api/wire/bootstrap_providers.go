package wire

import (
	"fmt"

	jwt "github.com/devpablocristo/monorepo/pkg/authe/jwt/v5"
	rabbit "github.com/devpablocristo/monorepo/pkg/brokers/rabbitmq/amqp091/producer"
	rdch "github.com/devpablocristo/monorepo/pkg/databases/cache/redis/v8"
	cass "github.com/devpablocristo/monorepo/pkg/databases/nosql/cassandra/gocql"
	mng "github.com/devpablocristo/monorepo/pkg/databases/nosql/mongodb/mongo-driver"
	gorm "github.com/devpablocristo/monorepo/pkg/databases/sql/gorm"
	pgdb "github.com/devpablocristo/monorepo/pkg/databases/sql/postgresql/pgxpool"
	resty "github.com/devpablocristo/monorepo/pkg/http/clients/resty"
	restymdw "github.com/devpablocristo/monorepo/pkg/http/middlewares/resty"
	ginsrv "github.com/devpablocristo/monorepo/pkg/http/servers/gin"
	ssmtp "github.com/devpablocristo/monorepo/pkg/notification/smtp"
	ws "github.com/devpablocristo/monorepo/pkg/websocket/gorilla"
)

func ProvideGormRepository() (gorm.Repository, error) {
	repo, err := gorm.Bootstrap("", "", "", "", "", 0)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Gorm: %w", err)
	}

	return repo, nil
}

func ProvideGinServer() (ginsrv.Server, error) {
	isTest := false
	server, err := ginsrv.Bootstrap("", "", isTest)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Gin server: %w", err)
	}
	return server, nil
}

func ProvideMongoDbRepository() (mng.Repository, error) {
	repo, err := mng.Bootstrap("", "", "", "", "")
	if err != nil {
		return nil, fmt.Errorf("failed to initialize MongoDB client: %w", err)
	}

	return repo, nil
}

func ProvidePostgresRepository() (pgdb.Repository, error) {
	repo, err := pgdb.Bootstrap("", "", "", "", "", "")
	if err != nil {
		return nil, fmt.Errorf("failed to bootstrap PostgreSQL repository: %w", err)
	}
	return repo, nil
}

func ProvideRedisCache() (rdch.Cache, error) {
	cache, err := rdch.Bootstrap("", "", 0)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Redis: %w", err)
	}
	return cache, nil
}

func ProvideSmtpService() (ssmtp.Service, error) {
	ssmtp, err := ssmtp.Bootstrap("", "", "", "", "", "")
	if err != nil {
		return nil, fmt.Errorf("failed to initialize SMTP service: %w", err)
	}

	return ssmtp, nil
}

func ProvideHttpClient() (resty.Client, error) {
	// Inicializar el cliente con la configuración adecuada
	httpc, err := resty.Bootstrap("", 0)
	if err != nil {
		return nil, err
	}

	// Añadir middleware de header personalizado
	restymdw.AddHeaderMiddleware(httpc, "X-Custom-Header", "custom-value")

	// Añadir middleware de logging
	logger := &resty.SimpleLogger{}
	restymdw.AddLoggingMiddleware(httpc, logger)

	return httpc, nil
}

func ProvideJwtService() (jwt.Service, error) {
	jwtSrv, err := jwt.Bootstrap("", 0, 0)
	if err != nil {
		return nil, err
	}

	return jwtSrv, nil
}

func ProvideRabbitProducer() (rabbit.Producer, error) {
	prod, err := rabbit.Bootstrap()
	if err != nil {
		return nil, err
	}

	return prod, nil
}

func ProvideCassandraRepository() (cass.Repository, error) {
	repo, err := cass.Bootstrap()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Cassandra repository: %w", err)
	}

	return repo, nil
}

// func ProvideWebSocketHandler() (ws.Server, error) {
// 	server, err := ws.Bootstrap("", "", "", "")
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to initialize WebSocket server: %w", err)
// 	}
// 	return server, nil
// }

func ProvideWebSocketUpgrader() (ws.Upgrader, error) {
	server, err := ws.Bootstrap()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize WebSocket Upgrader: %w", err)
	}
	return server, nil
}
