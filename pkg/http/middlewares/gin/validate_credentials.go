package pkgmwr

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	pkgtypes "github.com/devpablocristo/monorepo/pkg/types"
)

// Constantes para los mensajes de error
const (
	errInvalidPayload = "Invalid request payload"
	errMissingField   = "Either username or email is required"
)

// ValidateCredentials middleware para validar el payload del login
func ValidateCredentials() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		log.Println("Credentials Middleware: Starting...")
		var credentials pkgtypes.LoginCredentials

		// Intentar enlazar el JSON al struct
		if err := ctx.ShouldBindJSON(&credentials); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"message": errInvalidPayload,
				"error":   err.Error(),
			})
			ctx.Abort()
			return
		}

		// Validar que al menos uno de los campos opcionales esté presente
		if credentials.Username == "" && credentials.Email == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"message": errMissingField,
			})
			ctx.Abort()
			return
		}

		// Guardar las credenciales validadas en el contexto
		ctx.Set("credentials", credentials)
		ctx.Next()
	}
}
