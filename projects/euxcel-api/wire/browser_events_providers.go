package wire

import (
	"errors"

	mng "github.com/alphacodinggroup/euxcel-backend/pkg/databases/nosql/mongodb/mongo-driver"
	mdw "github.com/alphacodinggroup/euxcel-backend/pkg/http/middlewares/gin"
	ginsrv "github.com/alphacodinggroup/euxcel-backend/pkg/http/servers/gin"
	ws "github.com/alphacodinggroup/euxcel-backend/pkg/websocket/gorilla"

	browserevent "github.com/alphacodinggroup/euxcel-backend/internal/browser-events"
)

func ProvideBrowserEventsRepository(repo mng.Repository) (browserevent.Repository, error) {
	if repo == nil {
		return nil, errors.New("mongoDB repository cannot be nil")
	}
	return browserevent.NewRepository(repo), nil
}

// ProvideBrowserEventsUseCases retorna browserevent.UseCases a partir del repositorio.
func ProvideBrowserEventsUseCases(repo browserevent.Repository) browserevent.UseCases {
	return browserevent.NewUseCases(repo)
}

func ProvideBrowserEventsWebsocket(
	useCases browserevent.UseCases,
	upgrader ws.Upgrader,
) browserevent.WebSocket {
	return browserevent.NewWebSocket(useCases, upgrader)
}

// ProvideBrowserEventsHandler retorna el Handler de browserevent inyectando el servidor Gin,
// el servidor WebSocket, los casos de uso y los middlewares.
func ProvideBrowserEventsHandler(
	ginSrv ginsrv.Server,
	usecases browserevent.UseCases,
	middlewares *mdw.Middlewares,
	websocket browserevent.WebSocket,
) *browserevent.Handler {
	return browserevent.NewHandler(ginSrv, usecases, middlewares, websocket)
}
