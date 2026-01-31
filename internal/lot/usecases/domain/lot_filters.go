package domain

// LotListFilter permite filtrar listados de lotes por workspace.
type LotListFilter struct {
	CustomerID *int64
	ProjectID  *int64
	CampaignID *int64
	FieldID    *int64
	CropID     *int64
}
