package dto

import (
	domain "github.com/devpablocristo/ponti-backend/internal/crop/usecases/domain"
	types "github.com/devpablocristo/ponti-backend/internal/shared/types"
)

// CropResponse es el DTO de salida para un crop individual.
type CropResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// CropFromDomain convierte un domain.Crop a CropResponse.
func CropFromDomain(d *domain.Crop) CropResponse {
	return CropResponse{ID: d.ID, Name: d.Name}
}

// ListCropsResponse es la respuesta paginada para el listado de crops.
type ListCropsResponse struct {
	Data     []CropResponse `json:"data"`
	PageInfo types.PageInfo `json:"page_info"`
}

// NewListCropsResponse construye la respuesta paginada de crops.
func NewListCropsResponse(crops []domain.Crop, page, perPage int, total int64) ListCropsResponse {
	data := make([]CropResponse, 0, len(crops))
	for i := range crops {
		data = append(data, CropFromDomain(&crops[i]))
	}
	return ListCropsResponse{
		Data:     data,
		PageInfo: types.NewPageInfo(page, perPage, total),
	}
}
