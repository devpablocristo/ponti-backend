package dto

type CreateLaborsResponse struct {
	Labors  []CreateLabor `json:"labors_ids"`
	Message string        `json:"message"`
}

type CreateLabor struct {
	LaborName   string `json:"labor_name"`
	LaborID     int64  `json:"labor_id"`
	IsSaved     bool   `json:"is_saved"`
	ErrorDetail string `json:"error_detail"`
}
