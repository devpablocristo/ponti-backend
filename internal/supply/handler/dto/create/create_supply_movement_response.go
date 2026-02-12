package create

type CreateSupplyMovementResponse struct {
	SupplyMovementID int64  `json:"supply_movement_id"`
	IsSaved          bool   `json:"is_saved"`
	ErrorDetail      string `json:"error_detail"`
}

type SupplyMovementSuccess struct {
	Index            int   `json:"index"`
	SupplyID         int64 `json:"supply_id"`
	SupplyMovementID int64 `json:"supply_movement_id"`
}

type SupplyMovementFailure struct {
	Index    int    `json:"index"`
	SupplyID int64  `json:"supply_id"`
	Code     string `json:"code"`
	Message  string `json:"message"`
}

type SupplyMovementSkipped struct {
	Index    int    `json:"index"`
	SupplyID int64  `json:"supply_id"`
	Reason   string `json:"reason"`
}

type CreateSupplyMovementBulkResponse struct {
	Success         bool                           `json:"success"`
	Mode            string                         `json:"mode"`
	Total           int                            `json:"total"`
	Applied         int                            `json:"applied"`
	Failed          int                            `json:"failed"`
	Successes       []SupplyMovementSuccess        `json:"successes"`
	Failures        []SupplyMovementFailure        `json:"failures"`
	Skipped         []SupplyMovementSkipped        `json:"skipped"`
	Warnings        []string                       `json:"warnings"`
	SupplyMovements []CreateSupplyMovementResponse `json:"supply_movements"`
}

func NewSuccessfulCreateSupplyMovementResponse(supplyMovementID int64) CreateSupplyMovementResponse {
	return CreateSupplyMovementResponse{
		SupplyMovementID: supplyMovementID,
		IsSaved:          true,
	}
}

func NewErrorCreateSupplyMovementResponse(errorDetail string) CreateSupplyMovementResponse {
	return CreateSupplyMovementResponse{
		IsSaved:     false,
		ErrorDetail: errorDetail,
	}
}
