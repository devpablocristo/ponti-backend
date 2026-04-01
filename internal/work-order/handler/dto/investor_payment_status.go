package dto

type UpdateInvestorPaymentStatusRequest struct {
	PaymentStatus string `json:"payment_status" binding:"required"`
}
