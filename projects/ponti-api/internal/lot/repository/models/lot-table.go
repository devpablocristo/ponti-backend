package models

import domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/lot/usecases/domain"

type LotTable struct {
	ProjectName    string
	FieldName      string
	LotName        string
	PreviousCrop   string
	CurrentCrop    string
	Variety        string
	SowedArea      float64
	SowingDate     string
	CostPerHectare float64
}

func (m *LotTable) ToDomain() domain.LotTable {
	return domain.LotTable{
		ProjectName:    m.ProjectName,
		FieldName:      m.FieldName,
		LotName:        m.LotName,
		PreviousCrop:   m.PreviousCrop,
		CurrentCrop:    m.CurrentCrop,
		Variety:        m.Variety,
		SowedArea:      m.SowedArea,
		SowingDate:     m.SowingDate,
		CostPerHectare: m.CostPerHectare,
	}
}
