package pkgmwr

import (
	"crypto/rsa"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	pkgutils "github.com/alphacodinggroup/euxcel-backend/pkg/utils"
)

// Validate retorna un gin.HandlerFunc que valida el JWT usando la lógica común.
func Validate(cfg pkgutils.Config) gin.HandlerFunc {
	// Parsear la clave RSA si se proporciona.
	var rsaPublicKey *rsa.PublicKey
	if cfg.PublicKeyPEM != "" {
		key, err := pkgutils.ParseRSAPublicKey(cfg.PublicKeyPEM)
		if err != nil {
			log.Fatalf("failed to parse RSA public key: %v", err)
		}
		rsaPublicKey = key
	}

	return func(c *gin.Context) {
		tokenStr, err := pkgutils.ExtractTokenFromRequest(c.Request, cfg)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		// Parsear el token sin verificar para conocer el método de firma.
		unverifiedToken, _, err := new(jwt.Parser).ParseUnverified(tokenStr, jwt.MapClaims{})
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": fmt.Sprintf("invalid token: %v", err)})
			c.Abort()
			return
		}

		keyFunc := pkgutils.SelectKeyFunc(unverifiedToken, cfg.SecretKey, rsaPublicKey)
		if keyFunc == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unexpected signing method"})
			c.Abort()
			return
		}

		parsedToken, err := jwt.Parse(tokenStr, keyFunc)
		if err != nil || !parsedToken.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": fmt.Sprintf("invalid token: %v", err)})
			c.Abort()
			return
		}

		// Guardar el token y los claims en el contexto de Gin.
		c.Set(cfg.ContextKey, parsedToken)
		c.Set(pkgutils.GetClaimsKey(cfg.ContextKey), parsedToken.Claims)
		c.Next()
	}
}
