package dto

import "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/lot/usecases/domain"

// Lot es el DTO para un lote (transferencia de datos).
type Lot struct {
	ID             int64   `json:"id"`
	Name           string  `json:"identifier"`
	FieldID        int64   `json:"field_id"`
	ProjectID      int64   `json:"project_id"`
	CurrentCropID  *int64  `json:"current_crop_id,omitempty"`
	PreviousCropID *int64  `json:"previous_crop_id,omitempty"`
	Variety        string  `json:"variety"`
	Area           float64 `json:"area"`
}

// ToDomain convierte el DTO Lot en la entidad del dominio.
func (l Lot) ToDomain() *domain.Lot {
	return &domain.Lot{
		ID:             l.ID,
		Name:           l.Name,
		FieldID:        l.FieldID,
		ProjectID:      l.ProjectID,
		CurrentCropID:  l.CurrentCropID,
		PreviousCropID: l.PreviousCropID,
		Variety:        l.Variety,
		Area:           l.Area,
	}
}

// FromDomain convierte una entidad del dominio a un DTO Lot.
func FromDomain(d domain.Lot) *Lot {
	return &Lot{
		ID:             d.ID,
		Name:           d.Name,
		FieldID:        d.FieldID,
		ProjectID:      d.ProjectID,
		CurrentCropID:  d.CurrentCropID,
		PreviousCropID: d.PreviousCropID,
		Variety:        d.Variety,
		Area:           d.Area,
	}
}
