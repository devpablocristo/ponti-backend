package wire

import (
	"context"

	firebase "firebase.google.com/go/v4"
	"github.com/google/wire"

	config "github.com/devpablocristo/ponti-backend/cmd/config"
	admin "github.com/devpablocristo/ponti-backend/internal/admin"
	adminidp "github.com/devpablocristo/ponti-backend/internal/admin/idp"
	pgorm "github.com/devpablocristo/ponti-backend/internal/platform/persistence/gorm"
)

func ProvideFirebaseApp(cfg *config.Config) (*firebase.App, error) {
	// Uses ADC in Cloud Run; for local, rely on GOOGLE_APPLICATION_CREDENTIALS.
	// When AUTH is disabled (local mode), don't require ADC just to boot the API.
	if !cfg.Auth.Enabled {
		return nil, nil
	}
	return firebase.NewApp(context.Background(), &firebase.Config{
		ProjectID: cfg.Auth.IdentityProjectID,
	})
}

func ProvideIdentityAdmin(cfg *config.Config, app *firebase.App) (adminidp.AdminClient, error) {
	if !cfg.Auth.Enabled || app == nil {
		return &adminidp.NoopAdmin{}, nil
	}
	return adminidp.NewFirebaseAdmin(app)
}

func ProvideAdminHandler(
	repo *pgorm.Repository,
	idpAdmin adminidp.AdminClient,
	srv GinEnginePort,
	acf *config.API,
	mws MiddlewaresEnginePort,
) *admin.Handler {
	return admin.NewHandler(repo.Client(), idpAdmin, srv, acf, mws)
}

var AdminSet = wire.NewSet(
	ProvideFirebaseApp,
	ProvideIdentityAdmin,
	ProvideAdminHandler,
)
