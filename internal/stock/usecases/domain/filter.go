package domain

// StockFilter permite filtros de workspace para listados de stock.
type StockFilter struct {
	CustomerID *int64
	ProjectID  *int64
	CampaignID *int64
	FieldID    *int64
}
