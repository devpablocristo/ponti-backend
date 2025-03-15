package wire

import (
	"errors"

	rabbit "github.com/devpablocristo/monorepo/pkg/brokers/rabbitmq/amqp091/producer"
	redis "github.com/devpablocristo/monorepo/pkg/databases/cache/redis/v8"
	cass "github.com/devpablocristo/monorepo/pkg/databases/nosql/cassandra/gocql"
	mdw "github.com/devpablocristo/monorepo/pkg/http/middlewares/gin"
	ginsrv "github.com/devpablocristo/monorepo/pkg/http/servers/gin"

	tweet "github.com/devpablocristo/monorepo/projects/qh/internal/tweet"
	user "github.com/devpablocristo/monorepo/projects/qh/internal/user"
)

func ProvideTweetRepository(repo cass.Repository) (tweet.Repository, error) {
	if repo == nil {
		return nil, errors.New("cassandra repository cannot be nil")
	}
	return tweet.NewRepository(repo), nil
}

func ProvideTweetCache(cache redis.Cache) (tweet.Cache, error) {
	if cache == nil {
		return nil, errors.New("redis cache cannot be nil")
	}
	return tweet.NewCache(cache), nil
}

func ProvideTweetBroker(prod rabbit.Producer) (tweet.Broker, error) {
	if prod == nil {
		return nil, errors.New("rabbit producer cannot be nil")
	}

	return tweet.NewBroker(prod, ""), nil
}

func ProvideTweetUseCases(repo tweet.Repository, usruc user.UseCases, cache tweet.Cache, prod tweet.Broker) tweet.UseCases {
	return tweet.NewUseCases(repo, usruc, cache, prod)
}

func ProvideTweetHandler(server ginsrv.Server, usecases tweet.UseCases, middlewares *mdw.Middlewares) *tweet.Handler {
	return tweet.NewHandler(server, usecases, middlewares)
}
