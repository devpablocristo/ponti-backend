package pkgmwr

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	pkgtypes "github.com/devpablocristo/monorepo/pkg/types"
)

// ErrorHandlingMiddleware captura errores añadidos al contexto y responde de manera adecuada
func ErrorHandlingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Registro contextual con método y ruta
		log.Printf("[ErrorHandlingMiddleware] Iniciando manejo de errores para %s %s", c.Request.Method, c.Request.URL.Path)

		c.Next() // Procesa la solicitud

		// Si ya se escribió una respuesta, no volvemos a escribir
		if c.Writer.Written() {
			return
		}

		// Si hay errores en el contexto, procesamos el primero
		if len(c.Errors) > 0 {
			log.Printf("[ErrorHandlingMiddleware] Se encontraron %d error(es)", len(c.Errors))

			// Tomamos el primer error para responder
			ginErr := c.Errors[0]
			log.Printf("[ErrorHandlingMiddleware] Error: %v", ginErr.Err)

			var status int
			var response interface{}

			// Manejo de errores del dominio (errores personalizados)
			if apiErr, ok := ginErr.Err.(*pkgtypes.Error); ok {
				response = apiErr.ToJSON() // Convertir a formato JSON
				status = mapErrorTypeToStatus(apiErr.Type)
			} else {
				// Para errores desconocidos, devolvemos un error interno con mensaje genérico
				response = gin.H{
					"error":   "INTERNAL_ERROR",
					"message": "Ha ocurrido un error interno, por favor intente más tarde.",
				}
				status = http.StatusInternalServerError
			}

			// Enviar respuesta JSON con el código de estado adecuado
			c.JSON(status, response)
			// Abortamos para asegurarnos que no se ejecute más lógica
			c.Abort()
			return
		}
	}
}

// mapErrorTypeToStatus mapea los tipos de errores a códigos HTTP
func mapErrorTypeToStatus(errType pkgtypes.ErrorType) int {
	switch errType {
	case pkgtypes.ErrNotFound:
		return http.StatusNotFound
	case pkgtypes.ErrValidation:
		return http.StatusBadRequest
	case pkgtypes.ErrConflict:
		return http.StatusConflict
	case pkgtypes.ErrAuthentication, pkgtypes.ErrAuthorization:
		return http.StatusUnauthorized
	case pkgtypes.ErrUnavailable:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}
