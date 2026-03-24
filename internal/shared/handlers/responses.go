package sharedhandlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	ginmw "github.com/devpablocristo/core/backend/gin/go"
)

// CreateResponse es la respuesta canónica para operaciones de creación.
// Se usa en todas las entidades para evitar DTOs individuales de create.
type CreateResponse struct {
	ID int64 `json:"id"`
}

// RespondCreated responde con 201 Created y el ID del recurso creado.
func RespondCreated(c *gin.Context, id int64) {
	ginmw.WriteCreated(c, CreateResponse{ID: id})
}

// RespondNoContent responde con 204 No Content (update, delete, archive, restore).
func RespondNoContent(c *gin.Context) {
	ginmw.WriteNoContent(c)
}

// RespondOK responde con 200 OK y el payload dado.
func RespondOK(c *gin.Context, data any) {
	ginmw.WriteJSON(c, http.StatusOK, data)
}
