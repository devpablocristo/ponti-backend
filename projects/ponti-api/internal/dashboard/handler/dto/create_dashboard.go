package dto

import (
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/dashboard/usecases/domain"
)

// CreateDashboard represents the request to create a dashboard.
type CreateDashboard struct {
	// Add fields as needed for dashboard creation
}

// ToDomain converts the DTO CreateDashboard to the domain entity.
func (c CreateDashboard) ToDomain() *domain.Dashboard {
	return &domain.Dashboard{
		// Add fields as needed for dashboard creation
	}
}

// CreateDashboardResponse represents the response after creating a dashboard.
type CreateDashboardResponse struct {
	Message string `json:"message"`
	ID      int64  `json:"id"`
}
