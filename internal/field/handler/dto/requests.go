package dto

import (
	"github.com/shopspring/decimal"

	cropdom "github.com/devpablocristo/ponti-backend/internal/crop/usecases/domain"
	fielddom "github.com/devpablocristo/ponti-backend/internal/field/usecases/domain"
	leasetypedom "github.com/devpablocristo/ponti-backend/internal/lease-type/usecases/domain"
	lotdom "github.com/devpablocristo/ponti-backend/internal/lot/usecases/domain"
)

type LotInput struct {
	Name           string          `json:"name" binding:"required"`
	Hectares       decimal.Decimal `json:"hectares" binding:"required"`
	PreviousCropID int64           `json:"previous_crop_id" binding:"required"`
	CurrentCropID  int64           `json:"current_crop_id" binding:"required"`
	Season         string          `json:"season" binding:"required"`
}

type CreateFieldRequest struct {
	Name        string     `json:"name" binding:"required"`
	LeaseTypeID int64      `json:"lease_type_id" binding:"required"`
	Lots        []LotInput `json:"lots" binding:"required,dive,required"`
}

func (r *CreateFieldRequest) ToDomain() *fielddom.Field {
	f := &fielddom.Field{
		Name:      r.Name,
		LeaseType: &leasetypedom.LeaseType{ID: r.LeaseTypeID},
	}
	for _, l := range r.Lots {
		f.Lots = append(f.Lots, lotdom.Lot{
			Name:         l.Name,
			Hectares:     l.Hectares,
			PreviousCrop: cropdom.Crop{ID: l.PreviousCropID},
			CurrentCrop:  cropdom.Crop{ID: l.CurrentCropID},
			Season:       l.Season,
		})
	}
	return f
}

type UpdateFieldRequest struct {
	Name        string `json:"name" binding:"required"`
	LeaseTypeID int64  `json:"lease_type_id" binding:"required"`
}

func (r *UpdateFieldRequest) ToDomain(id int64) *fielddom.Field {
	return &fielddom.Field{
		ID:        id,
		Name:      r.Name,
		LeaseType: &leasetypedom.LeaseType{ID: r.LeaseTypeID},
	}
}
