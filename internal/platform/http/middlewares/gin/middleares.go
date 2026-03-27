package pkgmwr

import (
	"time"

	coreginmw "github.com/devpablocristo/core/http/gin/go"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Middlewares struct {
	global     []gin.HandlerFunc
	validation []gin.HandlerFunc
	protected  []gin.HandlerFunc
}

type BuildConfig struct {
	DB   *gorm.DB
	Auth IdentityAuthConfig
}

func NewDefaultMiddlewares(cfg BuildConfig) *Middlewares {
	global := []gin.HandlerFunc{
		ErrorHandling(),
		RequestAndResponseLogger(HttpLoggingOptions{
			LogLevel:       "info",
			IncludeHeaders: true,
			IncludeBody:    false,
			ExcludedPaths:  []string{"/health", "/ping", "/swagger/spec", "/swagger/ui/index.html"},
		}),
	}
	validation := []gin.HandlerFunc{coreginmw.RequireAPIKeyFromEnv("X_API_KEY")}
	if cfg.Auth.Enabled {
		validation = append(validation, RequireIdentityPlatformAuthz(cfg.Auth, cfg.DB))
	} else {
		validation = append(validation, RequireLocalDevAuthz(cfg.Auth, cfg.DB))
	}
	protected := []gin.HandlerFunc{}

	if cfg.Auth.CacheTTL <= 0 {
		cfg.Auth.CacheTTL = 5 * time.Minute
	}

	return &Middlewares{
		global:     global,
		validation: validation,
		protected:  protected,
	}
}
func (m *Middlewares) GetGlobal() []gin.HandlerFunc     { return m.global }
func (m *Middlewares) GetValidation() []gin.HandlerFunc { return m.validation }
func (m *Middlewares) GetProtected() []gin.HandlerFunc  { return m.protected }
