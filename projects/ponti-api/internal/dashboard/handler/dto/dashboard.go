package dto

import (
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/dashboard/usecases/domain"
)

// Dashboard represents a dashboard for data transfer.
type Dashboard struct {
	ID int64 `json:"id"`
}

// ToDomain converts the DTO Dashboard to the domain entity.
func (d Dashboard) ToDomain() *domain.Dashboard {
	return &domain.Dashboard{
		ID: d.ID,
	}
}

// FromDomain converts a domain Dashboard to the DTO.
func FromDomain(d domain.Dashboard) *Dashboard {
	return &Dashboard{
		ID: d.ID,
	}
}

func NewListDashboardsResponse(dashboards []domain.Dashboard) *ListDashboardsResponse {
	var convertedDashboards []Dashboard
	for _, dashboard := range dashboards {
		convertedDashboards = append(convertedDashboards, *FromDomain(dashboard))
	}
	return &ListDashboardsResponse{
		Dashboards: convertedDashboards,
	}
}

// ListDashboardsResponse represents a list of dashboards for data transfer.
type ListDashboardsResponse struct {
	Dashboards []Dashboard `json:"data"`
}
