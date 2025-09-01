package dto

import (
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/dashboard/usecases/domain"
)

// CreateDashboard representa un request para crear un dashboard
type CreateDashboard struct {
	// Campos básicos del dashboard
}

// ToDomain convierte el DTO a entidad de dominio
func (c CreateDashboard) ToDomain() *domain.Dashboard {
	return &domain.Dashboard{
		// Campos básicos
	}
}

// CreateDashboardResponse representa la respuesta de creación de dashboard
type CreateDashboardResponse struct {
	Message string `json:"message"`
	ID      int64  `json:"id"`
}

// Dashboard representa un dashboard completo (para casos de actualización)
type Dashboard struct {
	ID int64 `json:"id"`
	// Otros campos según sea necesario
}

// ToDomain convierte el DTO a entidad de dominio
func (d Dashboard) ToDomain() *domain.Dashboard {
	return &domain.Dashboard{
		ID: d.ID,
		// Otros campos
	}
}

// NewListDashboardsResponse crea una respuesta para listar dashboards
func NewListDashboardsResponse(dashboards []domain.Dashboard) interface{} {
	// Implementar según sea necesario
	return map[string]interface{}{
		"dashboards": dashboards,
		"count":      len(dashboards),
	}
}
