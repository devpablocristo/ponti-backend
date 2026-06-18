package dto

type PendingLaborRequest struct {
    Name string `json:"name" binding:"required"`
}

type PendingLaborResponse struct {
    ID        int64  `json:"id"`
    Name      string `json:"name"`
    IsPending bool   `json:"is_pending"`
    Created   bool   `json:"created"`
}