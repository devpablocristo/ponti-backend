package pkgmwr

import (
	"crypto/rsa"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/devpablocristo/core/backend/go/domainerr"
	"github.com/devpablocristo/core/backend/go/httperr"

	pkgutils "github.com/devpablocristo/ponti-backend/internal/shared/utils"
)

// RequireJWT valida el token JWT en la solicitud.
func RequireJWT(cfg pkgutils.Config) gin.HandlerFunc {
	var rsaPublicKey *rsa.PublicKey
	if cfg.PublicKeyPEM != "" {
		key, err := pkgutils.ParseRSAPublicKey(cfg.PublicKeyPEM)
		if err != nil {
			panic(fmt.Errorf("failed to parse RSA public key: %w", err))
		}
		rsaPublicKey = key
	}

	return func(c *gin.Context) {
		tokenStr, err := pkgutils.ExtractTokenFromRequest(c.Request, cfg)
		if err != nil {
			domErr := domainerr.Unauthorized("invalid token")
			status, apiErr := httperr.Normalize(domErr)
			c.AbortWithStatusJSON(status, apiErr)
			return
		}
		unverifiedToken, _, err := new(jwt.Parser).ParseUnverified(tokenStr, jwt.MapClaims{})
		if err != nil {
			domErr := domainerr.Unauthorized("invalid token")
			status, apiErr := httperr.Normalize(domErr)
			c.AbortWithStatusJSON(status, apiErr)
			return
		}
		keyFunc := pkgutils.SelectKeyFunc(unverifiedToken, cfg.SecretKey, rsaPublicKey)
		if keyFunc == nil {
			domErr := domainerr.Unauthorized("unexpected signing method")
			status, apiErr := httperr.Normalize(domErr)
			c.AbortWithStatusJSON(status, apiErr)
			return
		}
		parsedToken, err := jwt.Parse(tokenStr, keyFunc)
		if err != nil || !parsedToken.Valid {
			domErr := domainerr.Unauthorized("invalid token")
			status, apiErr := httperr.Normalize(domErr)
			c.AbortWithStatusJSON(status, apiErr)
			return
		}
		c.Set(cfg.ContextKey, parsedToken)
		c.Set(pkgutils.GetClaimsKey(cfg.ContextKey), parsedToken.Claims)
		c.Next()
	}
}
