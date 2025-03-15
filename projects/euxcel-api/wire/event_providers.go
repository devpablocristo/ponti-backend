package wire

import (
	"errors"

	mng "github.com/devpablocristo/monorepo/pkg/databases/nosql/mongodb/mongo-driver"
	mdw "github.com/devpablocristo/monorepo/pkg/http/middlewares/gin"
	ginsrv "github.com/devpablocristo/monorepo/pkg/http/servers/gin"

	event "github.com/devpablocristo/monorepo/projects/qh/internal/event"
)

func ProvideEventRepository(repo mng.Repository) (event.Repository, error) {
	if repo == nil {
		return nil, errors.New("mongoDB repository cannot be nil")
	}
	return event.NewRepository(repo), nil
}

func ProvideEventUseCases(repo event.Repository) event.UseCases {
	return event.NewUseCases(repo)
}

func ProvideEventHandler(server ginsrv.Server, usecases event.UseCases, middlewares *mdw.Middlewares) *event.Handler {
	return event.NewHandler(server, usecases, middlewares)
}
