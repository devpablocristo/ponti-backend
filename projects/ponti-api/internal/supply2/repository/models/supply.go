package models

import (
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/supply/usecases/domain"
)

type Supply struct {
	ID           int64   `gorm:"primaryKey;autoIncrement;column:id"`
	ProjectID    int64   `gorm:"not null;index"`
	CampaignID   int64   `gorm:"index"`
	FieldID      int64   `gorm:"index"`
	InvestorID   int64   `gorm:"index"`
	DeliveryNote string  `gorm:"type:varchar(30)"`
	Date         string  `gorm:"type:date"`
	EntryType    string  `gorm:"type:varchar(20)"`
	Name         string  `gorm:"type:varchar(100);not null"`
	Unit         string  `gorm:"type:varchar(20);not null"`
	Amount       float64 `gorm:"not null"`
	Price        float64 `gorm:"not null"`
	Category     string  `gorm:"type:varchar(50);not null"`
	Type         string  `gorm:"type:varchar(50);not null"`
	Provider     string  `gorm:"type:varchar(80)"`
	TotalUSD     float64 `gorm:"not null"`
}

func (m *Supply) ToDomain() *domain.Supply {
	return &domain.Supply{
		ID:           m.ID,
		ProjectID:    m.ProjectID,
		CampaignID:   m.CampaignID,
		FieldID:      m.FieldID,
		InvestorID:   m.InvestorID,
		DeliveryNote: m.DeliveryNote,
		Date:         m.Date,
		EntryType:    m.EntryType,
		Name:         m.Name,
		Unit:         m.Unit,
		Amount:       m.Amount,
		Price:        m.Price,
		Category:     m.Category,
		Type:         m.Type,
		Provider:     m.Provider,
		TotalUSD:     m.TotalUSD,
	}
}

func FromDomain(d *domain.Supply) *Supply {
	return &Supply{
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
