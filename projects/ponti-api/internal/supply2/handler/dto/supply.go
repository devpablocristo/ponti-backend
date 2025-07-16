package dto

import (
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/supply2/usecases/domain"
)

type Supply struct {
	ID           int64   `json:"id,omitempty"`
	ProjectID    int64   `json:"project_id"`
	CampaignID   int64   `json:"campaign_id,omitempty"`
	FieldID      int64   `json:"field_id,omitempty"`
	InvestorID   int64   `json:"investor_id,omitempty"`
	DeliveryNote string  `json:"delivery_note"`
	Date         string  `json:"date"`
	EntryType    string  `json:"entry_type"`
	Name         string  `json:"name"`
	Unit         string  `json:"unit"`
	Amount       float64 `json:"amount"`
	Price        float64 `json:"price"`
	Category     string  `json:"category"`
	Type         string  `json:"type"`
	Provider     string  `json:"provider"`
	TotalUSD     float64 `json:"total_usd"`
}

func (d *Supply) ToDomain() *domain.Supply {
	return &domain.Supply{
		ID:           d.ID,
		ProjectID:    d.ProjectID,
		CampaignID:   d.CampaignID,
		FieldID:      d.FieldID,
		InvestorID:   d.InvestorID,
		DeliveryNote: d.DeliveryNote,
		Date:         d.Date,
		EntryType:    d.EntryType,
		Name:         d.Name,
		Unit:         d.Unit,
		Amount:       d.Amount,
		Price:        d.Price,
		Category:     d.Category,
		Type:         d.Type,
		Provider:     d.Provider,
		TotalUSD:     d.TotalUSD,
	}
}

func FromDomain(s *domain.Supply) *Supply {
	return &Supply{
		ID:           s.ID,
		ProjectID:    s.ProjectID,
		CampaignID:   s.CampaignID,
		FieldID:      s.FieldID,
		InvestorID:   s.InvestorID,
		DeliveryNote: s.DeliveryNote,
		Date:         s.Date,
		EntryType:    s.EntryType,
		Name:         s.Name,
		Unit:         s.Unit,
		Amount:       s.Amount,
		Price:        s.Price,
		Category:     s.Category,
		Type:         s.Type,
		Provider:     s.Provider,
		TotalUSD:     s.TotalUSD,
	}
}
