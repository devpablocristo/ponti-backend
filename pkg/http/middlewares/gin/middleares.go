// File: pkgmwr/types.go
package pkgmwr

import "github.com/gin-gonic/gin"

// Middlewares implementa MiddlewaresPort agrupando tus slices.
type Middlewares struct {
	Global     []gin.HandlerFunc
	Validation []gin.HandlerFunc
	Auth       []gin.HandlerFunc
}

// GetGlobal devuelve los middlewares globales (logs, errores…)
func (m *Middlewares) GetGlobal() []gin.HandlerFunc {
	return m.Global
}

// GetValidation devuelve los middlewares de validación (payload, headers…)
func (m *Middlewares) GetValidation() []gin.HandlerFunc {
	return m.Validation
}

// GetProtected devuelve los middlewares de protección (JWT…)
func (m *Middlewares) GetProtected() []gin.HandlerFunc {
	return m.Auth
}
