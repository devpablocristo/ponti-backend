package create

import (
	"time"

	domainclasstype "github.com/alphacodinggroup/ponti-backend/internal/class-type/usecases/domain"

	shareddomain "github.com/alphacodinggroup/ponti-backend/internal/shared/domain"
	domain "github.com/alphacodinggroup/ponti-backend/internal/supply/usecases/domain"
	decimal "github.com/shopspring/decimal"
)

// DTO para entrada y salida (puedes separarlo si quieres)
type SupplyRequest struct {
	ID        int64           `json:"id,omitempty"`
	ProjectID int64           `json:"project_id"`
	Name      string          `json:"name"`
	Price     decimal.Decimal `json:"price"`

	UnitID     int64 `json:"unit_id"`
	CategoryID int64 `json:"category_id"`
	TypeID     int64 `json:"type_id"`

	// Compatibilidad con frontend legacy (ordenes de trabajo)
	Unit     int64 `json:"unit,omitempty"`
	Category int64 `json:"category,omitempty"`
	Type     int64 `json:"type,omitempty"`

	// Audit fields opcionales
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
	CreatedBy *int64    `json:"created_by,omitempty"`
	UpdatedBy *int64    `json:"updated_by,omitempty"`
}

// Para entrada (request): solo mapear IDs
func (d *SupplyRequest) ToDomain() *domain.Supply {
	unitID := d.UnitID
	if unitID == 0 {
		unitID = d.Unit
	}

	categoryID := d.CategoryID
	if categoryID == 0 {
		categoryID = d.Category
	}

	typeID := d.TypeID
	if typeID == 0 {
		typeID = d.Type
	}

	return &domain.Supply{
		ID:         d.ID,
		ProjectID:  d.ProjectID,
		Name:       d.Name,
		Price:      d.Price,
		UnitID:     unitID,
		CategoryID: categoryID,
		Type: domainclasstype.ClassType{
			ID: typeID,
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
		ID:         s.ID,
		ProjectID:  s.ProjectID,
		Name:       s.Name,
		Price:      s.Price,
		UnitID:     s.UnitID,
		CategoryID: s.CategoryID,
		TypeID:     s.Type.ID,
		CreatedAt:  s.CreatedAt,
		UpdatedAt:  s.UpdatedAt,
		CreatedBy:  s.CreatedBy,
		UpdatedBy:  s.UpdatedBy,
	}
}
