package models

import (
	"time"

	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/lot/usecases/domain"
	sharedmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/models"
)

type LotTable struct {
	ID             int64
	ProjectID      int64 `gorm:"project_id"`
	ProjectName    string
	FieldName      string
	LotName        string
	PreviousCrop   string
	PreviousCropID int64
	CurrentCrop    string
	CurrentCropID  int64
	Variety        string
	SowedArea      float64
	Season         string
	Tons           int
	UpdatedAt      *time.Time `gorm:"updated_at,omitempty"`
	CostPerHectare float64
}

type LotDates struct {
	LotID       int64      `gorm:"lot_id"`
	SowingDate  *time.Time `gorm:"sowing_date"`
	HarvestDate *time.Time `gorm:"harvest_date"`
	Sequence    int
	sharedmodels.Base
}

func (m *LotTable) ToDomain(dates []LotDates) domain.LotTable {
	var domainDates []domain.LotDates
	for _, date := range dates {
		domainDates = append(domainDates, domain.LotDates{
			SowingDate:  date.SowingDate,
			HarvestDate: date.HarvestDate,
			Sequence:    date.Sequence,
		})
	}

	return domain.LotTable{
		ID:             m.ID,
		ProjectID:      m.ProjectID,
		ProjectName:    m.ProjectName,
		FieldName:      m.FieldName,
		LotName:        m.LotName,
		PreviousCrop:   m.PreviousCrop,
		PreviousCropID: m.PreviousCropID,
		CurrentCrop:    m.CurrentCrop,
		CurrentCropID:  m.CurrentCropID,
		Variety:        m.Variety,
		SowedArea:      m.SowedArea,
		Dates:          domainDates,
		Season:         m.Season,
		Tons:           m.Tons,
		UpdatedAt:      m.UpdatedAt,
		CostPerHectare: m.CostPerHectare,
	}
}
