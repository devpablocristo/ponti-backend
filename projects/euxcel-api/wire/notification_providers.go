package wire

import (
	"errors"

	mdw "github.com/devpablocristo/monorepo/pkg/http/middlewares/gin"
	gin "github.com/devpablocristo/monorepo/pkg/http/servers/gin"
	smtp "github.com/devpablocristo/monorepo/pkg/notification/smtp"

	notification "github.com/devpablocristo/monorepo/projects/qh/internal/notification"
)

func ProvideNotificationSmtpService(smtp smtp.Service) (notification.SmtpService, error) {
	if smtp == nil {
		return nil, errors.New("smtp service cannot be nil")
	}
	return notification.NewSmtpService(smtp), nil
}

func ProvideNotificationUseCases(
	ssrv notification.SmtpService,
) notification.UseCases {
	return notification.NewUseCases(ssrv)
}

func ProvideNotificationHandler(
	server gin.Server,
	usecases notification.UseCases,
	middlewares *mdw.Middlewares,
) *notification.Handler {
	return notification.NewHandler(server, usecases, middlewares)
}
