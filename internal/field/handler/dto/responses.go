package dto

import (
	"time"

	"github.com/shopspring/decimal"

	fielddom "github.com/alphacodinggroup/ponti-backend/internal/field/usecases/domain"
	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
)

type LotResponse struct {
	ID               int64           `json:"id"`
	Name             string          `json:"name"`
	Hectares         decimal.Decimal `json:"hectares"`
	PreviousCropID   int64           `json:"previous_crop_id"`
	PreviousCropName string          `json:"previous_crop_name,omitempty"`
	CurrentCropID    int64           `json:"current_crop_id"`
	CurrentCropName  string          `json:"current_crop_name,omitempty"`
	Season           string          `json:"season"`
}

type FieldResponse struct {
	ID               int64            `json:"id"`
	ProjectID        int64            `json:"project_id"`
	Name             string           `json:"name"`
	LeaseTypeID      int64            `json:"lease_type_id"`
	LeaseTypePercent *decimal.Decimal `json:"lease_type_percent,omitempty"`
	LeaseTypeValue   *decimal.Decimal `json:"lease_type_value,omitempty"`
	Lots             []LotResponse    `json:"lots"`
	ArchivedAt       *time.Time       `json:"archived_at,omitempty"`
	CreatedAt        time.Time        `json:"created_at"`
	UpdatedAt        time.Time        `json:"updated_at"`
}

func FieldFromDomain(d *fielddom.Field) FieldResponse {
	var leaseTypeID int64
	if d.LeaseType != nil {
		leaseTypeID = d.LeaseType.ID
	}

	lots := make([]LotResponse, 0, len(d.Lots))
	for _, l := range d.Lots {
		lots = append(lots, LotResponse{
			ID:               l.ID,
			Name:             l.Name,
			Hectares:         l.Hectares,
			PreviousCropID:   l.PreviousCrop.ID,
			PreviousCropName: l.PreviousCrop.Name,
			CurrentCropID:    l.CurrentCrop.ID,
			CurrentCropName:  l.CurrentCrop.Name,
			Season:           l.Season,
		})
	}

	return FieldResponse{
		ID:               d.ID,
		ProjectID:        d.ProjectID,
		Name:             d.Name,
		LeaseTypeID:      leaseTypeID,
		LeaseTypePercent: d.LeaseTypePercent,
		LeaseTypeValue:   d.LeaseTypeValue,
		Lots:             lots,
		ArchivedAt:       d.ArchivedAt,
		CreatedAt:        d.CreatedAt,
		UpdatedAt:        d.UpdatedAt,
	}
}

type ListFieldsResponse struct {
	Data     []FieldResponse `json:"data"`
	PageInfo types.PageInfo  `json:"page_info"`
}

func NewListFieldsResponse(fields []fielddom.Field, page, perPage int, total int64) ListFieldsResponse {
	data := make([]FieldResponse, 0, len(fields))
	for i := range fields {
		data = append(data, FieldFromDomain(&fields[i]))
	}
	return ListFieldsResponse{
		Data:     data,
		PageInfo: types.NewPageInfo(page, perPage, total),
	}
}
