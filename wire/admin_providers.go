package wire

import (
	"context"

	firebase "firebase.google.com/go/v4"
	"github.com/google/wire"

	config "github.com/alphacodinggroup/ponti-backend/cmd/config"
	admin "github.com/alphacodinggroup/ponti-backend/internal/admin"
	adminidp "github.com/alphacodinggroup/ponti-backend/internal/admin/idp"
	pgorm "github.com/alphacodinggroup/ponti-backend/pkg/databases/sql/gorm"
)

func ProvideFirebaseApp(cfg *config.Config) (*firebase.App, error) {
	// Uses ADC in Cloud Run; for local, rely on GOOGLE_APPLICATION_CREDENTIALS.
	return firebase.NewApp(context.Background(), &firebase.Config{
		ProjectID: cfg.Auth.IdentityProjectID,
	})
}

func ProvideIdentityAdmin(app *firebase.App) (adminidp.AdminClient, error) {
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

