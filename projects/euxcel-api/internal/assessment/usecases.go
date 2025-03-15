package assessment

import (
	authe "github.com/alphacodinggroup/euxcel-backend/internal/authe"
	candidate "github.com/alphacodinggroup/euxcel-backend/internal/candidate"
	config "github.com/alphacodinggroup/euxcel-backend/internal/config"
	notification "github.com/alphacodinggroup/euxcel-backend/internal/notification"
	person "github.com/alphacodinggroup/euxcel-backend/internal/person"
)

// useCases implementa la interfaz UseCases
type useCases struct {
	repository     Repository
	config         config.Loader
	autheUc        authe.UseCases
	candidateUc    candidate.UseCases
	personUc       person.UseCases
	notificationUc notification.UseCases
}

// NewUseCases crea una instancia de useCases con las dependencias adecuadas
func NewUseCases(
	repo Repository,
	notif notification.UseCases,
	candidate candidate.UseCases,
	cfg config.Loader,
	au authe.UseCases,
	pn person.UseCases,
) UseCases {
	return &useCases{
		repository:     repo,
		notificationUc: notif,
		candidateUc:    candidate,
		config:         cfg,
		autheUc:        au,
		personUc:       pn,
	}
}
