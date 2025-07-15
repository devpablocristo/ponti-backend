package models

import (
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/base"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/labor/usecases/domain"
)

type Labor struct {
	ID              int64   `gorm:"primaryKey;autoIncrement"`
	Name            string  `gorm:"type:varchar(255);not null;column:name"`
	ContractorName  string  `gorm:"type:varchar(255);not null;column:contractor_name"`
	Price           float64 `gorm:"not null;column:price"`
	ProjectId       int64   `gorm:"not null;column:project_id"`
	LaborCategoryID int64   `gorm:"not null;column:category_id"`

	base.BaseModel
}

type LaborCategory struct {
	ID          int64  `gorm:"primaryKey;autoIncrement"`
	Name        string `gorm:"not null;column:name"`
	LaborTypeId int64  `gorm:"not null;column:name"`

	base.BaseModel
}

type LaborType struct {
	ID   int64  `gorm:"primaryKey;autoIncrement"`
	Name string `gorm:"not null;column:name"`

	base.BaseModel
}

func (l Labor) ToDomain() *domain.Labor {
	return &domain.Labor{
		ID:              l.ID,
		Name:            l.Name,
		ContractorName:  l.ContractorName,
		Price:           l.Price,
		ProjectId:       l.ProjectId,
		LaborCategoryId: l.LaborCategoryID,
	}
}

func FromDomain(d *domain.Labor) *Labor {
	return &Labor{
		ID:              d.ID,
		Name:            d.Name,
		ContractorName:  d.ContractorName,
		Price:           d.Price,
		ProjectId:       d.ProjectId,
		LaborCategoryID: d.LaborCategoryId,
	}
}
