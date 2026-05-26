package models

import (
	domain "github.com/devpablocristo/ponti-backend/internal/report/usecases/domain"
)

// ProjectInfo es el shape SQL/GORM que mapea a la query custom de project info.
// Tiene los `column:` tags porque GORM.Scan los necesita; el dominio
// (domain.ProjectInfo) queda sin saber de columnas SQL.
type ProjectInfo struct {
	ProjectID    int64  `gorm:"column:project_id"`
	ProjectName  string `gorm:"column:project_name"`
	CustomerID   int64  `gorm:"column:customer_id"`
	CustomerName string `gorm:"column:customer_name"`
	CampaignID   int64  `gorm:"column:campaign_id"`
	CampaignName string `gorm:"column:campaign_name"`
}

// ToDomain convierte el shape SQL al dominio puro.
func (p ProjectInfo) ToDomain() domain.ProjectInfo {
	return domain.ProjectInfo{
		ProjectID:    p.ProjectID,
		ProjectName:  p.ProjectName,
		CustomerID:   p.CustomerID,
		CustomerName: p.CustomerName,
		CampaignID:   p.CampaignID,
		CampaignName: p.CampaignName,
	}
}
