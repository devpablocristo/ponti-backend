package dto

// CreateCrop is the DTO for the create request of a crop.
// It embeds the base Crop DTO.
type CreateCrop struct {
	Crop
}

type CreateCropResponse struct {
	Message string `json:"message"`
	CropID  int64  `json:"crop_id"`
}
