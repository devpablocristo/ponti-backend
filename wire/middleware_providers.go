// File: wire/middleware_provider.go
package wire

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"

	config "github.com/alphacodinggroup/ponti-backend/cmd/config"
	pgorm "github.com/alphacodinggroup/ponti-backend/pkg/databases/sql/gorm"
	mwr "github.com/alphacodinggroup/ponti-backend/pkg/http/middlewares/gin"
)

type MiddlewaresEnginePort interface {
	GetGlobal() []gin.HandlerFunc
	GetValidation() []gin.HandlerFunc
	GetProtected() []gin.HandlerFunc
}

func ProvideMiddlewares(cfg *config.Config, repo *pgorm.Repository) *mwr.Middlewares {
	issuer := cfg.Auth.IdentityIssuer
	if issuer == "" && cfg.Auth.IdentityProjectID != "" {
		issuer = "https://securetoken.google.com/" + cfg.Auth.IdentityProjectID
	}

	audience := cfg.Auth.IdentityAudience
	if audience == "" {
		audience = cfg.Auth.IdentityProjectID
	}

	return mwr.NewDefaultMiddlewares(mwr.BuildConfig{
		DB: repo.Client(),
		Auth: mwr.IdentityAuthConfig{
			ProjectID:    cfg.Auth.IdentityProjectID,
			Issuer:       issuer,
			Audience:     audience,
			JWKSURL:      cfg.Auth.IdentityJWKSURL,
			CacheTTL:     time.Duration(cfg.Auth.IdentityJWKSCacheTTL) * time.Second,
			TenantHeader: cfg.Auth.TenantHeader,
			AutoProvision: cfg.Auth.AutoProvision,
			DefaultTenant: cfg.Auth.DefaultTenantName,
			DefaultRole:   cfg.Auth.DefaultRoleName,
		},
	})
}

// ProvideMiddlewaresEnginePort convierte el *mwr.Middlewares en la interfaz MiddlewaresEnginePort.
func ProvideMiddlewaresEnginePort(m *mwr.Middlewares) MiddlewaresEnginePort {
	return m
}

// MiddlewareSet expone los dos providers necesarios.
var MiddlewareSet = wire.NewSet(
	ProvideMiddlewares,
	ProvideMiddlewaresEnginePort,
)
