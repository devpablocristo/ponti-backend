package create

import (
	domainclasstype "github.com/devpablocristo/ponti-backend/internal/class-type/usecases/domain"
	"time"

	shareddomain "github.com/devpablocristo/ponti-backend/internal/shared/domain"
	domain "github.com/devpablocristo/ponti-backend/internal/supply/usecases/domain"
	decimal "github.com/shopspring/decimal"
)

// DTO para entrada y salida (puedes separarlo si quieres)
type SupplyRequest struct {
	ID             int64           `json:"id,omitempty"`
	ProjectID      int64           `json:"project_id"`
	Name           string          `json:"name"`
	Price          decimal.Decimal `json:"price"`
	IsPartialPrice *bool           `json:"is_partial_price"`

	UnitID     int64 `json:"unit_id"`
	CategoryID int64 `json:"category_id"`
	TypeID     int64 `json:"type_id"`

	// Audit fields opcionales
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
	CreatedBy *string   `json:"created_by,omitempty"`
	UpdatedBy *string   `json:"updated_by,omitempty"`
}

// Para entrada (request): solo mapear IDs
func (d *SupplyRequest) ToDomain() *domain.Supply {
	return &domain.Supply{
		ID:             d.ID,
		ProjectID:      d.ProjectID,
		Name:           d.Name,
		Price:          d.Price,
		IsPartialPrice: boolOrDefault(d.IsPartialPrice, false),
		UnitID:         d.UnitID,
		CategoryID:     d.CategoryID,
		Type: domainclasstype.ClassType{
			ID: d.TypeID,
		},
		Base: shareddomain.Base{
			CreatedBy: d.CreatedBy,
			UpdatedBy: d.UpdatedBy,
		},
	}
}

// Para salida (response): se pueden agregar los nombres si tienes preload
func FromDomain(s *domain.Supply) *SupplyRequest {
	return &SupplyRequest{
		ID:             s.ID,
		ProjectID:      s.ProjectID,
		Name:           s.Name,
		Price:          s.Price,
		IsPartialPrice: boolPtr(s.IsPartialPrice),
		UnitID:         s.UnitID,
		CategoryID:     s.CategoryID,
		TypeID:         s.Type.ID,
		CreatedAt:      s.CreatedAt,
		UpdatedAt:      s.UpdatedAt,
		CreatedBy:      s.CreatedBy,
		UpdatedBy:      s.UpdatedBy,
	}
}

func boolOrDefault(v *bool, fallback bool) bool {
	if v == nil {
		return fallback
	}
	return *v
}

func boolPtr(v bool) *bool {
	value := v
	return &value
}
