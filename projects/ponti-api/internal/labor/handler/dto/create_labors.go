package dto

type CreateLaborsResponse struct {
	LaborsIds []int64 `json:"labors_ids"`
	Message   string  `json:"message"`
}

type CreateLabor
