package create

type PendingSupplyRequest struct {
	ProjectID int64  `json:"project_id" binding:"required"`
	Name      string `json:"name" binding:"required"`
}

type PendingSupplyResponse struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	IsPending bool   `json:"is_pending"`
	Created   bool   `json:"created"`
}
